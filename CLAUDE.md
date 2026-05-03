# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

ssher — Go CLI for managing encrypted SSH server configurations, plus a drop-in `sshpass` replacement. Single binary, single entry point: `main.go` → `cmd.Execute()` → cobra command tree.

The repo was previously a Python package; the Python tree was removed in the Go rewrite. The full Python implementation is preserved in git history at commit `8527ca2 init files` if you need to compare behaviour.

## Common commands

```bash
go build ./...            # compile
go test ./...             # run tests (vault + legacy packages have real coverage)
go test ./internal/vault  # run a single package's tests
go run . --help           # invoke the CLI from source
go run . vault status     # smoke-test
go install .              # install ssher to $GOBIN

# Release-time
goreleaser release --snapshot --clean  # local cross-compile dry run
```

There is **no linter configured**. `go vet ./...` is the bar. Don't introduce golangci-lint without asking.

## Architecture

The package boundary that matters most: **process state vs. the world**. Everything in `internal/` is testable without a network or a real terminal; the cobra commands in `cmd/` are the thin shell that wires interactive I/O onto pure logic.

### `internal/` packages

- **`paths/`** — single source of truth for `~/.ssher/...` resolution and file-mode constants. Never hardcode a path under `~/.ssher/` outside this package; import the appropriate accessor (`paths.VaultFile()`, `paths.SessionFile()`, …) and let it return the absolute path.
- **`vault/`** — pure crypto + on-disk format. AES-256-GCM with Argon2id KDF, header carries the KDF parameters so we can ratchet costs upward without breaking existing files. `Encrypt`/`Decrypt` know nothing about prompts or files; `LoadFile`/`SaveFile`/`SaveSession` are the I/O layer. The session file binds the cached vault key to a per-host fingerprint via HKDF — defense-in-depth against `~/.ssher/.session` being lifted off-host.
- **`legacy/`** — one-shot reader for the Python Fernet vault format. Used only by `cmd/import_legacy.go`. Has its own tests with a synthesised Fernet writer so we can round-trip without depending on a Python install. We never *write* Fernet.
- **`store/`** — data model (`Server`, `Profile`, `Alias`, `HistoryEntry`, `Vault`) and the cached `Saved{Vault, Key, Salt}` bundle that flows through every command. `LoadFromSession`/`LoadWithPassword`/`InitialiseEmpty` are the three entry points; everything else mutates `*Vault` and calls `Saved.SaveAndRefreshSession`.
- **`ssh/`** — wraps `golang.org/x/crypto/ssh`. `Dial` walks the JumpHost chain (single hop — the data model doesn't represent multi-hop chains). `Interactive(c, opts)` opens a PTY session with optional output teeing; `Run(c, cmd)` is the non-interactive path used by `exec`. **Never shell out to `/usr/bin/ssh`** from this package — that's the architectural choice the rewrite was built on.
- **`transfer/`** — SFTP via `pkg/sftp`, rsync by shelling out (no Go-native rsync exists). **rsync currently requires key auth** because we don't inject passwords into rsync's child SSH process; documented limitation, easy to extend later by routing through `internal/pty` if needed.
- **`pty/`** — the *only* place we use `creack/pty` and password-prompt scanning. Used exclusively by `cmd/wrap.go`, which is the sshpass replacement and has no choice but to wrap an arbitrary user-supplied command. **Do not grow another PTY/expect path elsewhere** — route through here.
- **`recording/`** — asciicast v2 reader/writer. The `Writer` is wired into `Interactive` via `InteractiveOptions.TeeOutput` so `connect --record` can capture session output without a separate code path.
- **`ui/`** — lipgloss styles and the `huh`-based form helpers. `LoadVault()` is the canonical "open the vault, prompting only if necessary" entry point that every command uses.
- **`clipboard/`**, **`pwgen/`** — small, self-contained utilities.

### `cmd/` packages

One file per subcommand, plus `root.go` and `interactive.go`. Conventions:

- `RunE` always: `saved, err := ui.LoadVault()` first, then mutate, then `saved.SaveAndRefreshSession()`. Never call into `vault.SaveFile` directly from a command — go through the `Saved` bundle.
- A bare `ssher <name-or-number>` falls through to `runConnect` from `rootCmd.RunE`. Preserve this when adding subcommands; it's how `ssher prod` works without a `connect` keyword.
- `wrap`, `generate-password`, and `completion` are the subcommands that **don't** call `ui.LoadVault()` (no encryption needed). Don't add them to a vault-required path.

### Key invariants

- All persistent files under `~/.ssher/` are written `0600`; the directory is `0700`. The constants live in `paths.FileMode` / `paths.DirMode`.
- The master password is never written to disk — only an Argon2id-derived key is cached in the session file (and that key is itself wrapped with a host-bound HKDF key).
- Plaintext exports (`export-csv`, default `export-json`) **omit passwords**. `--include-passwords` is the explicit opt-in for `export-json`. Don't change this default; it's a real security boundary.

## Deps that aren't obvious from go.mod

- `github.com/charmbracelet/huh` — interactive forms (used in `add`, `edit`, `delete` confirmation, vault password setup, interactive mode menu).
- `github.com/charmbracelet/lipgloss` — styling only; we are intentionally **not** using `bubbletea` (that would be a redesign of the UX, not a port).
- `github.com/creack/pty` — only in `internal/pty`, only for `wrap`.
- `golang.org/x/crypto/ssh` + `pkg/sftp` — the SSH stack.
