package ui

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"

	"github.com/johnniewhite/ssher/internal/store"
	"github.com/johnniewhite/ssher/internal/vault"
)

const minMasterPasswordLength = 12

// LoadVault is the canonical entry point for every command that touches the
// encrypted store. It tries (in order):
//  1. an active session (no prompt)
//  2. interactive password prompt
//  3. first-time setup if no vault exists yet
//
// On success, callers receive a *store.Saved they can mutate and re-save.
// The session is refreshed on every successful unlock.
func LoadVault() (*store.Saved, error) {
	exists, err := vault.Exists()
	if err != nil {
		return nil, err
	}

	if !exists {
		return firstTimeSetup()
	}

	if s, err := store.LoadFromSession(); err == nil {
		return s, nil
	} else if !errors.Is(err, vault.ErrNoSession) {
		// Non-recoverable error reading session; fall through to password.
		fmt.Fprintln(os.Stderr, Warnf("session unreadable, falling back to password prompt"))
	}

	return promptUnlock()
}

func firstTimeSetup() (*store.Saved, error) {
	fmt.Println(Title.Render("First-time setup"))
	fmt.Println(Muted.Render("Choose a master password to encrypt your SSH configurations."))
	fmt.Println()

	var pw, confirm string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Master password").
				EchoMode(huh.EchoModePassword).
				Value(&pw).
				Validate(func(s string) error {
					if len(s) < minMasterPasswordLength {
						return fmt.Errorf("password must be at least %d characters", minMasterPasswordLength)
					}
					return nil
				}),
			huh.NewInput().
				Title("Confirm password").
				EchoMode(huh.EchoModePassword).
				Value(&confirm).
				Validate(func(s string) error {
					if s != pw {
						return fmt.Errorf("passwords do not match")
					}
					return nil
				}),
		),
	)
	if err := form.Run(); err != nil {
		return nil, err
	}

	saved, err := store.InitialiseEmpty([]byte(pw))
	if err != nil {
		return nil, err
	}
	if err := vault.SaveSession(saved.Key, saved.Salt); err != nil {
		fmt.Fprintln(os.Stderr, Warnf(fmt.Sprintf("could not write session: %v", err)))
	}
	fmt.Println(Successf("master password set"))
	return saved, nil
}

func promptUnlock() (*store.Saved, error) {
	pw, err := readPassword("Master password: ")
	if err != nil {
		return nil, err
	}
	saved, err := store.LoadWithPassword([]byte(pw))
	if err != nil {
		if errors.Is(err, vault.ErrBadPassword) {
			return nil, errors.New("incorrect master password")
		}
		return nil, err
	}
	if err := vault.SaveSession(saved.Key, saved.Salt); err != nil {
		fmt.Fprintln(os.Stderr, Warnf(fmt.Sprintf("could not refresh session: %v", err)))
	}
	return saved, nil
}

// readPassword prefers golang.org/x/term for echo suppression. If stdin isn't
// a TTY we fall back to plain ReadString -- mostly relevant for piped input
// in scripted use.
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	if term.IsTerminal(int(os.Stdin.Fd())) {
		b, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", fmt.Errorf("read password: %w", err)
		}
		return string(b), nil
	}
	var s string
	if _, err := fmt.Scanln(&s); err != nil {
		return "", err
	}
	return s, nil
}

// PromptNewPassword runs the same two-field flow as first-time setup.
// Used by `vault change-password`.
func PromptNewPassword(title string) (string, error) {
	var pw, confirm string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(title).
				EchoMode(huh.EchoModePassword).
				Value(&pw).
				Validate(func(s string) error {
					if len(s) < minMasterPasswordLength {
						return fmt.Errorf("password must be at least %d characters", minMasterPasswordLength)
					}
					return nil
				}),
			huh.NewInput().
				Title("Confirm").
				EchoMode(huh.EchoModePassword).
				Value(&confirm).
				Validate(func(s string) error {
					if s != pw {
						return fmt.Errorf("passwords do not match")
					}
					return nil
				}),
		),
	)
	if err := form.Run(); err != nil {
		return "", err
	}
	return pw, nil
}

// PromptPassword does a one-shot password read (no confirmation).
// Used for legacy import where we need the OLD vault password.
func PromptPassword(prompt string) (string, error) {
	return readPassword(prompt)
}
