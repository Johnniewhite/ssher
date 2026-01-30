"""
File transfer interactive menu.
Developed by Inioluwa Adeyinka
"""

from ssher.formatting import colored, Colors, error, warning, info


class TransferUIMixin:
    """Mixin for file transfer UI."""

    def prompt_file_transfer(self):
        """File transfer menu."""
        if not self.manager.servers:
            print(warning("No servers available."))
            return

        print(f"\n{colored('[File Transfer]', Colors.MAGENTA, bold=True)}")
        print(f"  [1] Upload (SCP)")
        print(f"  [2] Download (SCP)")
        print(f"  [3] Rsync to server")
        print(f"  [4] Rsync from server")
        print(f"  [5] Interactive SFTP")

        choice = input(f"\n{colored('Select:', Colors.CYAN)} ").strip()

        self.display_servers(show_groups=False)
        server_choice = input(f"\n{colored('Server #:', Colors.CYAN)} ").strip()

        if not server_choice.isdigit() or int(server_choice) < 1 or int(server_choice) > len(self.manager.servers):
            print(error("Invalid server."))
            return

        server = self.manager.servers[int(server_choice) - 1]

        if choice == '1':
            local = input(f"{colored('Local file/dir:', Colors.CYAN)} ").strip()
            remote = input(f"{colored('Remote path:', Colors.CYAN)} ").strip()
            recursive = input("Recursive? [y/N]: ").strip().lower() == 'y'
            self.transfer.upload(server, local, remote, recursive)

        elif choice == '2':
            remote = input(f"{colored('Remote file/dir:', Colors.CYAN)} ").strip()
            local = input(f"{colored('Local path:', Colors.CYAN)} ").strip()
            recursive = input("Recursive? [y/N]: ").strip().lower() == 'y'
            self.transfer.download(server, remote, local, recursive)

        elif choice == '3':
            source = input(f"{colored('Local source:', Colors.CYAN)} ").strip()
            dest = input(f"{colored('Remote destination:', Colors.CYAN)} ").strip()
            delete = input("Delete extraneous files? [y/N]: ").strip().lower() == 'y'
            dry_run = input("Dry run first? [Y/n]: ").strip().lower() != 'n'
            if dry_run:
                print(info("Performing dry run..."))
                self.transfer.rsync(server, source, dest, True, delete, True)
                if input("Proceed with actual sync? [y/N]: ").strip().lower() == 'y':
                    self.transfer.rsync(server, source, dest, True, delete, False)
            else:
                self.transfer.rsync(server, source, dest, True, delete, False)

        elif choice == '4':
            source = input(f"{colored('Remote source:', Colors.CYAN)} ").strip()
            dest = input(f"{colored('Local destination:', Colors.CYAN)} ").strip()
            delete = input("Delete extraneous files? [y/N]: ").strip().lower() == 'y'
            self.transfer.rsync(server, source, dest, False, delete, False)

        elif choice == '5':
            self.transfer.sftp_interactive(server)
