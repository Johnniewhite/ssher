# SSHer v3.0.0

**Ultimate SSH Configuration Manager**

A powerful CLI tool for managing SSH server configurations. Save your server credentials once and connect with a simple selection. Drop-in sshpass replacement with encrypted credential storage.

**Developed by Inioluwa Adeyinka**

## Features

### Core
- Save unlimited SSH server configurations
- Custom names, groups, tags, and favorites for easy organization
- Encrypted storage with master password protection (AES-256 via Fernet)
- Support for SSH key and password authentication
- Fuzzy search for quick server selection
- Jump host / bastion support
- Port forwarding (local, remote, X11)

### Connection Management
- Connect by number, name, or fuzzy search
- Auto-reconnect with exponential backoff
- Connection profiles (reusable SSH option sets)
- Server aliases (multiple names for one server)
- Connection history tracking

### File Transfer
- SCP upload/download (recursive supported)
- Rsync sync with dry-run support
- Interactive SFTP sessions

### Multi-Server Operations
- Execute commands on multiple servers in parallel
- Select by numbers, ranges, groups, or all servers

### sshpass Replacement
- `ssher wrap -e ssh user@host` — password from `$SSHPASS` env var
- `ssher wrap -f /path/to/file ssh user@host` — password from file
- `ssher wrap -d 3 ssh user@host` — password from file descriptor
- `ssher wrap -P "prompt:" ssh user@host` — custom prompt matching
- Works with any SSH-based command (ssh, scp, rsync, git)

### Vault Management
- `ssher vault lock` — clear session, require re-auth
- `ssher vault unlock` — authenticate and create session
- `ssher vault change-password` — re-encrypt with new master password
- `ssher vault status` — show lock state and session time remaining

### New in v3.0.0
- **Shell Tab Completion** — bash and zsh completion scripts
- **Clipboard Support** — copy host/user/password/SSH command
- **SSH Config Export** — write servers to `~/.ssh/config` format
- **Connection Profiles** — reusable SSH option presets
- **Server Aliases** — multiple names for one server
- **Session Recording** — capture and replay terminal sessions
- **Password Generator** — cryptographically secure password generation
- **Batch Import/Export** — CSV and JSON import/export
- **Two-Column Display** — wide terminal support (>=140 chars)
- **Modular Architecture** — proper Python package structure

## Installation

```bash
# Clone or download the repository
cd ssher

# Run the installation script
./install.sh
```

Or install manually:

```bash
pip3 install -r requirements.txt
pip3 install -e .
```

## Usage

### Interactive Mode

```bash
ssher
```

### First Run

On first run, you'll be prompted to create a master password that encrypts all your server configurations.

### CLI Commands

```bash
# Server management
ssher                          # Interactive mode
ssher list                     # List all servers
ssher add                      # Add a new server
ssher <number>                 # Connect to server by number
ssher <name>                   # Connect by name (fuzzy search)
ssher ping                     # Check connectivity for all servers
ssher groups                   # List server groups

# Connection options
ssher <name> --reconnect       # Auto-reconnect on disconnect
ssher <name> --record          # Record the session

# File transfer
ssher upload <local> <remote> -s <server>
ssher download <remote> <local> -s <server>

# Multi-server execution
ssher exec "uptime" --all
ssher exec "df -h" -g production
ssher exec "whoami" -s web1,web2

# sshpass replacement
ssher wrap -e ssh user@host
ssher wrap -f /path/to/password ssh user@host

# Vault
ssher vault status
ssher vault lock
ssher vault unlock
ssher vault change-password

# Profiles
ssher profile list
ssher profile add high-latency --timeout 60 --keepalive 120
ssher profile apply high-latency --server prod

# Aliases
ssher alias add production-web pw
ssher alias list
ssher alias remove pw

# Clipboard
ssher copy <server> --field host
ssher copy <server> --field password
ssher copy <server> --field command

# SSH Config Export
ssher export-config
ssher export-config --append
ssher export-config --output /path/to/config

# Session Recording
ssher record list
ssher record replay <file>

# Password Generator
ssher generate-password
ssher generate-password --length 32 --no-symbols

# Batch Import/Export
ssher import-csv servers.csv
ssher import-json servers.json
ssher export-csv
ssher export-json

# Shell Completion
ssher completion bash
ssher completion zsh

# Backup
ssher backup
ssher history
```

