"""
Full help/command reference.
Developed by Inioluwa Adeyinka
"""

from ssher.formatting import colored, Colors


class HelpUIMixin:
    """Mixin for help display."""

    def show_full_help(self):
        """Show comprehensive help."""
        divider = colored("\u2550" * 60, Colors.CYAN)
        print(f"\n{divider}")
        print(colored("  SSHer - Complete Command Reference", Colors.CYAN, bold=True))
        print(divider)

        print(f"\n{colored('Connection:', Colors.YELLOW, bold=True)}")
        print("  [number]     Connect to server by number (e.g., 1, 2, 3)")
        print("  [name]       Connect by fuzzy search (e.g., 'prod', 'web')")

        print(f"\n{colored('Server Management:', Colors.YELLOW, bold=True)}")
        print("  a            Add new server")
        print("  e            Edit existing server")
        print("  d            Delete server")
        print("  f            Toggle favorite status")
        print("  g            Manage groups (move servers between groups)")

        print(f"\n{colored('Search & Discovery:', Colors.YELLOW, bold=True)}")
        print("  s            Search servers by name, host, tag, or group")
        print("  p            Ping all servers (check connectivity)")

        print(f"\n{colored('File Transfer:', Colors.YELLOW, bold=True)}")
        print("  t            Open file transfer menu")
        print("               - SCP upload/download")
        print("               - Rsync sync")
        print("               - Interactive SFTP")

        print(f"\n{colored('Multi-Server:', Colors.YELLOW, bold=True)}")
        print("  x            Execute command on multiple servers")
        print("               - Select by numbers, ranges, or groups")
        print("               - Parallel or sequential execution")

        print(f"\n{colored('New Features:', Colors.YELLOW, bold=True)}")
        print("  c            Copy server details to clipboard")
        print("  v            Vault management (lock/unlock)")

        print(f"\n{colored('Data Management:', Colors.YELLOW, bold=True)}")
        print("  h            View connection history")
        print("  i            Import from ~/.ssh/config")
        print("  b            Backup/restore configurations")

        print(f"\n{colored('CLI Usage:', Colors.YELLOW, bold=True)}")
        print("  ssher                          Interactive mode")
        print("  ssher list                     List all servers")
        print("  ssher <n>                      Connect to server #n")
        print("  ssher <name>                   Connect by name (fuzzy)")
        print("  ssher add                      Add new server")
        print("  ssher ping                     Check all servers")
        print("  ssher exec <cmd>               Run command on servers")
        print("  ssher upload <l> <r>           Upload file")
        print("  ssher download <r> <l>         Download file")
        print("  ssher wrap -e ssh user@host    sshpass-compatible mode")
        print("  ssher vault status             Show vault state")
        print("  ssher completion bash           Output bash completions")
        print("  ssher generate-password        Generate secure password")
        print("  ssher export-config            Export to ~/.ssh/config")
        print("  ssher copy <server> --field x  Copy server detail")
        print("  ssher profile list             Manage connection profiles")
        print("  ssher alias add <alias> <srv>  Create server alias")
        print("  ssher record list              Session recordings")
        print("  ssher import-csv <file>        Import from CSV")
        print("  ssher export-json              Export to JSON")

        print(f"\n{divider}")
