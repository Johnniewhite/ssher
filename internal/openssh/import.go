// Package openssh imports concrete hosts from OpenSSH client configuration.
// It asks the installed OpenSSH client to resolve each host (`ssh -G`) so
// Include files, Host precedence, tokens, platform defaults, and Match rules
// follow the user's actual SSH implementation instead of a partial clone.
package openssh

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/johnniewhite/ssher/internal/store"
)

const resolveTimeout = 8 * time.Second

type Result struct {
	Servers  []store.Server
	Warnings []string
}

type Resolver func(context.Context, string, string) ([]byte, error)

func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".ssh", "config"), nil
}

func Import(ctx context.Context, configPath string, resolver Resolver) (Result, error) {
	if strings.TrimSpace(configPath) == "" {
		var err error
		configPath, err = DefaultConfigPath()
		if err != nil {
			return Result{}, err
		}
	}
	configPath, err := filepath.Abs(expandHome(configPath))
	if err != nil {
		return Result{}, fmt.Errorf("resolve SSH config path: %w", err)
	}
	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Result{}, fmt.Errorf("OpenSSH config not found at %s", configPath)
		}
		return Result{}, fmt.Errorf("read OpenSSH config: %w", err)
	}
	aliases, warnings, err := discoverAliases(configPath)
	if err != nil {
		return Result{}, err
	}
	if resolver == nil {
		resolver = resolveWithOpenSSH
	}
	result := Result{Warnings: warnings}
	for _, alias := range aliases {
		resolved, err := resolver(ctx, configPath, alias)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %v", alias, err))
			continue
		}
		server, err := ParseResolved(alias, bytes.NewReader(resolved))
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("%s: %v", alias, err))
			continue
		}
		result.Servers = append(result.Servers, server)
	}
	if len(result.Servers) == 0 && len(aliases) > 0 {
		return result, errors.New("OpenSSH could not resolve any concrete Host entries")
	}
	return result, nil
}

func resolveWithOpenSSH(ctx context.Context, configPath, alias string) ([]byte, error) {
	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return nil, errors.New("OpenSSH client not found; install ssh or pass a config through a machine that has OpenSSH")
	}
	resolveCtx, cancel := context.WithTimeout(ctx, resolveTimeout)
	defer cancel()
	command := exec.CommandContext(resolveCtx, sshPath, "-G", "-F", configPath, alias)
	command.Stdin = nil
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if resolveCtx.Err() != nil {
		return nil, fmt.Errorf("resolution timed out after %s", resolveTimeout)
	}
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("ssh -G failed: %s", message)
	}
	return output, nil
}

func ParseResolved(alias string, input io.Reader) (store.Server, error) {
	values := map[string][]string{}
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		key, value = strings.ToLower(strings.TrimSpace(key)), strings.TrimSpace(value)
		values[key] = append(values[key], value)
	}
	if err := scanner.Err(); err != nil {
		return store.Server{}, fmt.Errorf("read resolved SSH config: %w", err)
	}
	host := first(values, "hostname")
	user := first(values, "user")
	if host == "" || user == "" {
		return store.Server{}, errors.New("resolved config is missing HostName or User")
	}
	port := positiveInt(first(values, "port"), 22)
	server := store.Server{
		Name: alias, Host: host, User: user, Port: port, AuthType: store.AuthKey,
		Group: "openssh", KeepAlive: positiveInt(first(values, "serveraliveinterval"), 0),
		ConnectionTimeout: positiveInt(first(values, "connecttimeout"), 30),
		JumpHost:          normalizeNone(first(values, "proxyjump")),
		X11Forward:        strings.EqualFold(first(values, "forwardx11"), "yes"),
		CustomOptions:     map[string]string{"source": "openssh"},
	}
	if strings.EqualFold(first(values, "pubkeyauthentication"), "no") && strings.EqualFold(first(values, "passwordauthentication"), "yes") {
		server.AuthType = store.AuthPassword
	}
	for _, identity := range values["identityfile"] {
		candidate := expandHome(strings.Trim(identity, `"`))
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			server.KeyPath = candidate
			break
		}
	}
	server.LocalForwards = parseForwards(values["localforward"])
	server.RemoteForwards = parseForwards(values["remoteforward"])
	server.Touch()
	return server, nil
}

