"""
Reusable connection profiles.
Developed by Inioluwa Adeyinka
"""

import json
from typing import List, Optional

from ssher.config import PROFILES_FILE
from ssher.models import ConnectionProfile
from ssher.formatting import error


class ProfileManager:
    """Manages reusable connection profiles."""

    def __init__(self):
        self.profiles: List[ConnectionProfile] = []
        self._load()

    def _load(self):
        """Load profiles from file."""
        if not PROFILES_FILE.exists():
            return
        try:
            data = json.loads(PROFILES_FILE.read_text())
            self.profiles = [ConnectionProfile.from_dict(p) for p in data]
        except Exception:
            self.profiles = []

    def _save(self):
        """Save profiles to file."""
        data = [p.to_dict() for p in self.profiles]
        PROFILES_FILE.write_text(json.dumps(data, indent=2))
        PROFILES_FILE.chmod(0o600)

    def list_profiles(self) -> List[ConnectionProfile]:
        """List all profiles."""
        return self.profiles

    def get_profile(self, name: str) -> Optional[ConnectionProfile]:
        """Get a profile by name."""
        for p in self.profiles:
            if p.name.lower() == name.lower():
                return p
        return None

    def add_profile(self, profile: ConnectionProfile):
        """Add a new profile."""
        # Remove existing with same name
        self.profiles = [p for p in self.profiles if p.name.lower() != profile.name.lower()]
        self.profiles.append(profile)
        self._save()

    def remove_profile(self, name: str) -> bool:
        """Remove a profile by name."""
        initial = len(self.profiles)
        self.profiles = [p for p in self.profiles if p.name.lower() != name.lower()]
        if len(self.profiles) < initial:
            self._save()
            return True
        return False

    def apply_to_server(self, profile_name: str, server_name: str, manager) -> bool:
        """Apply a profile to a server."""
        profile = self.get_profile(profile_name)
        if not profile:
            print(error(f"Profile '{profile_name}' not found."))
            return False

        server = manager.get_by_name(server_name)
        if not server:
            server = manager.resolve_server(server_name)
        if not server:
            print(error(f"Server '{server_name}' not found."))
            return False

        # Apply profile settings
        server.connection_timeout = profile.connection_timeout
        server.keep_alive = profile.keep_alive
        server.x11_forward = profile.x11_forward
        server.auto_reconnect = profile.auto_reconnect
        server.max_reconnect_retries = profile.max_reconnect_retries
        server.profile = profile.name

        for key, value in profile.custom_options.items():
            server.custom_options[key] = value

        # Find and update in manager
        for idx, s in enumerate(manager.servers):
            if s.name == server.name:
                manager.update(idx, server)
                return True

        return False
