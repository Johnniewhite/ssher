<p align="center">
  <img src="logo.png" alt="ssher — Ultimate SSH Configuration Manager" width="480">
</p>

<p align="center">
  <em>The SSH config manager that <b>actually</b> remembers your servers.</em>
</p>

<p align="center">
  <a href="https://getssher.com"><img alt="Website" src="https://img.shields.io/badge/website-getssher.com-49f58a"></a>
  <a href="https://pkg.go.dev/github.com/johnniewhite/ssher"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/johnniewhite/ssher.svg"></a>
  <a href="https://github.com/Johnniewhite/ssher/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/Johnniewhite/ssher/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://goreportcard.com/report/github.com/johnniewhite/ssher"><img alt="Go Report" src="https://goreportcard.com/badge/github.com/johnniewhite/ssher"></a>
  <a href="LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
  <img alt="Platforms" src="https://img.shields.io/badge/platforms-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey">
</p>

---

`ssher` is a single-binary SSH companion. It keeps your servers in an encrypted vault, lets you connect with one word, runs commands across fleets in parallel, transfers files over native SFTP, and ships a drop-in replacement for `sshpass` — all without spawning the system `ssh` binary or shelling passwords through `pexpect`.

```text
$ ssher prod
[i] (matched alias -> production-web)
[ok] connected to production-web (deploy@10.0.4.21:22)
deploy@web1:~$
```

## Highlights

- **One word to connect.** `ssher prod`, `ssher 3`, `ssher pw` — name, index, alias, or fuzzy match.
- **Native SSH stack.** Authenticates with `golang.org/x/crypto/ssh` directly. No spawned `ssh`, no pexpect, no shell-string injection.
- **Encrypted vault.** AES‑256‑GCM with Argon2id KDF; parameters live in the file header so costs can ratchet upward over time.
- **Parallel fleet exec.** `ssher exec "uptime" --all` fans out across servers via goroutines.
- **`sshpass` replacement.** `ssher wrap` wraps any SSH-shaped command and feeds the password through a PTY at the prompt.
- **Recordings.** `ssher prod --record` captures asciicast v2 you can replay with `ssher record replay`.
- **Quality of life.** Profiles, aliases, favorites, groups, clipboard copy, CSV/JSON import-export, shell completion, `~/.ssh/config` export.
- **Drop-in upgrade path.** `ssher import-legacy` migrates a Python-version Fernet vault in one shot.
- **Encrypted team sync.** Link to SSHer Cloud and share server records without sending plaintext credentials to the service.

## Install

### Homebrew (macOS — recommended)

```bash
brew tap johnniewhite/ssher
brew install ssher
```

