"""
Add/edit/delete/search/favorite/group prompts.
Developed by Inioluwa Adeyinka
"""

import getpass
from datetime import datetime, timedelta
from pathlib import Path
from typing import Optional

from ssher.formatting import colored, Colors, success, error, warning, info
from ssher.models import Server


class PromptsMixin:
    """Mixin for server management prompts."""

    def prompt_add_server(self) -> Optional[Server]:
        """Interactive prompt to add a server."""
        print(f"\n{colored('[Add New Server]', Colors.GREEN, bold=True)}\n")

        name = input(f"{colored('Server name', Colors.CYAN)} (custom label): ").strip()
        if not name:
            print(error("Server name is required."))
            return None

        if self.manager.get_by_name(name):
            print(error(f"Server '{name}' already exists."))
            return None

        host = input(f"{colored('Hostname/IP', Colors.CYAN)}: ").strip()
        if not host:
            print(error("Hostname is required."))
            return None

        import os
        default_user = os.getenv('USER', 'root')
        user = input(f"{colored('Username', Colors.CYAN)} [{default_user}]: ").strip() or default_user

        port_str = input(f"{colored('Port', Colors.CYAN)} [22]: ").strip()
        port = int(port_str) if port_str.isdigit() else 22

        existing_groups = list(self.manager.get_groups().keys())
        if existing_groups:
            print(f"\n{colored('Existing groups:', Colors.BRIGHT_BLACK)} {', '.join(existing_groups)}")
        group = input(f"{colored('Group', Colors.CYAN)} [default]: ").strip() or "default"

        tags_str = input(f"{colored('Tags', Colors.CYAN)} (comma-separated): ").strip()
        tags = [t.strip() for t in tags_str.split(",") if t.strip()]

        print(f"\n{colored('Authentication method:', Colors.YELLOW)}")
        print(f"  [1] SSH Key {colored('(recommended)', Colors.GREEN)}")
        print(f"  [2] Password")
        auth_choice = input("Choose [1/2]: ").strip()

        auth_type = 'password' if auth_choice == '2' else 'key'
        password = ""
        key_path = str(Path.home() / ".ssh" / "id_rsa")

        if auth_type == 'password':
            # Offer password generation
            print(f"\n{colored('Password options:', Colors.YELLOW)}")
            print(f"  [1] Enter password manually")
            print(f"  [2] Generate a secure password")
            pw_choice = input("Choose [1/2]: ").strip()

            if pw_choice == '2':
                from ssher.password_gen import generate_password
                password = generate_password()
                print(f"{success('Generated password:')} {colored(password, Colors.BRIGHT_WHITE, bold=True)}")
                print(colored("  Make sure to save this password!", Colors.YELLOW))
            else:
                password = getpass.getpass(f"{colored('Password', Colors.CYAN)}: ")

            print(f"\n{colored('Password expiry (optional):', Colors.YELLOW)}")
            print("  Enter number of days until password expires, or leave blank for none.")
            expiry_days = input("Days until expiry: ").strip()
            password_expires = ""
            if expiry_days.isdigit():
                password_expires = (datetime.now() + timedelta(days=int(expiry_days))).isoformat()
        else:
            default_key = str(Path.home() / ".ssh" / "id_rsa")
            key_path = input(f"{colored('SSH key path', Colors.CYAN)} [{default_key}]: ").strip() or default_key
            password_expires = ""

        notes = input(f"{colored('Notes', Colors.CYAN)} (optional): ").strip()

        print(f"\n{colored('Configure advanced options?', Colors.YELLOW)} [y/N]: ", end="")
        if input().strip().lower() == 'y':
            return self._prompt_advanced_options(Server(
                name=name, host=host, user=user, port=port,
                auth_type=auth_type, password=password, key_path=key_path,
                group=group, tags=tags, notes=notes, password_expires=password_expires
            ))

        return Server(
            name=name, host=host, user=user, port=port,
            auth_type=auth_type, password=password, key_path=key_path,
            group=group, tags=tags, notes=notes, password_expires=password_expires
        )

    def _prompt_advanced_options(self, server: Server) -> Server:
        """Prompt for advanced server options."""
        print(f"\n{colored('[Advanced Options]', Colors.MAGENTA, bold=True)}\n")

        if self.manager.servers:
            print(f"{colored('Available jump hosts:', Colors.BRIGHT_BLACK)}")
            for idx, s in enumerate(self.manager.servers, 1):
                print(f"  [{idx}] {s.name}")
        jump = input(f"{colored('Jump host', Colors.CYAN)} (server name or number, blank for none): ").strip()
        if jump:
            if jump.isdigit():
                jump_server = self.manager.get_by_index(int(jump))
                if jump_server:
                    server.jump_host = jump_server.name
            else:
                server.jump_host = jump

        print(f"\n{colored('Local port forwarding', Colors.CYAN)} (e.g., 8080:80 to forward local 8080 to remote 80)")
        local_fwd = input("Format local:remote (blank for none): ").strip()
        if local_fwd and ':' in local_fwd:
            parts = local_fwd.split(':')
            if len(parts) == 2 and parts[0].isdigit() and parts[1].isdigit():
                server.local_forwards.append({'local': int(parts[0]), 'remote': int(parts[1])})

        print(f"\n{colored('Remote port forwarding', Colors.CYAN)} (e.g., 9000:3000 to forward remote 9000 to local 3000)")
        remote_fwd = input("Format remote:local (blank for none): ").strip()
        if remote_fwd and ':' in remote_fwd:
            parts = remote_fwd.split(':')
            if len(parts) == 2 and parts[0].isdigit() and parts[1].isdigit():
                server.remote_forwards.append({'remote': int(parts[0]), 'local': int(parts[1])})

        x11 = input(f"\n{colored('Enable X11 forwarding?', Colors.CYAN)} [y/N]: ").strip().lower()
        server.x11_forward = x11 == 'y'

        keep_alive = input(f"{colored('Keep-alive interval', Colors.CYAN)} (seconds) [60]: ").strip()
        server.keep_alive = int(keep_alive) if keep_alive.isdigit() else 60

        timeout = input(f"{colored('Connection timeout', Colors.CYAN)} (seconds) [30]: ").strip()
        server.connection_timeout = int(timeout) if timeout.isdigit() else 30

        reconnect = input(f"{colored('Enable auto-reconnect?', Colors.CYAN)} [y/N]: ").strip().lower()
        server.auto_reconnect = reconnect == 'y'

        return server

    def prompt_edit_server(self) -> bool:
        """Prompt to edit a server."""
        if not self.manager.servers:
            print(warning("No servers to edit."))
            return False

        self.display_servers(show_groups=False)
        choice = input(f"\n{colored('Enter server number to edit:', Colors.CYAN)} ").strip()

        if not choice.isdigit() or int(choice) < 1 or int(choice) > len(self.manager.servers):
            print(error("Invalid selection."))
            return False

        idx = int(choice) - 1
        server = self.manager.servers[idx]

        print(f"\n{colored(f'[Editing: {server.name}]', Colors.YELLOW, bold=True)}")
        print(colored("Press Enter to keep current value.\n", Colors.BRIGHT_BLACK))

        name = input(f"Server name [{server.name}]: ").strip()
        if name and name != server.name:
            if self.manager.get_by_name(name):
                print(error(f"Server '{name}' already exists."))
            else:
                server.name = name

        host = input(f"Hostname [{server.host}]: ").strip()
        if host:
            server.host = host

        user = input(f"Username [{server.user}]: ").strip()
        if user:
            server.user = user

        port = input(f"Port [{server.port}]: ").strip()
        if port.isdigit():
            server.port = int(port)

        group = input(f"Group [{server.group}]: ").strip()
        if group:
            server.group = group

        tags = input(f"Tags [{', '.join(server.tags)}]: ").strip()
        if tags:
            server.tags = [t.strip() for t in tags.split(",") if t.strip()]

        notes = input(f"Notes [{server.notes[:30]}...]: ").strip()
        if notes:
            server.notes = notes

        if input(f"\n{colored('Update authentication?', Colors.YELLOW)} [y/N]: ").strip().lower() == 'y':
            print(f"\n  [1] SSH Key")
            print(f"  [2] Password")
            auth_choice = input("Choose [1/2]: ").strip()

            if auth_choice == '2':
                server.auth_type = 'password'
                server.password = getpass.getpass("New password: ")
                server.key_path = ""
            else:
                server.auth_type = 'key'
                server.key_path = input(f"SSH key path [{server.key_path}]: ").strip() or server.key_path
                server.password = ""

        if input(f"\n{colored('Edit advanced options?', Colors.YELLOW)} [y/N]: ").strip().lower() == 'y':
            server = self._prompt_advanced_options(server)

        self.manager.update(idx, server)
        print(success(f"Server '{server.name}' updated!"))
        return True

    def prompt_delete_server(self) -> bool:
        """Prompt to delete a server."""
        if not self.manager.servers:
            print(warning("No servers to delete."))
            return False

        self.display_servers(show_groups=False)
        choice = input(f"\n{colored('Enter server number to delete:', Colors.CYAN)} ").strip()

        if not choice.isdigit() or int(choice) < 1 or int(choice) > len(self.manager.servers):
            print(error("Invalid selection."))
            return False

        idx = int(choice) - 1
        server = self.manager.servers[idx]

        confirm = input(f"{colored('Delete', Colors.RED)} '{server.name}'? [y/N]: ").strip().lower()
        if confirm == 'y':
            self.manager.delete(idx)
            print(success(f"Server '{server.name}' deleted."))
            return True

        print(info("Deletion cancelled."))
        return False

    def prompt_search(self):
        """Search and connect to a server."""
        query = input(f"\n{colored('Search:', Colors.CYAN)} ").strip()
        if not query:
            return

        results = self.manager.search(query)
        if not results:
            print(warning("No servers found."))
            return

        print(f"\n{colored('Search Results:', Colors.GREEN, bold=True)}")
        for i, (idx, server, score) in enumerate(results[:10], 1):
            fav = colored("\u2605", Colors.YELLOW) if server.is_favorite else " "
            print(f"  [{i}]{fav} {server.name} ({server.user}@{server.host}) - {server.group}")

        choice = input(f"\n{colored('Connect to:', Colors.CYAN)} ").strip()
        if choice.isdigit() and 1 <= int(choice) <= len(results):
            _, server, _ = results[int(choice) - 1]
            self.ssh.connect(server)

    def prompt_toggle_favorite(self):
        """Toggle server favorite status."""
        if not self.manager.servers:
            print(warning("No servers available."))
            return

        self.display_servers(show_groups=False)
        choice = input(f"\n{colored('Toggle favorite for server #:', Colors.CYAN)} ").strip()

        if choice.isdigit() and 1 <= int(choice) <= len(self.manager.servers):
            idx = int(choice) - 1
            server = self.manager.servers[idx]
            server.is_favorite = not server.is_favorite
            self.manager.update(idx, server)
            status = "added to" if server.is_favorite else "removed from"
            print(success(f"'{server.name}' {status} favorites."))

    def prompt_manage_groups(self):
        """Manage server groups."""
        if not self.manager.servers:
            print(warning("No servers available."))
            return

        print(f"\n{colored('[Manage Groups]', Colors.MAGENTA, bold=True)}")

        groups = self.manager.get_groups()
        print(f"\n{colored('Current groups:', Colors.CYAN)}")
        for group, servers in sorted(groups.items()):
            print(f"  \u2022 {group}: {len(servers)} servers")

        print(f"\n  [1] Move server to group")
        print(f"  [2] Rename group")

        choice = input(f"\n{colored('Select:', Colors.CYAN)} ").strip()

        if choice == '1':
            self.display_servers(show_groups=False)
            server_num = input(f"\n{colored('Server #:', Colors.CYAN)} ").strip()
            if server_num.isdigit() and 1 <= int(server_num) <= len(self.manager.servers):
                idx = int(server_num) - 1
                server = self.manager.servers[idx]
                new_group = input(f"{colored('New group name:', Colors.CYAN)} ").strip()
                if new_group:
                    server.group = new_group
                    self.manager.update(idx, server)
                    print(success(f"Moved '{server.name}' to group '{new_group}'."))

        elif choice == '2':
            old_name = input(f"{colored('Current group name:', Colors.CYAN)} ").strip()
            new_name = input(f"{colored('New group name:', Colors.CYAN)} ").strip()
            if old_name and new_name:
                count = 0
                for idx, server in enumerate(self.manager.servers):
                    if server.group == old_name:
                        server.group = new_name
                        self.manager.servers[idx] = server
                        count += 1
                if count > 0:
                    self.manager.save()
                    print(success(f"Renamed group '{old_name}' to '{new_name}' ({count} servers)."))
                else:
                    print(warning(f"No servers in group '{old_name}'."))
