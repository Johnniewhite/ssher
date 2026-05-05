<p align="center">
  <img src="logo.png" alt="ssher — Ultimate SSH Configuration Manager" width="480">
</p>

<p align="center">
  <em>The SSH config manager that <b>actually</b> remembers your servers.</em>
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/johnniewhite/ssher"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/johnniewhite/ssher.svg"></a>
  <a href="https://goreportcard.com/report/github.com/johnniewhite/ssher"><img alt="Go Report" src="https://goreportcard.com/badge/github.com/johnniewhite/ssher"></a>
  <a href="LICENSE"><img alt="License: MIT" src="https://img.shields.io/badge/license-MIT-blue.svg"></a>
  <img alt="Platforms" src="https://img.shields.io/badge/platforms-macOS%20%7C%20Linux-lightgrey">
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

## Install

### Homebrew (macOS — recommended)

```bash
brew tap johnniewhite/ssher
brew install ssher
```

Upgrade with `brew update && brew upgrade ssher`. Also works on Linux via [Homebrew on Linux](https://docs.brew.sh/Homebrew-on-Linux).

### Pre-built binary

```bash
# macOS arm64 shown; swap the asset name for your OS/arch
# (Darwin_arm64, Darwin_amd64, Linux_amd64, Linux_arm64)
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
ssher exec "uptime" --all
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

Run `ssher --help` to see every subcommand.

## How it works

`ssher` talks SSH directly via [`golang.org/x/crypto/ssh`](https://pkg.go.dev/golang.org/x/crypto/ssh) — there is no embedded or spawned `ssh` process for the connect, exec, and SFTP paths. That choice gives you native auth (password, key file, `ssh-agent`), real goroutine concurrency, jump-host chains via `ssh.Dial` over an existing connection, and native local/remote port forwarding.

The one exception is `ssher wrap`, which by definition wraps an arbitrary user-supplied command and must use a PTY + prompt-scanner. That logic is fully isolated to `internal/pty/` and used nowhere else in the codebase.

### Trade-offs you should know

| Limitation | Why | Workaround |
| --- | --- | --- |
| `~/.ssh/config` directives (`ProxyCommand`, `ControlMaster`, `Match`, etc.) aren't honoured | Native dial, no `ssh(1)` involvement | `ssher export-config` writes an `~/.ssh/config` snippet for tools that *do* read it |
| `rsync` requires SSH key auth | We shell out to `rsync(1)` and don't currently inject passwords into its child SSH | Use `ssher upload` / `ssher download` (SFTP) for password-auth servers |
| Encrypted private keys aren't unlocked | No passphrase prompt yet | Use an unencrypted key, or load it into `ssh-agent` first |

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
- Argon2id parameters are stored in the header, so costs can ratchet upward without breaking existing vaults.
- The master password is **never** written to disk. The session file caches an Argon2id-derived key, wrapped with a host-bound HKDF key — exfiltrating `~/.ssher/.session` alone doesn't get you the vault.
- Plaintext exports (`export-csv`, default `export-json`) **omit passwords**. `--include-passwords` is the explicit opt-in.

| Path | Mode | Contents |
| --- | --- | --- |
| `~/.ssher/vault.bin` | `0600` | the encrypted vault |
| `~/.ssher/.session` | `0600` | cached vault key, host-bound, 30-minute expiry |
| `~/.ssher/recordings/*.cast` | `0600` | asciicast v2 session recordings |
| `~/.ssher/backups/vault-*.bin` | `0600` | `ssher backup` snapshots |

## Roadmap

- **Workspaces** — share an encrypted server list across a small team, in design now.
- **Encrypted-key passphrase prompts** for the connect path.
- **Password injection for `rsync`** via the same PTY layer that powers `wrap`.

## License

[MIT](LICENSE).

## Author

Built by [@johnniewhite](https://github.com/johnniewhite).