func discoverAliases(root string) ([]string, []string, error) {
	seenFiles, aliases, seenAliases := map[string]bool{}, []string{}, map[string]bool{}
	warnings := []string{}
	var walk func(string) error
	walk = func(path string) error {
		path = filepath.Clean(expandHome(path))
		if seenFiles[path] {
			return nil
		}
		seenFiles[path] = true
		file, err := os.Open(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				warnings = append(warnings, fmt.Sprintf("included SSH config not found: %s", path))
				return nil
			}
			return fmt.Errorf("open SSH config %s: %w", path, err)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		lineNumber := 0
		for scanner.Scan() {
			lineNumber++
			fields, err := splitFields(scanner.Text())
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s:%d: %v", path, lineNumber, err))
				continue
			}
			if len(fields) < 2 {
				continue
			}
			switch strings.ToLower(fields[0]) {
			case "host":
				for _, alias := range fields[1:] {
					if alias == "" || strings.HasPrefix(alias, "!") || strings.ContainsAny(alias, "*?!") || strings.HasPrefix(alias, "-") {
						continue
					}
					key := strings.ToLower(alias)
					if !seenAliases[key] {
						seenAliases[key] = true
						aliases = append(aliases, alias)
					}
				}
			case "include":
				for _, pattern := range fields[1:] {
					pattern = expandHome(pattern)
					if !filepath.IsAbs(pattern) {
						pattern = filepath.Join(filepath.Dir(path), pattern)
					}
					matches, globErr := filepath.Glob(pattern)
					if globErr != nil {
						warnings = append(warnings, fmt.Sprintf("%s:%d: invalid Include %q", path, lineNumber, pattern))
						continue
					}
					sort.Strings(matches)
					for _, match := range matches {
						if err := walk(match); err != nil {
							return err
						}
					}
				}
			}
		}
		return scanner.Err()
	}
	if err := walk(root); err != nil {
		return nil, warnings, err
	}
	return aliases, warnings, nil
}

func splitFields(line string) ([]string, error) {
	fields, current := []string{}, strings.Builder{}
	var quote rune
	escaped := false
	flush := func() {
		if current.Len() > 0 {
			fields = append(fields, current.String())
			current.Reset()
		}
	}
	for _, char := range line {
		if escaped {
			current.WriteRune(char)
			escaped = false
			continue
		}
		if char == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if char == quote {
				quote = 0
			} else {
				current.WriteRune(char)
			}
			continue
		}
		if char == '\'' || char == '"' {
			quote = char
			continue
		}
		if char == '#' {
			break
		}
		if char == ' ' || char == '\t' {
			flush()
			continue
		}
		current.WriteRune(char)
	}
	if escaped || quote != 0 {
		return nil, errors.New("unterminated quote or escape")
	}
	flush()
	return fields, nil
}

func first(values map[string][]string, key string) string {
	if len(values[key]) == 0 {
		return ""
	}
	return values[key][0]
}

func positiveInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func normalizeNone(value string) string {
	if strings.EqualFold(value, "none") {
		return ""
	}
	return value
}

func parseForwards(values []string) []store.PortForward {
	result := []store.PortForward{}
	for _, value := range values {
		parts := strings.Fields(value)
		if len(parts) != 2 {
			continue
		}
		local, err := strconv.Atoi(strings.TrimPrefix(parts[0], "localhost:"))
		host, remoteText, ok := strings.Cut(parts[1], ":")
		remote, remoteErr := strconv.Atoi(remoteText)
		if err == nil && ok && remoteErr == nil && local > 0 && remote > 0 {
			result = append(result, store.PortForward{LocalPort: local, RemoteHost: host, RemotePort: remote})
		}
	}
	return result
}

func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		if home, err := os.UserHomeDir(); err == nil {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
