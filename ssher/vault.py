"""
Vault management: lock/unlock/change-password/status.
Developed by Inioluwa Adeyinka
"""

import getpass

from ssher.formatting import colored, Colors, success, error, warning, info
from ssher.crypto import EncryptionManager


class VaultManager:
    """Interactive vault management (wraps EncryptionManager with UI)."""

    def __init__(self, encryption: EncryptionManager):
        self.encryption = encryption

    def lock(self):
        """Lock the vault by clearing the session."""
        self.encryption.clear_session()
        print(success("Vault locked. You will need to re-authenticate."))

    def unlock(self):
        """Unlock the vault interactively."""
        password = getpass.getpass(colored("Enter master password: ", Colors.YELLOW))
        if self.encryption.initialize_with_password(password):
            print(success("Vault unlocked."))
            return True
        else:
            print(error("Incorrect master password!"))
            return False

    def status(self):
        """Show vault status."""
        session_info = self.encryption.get_session_info()

        print(f"\n{colored('[Vault Status]', Colors.CYAN, bold=True)}")

        if session_info:
            remaining = session_info['remaining_seconds']
            minutes = remaining // 60
            seconds = remaining % 60
            print(f"  Status:    {colored('Unlocked', Colors.GREEN, bold=True)}")
            print(f"  Session:   {minutes}m {seconds}s remaining")
            print(f"  Timeout:   {session_info['timeout']}s")
        else:
            print(f"  Status:    {colored('Locked', Colors.RED, bold=True)}")
            print(f"  Session:   No active session")

        print(f"  First-time: {'Yes' if self.encryption.is_first_time() else 'No'}")

    def interactive_change_password(self):
        """Interactively change the master password."""
        print(f"\n{colored('[Change Master Password]', Colors.YELLOW, bold=True)}\n")

        old_password = getpass.getpass(colored("Current master password: ", Colors.YELLOW))

        while True:
            new_password = getpass.getpass(colored("New master password: ", Colors.YELLOW))
            if len(new_password) < 4:
                print(error("Password must be at least 4 characters."))
                continue
            confirm = getpass.getpass(colored("Confirm new password: ", Colors.YELLOW))
            if new_password != confirm:
                print(error("Passwords don't match. Try again."))
                continue
            break

        if self.encryption.change_password(old_password, new_password):
            print(success("Master password changed successfully!"))
        else:
            print(error("Failed to change password. Check your current password."))
