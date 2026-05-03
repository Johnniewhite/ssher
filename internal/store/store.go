package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/johnniewhite/ssher/internal/vault"
)

// Saved bundles everything callers need to write the vault back without
// re-deriving the Argon2id key on every command.
type Saved struct {
	Vault *Vault
	Key   []byte
	Salt  [16]byte
}

// LoadWithPassword decrypts the vault using a master password (cold path).
func LoadWithPassword(password []byte) (*Saved, error) {
	pt, h, key, err := vault.LoadFile(password)
	if err != nil {
		return nil, err
	}
	v, err := unmarshal(pt)
	if err != nil {
		return nil, err
	}
	return &Saved{Vault: v, Key: key, Salt: h.SaltBytes()}, nil
}

// LoadFromSession decrypts the vault using a cached session key. Returns
// vault.ErrNoSession if there isn't a usable session, in which case callers
// should prompt for the master password and call LoadWithPassword.
func LoadFromSession() (*Saved, error) {
	key, salt, err := vault.LoadSession()
	if err != nil {
		return nil, err
	}
	exists, err := vault.Exists()
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("store: session present but vault file missing")
	}
	// Need a fresh read of the bytes to AEAD-open with the cached key.
	blob, h, err := vault.OpenBlobWithKey(key)
	if err != nil {
		return nil, err
	}
	if h.SaltBytes() != salt {
		// Vault was rewritten with a new salt while a session was live;
		// invalidate the session and force re-auth.
		_ = vault.ClearSession()
		return nil, vault.ErrNoSession
	}
	v, err := unmarshal(blob)
	if err != nil {
		return nil, err
	}
	return &Saved{Vault: v, Key: key, Salt: salt}, nil
}

// Save writes the vault back atomically using the cached key.
func (s *Saved) Save() error {
	pt, err := marshal(s.Vault)
	if err != nil {
		return err
	}
	return vault.SaveFile(pt, s.Key, s.Salt, vault.DefaultParams)
}

// SaveAndRefreshSession persists the vault and bumps the session expiry.
func (s *Saved) SaveAndRefreshSession() error {
	if err := s.Save(); err != nil {
		return err
	}
	return vault.SaveSession(s.Key, s.Salt)
}

// InitialiseEmpty writes a brand-new empty vault with the supplied password.
// Used on first-time setup.
func InitialiseEmpty(password []byte) (*Saved, error) {
	v := New()
	pt, err := marshal(v)
	if err != nil {
		return nil, err
	}
	if err := vault.SaveFileFromPassword(pt, password); err != nil {
		return nil, err
	}
	return LoadWithPassword(password)
}

func marshal(v *Vault) ([]byte, error) {
	if v.Version == 0 {
		v.Version = CurrentVaultVersion
	}
	if v.Aliases == nil {
		v.Aliases = map[string]string{}
	}
	return json.Marshal(v)
}

func unmarshal(b []byte) (*Vault, error) {
	v := &Vault{}
	if err := json.Unmarshal(b, v); err != nil {
		return nil, fmt.Errorf("decode vault json: %w", err)
	}
	if v.Aliases == nil {
		v.Aliases = map[string]string{}
	}
	return v, nil
}

// ----- Server operations ------------------------------------------------

func (v *Vault) AddServer(s Server) error {
	s.Touch()
	if s.Name == "" {
		return errors.New("server name is required")
	}
	if v.findIndex(s.Name) >= 0 {
		return fmt.Errorf("server %q already exists", s.Name)
	}
	v.Servers = append(v.Servers, s)
	return nil
}

func (v *Vault) UpdateServer(name string, mutate func(*Server)) error {
	idx := v.findIndex(name)
	if idx < 0 {
		return fmt.Errorf("server %q not found", name)
	}
	mutate(&v.Servers[idx])
	return nil
}

func (v *Vault) DeleteServer(name string) error {
	idx := v.findIndex(name)
	if idx < 0 {
		return fmt.Errorf("server %q not found", name)
	}
	v.Servers = append(v.Servers[:idx], v.Servers[idx+1:]...)
	// Drop aliases that point at the deleted server.
	for alias, target := range v.Aliases {
		if target == name {
			delete(v.Aliases, alias)
		}
	}
	return nil
}

func (v *Vault) findIndex(name string) int {
	for i := range v.Servers {
		if v.Servers[i].Name == name {
			return i
		}
	}
	return -1
}

// SortedServers returns a copy ordered favourites-first, then by name.
// The Python UI uses the same ordering -- preserves muscle memory for
// number-based selection.
func (v *Vault) SortedServers() []Server {
	out := append([]Server(nil), v.Servers...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].IsFavorite != out[j].IsFavorite {
			return out[i].IsFavorite
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

// ResolveTarget locates a server by name, alias, 1-based index in the sorted
// view, or unique fuzzy substring match (case-insensitive). Returns the
// canonical server, plus a small explanation of how it was matched (useful
// for confirmations like "matched alias 'pw' -> production-web").
func (v *Vault) ResolveTarget(target string) (*Server, string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, "", errors.New("empty target")
	}
	sorted := v.SortedServers()

	// 1-based index into the sorted view.
	if n, err := strconv.Atoi(target); err == nil {
		if n < 1 || n > len(sorted) {
			return nil, "", fmt.Errorf("server index %d out of range (1..%d)", n, len(sorted))
		}
		idx := v.findIndex(sorted[n-1].Name)
		return &v.Servers[idx], fmt.Sprintf("by index %d", n), nil
	}

	// Exact name match.
	if idx := v.findIndex(target); idx >= 0 {
		return &v.Servers[idx], "by name", nil
	}

	// Alias match.
	if canonical, ok := v.Aliases[target]; ok {
		if idx := v.findIndex(canonical); idx >= 0 {
			return &v.Servers[idx], fmt.Sprintf("alias -> %s", canonical), nil
		}
	}

	// Case-insensitive fuzzy: unique substring of name OR alias.
	lower := strings.ToLower(target)
	var matches []int
	for i, s := range v.Servers {
		if strings.Contains(strings.ToLower(s.Name), lower) {
			matches = append(matches, i)
			continue
		}
		for _, a := range s.Aliases {
			if strings.Contains(strings.ToLower(a), lower) {
				matches = append(matches, i)
				break
			}
		}
	}
	switch len(matches) {
	case 0:
		return nil, "", fmt.Errorf("no server matches %q", target)
	case 1:
		return &v.Servers[matches[0]], "fuzzy match", nil
	default:
		names := make([]string, 0, len(matches))
		for _, i := range matches {
			names = append(names, v.Servers[i].Name)
		}
		return nil, "", fmt.Errorf("ambiguous match %q: %s", target, strings.Join(names, ", "))
	}
}

// Groups returns the set of group names in stable order.
func (v *Vault) Groups() []string {
	seen := map[string]struct{}{}
	for _, s := range v.Servers {
		seen[s.Group] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for g := range seen {
		out = append(out, g)
	}
	sort.Strings(out)
	return out
}

// ServersInGroup returns servers belonging to a group.
func (v *Vault) ServersInGroup(group string) []Server {
	var out []Server
	for _, s := range v.Servers {
		if s.Group == group {
			out = append(out, s)
		}
	}
	return out
}

// RecordHistory appends an entry, keeping only the most recent 1000.
func (v *Vault) RecordHistory(e HistoryEntry) {
	v.History = append(v.History, e)
	if len(v.History) > 1000 {
		v.History = v.History[len(v.History)-1000:]
	}
}
