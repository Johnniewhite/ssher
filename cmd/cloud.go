package cmd

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cloudapi "github.com/johnniewhite/ssher/internal/cloud"
	"github.com/johnniewhite/ssher/internal/paths"
	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
)

var (
	cloudAPIURL       = cloudapi.DefaultAPIURL
	cloudAppURL       = cloudapi.DefaultAppURL
	cloudOrganization string
	cloudServerTarget string
	cloudIncludeKeys  bool
)

var cloudCmd = &cobra.Command{Use: "cloud", Short: "Sync servers through end-to-end encrypted SSHer Cloud"}

var cloudLoginCmd = &cobra.Command{
	Use: "login", Short: "Authorize this device with SSHer Cloud",
	RunE: func(c *cobra.Command, _ []string) error {
		key, err := cloudapi.LoadOrCreateDeviceKey()
		if err != nil {
			return err
		}
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "SSHer CLI"
		}
		client := cloudapi.NewClient(cloudAPIURL, "")
		deviceCode, err := client.CreateDeviceCode(c.Context(), hostname, runtime.GOOS+"/"+runtime.GOARCH, key.PublicKey().Bytes())
		if err != nil {
			return err
		}
		verificationURL := deviceCode.VerificationURI + "?code=" + strings.ReplaceAll(deviceCode.UserCode, "-", "")
		fmt.Println(ui.Title.Render("Connect SSHer Cloud"))
		fmt.Printf("Open %s\n\n", verificationURL)
		fmt.Printf("Enter code: %s\n", deviceCode.UserCode)
		fmt.Println(ui.Muted.Render("Waiting for approval…"))
		expires, err := time.Parse(time.RFC3339, deviceCode.ExpiresAt)
		if err != nil {
			expires = time.Now().Add(10 * time.Minute)
		}
		interval := time.Duration(deviceCode.Interval) * time.Second
		if interval < 2*time.Second {
			interval = 2 * time.Second
		}
		for time.Now().Before(expires) {
			select {
			case <-c.Context().Done():
				return c.Context().Err()
			case <-time.After(interval):
			}
			result, pending, err := client.ExchangeDeviceCode(c.Context(), deviceCode.DeviceCode)
			if err != nil {
				return err
			}
			if pending {
				continue
			}
			cfg := &cloudapi.Config{APIURL: cloudAPIURL, AppURL: cloudAppURL, SessionToken: result.Session.Token, ExpiresAt: result.Session.ExpiresAt, UserID: result.User.ID, UserEmail: result.User.Email, DeviceID: result.Device.ID}
			authClient := cloudapi.NewClient(cfg.APIURL, cfg.SessionToken)
			organizations, listErr := authClient.Organizations(c.Context())
			if listErr == nil && len(organizations) == 1 {
				cfg.OrganizationID = organizations[0].ID
				cfg.OrganizationName = organizations[0].Name
			}
			if err := cloudapi.SaveConfig(cfg); err != nil {
				return err
			}
			fmt.Println(ui.Successf(fmt.Sprintf("signed in as %s", result.User.Email)))
			if cfg.OrganizationID != "" {
				fmt.Println(ui.Successf(fmt.Sprintf("linked to %s", cfg.OrganizationName)))
			} else {
				fmt.Println(ui.Infof("run 'ssher cloud link --organization <slug>' to select a workspace"))
			}
			return nil
		}
		return errors.New("device code expired; run 'ssher cloud login' again")
	},
}

var cloudLinkCmd = &cobra.Command{
	Use: "link", Short: "Link this CLI to a workspace",
	RunE: func(c *cobra.Command, _ []string) error {
		cfg, err := cloudapi.LoadConfig()
		if err != nil {
			return err
		}
		organizations, err := cloudapi.NewClient(cfg.APIURL, cfg.SessionToken).Organizations(c.Context())
		if err != nil {
			return err
		}
		if cloudOrganization == "" {
			if len(organizations) == 1 {
				cloudOrganization = organizations[0].ID
			} else {
				fmt.Println("Available workspaces:")
				for _, org := range organizations {
					fmt.Printf("  %-24s %-18s %s\n", org.Name, org.Slug, org.Role)
				}
				return errors.New("choose one with --organization <name, slug, or ID>")
			}
		}
		for _, org := range organizations {
			if strings.EqualFold(cloudOrganization, org.ID) || strings.EqualFold(cloudOrganization, org.Slug) || strings.EqualFold(cloudOrganization, org.Name) {
				cfg.OrganizationID, cfg.OrganizationName = org.ID, org.Name
				if err := cloudapi.SaveConfig(cfg); err != nil {
					return err
				}
				fmt.Println(ui.Successf(fmt.Sprintf("linked to %s", org.Name)))
				return nil
			}
		}
		return fmt.Errorf("workspace %q not found", cloudOrganization)
	},
}