Upgrade with `brew update && brew upgrade ssher`. Also works on Linux via [Homebrew on Linux](https://docs.brew.sh/Homebrew-on-Linux).

### Windows (PowerShell)

Run this in PowerShell. The installer selects x64 or ARM64, verifies the
release checksum, installs to your user profile, and adds `ssher` to your user
`PATH` (no administrator access required):

```powershell
irm https://getssher.com/install.ps1 | iex
```

Open a new PowerShell window, then run `ssher`. Windows 10 version 1809,
Windows 11, and Windows Server 2019 or later are supported.

### Pre-built binary

```bash
# macOS arm64 shown; swap the asset name for your OS/arch
# (Darwin_arm64, Darwin_amd64, Linux_amd64, Linux_arm64,
#  Windows_amd64.zip, Windows_arm64.zip)
curl -LO https://github.com/johnniewhite/ssher/releases/latest/download/ssher_Darwin_arm64.tar.gz
tar -xzf ssher_Darwin_arm64.tar.gz && sudo mv ssher /usr/local/bin/
```

### From source

```bash
go install github.com/johnniewhite/ssher@latest
```

If `ssher` isn't found afterwards, `$GOBIN` isn't on `$PATH`. Add to `~/.zshrc` (or your shell's equivalent):

```bash
export PATH="$HOME/go/bin:$PATH"
```

Enable shell completion:

```bash
# zsh
eval "$(ssher completion zsh)"
# bash
eval "$(ssher completion bash)"

# PowerShell
ssher completion powershell | Out-String | Invoke-Expression
```

## Quick tour

```bash
ssher                          # interactive menu
ssher add                      # add a server with a form
ssher list                     # see what you've got
ssher prod                     # connect by name / alias / fuzzy / index
ssher prod --reconnect         # auto-reconnect on disconnect
ssher prod --record            # capture asciicast v2 recording
```

### Run things across servers

```bash
ssher exec "uptime" --all --timeout 2m
ssher exec "df -h" --group production
ssher exec "whoami" --servers web1,web2,db1
```

### Transfer files

```bash
ssher upload  ./build.tar.gz /var/tmp/  -s prod
ssher download /var/log/syslog ./syslog -s prod
ssher sync   ./public/ /srv/www/ -s prod --delete   # rsync, key auth only
```

### Drop-in `sshpass`

```bash
ssher wrap -e ssh user@host                       # password from $SSHPASS
ssher wrap -f /etc/secret/pw scp file user@host:  # password from a file
ssher wrap -P "Enter passphrase:" git push        # custom prompt
```

### Vault management

```bash
ssher vault status
ssher vault lock | unlock | change-password
ssher backup                                # snapshot ~/.ssher/vault.bin
ssher import-legacy                         # one-shot Fernet -> new format
```

### End-to-end encrypted cloud sync

```bash
ssher cloud login                           # approve this device in the browser
ssher cloud link --organization my-team    # choose a workspace
ssher cloud pull                            # decrypt cloud servers into this vault
ssher cloud push                            # encrypt local changes and upload
ssher cloud sync                            # pull, then push non-conflicting changes
```

Private-key file contents are not uploaded by default. Use `ssher cloud push
--include-keys` only when the key itself should be shared; it is encrypted
inside the server payload before upload. Passwords and all server fields are
always protected by the workspace key.

Run `ssher --help` to see every subcommand.

## How it works

`ssher` talks SSH directly via [`golang.org/x/crypto/ssh`](https://pkg.go.dev/golang.org/x/crypto/ssh) — there is no embedded or spawned `ssh` process for the connect, exec, and SFTP paths. That choice gives you native auth (password, encrypted or unencrypted key file, `ssh-agent`), real goroutine concurrency, one saved jump host, keepalives, and native local/remote port forwarding.

The one exception is `ssher wrap`, which by definition wraps an arbitrary user-supplied command and must use a PTY + prompt-scanner. That logic is fully isolated to `internal/pty/` and used nowhere else in the codebase.

### Trade-offs you should know

| Limitation | Why | Workaround |
| --- | --- | --- |
| `~/.ssh/config` directives (`ProxyCommand`, `ControlMaster`, `Match`, etc.) aren't honoured | Native dial, no `ssh(1)` involvement | `ssher export-config` writes an `~/.ssh/config` snippet for tools that *do* read it |
| `rsync` requires SSH key auth | We shell out to `rsync(1)` and don't currently inject passwords into its child SSH | Use `ssher upload` / `ssher download` (SFTP) for password-auth servers |
| X11 and arbitrary OpenSSH options are export-only | Native connections do not implement OpenSSH's X11/config machinery | Run `ssher export-config` and connect with OpenSSH when those options are required |
| `rsync` is not included with Windows | ssher shells out to an existing `rsync` for directory synchronization | Use native `ssher upload` / `ssher download`, or install rsync through WSL/MSYS2 |

## Security model

```text
[ magic       "SSHV"            4 B ]
[ version     1                  1 B ]
[ kdf_id      argon2id (=1)      1 B ]
[ argon2_time 3                  4 B ]
[ argon2_mem  64 MiB (in KiB)    4 B ]
[ argon2_par  4                  1 B ]
[ salt        random            16 B ]
[ nonce       random            12 B ]
[ ciphertext  AES-256-GCM(gzip(json)) + 16 B tag ]
```

- The plaintext payload is gzipped JSON containing servers, profiles, aliases, and history together.
- Argon2id parameters are stored in the header. Weak parameters are ratcheted during a password unlock, when a matching replacement key can be derived safely.
- New master passwords must be at least 12 characters; existing vault passwords remain compatible.
- The master password is **never** written to disk. The session file caches an Argon2id-derived key, wrapped with a host-bound HKDF key. This machine binding is defense in depth; protect access to the local user account.
- Plaintext exports (`export-csv`, default `export-json`) **omit passwords**. `--include-passwords` is the explicit opt-in.
- SSH host keys use trust on first use through `~/.ssh/known_hosts`; changed keys are rejected.
- SSHer Cloud uses a P-256 device key to unwrap an organization workspace key.
  Server records are encrypted locally with AES-256-GCM and revision-bound AAD;
  the Cloud API stores ciphertext and team access metadata only.

| Path | Mode | Contents |
| --- | --- | --- |
| `~/.ssher/vault.bin` | `0600` | the encrypted vault |
| `~/.ssher/.session` | `0600` | cached vault key, host-bound, 30-minute expiry |
| `~/.ssher/recordings/*.cast` | `0600` | asciicast v2 session recordings |
| `~/.ssher/backups/vault-*.bin` | `0600` | `ssher backup` snapshots |
| `~/.ssher/cloud.json` | `0600` | revocable Cloud session and linked workspace |
| `~/.ssher/cloud-device-key.pem` | `0600` | P-256 device private key |
| `~/.ssher/cloud-keys/*` | `0600` | private keys decrypted from managed Cloud records |

On Windows, the same files live under `%USERPROFILE%\.ssher`. ACLs on the
user profile provide the platform equivalent of Unix `0600` permissions, and
vault/session updates use atomic replace semantics.

## Roadmap

- **Workspace recovery and key rotation** for end-to-end encrypted team sync.
- **Password injection for `rsync`** via the same PTY layer that powers `wrap`.
- **Native X11 forwarding** with local display authentication and channel proxying.

## Contributing and security

Contributions are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md) and the [architecture guide](docs/ARCHITECTURE.md). Please follow the [Code of Conduct](CODE_OF_CONDUCT.md).

Do not report vulnerabilities or submit real credentials through public issues. Follow the private process in [SECURITY.md](SECURITY.md). Release history is maintained in [CHANGELOG.md](CHANGELOG.md).

The source and deployment notes for [getssher.com](https://getssher.com) are in [`website/`](website/) and [the website guide](docs/WEBSITE.md).

## License

[MIT](LICENSE).

## Author

Built by [@johnniewhite](https://github.com/johnniewhite).
