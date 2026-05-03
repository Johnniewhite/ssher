# ssher

**Ultimate SSH Configuration Manager** — a single-binary CLI for managing SSH server configurations, with native SSH auth and a drop-in `sshpass` replacement.

Save your servers once, connect with `ssher prod` or `ssher 3`, and run commands across many servers in parallel. The vault is encrypted with AES-256-GCM and Argon2id.

## Install

### From source

```bash
go install github.com/johnniewhite/ssher@latest
```

### Pre-built binaries

Grab a release archive from the [GitHub releases page](https://github.com/johnniewhite/ssher/releases) for your OS/arch:

```bash
# macOS arm64 example
curl -LO https://github.com/johnniewhite/ssher/releases/latest/download/ssher_Darwin_arm64.tar.gz
tar -xzf ssher_Darwin_arm64.tar.gz
sudo mv ssher /usr/local/bin/
```

## Usage

```bash
ssher                          # interactive menu
ssher add                      # add a server (form prompts)
ssher list                     # list all servers
ssher 3                        # connect to server #3 in the list
ssher prod                     # connect by name / alias / fuzzy match
ssher prod --reconnect         # auto-reconnect on disconnect
ssher prod --record            # capture asciicast v2 recording

# Multi-server execution (parallel)
ssher exec "uptime" --all
ssher exec "df -h" -g production
ssher exec "whoami" -s web1,web2

# Transfers
ssher upload local.tar.gz /var/tmp/ -s prod
ssher download /var/log/syslog ./syslog -s prod
ssher sync ./build/ /srv/app/ -s prod --delete

# sshpass replacement
ssher wrap -e ssh user@host                  # password from $SSHPASS
ssher wrap -f /path/to/pw scp file user@host:
ssher wrap -P "Enter passphrase:" git push   # custom prompt

# Vault
ssher vault status
ssher vault unlock | lock | change-password

# Migrating from the Python version
ssher import-legacy                          # one-shot Fernet -> new format
```

## Why a Go rewrite?

The Python predecessor at `8527ca2` shipped a working tool but every SSH operation went through `pexpect` to inject passwords into a spawned `ssh` binary. The Go version uses `golang.org/x/crypto/ssh` and authenticates natively — no pexpect, no spawned `ssh`, no shell-string assembly. That gives us:

- A single static binary (no `pip install` dance).
- Native password and key auth, including via `ssh-agent`.
- Real concurrency for `exec` across many servers (goroutines, not Python threads).
- Proper jump-host chains via `ssh.Dial` over an existing connection.
- Native local/remote port forwarding.

The one place pexpect-equivalent logic survives is `ssher wrap`, which by definition wraps an arbitrary user-supplied SSH-shaped command and has no other choice. That code lives in `internal/pty/` and is the *only* PTY-injection code in the project.

### What you give up

`golang.org/x/crypto/ssh` does not honour `~/.ssh/config`. Host aliases, `ProxyJump`, `ControlMaster`, `Include`, `Match`, etc. defined there will be ignored when ssher dials directly. ssher implements what its data model needs (jump host as a saved server, port forwarding from per-server `local_forwards`/`remote_forwards`) and exposes the rest via `ssher export-config`, which writes an `~/.ssh/config` snippet you can paste in.

`rsync` shells out to the system `rsync`, which means **rsync currently requires SSH key auth**. Password-auth servers will see rsync prompt and fail. SCP/SFTP work fine for password-auth servers — the password is fed through the native SSH session, not a child process.

## Vault format (v1)

```
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

Plaintext is gzipped JSON containing servers, profiles, aliases, and history together. Argon2id parameters are stored in the header so we can ratchet costs over time without breaking existing vaults.

State on disk:

| Path | Mode | Contents |
| --- | --- | --- |
| `~/.ssher/vault.bin` | `0600` | the encrypted vault |
| `~/.ssher/.session` | `0600` | cached vault key, host-bound, 30-min expiry |
| `~/.ssher/recordings/*.cast` | `0600` | asciicast v2 session recordings |
| `~/.ssher/backups/vault-*.bin` | `0600` | snapshots from `ssher backup` |

## License

MIT — see `LICENSE`.

## Author

Originally written in Python by Inioluwa Adeyinka. Go rewrite in this tree.
