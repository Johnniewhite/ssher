"""
Base UI: terminal width, header, shared state.
Developed by Inioluwa Adeyinka
"""

import os

from ssher import APP_NAME, DEVELOPER
from ssher.config import VERSION
from ssher.formatting import colored, Colors


class BaseUI:
    """Base class with shared UI state and utilities."""

    def __init__(self, server_manager, ssh, transfer):
        self.manager = server_manager
        self.ssh = ssh
        self.transfer = transfer

    def get_terminal_width(self) -> int:
        """Get terminal width."""
        try:
            return os.get_terminal_size().columns
        except OSError:
            return 80

    def print_header(self):
        """Print application header."""
        width = min(self.get_terminal_width(), 70)

        print()
        print(colored("\u2550" * width, Colors.CYAN))
        print(colored(f"  {APP_NAME} v{VERSION}".center(width), Colors.CYAN, bold=True))
        print(colored("  Ultimate SSH Configuration Manager".center(width), Colors.CYAN))
        print(colored(f"  Developed by {DEVELOPER}".center(width), Colors.BRIGHT_BLACK))
        print(colored("\u2550" * width, Colors.CYAN))
