# Architecture

ssher is a single Go binary. `main.go` delegates to the Cobra command tree in `cmd/`; commands load the encrypted store, call an implementation package under `internal/`, and render a result.

## Boundaries

| Package | Responsibility |
| --- | --- |
| `cmd` | CLI parsing, prompts, orchestration, and user-facing output |
| `internal/store` | Vault document model, target resolution, aliases, history, and save state |
| `internal/vault` | AES-GCM format, Argon2id KDF, atomic files, and cached sessions |
| `internal/ssh` | Authentication, host-key verification, jump host, sessions, keepalives, and forwarding |
| `internal/transfer` | Native SFTP and external `rsync` integration |
| `internal/openssh` | OpenSSH config discovery and canonical `ssh -G` resolution for portable imports |
| `internal/pty` | Password-prompt handling used only by `ssher wrap`, backed by Unix PTYs or Windows ConPTY |
| `internal/recording` | Asciicast v2 writing and replay |
| `internal/ui` | Vault authentication, forms, tables, and styles |
| `internal/paths` | Canonical paths, platform-safe atomic replacement, permissions, and host fingerprint |

## Vault lifecycle

```text
password unlock ──Argon2id──> key ──AES-GCM──> JSON vault
                                  │
                                  └──machine-bound wrap──> .session

later command ──.session──> cached key ──AES-GCM──> JSON vault
```

The vault header stores the KDF parameters, salt, and nonce. A cached key must always be saved with the parameters that produced it. Weak parameters are upgraded only during password-based unlock, when a new matching key can be derived. Vault and session writes use a temporary file, `fsync`, and atomic replacement (`rename` on Unix and `MoveFileEx` on Windows).

`Vault.Aliases` is canonical. Legacy per-server aliases are migrated on load. The JSON vault format remains at version 1.

## SSH lifecycle

1. Resolve the target by index, exact name, alias, or unique fuzzy match.
2. Build authentication from the stored password, configured key, ssh-agent, or standard default keys.
3. verify `~/.ssh/known_hosts`; append an unknown key using trust on first use, but reject changed keys.
4. Dial directly or through one saved jump host.
5. Start keepalives, forwarding, and either an interactive PTY or command session.
6. Record history and persist successful-connection metadata.

`import-ssh-config` discovers concrete aliases and delegates OpenSSH precedence,
`Include`, and `Match` evaluation to the installed `ssh -G`. It then stores the
portable subset understood by the native stack. The native dialer still does
not execute arbitrary OpenSSH directives; X11 and custom options remain
available through `export-config` and external OpenSSH tools.

## Compatibility rules

- Keep existing vaults readable.
- Preserve KDF header/key consistency.
- Write files under `~/.ssher` with `0600` and directories with `0700`.
- On Windows, map that directory to `%USERPROFILE%\.ssher`, rely on user-profile ACLs, and use MachineGuid for session binding.
- Omit passwords from plaintext exports unless explicitly requested.
- Keep PTY prompt injection isolated to `internal/pty`.
- Do not shell out to `ssh` for native connect, exec, or SFTP operations.