var cloudStatusCmd = &cobra.Command{
	Use: "status", Short: "Show cloud account and sync status",
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := cloudapi.LoadConfig()
		if err != nil {
			return err
		}
		fmt.Printf("Account:   %s\nDevice:    %s\nWorkspace: %s\nAPI:       %s\n", cfg.UserEmail, cfg.DeviceID, valueOr(cfg.OrganizationName, "not linked"), cfg.APIURL)
		return nil
	},
}

var cloudLogoutCmd = &cobra.Command{
	Use: "logout", Short: "Remove the local SSHer Cloud session",
	RunE: func(_ *cobra.Command, _ []string) error {
		if err := cloudapi.RemoveConfig(); err != nil {
			return err
		}
		fmt.Println(ui.Successf("signed out of SSHer Cloud"))
		return nil
	},
}

var cloudPullCmd = &cobra.Command{
	Use: "pull", Short: "Pull and decrypt workspace servers into the local vault",
	RunE: func(c *cobra.Command, _ []string) error { _, err := pullCloud(c.Context()); return err },
}

var cloudPushCmd = &cobra.Command{
	Use: "push", Short: "Encrypt and push local server changes",
	RunE: func(c *cobra.Command, _ []string) error { _, err := pushCloud(c.Context()); return err },
}

var cloudSyncCmd = &cobra.Command{
	Use: "sync", Short: "Pull remote changes, then push local changes",
	RunE: func(c *cobra.Command, _ []string) error {
		pulled, err := pullCloud(c.Context())
		if err != nil {
			return err
		}
		pushed, err := pushCloud(c.Context())
		if err != nil {
			return err
		}
		fmt.Println(ui.Successf(fmt.Sprintf("sync complete (%d pulled, %d pushed)", pulled, pushed)))
		return nil
	},
}

func cloudWorkspace(ctx context.Context) (*cloudapi.Config, *cloudapi.Client, []byte, error) {
	cfg, err := cloudapi.LoadConfig()
	if err != nil {
		return nil, nil, nil, err
	}
	if cfg.OrganizationID == "" {
		return nil, nil, nil, errors.New("no workspace linked; run 'ssher cloud link'")
	}
	key, err := cloudapi.LoadOrCreateDeviceKey()
	if err != nil {
		return nil, nil, nil, err
	}
	client := cloudapi.NewClient(cfg.APIURL, cfg.SessionToken)
	envelope, err := client.WorkspaceEnvelope(ctx, cfg.OrganizationID, cfg.DeviceID)
	if apiErr := new(cloudapi.APIError); errors.As(err, &apiErr) && apiErr.Status == 404 {
		return nil, nil, nil, errors.New("this device is awaiting encrypted workspace access; ask an owner or admin to open the Servers page")
	}
	if err != nil {
		return nil, nil, nil, err
	}
	workspaceKey, err := cloudapi.UnwrapWorkspaceKey(envelope, key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unlock workspace key: %w", err)
	}
	return cfg, client, workspaceKey, nil
}

func pullCloud(ctx context.Context) (int, error) {
	cfg, client, key, err := cloudWorkspace(ctx)
	if err != nil {
		return 0, err
	}
	records, err := client.Servers(ctx, cfg.OrganizationID)
	if err != nil {
		return 0, err
	}
	saved, err := ui.LoadVault()
	if err != nil {
		return 0, err
	}
	changed, conflicts := 0, []string{}
	for _, record := range records {
		index := cloudServerIndex(saved.Vault, record.ID)
		if record.DeletedAt != nil {
			if index >= 0 {
				local := saved.Vault.Servers[index]
				if local.CloudSyncedHash != "" && cloudapi.ServerHash(local) != local.CloudSyncedHash {
					conflicts = append(conflicts, local.Name+" (deleted remotely)")
					continue
				}
				saved.Vault.Servers = append(saved.Vault.Servers[:index], saved.Vault.Servers[index+1:]...)
				changed++
			}
			continue
		}
		if index >= 0 {
			local := saved.Vault.Servers[index]
			if record.Revision <= local.CloudRevision {
				continue
			}
			if local.CloudSyncedHash != "" && cloudapi.ServerHash(local) != local.CloudSyncedHash {
				conflicts = append(conflicts, local.Name)
				continue
			}
		}
		payload, err := cloudapi.DecryptServer(record, key)
		if err != nil {
			return changed, fmt.Errorf("decrypt server %s: %w", record.ID, err)
		}
		server := payload.Server
		server.CloudID, server.CloudRevision, server.CloudTeamIDs = record.ID, record.Revision, record.TeamIDs
		if payload.PrivateKey != "" {
			path, err := saveManagedPrivateKey(record.ID, []byte(payload.PrivateKey))
			if err != nil {
				return changed, err
			}
			server.KeyPath = path
		}
		server.Touch()
		server.CloudSyncedHash = cloudapi.ServerHash(server)
		if index >= 0 {
			saved.Vault.Servers[index] = server
		} else if nameIndex := serverNameIndex(saved.Vault, server.Name); nameIndex >= 0 && saved.Vault.Servers[nameIndex].CloudID == "" {
			saved.Vault.Servers[nameIndex] = server
		} else if err := saved.Vault.AddServer(server); err != nil {
			return changed, err
		}
		changed++
	}
	if changed > 0 {
		if err := saved.SaveAndRefreshSession(); err != nil {
			return changed, err
		}
	}
	if len(conflicts) > 0 {
		return changed, fmt.Errorf("local changes conflict with newer cloud data: %s; push or resolve them first", strings.Join(conflicts, ", "))
	}
	fmt.Println(ui.Successf(fmt.Sprintf("pulled %d server change(s) from %s", changed, cfg.OrganizationName)))
	return changed, nil
}

