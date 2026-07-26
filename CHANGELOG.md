# Changelog

Notable changes to ssher are documented here. The project follows [Semantic Versioning](https://semver.org/).

## [Unreleased]

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

[Unreleased]: https://github.com/Johnniewhite/ssher/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/Johnniewhite/ssher/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/Johnniewhite/ssher/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Johnniewhite/ssher/releases/tag/v0.1.0
