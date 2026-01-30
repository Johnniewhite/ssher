"""
Server table display, two-column layout.
Developed by Inioluwa Adeyinka
"""

from datetime import datetime, timedelta
from typing import List

from ssher.formatting import colored, Colors, warning
from ssher.models import Server


def server_display_name(server: Server) -> str:
    """Get display name with favorite indicator."""
    star = colored("\u2605", Colors.YELLOW) if server.is_favorite else " "
    return f"{star} {server.name}"


def server_status_color(server: Server) -> str:
    """Get color based on password expiry status."""
    if server.password_expires:
        try:
            expires = datetime.fromisoformat(server.password_expires)
            if datetime.now() > expires:
                return Colors.RED
            elif datetime.now() + timedelta(days=7) > expires:
                return Colors.YELLOW
        except ValueError:
            pass
    return Colors.GREEN


class DisplayMixin:
    """Mixin for server display functionality."""

    def display_servers(self, servers: List[Server] = None, show_groups: bool = True):
        """Display server list."""
        if servers is None:
            servers = self.manager.servers

        if not servers:
            print(f"\n{warning('No servers configured yet.')} Press 'a' to add one.")
            return

        width = min(self.get_terminal_width(), 100)
        use_two_columns = self.get_terminal_width() >= 140

        # Show favorites first
        favorites = [s for s in servers if s.is_favorite]
        if favorites:
            fav_header = colored("\u2605 Favorites", Colors.YELLOW, bold=True)
            print(f"\n{fav_header}")
            if use_two_columns:
                self._display_server_table_two_col(favorites, width)
            else:
                self._display_server_table(favorites, width)

        # Show recent
        recent = self.manager.get_recent(3)
        recent_names = {s.name for s in recent}
        recent_non_fav = [s for s in recent if not s.is_favorite]
        if recent_non_fav:
            recent_header = colored("\u23f1 Recent", Colors.CYAN, bold=True)
            print(f"\n{recent_header}")
            if use_two_columns:
                self._display_server_table_two_col(recent_non_fav, width)
            else:
                self._display_server_table(recent_non_fav, width)

        # Show by groups
        if show_groups:
            groups = self.manager.get_groups()
            for group_name, group_servers in sorted(groups.items()):
                remaining = [s for s in group_servers
                             if not s.is_favorite and s.name not in recent_names]
                if remaining:
                    group_label = "\U0001f4c1 " + group_name.title()
                    group_header = colored(group_label, Colors.MAGENTA, bold=True)
                    print(f"\n{group_header}")
                    if use_two_columns:
                        self._display_server_table_two_col(remaining, width)
                    else:
                        self._display_server_table(remaining, width)
        else:
            print(f"\n{colored('All Servers', Colors.BLUE, bold=True)}")
            if use_two_columns:
                self._display_server_table_two_col(servers, width)
            else:
                self._display_server_table(servers, width)

    def _display_server_table(self, servers: List[Server], width: int):
        """Display servers in table format."""
        print(colored("-" * width, Colors.BRIGHT_BLACK))
        header = f"{'#':<4} {'Name':<22} {'Host':<28} {'User':<12} {'Group':<12}"
        print(colored(header, Colors.BRIGHT_WHITE, bold=True))
        print(colored("-" * width, Colors.BRIGHT_BLACK))

        for server in servers:
            try:
                idx = self.manager.servers.index(server) + 1
            except ValueError:
                idx = 0

            fav = colored("\u2605", Colors.YELLOW) if server.is_favorite else " "
            name = f"{fav}{server.name}"[:20]
            host = server.host[:26]
            if server.port != 22:
                host = f"{host}:{server.port}"[:26]
            user = server.user[:10]
            group = server.group[:10]

            status_color = server_status_color(server)

            line = f"{idx:<4} {name:<22} {host:<28} {user:<12} {group:<12}"
            print(colored(line, status_color))

        print(colored("-" * width, Colors.BRIGHT_BLACK))

    def _display_server_table_two_col(self, servers: List[Server], width: int):
        """Display servers in two-column layout for wide terminals."""
        col_width = 65
        print(colored("-" * (col_width * 2 + 3), Colors.BRIGHT_BLACK))
        header = f"{'#':<4} {'Name':<20} {'Host':<22} {'User':<10} {'Grp':<8}"
        print(colored(f"{header} | {header}", Colors.BRIGHT_WHITE, bold=True))
        print(colored("-" * (col_width * 2 + 3), Colors.BRIGHT_BLACK))

        half = (len(servers) + 1) // 2
        left = servers[:half]
        right = servers[half:]

        for i in range(half):
            left_line = self._format_server_cell(left[i], col_width)
            if i < len(right):
                right_line = self._format_server_cell(right[i], col_width)
            else:
                right_line = " " * col_width
            print(f"{left_line} | {right_line}")

        print(colored("-" * (col_width * 2 + 3), Colors.BRIGHT_BLACK))

    def _format_server_cell(self, server: Server, col_width: int) -> str:
        """Format a single server cell for two-column display."""
        try:
            idx = self.manager.servers.index(server) + 1
        except ValueError:
            idx = 0

        fav = colored("\u2605", Colors.YELLOW) if server.is_favorite else " "
        name = f"{fav}{server.name}"[:18]
        host = server.host[:20]
        if server.port != 22:
            host = f"{host}:{server.port}"[:20]
        user = server.user[:8]
        group = server.group[:6]

        status_color = server_status_color(server)
        return colored(f"{idx:<4} {name:<20} {host:<22} {user:<10} {group:<8}", status_color)