func pushCloud(ctx context.Context) (int, error) {
	cfg, client, key, err := cloudWorkspace(ctx)
	if err != nil {
		return 0, err
	}
	saved, err := ui.LoadVault()
	if err != nil {
		return 0, err
	}
	selectedName := ""
	if cloudServerTarget != "" {
		target, _, err := saved.Vault.ResolveTarget(cloudServerTarget)
		if err != nil {
			return 0, err
		}
		selectedName = target.Name
	}
	pushed := 0
	for index := range saved.Vault.Servers {
		server := &saved.Vault.Servers[index]
		if selectedName != "" && server.Name != selectedName {
			continue
		}
		currentHash := cloudapi.ServerHash(*server)
		if server.CloudID != "" && server.CloudSyncedHash == currentHash && !cloudIncludeKeys {
			continue
		}
		if server.CloudID == "" {
			server.CloudID, err = newUUID()
			if err != nil {
				return pushed, err
			}
		}
		payload := cloudapi.PayloadForServer(*server)
		if cloudIncludeKeys && server.AuthType == store.AuthKey && server.KeyPath != "" {
			keyBytes, err := readPrivateKey(server.KeyPath)
			if err != nil {
				return pushed, fmt.Errorf("read key for %s: %w", server.Name, err)
			}
			payload.PrivateKey = string(keyBytes)
		}
		nextRevision := server.CloudRevision + 1
		if nextRevision < 1 {
			nextRevision = 1
		}
		record, err := cloudapi.EncryptServer(payload, key, cfg.OrganizationID, server.CloudID, nextRevision)
		if err != nil {
			return pushed, err
		}
		record.TeamIDs = server.CloudTeamIDs
		if server.CloudRevision == 0 {
			record, err = client.CreateServer(ctx, cfg.OrganizationID, record)
		} else {
			record, err = client.UpdateServer(ctx, cfg.OrganizationID, record, server.CloudRevision)
		}
		if err != nil {
			return pushed, fmt.Errorf("push %s: %w", server.Name, err)
		}
		server.CloudRevision, server.CloudSyncedHash = record.Revision, cloudapi.ServerHash(*server)
		if err := saved.SaveAndRefreshSession(); err != nil {
			return pushed, err
		}
		pushed++
	}
	fmt.Println(ui.Successf(fmt.Sprintf("pushed %d server change(s) to %s", pushed, cfg.OrganizationName)))
	return pushed, nil
}

func cloudServerIndex(vault *store.Vault, id string) int {
	for i := range vault.Servers {
		if vault.Servers[i].CloudID == id {
			return i
		}
	}
	return -1
}
func serverNameIndex(vault *store.Vault, name string) int {
	for i := range vault.Servers {
		if vault.Servers[i].Name == name {
			return i
		}
	}
	return -1
}
func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

func readPrivateKey(path string) ([]byte, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > 1<<20 {
		return nil, errors.New("private key exceeds 1 MiB")
	}
	return os.ReadFile(path)
}

func saveManagedPrivateKey(serverID string, contents []byte) (string, error) {
	dir, err := paths.CloudKeysDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, paths.DirMode); err != nil {
		return "", err
	}
	path := filepath.Join(dir, serverID)
	tmp, err := os.CreateTemp(dir, ".key-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(paths.FileMode); err != nil {
		tmp.Close()
		return "", err
	}
	if _, err := tmp.Write(contents); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return "", err
	}
	return path, nil
}

func init() {
	cloudLoginCmd.Flags().StringVar(&cloudAPIURL, "api-url", cloudapi.DefaultAPIURL, "SSHer Cloud API URL")
	cloudLoginCmd.Flags().StringVar(&cloudAppURL, "app-url", cloudapi.DefaultAppURL, "SSHer Cloud app URL")
	cloudLinkCmd.Flags().StringVarP(&cloudOrganization, "organization", "o", "", "workspace name, slug, or ID")
	cloudPushCmd.Flags().StringVarP(&cloudServerTarget, "server", "s", "", "push only one server")
	cloudPushCmd.Flags().BoolVar(&cloudIncludeKeys, "include-keys", false, "include private-key contents in the encrypted payload")
	cloudCmd.AddCommand(cloudLoginCmd, cloudLinkCmd, cloudStatusCmd, cloudPullCmd, cloudPushCmd, cloudSyncCmd, cloudLogoutCmd)
	rootCmd.AddCommand(cloudCmd)
}
