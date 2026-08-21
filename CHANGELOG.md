# Changelog

Notable changes to ssher are documented here. The project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.3.3] - 2026-08-21

### Added

- Add `ssher cloud authorize-devices` so an authorized CLI can grant pending
  workspace devices encrypted access without exposing the workspace key to the
  Cloud API.

### Fixed

- Restore browser access safely when a replacement device needs a fresh
  workspace-key envelope.

## [0.3.2] - 2026-08-21

### Added

- Add anonymous, privacy-safe installation registration so project owners can
  measure released CLI adoption by version and platform without collecting
  commands, server details, account data, hostnames, or email addresses.
- Add a complete documentation hub covering setup, commands, Cloud sync,
  configuration, and security.

### Changed

- Redesign getssher.com around the visual language of ssher Cloud and give the
  CLI, browser terminals, team workspaces, and security model a unified home.

## [0.3.1] - 2026-08-21

### Fixed

- Submit injected password prompts with the Windows ConPTY Enter sequence.
- Close ConPTY after a wrapped no-prompt command exits so short-lived and
  key-authenticated commands return normally instead of waiting indefinitely.
- Exercise ConPTY command output, password injection, process exit, and the
  complete Windows build in native Windows CI.

## [0.3.0] - 2026-08-21

### Added

- Add native Windows x64 and ARM64 builds packaged as verified ZIP archives.
- Add Windows ConPTY support for interactive `ssher wrap` commands, PowerShell
  clipboard integration, OpenSSH agent named-pipe authentication, automatic
  Cloud login browser launch, and PowerShell completion.
- Add a checksum-verifying, per-user PowerShell installer and Windows install
  option on getssher.com.
- Add a Windows CI job that vets, tests, and builds the application on a real
  Windows runner.

### Security

- Bind cached vault sessions to the Windows MachineGuid.
- Use replace-existing atomic file moves for vault, session, Cloud config, and
  managed private-key updates on Windows.

## [0.2.0] - 2026-08-20

### Added

- Add end-to-end encrypted ssher Cloud sync with browser-based device approval,
  workspace linking, revision-aware pull/push/sync, and optional encrypted
  private-key sharing.
- Add cloud identity and team-assignment metadata to local server records while
  keeping the existing encrypted vault format backward compatible.
- Introduce ssher Cloud on the project website with browser SSH, team
  workspaces, and resumable terminal sessions.
- Add a 30-second Remotion announcement film with an original stereo score,
  message tones, keyboard detail, and launch sound design.

### Security

- Generate a per-device P-256 identity and unwrap workspace keys locally.
- Encrypt server payloads with AES-256-GCM and bind ciphertext to the workspace,
  server identity, and revision through authenticated additional data.
- Keep private-key files local by default; `--include-keys` is required before
  key contents can be encrypted into a cloud server record.

## [0.1.2] - 2026-07-26

### Security

- Preserve the Argon2id parameters associated with cached vault keys and safely ratchet weak vaults during password-based unlock.
- Reject malformed or excessive Argon2id header costs before key derivation and require stronger new master passwords.
- Re-key complete vault documents atomically without an intermediate empty vault.
- Serialize trust-on-first-use host-key updates and enforce private permissions on plaintext export files.

### Fixed

- Support agent-only authentication and standard Ed25519, ECDSA, and RSA default key paths.
- Honor keepalive and reconnect settings, bare-target connection flags, fleet execution timeouts, and remote failure exit status.
- Normalize aliases, remote forwarding, connection statistics, PTY no-prompt operation, and recording filenames.
- Treat Linux PTY closure as end-of-file so no-prompt wrapped commands exit cleanly.

### Documentation

- Add contributor, security, conduct, architecture, issue, pull-request, and CI documentation.
- Add the open-source [getssher.com](https://getssher.com) project website and reproducible Caddy deployment.

## [0.1.1] - 2026-06-20

- Harden password clipboard handling and SSH host-key verification.
- Align fleet execution and ping parallelism.
- Improve CLI consistency, transfers, and interactive workflows.

## [0.1.0] - 2026-05-05

- Initial Go rewrite with encrypted vaults, native SSH/SFTP, fleet execution, recordings, import/export, and release automation.

[Unreleased]: https://github.com/Johnniewhite/ssher/compare/v0.3.3...HEAD
[0.3.3]: https://github.com/Johnniewhite/ssher/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/Johnniewhite/ssher/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/Johnniewhite/ssher/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/Johnniewhite/ssher/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/Johnniewhite/ssher/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/Johnniewhite/ssher/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/Johnniewhite/ssher/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Johnniewhite/ssher/releases/tag/v0.1.0
