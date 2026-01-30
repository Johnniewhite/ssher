"""
Multi-server execution UI.
Developed by Inioluwa Adeyinka
"""

from ssher.formatting import colored, Colors, error, warning


class MultiExecUIMixin:
    """Mixin for multi-server execution UI."""

    def prompt_execute_multiple(self):
        """Execute command on multiple servers."""
        if not self.manager.servers:
            print(warning("No servers available."))
            return

        print(f"\n{colored('[Execute on Multiple Servers]', Colors.MAGENTA, bold=True)}")
        print(f"\n{colored('Select servers:', Colors.CYAN)}")
        print(f"  - Enter numbers separated by commas (e.g., 1,3,5)")
        print(f"  - Use ranges (e.g., 1-5)")
        print(f"  - Use 'all' for all servers")
        print(f"  - Use '@group' for a group (e.g., @production)")

        self.display_servers(show_groups=False)

        selection = input(f"\n{colored('Servers:', Colors.CYAN)} ").strip().lower()

        selected = []
        if selection == 'all':
            selected = self.manager.servers[:]
        elif selection.startswith('@'):
            group = selection[1:]
            selected = [s for s in self.manager.servers if s.group.lower() == group]
        else:
            for part in selection.split(','):
                part = part.strip()
                if '-' in part:
                    try:
                        start, end = part.split('-')
                        for i in range(int(start), int(end) + 1):
                            if 1 <= i <= len(self.manager.servers):
                                selected.append(self.manager.servers[i - 1])
                    except ValueError:
                        pass
                elif part.isdigit():
                    i = int(part)
                    if 1 <= i <= len(self.manager.servers):
                        selected.append(self.manager.servers[i - 1])

        if not selected:
            print(error("No servers selected."))
            return

        print(f"\n{colored('Selected servers:', Colors.GREEN)} {', '.join(s.name for s in selected)}")
        command = input(f"{colored('Command to execute:', Colors.CYAN)} ").strip()

        if not command:
            print(error("No command specified."))
            return

        parallel = input("Execute in parallel? [Y/n]: ").strip().lower() != 'n'

        print(f"\n{colored('Executing...', Colors.YELLOW)}")
        results = self.ssh.execute_on_multiple(selected, command, parallel)

        print(f"\n{colored('Results:', Colors.GREEN, bold=True)}")
        for server_name, output in results.items():
            banner = colored("\u2550\u2550\u2550 " + server_name + " \u2550\u2550\u2550", Colors.CYAN, bold=True)
            print(f"\n{banner}")
            print(output if output.strip() else colored("(no output)", Colors.BRIGHT_BLACK))
