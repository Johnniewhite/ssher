# Contributing to ssher

Thanks for helping improve ssher. Bug reports, documentation improvements, tests, and focused code changes are welcome.

## Before opening an issue

- Search existing issues and the changelog.
- Use the bug or feature template.
- Remove credentials, private keys, vault files, recordings, usernames, and private server addresses.
- Report security vulnerabilities privately as described in [SECURITY.md](SECURITY.md).

## Development

Requirements:

- Go version declared in `go.mod`
- macOS or Linux
- `rsync` only when testing the `sync` command

```bash
git clone git@github.com:Johnniewhite/ssher.git
cd ssher
go test ./...
go vet ./...
go build ./...
```

The command entry point is `main.go`; the Cobra commands live in `cmd/`, and testable implementation packages live in `internal/`. See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) before changing vault, SSH, or persistence behavior.

## Pull requests

1. Create a focused branch from `main`.
2. Add regression tests for behavioral fixes.
3. Run `gofmt`, `go vet ./...`, and `go test -race ./...`.
4. Update README or architecture documentation when behavior changes.
5. Explain compatibility and security implications in the pull request.

Do not change the encrypted vault format casually. New readers must remain compatible with existing vaults, and any KDF change must keep the header parameters and derived encryption key consistent.

## Commit style

Use concise imperative subjects, for example:

```text
Fix KDF parameter preservation during vault saves
Add timeout handling to fleet execution
Document native SSH limitations
```

By participating, you agree to follow the [Code of Conduct](CODE_OF_CONDUCT.md).
