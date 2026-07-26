package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/johnniewhite/ssher/internal/legacy"
	"github.com/johnniewhite/ssher/internal/paths"
	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/ui"
	"github.com/johnniewhite/ssher/internal/vault"
)

var importLegacyKeepFiles bool

var importLegacyCmd = &cobra.Command{
	Use:   "import-legacy",
	Short: "Import a Python-version vault into the new format (one-shot)",
	Long: `Reads ~/.ssher/servers.enc (Python Fernet vault) plus the plaintext
sidecar files (history.json, profiles.json, aliases.json), prompts for the
old master password, prompts for a new master password, and writes the new
vault.bin in the modern AES-256-GCM + Argon2id format.

Original files are left in place by default so you can verify the import
before deleting them. Pass --replace to remove them on success.`,
	RunE: func(c *cobra.Command, args []string) error {
		exists, err := vault.Exists()
		if err != nil {
			return err
		}
		if exists {
			return errors.New("a new-format vault already exists at ~/.ssher/vault.bin -- aborting to avoid clobbering it")
		}
		legacyVault, err := paths.LegacyServers()
		if err != nil {
			return err
		}
		if _, err := os.Stat(legacyVault); err != nil {
			return fmt.Errorf("no legacy vault at %s", legacyVault)
		}
		legacySalt, err := paths.LegacySalt()
		if err != nil {
			return err
		}
		salt, err := os.ReadFile(legacySalt)
		if err != nil {
			return fmt.Errorf("read legacy salt (%s): %w", legacySalt, err)
		}

		oldPw, err := ui.PromptPassword("Old (Python) master password: ")
		if err != nil {
			return err
		}
		key := legacy.DeriveKey([]byte(oldPw), salt)

		// Verification token is encrypted "SSHER_VERIFIED".
		if err := verifyLegacyKey(key); err != nil {
			return fmt.Errorf("verification failed: %w", err)
		}

		blob, err := os.ReadFile(legacyVault)
		if err != nil {
			return err
		}
		pt, err := legacy.Decrypt(blob, key)
		if err != nil {
			return fmt.Errorf("decrypt legacy vault: %w", err)
		}

		var legacyServers []store.Server
		if err := json.Unmarshal(pt, &legacyServers); err != nil {
			// The Python tool wrote either a list or {"servers":[...]}. Try both.
			var wrapped struct {
				Servers []store.Server `json:"servers"`
			}
			if err2 := json.Unmarshal(pt, &wrapped); err2 != nil {
				return fmt.Errorf("parse decrypted servers: %w", err)
			}
			legacyServers = wrapped.Servers
		}

		newVault := store.New()
		newVault.Servers = legacyServers
		readSidecarHistory(newVault)
		readSidecarProfiles(newVault)
		readSidecarAliases(newVault)

		newPw, err := ui.PromptNewPassword("New master password")
		if err != nil {
			return err
		}

		// Write the complete imported document atomically so interruption can
		// never leave a new-format but empty vault blocking a retry.
		fresh, err := store.Initialise(newVault, []byte(newPw))
		if err != nil {
			return err
		}
		if err := vault.SaveSession(fresh.Key, fresh.Salt); err != nil {
			fmt.Println(ui.Warnf(fmt.Sprintf("import succeeded, but session refresh failed: %v", err)))
		}

		fmt.Println(ui.Successf(fmt.Sprintf(
			"imported %d server(s), %d profile(s), %d alias(es), %d history entr(ies)",
			len(newVault.Servers), len(newVault.Profiles),
			len(newVault.Aliases), len(newVault.History))))

		if importLegacyKeepFiles {
			fmt.Println(ui.Muted.Render("legacy files left in place; remove them once you're satisfied"))
			return nil
		}
		removeLegacyFiles()
		fmt.Println(ui.Successf("removed legacy files"))
		return nil
	},
}

func init() {
	importLegacyCmd.Flags().BoolVar(&importLegacyKeepFiles, "keep", true,
		"keep legacy files (servers.enc/.salt/.key/...) on disk after import (default true; pass --keep=false to delete)")
	rootCmd.AddCommand(importLegacyCmd)
}

func verifyLegacyKey(key []byte) error {
	keyFile, err := paths.LegacyKey()
	if err != nil {
		return err
	}
	token, err := os.ReadFile(keyFile)
	if err != nil {
		return fmt.Errorf("read verification token: %w", err)
	}
	pt, err := legacy.Decrypt(token, key)
	if err != nil {
		return fmt.Errorf("decrypt verification token: %w", err)
	}
	if strings.TrimSpace(string(pt)) != "SSHER_VERIFIED" {
		return errors.New("verification marker mismatch (vault may be corrupted)")
	}
	return nil
}

func readSidecarHistory(v *store.Vault) {
	path, err := paths.LegacyHistory()
	if err != nil {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var entries []store.HistoryEntry
	if err := json.Unmarshal(b, &entries); err != nil {
		// Python wrote {"connections":[...]} in some versions; try that.
		var wrapped struct {
			Connections []store.HistoryEntry `json:"connections"`
		}
		if err2 := json.Unmarshal(b, &wrapped); err2 == nil {
			entries = wrapped.Connections
		}
	}
	v.History = entries
}

func readSidecarProfiles(v *store.Vault) {
	path, err := paths.LegacyProfiles()
	if err != nil {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var profiles []store.Profile
	if err := json.Unmarshal(b, &profiles); err != nil {
		var wrapped struct {
			Profiles []store.Profile `json:"profiles"`
		}
		if err2 := json.Unmarshal(b, &wrapped); err2 == nil {
			profiles = wrapped.Profiles
		}
	}
	v.Profiles = profiles
}

func readSidecarAliases(v *store.Vault) {
	path, err := paths.LegacyAliases()
	if err != nil {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	aliases := map[string]string{}
	if err := json.Unmarshal(b, &aliases); err != nil {
		return
	}
	v.Aliases = aliases
}

func removeLegacyFiles() {
	for _, get := range []func() (string, error){
		paths.LegacyServers, paths.LegacySalt, paths.LegacyKey,
		paths.LegacyHistory, paths.LegacyProfiles, paths.LegacyAliases,
	} {
		if p, err := get(); err == nil {
			_ = os.Remove(p)
		}
	}
}