### Shell Completion

Enable tab completion by adding to your shell config:

```bash
# Bash (~/.bashrc)
eval "$(ssher completion bash)"

# Zsh (~/.zshrc)
eval "$(ssher completion zsh)"
```

### Interactive Menu

```
Commands:
  [1-N]    Connect to server by number
  [name]   Connect by server name (fuzzy search)
  a        Add new server
  e        Edit server
  d        Delete server
  g        Manage groups
  f        Toggle favorite
  s        Search servers
  t        File transfer (SCP/SFTP)
  x        Execute command on multiple servers
  p        Ping/check all servers
  c        Copy server details to clipboard
  v        Vault management
  h        View connection history
  i        Import from ~/.ssh/config
  b        Backup/Export
  ?        Show all commands
  q        Quit
```

## Security

- All configurations are encrypted using Fernet (AES-128-CBC with HMAC)
- Master password is never stored; only a verification token
- Key derivation uses PBKDF2 with 480,000 iterations
- Configuration files are stored with restricted permissions (600)
- Passwords are encrypted at rest
- Plaintext exports (CSV/JSON) intentionally exclude passwords
- Session timeout after 30 minutes of inactivity

## Architecture

SSHer v3.0.0 uses a modular package architecture:

```
ssher/
├── __init__.py          # Version, app name, developer
├── __main__.py          # python -m ssher support
├── config.py            # Constants and paths
├── formatting.py        # Terminal colors and formatting
├── models.py            # Server, ConnectionHistory, ConnectionProfile
├── crypto.py            # Encryption (pure crypto logic)
├── vault.py             # Vault management (lock/unlock/change-password)
├── servers.py           # Server CRUD, search, import/export
├── password_auth.py     # Shared pexpect password injection
├── connection.py        # SSH connections, multi-exec, health check
├── transfer.py          # SCP, SFTP, rsync
├── sshpass_compat.py    # sshpass-compatible wrapper
├── ssh_config_export.py # Export to ~/.ssh/config
├── profiles.py          # Connection profiles
├── completion.py        # Shell completion generator
├── clipboard.py         # Clipboard support
├── recording.py         # Session recording/replay
├── password_gen.py      # Password generator
├── batch_import.py      # CSV/JSON import/export
├── ui/                  # Interactive UI (mixin-based)
│   ├── __init__.py      # InteractiveUI composition
│   ├── base.py          # Terminal width, header
│   ├── display.py       # Server table, two-column layout
│   ├── prompts.py       # Add/edit/delete/search/favorite/group
│   ├── transfer_ui.py   # File transfer menu
│   ├── multi_exec_ui.py # Multi-server execution
│   ├── status_ui.py     # Ping, history, backup
│   ├── help_ui.py       # Command reference
│   └── main_loop.py     # Main interactive loop
└── cli/                 # CLI layer
    ├── __init__.py      # main() entry point
    ├── parser.py        # Argparse with all subcommands
    ├── commands.py      # Command dispatch
    └── wrap.py          # sshpass wrap handler
```

## Requirements

- Python 3.7+
- cryptography library
- pexpect library

## File Locations

Configurations are stored in `~/.ssher/`:

- `servers.enc` — Encrypted server configurations
- `.key` — Verification token
- `.salt` — Encryption salt
- `history.json` — Connection history
- `.session` — Session state
- `profiles.json` — Connection profiles
- `aliases.json` — Server aliases
- `backups/` — Encrypted backups
- `recordings/` — Session recordings

## License

MIT License

## Author

Developed by **Inioluwa Adeyinka**
