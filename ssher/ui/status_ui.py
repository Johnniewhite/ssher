"""
Ping, history, import, backup prompts.
Developed by Inioluwa Adeyinka
"""

from datetime import datetime

from ssher.config import BACKUP_DIR
from ssher.formatting import colored, Colors, success, error, warning, info


class StatusUIMixin:
    """Mixin for status/data management UI."""

    def prompt_ping_all(self):
        """Check connectivity to all servers."""
        if not self.manager.servers:
            print(warning("No servers available."))
            return

        print(f"\n{colored('[Checking Server Connectivity]', Colors.CYAN, bold=True)}")
        print(colored("Please wait...", Colors.BRIGHT_BLACK))

        results = self.ssh.check_all_connectivity()

        width = min(self.get_terminal_width(), 70)
        print("\n" + "-" * width)
        print(f"{'Server':<25} {'Status':<15} {'Latency':<15}")
        print("-" * width)

        online = 0
        for name, (reachable, latency) in sorted(results.items()):
            if reachable:
                status = colored("\u25cf Online", Colors.GREEN)
                latency_str = f"{latency}ms"
                online += 1
            else:
                status = colored("\u25cf Offline", Colors.RED)
                latency_str = "-"

            print(f"{name:<25} {status:<24} {latency_str:<15}")

        print("-" * width)
        print(f"\n{colored('Summary:', Colors.YELLOW)} {online}/{len(results)} servers online")

    def prompt_view_history(self):
        """View connection history."""
        self.manager.load_history()

        if not self.manager.history:
            print(warning("No connection history."))
            return

        print(f"\n{colored('[Connection History]', Colors.CYAN, bold=True)}")

        width = min(self.get_terminal_width(), 90)
        print("-" * width)
        print(f"{'Time':<20} {'Server':<20} {'User@Host':<25} {'Duration':<10} {'Status':<10}")
        print("-" * width)

        for entry in reversed(self.manager.history[-20:]):
            try:
                dt = datetime.fromisoformat(entry.timestamp)
                time_str = dt.strftime("%Y-%m-%d %H:%M")
            except ValueError:
                time_str = entry.timestamp[:16]

            duration = f"{entry.duration}s" if entry.duration else "-"
            status = colored("\u2713", Colors.GREEN) if entry.success else colored("\u2717", Colors.RED)

            print(f"{time_str:<20} {entry.server_name:<20} {entry.user}@{entry.host:<23} {duration:<10} {status:<10}")

        print("-" * width)

    def prompt_import(self):
        """Import from SSH config."""
        config_path = input(f"{colored('SSH config path', Colors.CYAN)} [~/.ssh/config]: ").strip()
        config_path = config_path or None

        count = self.manager.import_ssh_config(config_path)
        if count > 0:
            print(success(f"Imported {count} servers."))
        else:
            print(info("No new servers to import."))

    def prompt_backup(self):
        """Backup/export menu."""
        print(f"\n{colored('[Backup & Export]', Colors.MAGENTA, bold=True)}")
        print(f"  [1] Create backup")
        print(f"  [2] Restore from backup")
        print(f"  [3] List backups")

        choice = input(f"\n{colored('Select:', Colors.CYAN)} ").strip()

        if choice == '1':
            path = self.manager.export_backup()
            print(success(f"Backup created: {path}"))

        elif choice == '2':
            backups = sorted(BACKUP_DIR.glob("*.enc"), reverse=True)
            if not backups:
                print(warning("No backups found."))
                return

            print(f"\n{colored('Available backups:', Colors.CYAN)}")
            for i, backup in enumerate(backups[:10], 1):
                print(f"  [{i}] {backup.name}")

            choice = input(f"\n{colored('Select backup:', Colors.CYAN)} ").strip()
            if choice.isdigit() and 1 <= int(choice) <= len(backups):
                self.manager.import_backup(str(backups[int(choice) - 1]))

        elif choice == '3':
            backups = sorted(BACKUP_DIR.glob("*.enc"), reverse=True)
            if not backups:
                print(warning("No backups found."))
            else:
                print(f"\n{colored('Backups:', Colors.CYAN)}")
                for backup in backups:
                    size = backup.stat().st_size / 1024
                    print(f"  \u2022 {backup.name} ({size:.1f} KB)")
