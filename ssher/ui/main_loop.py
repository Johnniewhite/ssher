"""
Main interactive loop.
Developed by Inioluwa Adeyinka
"""

import sys

from ssher.formatting import colored, Colors, error, info


class MainLoopMixin:
    """Mixin for the main interactive loop."""

    def print_menu(self):
        """Print the main menu."""
        print(f"\n{colored('Commands:', Colors.YELLOW, bold=True)}")
        print(f"  {colored('[1-N]', Colors.GREEN)}    Connect to server by number")
        print(f"  {colored('[name]', Colors.GREEN)}   Connect by server name (fuzzy search)")
        print(f"  {colored('a', Colors.CYAN)}        Add new server")
        print(f"  {colored('e', Colors.CYAN)}        Edit server")
        print(f"  {colored('d', Colors.CYAN)}        Delete server")
        print(f"  {colored('g', Colors.CYAN)}        Manage groups")
        print(f"  {colored('f', Colors.CYAN)}        Toggle favorite")
        print(f"  {colored('s', Colors.CYAN)}        Search servers")
        print(f"  {colored('t', Colors.CYAN)}        File transfer (SCP/SFTP)")
        print(f"  {colored('x', Colors.CYAN)}        Execute command on multiple servers")
        print(f"  {colored('p', Colors.CYAN)}        Ping/check all servers")
        print(f"  {colored('c', Colors.CYAN)}        Copy server details to clipboard")
        print(f"  {colored('v', Colors.CYAN)}        Vault management")
        print(f"  {colored('h', Colors.CYAN)}        View connection history")
        print(f"  {colored('i', Colors.CYAN)}        Import from ~/.ssh/config")
        print(f"  {colored('b', Colors.CYAN)}        Backup/Export")
        print(f"  {colored('?', Colors.CYAN)}        Show all commands")
        print(f"  {colored('q', Colors.CYAN)}        Quit")
        print()

    def run(self):
        """Run the interactive UI."""
        self.print_header()

        while True:
            self.display_servers()
            self.print_menu()

            choice = input(colored("\u276f ", Colors.GREEN, bold=True)).strip()

            if not choice:
                continue

            if choice.lower() == 'q':
                print(f"\n{colored('Goodbye!', Colors.CYAN)}\n")
                sys.exit(0)

            elif choice.lower() == 'a':
                from ssher.formatting import success as fmt_success
                server = self.prompt_add_server()
                if server:
                    self.manager.add(server)
                    print(fmt_success(f"Server '{server.name}' added!"))

            elif choice.lower() == 'e':
                self.prompt_edit_server()

            elif choice.lower() == 'd':
                self.prompt_delete_server()

            elif choice.lower() == 's':
                self.prompt_search()

            elif choice.lower() == 'f':
                self.prompt_toggle_favorite()

            elif choice.lower() == 'g':
                self.prompt_manage_groups()

            elif choice.lower() == 't':
                self.prompt_file_transfer()

            elif choice.lower() == 'x':
                self.prompt_execute_multiple()

            elif choice.lower() == 'p':
                self.prompt_ping_all()

            elif choice.lower() == 'h':
                self.prompt_view_history()

            elif choice.lower() == 'i':
                self.prompt_import()

            elif choice.lower() == 'b':
                self.prompt_backup()

            elif choice.lower() == 'c':
                self.prompt_clipboard()

            elif choice.lower() == 'v':
                self.prompt_vault()

            elif choice == '?':
                self.show_full_help()

            elif choice.isdigit():
                idx = int(choice)
                server = self.manager.get_by_index(idx)
                if server:
                    self.ssh.connect(server)
                else:
                    print(error(f"Invalid server number: {idx}"))

            else:
                # Fuzzy search and connect
                results = self.manager.search(choice)
                if results:
                    if len(results) == 1 or results[0][2] >= 80:
                        _, server, _ = results[0]
                        self.ssh.connect(server)
                    else:
                        print(f"\n{colored('Multiple matches found:', Colors.YELLOW)}")
                        for i, (_, server, score) in enumerate(results[:5], 1):
                            print(f"  [{i}] {server.name} ({server.host})")
                        sub_choice = input(f"\n{colored('Select:', Colors.CYAN)} ").strip()
                        if sub_choice.isdigit() and 1 <= int(sub_choice) <= len(results):
                            _, server, _ = results[int(sub_choice) - 1]
                            self.ssh.connect(server)
                else:
                    print(error(f"Unknown command or server: {choice}"))

    def prompt_clipboard(self):
        """Interactive clipboard copy prompt."""
        if not self.manager.servers:
            from ssher.formatting import warning as fmt_warning
            print(fmt_warning("No servers available."))
            return

        self.display_servers(show_groups=False)
        choice = input(f"\n{colored('Server # to copy:', Colors.CYAN)} ").strip()

        if not choice.isdigit() or int(choice) < 1 or int(choice) > len(self.manager.servers):
            print(error("Invalid selection."))
            return

        server = self.manager.servers[int(choice) - 1]

        print(f"\n{colored('Copy what?', Colors.YELLOW)}")
        print(f"  [1] Host ({server.host})")
        print(f"  [2] Username ({server.user})")
        print(f"  [3] SSH command (ssh {server.user}@{server.host})")
        print(f"  [4] Password")

        field_choice = input(f"{colored('Select:', Colors.CYAN)} ").strip()

        from ssher.clipboard import copy_to_clipboard
        from ssher.formatting import success as fmt_success

        if field_choice == '1':
            copy_to_clipboard(server.host)
            print(fmt_success("Host copied to clipboard."))
        elif field_choice == '2':
            copy_to_clipboard(server.user)
            print(fmt_success("Username copied to clipboard."))
        elif field_choice == '3':
            ssh_cmd = f"ssh {server.user}@{server.host}"
            if server.port != 22:
                ssh_cmd += f" -p {server.port}"
            copy_to_clipboard(ssh_cmd)
            print(fmt_success("SSH command copied to clipboard."))
        elif field_choice == '4':
            if server.password:
                copy_to_clipboard(server.password)
                print(fmt_success("Password copied to clipboard."))
            else:
                print(error("No password stored for this server."))

    def prompt_vault(self):
        """Interactive vault management prompt."""
        from ssher.vault import VaultManager

        vault = VaultManager(self.manager.encryption)

        print(f"\n{colored('[Vault Management]', Colors.MAGENTA, bold=True)}")
        print(f"  [1] Lock vault (clear session)")
        print(f"  [2] Vault status")
        print(f"  [3] Change master password")

        choice = input(f"\n{colored('Select:', Colors.CYAN)} ").strip()

        if choice == '1':
            vault.lock()
        elif choice == '2':
            vault.status()
        elif choice == '3':
            vault.interactive_change_password()
