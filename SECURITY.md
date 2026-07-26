# Security Policy

## Supported versions

Security fixes are provided for the latest published release. Users should upgrade through Homebrew or the newest GitHub release before reporting an issue that may already be fixed.

## Reporting a vulnerability

Do not open a public issue for suspected vulnerabilities.

Use [GitHub private vulnerability reporting](https://github.com/Johnniewhite/ssher/security/advisories/new) and include:

- affected version and operating system;
- the relevant command or component;
- reproduction steps using non-production credentials;
- impact and any suggested mitigation.

Do not submit real vaults, master passwords, private keys, recordings, or production server details. Construct a minimal test vault instead.

You should receive an acknowledgement within seven days. Valid reports will be investigated privately, with fixes and disclosure coordinated through a GitHub security advisory.

## Security boundaries

- `vault.bin` is encrypted with AES-256-GCM using an Argon2id-derived key.
- `.session` contains the derived vault key encrypted with a machine-bound wrapping key and expires after 30 minutes. Machine binding is defense in depth, not a substitute for protecting the local account.
- A process running as the same user while the vault is unlocked may be able to access credentials.
- Session recordings and explicit plaintext exports can contain secrets even though they are written with restrictive permissions.
- New SSH host keys use trust on first use and are stored in `~/.ssh/known_hosts`; changed keys are rejected.
- `wrap` intentionally supplies a password to another local process through a PTY. Prefer native `connect`, `exec`, and SFTP commands when possible.
