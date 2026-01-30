"""
Server management (CRUD, search, import/export).
Developed by Inioluwa Adeyinka
"""

import json
import os
import shutil
from datetime import datetime
from pathlib import Path
from typing import Optional, List, Dict

from ssher.config import CONFIG_FILE, HISTORY_FILE, BACKUP_DIR, ALIASES_FILE
from ssher.formatting import error, success
from ssher.models import Server, ConnectionHistory
from ssher.crypto import EncryptionManager


class ServerManager:
    """Manages server configurations."""

    def __init__(self, encryption: EncryptionManager):
        self.encryption = encryption
        self.servers: List[Server] = []
        self.history: List[ConnectionHistory] = []
        self._aliases: Dict[str, str] = {}  # alias -> server name

    def load(self) -> bool:
        """Load servers from encrypted config."""
        if not CONFIG_FILE.exists():
            return True

        try:
            encrypted_data = CONFIG_FILE.read_bytes()
            decrypted = self.encryption.decrypt(encrypted_data)
            data = json.loads(decrypted)
            self.servers = [Server.from_dict(s) for s in data]
            return True
        except Exception as e:
            print(error(f"Failed to load configuration: {e}"))
            return False

    def save(self):
        """Save servers to encrypted config."""
        data = json.dumps([s.to_dict() for s in self.servers], indent=2)
        encrypted = self.encryption.encrypt(data)
        CONFIG_FILE.write_bytes(encrypted)
        CONFIG_FILE.chmod(0o600)

    def load_history(self):
        """Load connection history."""
        if not HISTORY_FILE.exists():
            return

        try:
            data = json.loads(HISTORY_FILE.read_text())
            self.history = [ConnectionHistory.from_dict(h) for h in data]
        except Exception:
            self.history = []

    def save_history(self):
        """Save connection history."""
        data = [h.to_dict() for h in self.history[-100:]]
        HISTORY_FILE.write_text(json.dumps(data, indent=2))
        HISTORY_FILE.chmod(0o600)

    def add_history(self, entry: ConnectionHistory):
        """Add a history entry."""
        self.history.append(entry)
        self.save_history()

    def add(self, server: Server):
        """Add a new server."""
        self.servers.append(server)
        self.save()

    def update(self, index: int, server: Server):
        """Update a server at index."""
        if 0 <= index < len(self.servers):
            self.servers[index] = server
            self.save()

    def delete(self, index: int) -> Optional[Server]:
        """Delete a server at index."""
        if 0 <= index < len(self.servers):
            server = self.servers.pop(index)
            self.save()
            return server
        return None

    def get_by_name(self, name: str) -> Optional[Server]:
        """Get server by name (case-insensitive)."""
        name_lower = name.lower()
        for server in self.servers:
            if server.name.lower() == name_lower:
                return server
        return None

    def get_by_alias(self, alias: str) -> Optional[Server]:
        """Get server by alias."""
        self.load_aliases()
        alias_lower = alias.lower()
        # Check server-level aliases first
        for server in self.servers:
            if any(a.lower() == alias_lower for a in server.aliases):
                return server
        # Check global alias file
        if alias_lower in self._aliases:
            return self.get_by_name(self._aliases[alias_lower])
        return None

    def resolve_server(self, identifier: str) -> Optional[Server]:
        """Resolve server by exact name -> alias -> fuzzy search."""
        # Exact name match
        server = self.get_by_name(identifier)
        if server:
            return server

        # Alias match
        server = self.get_by_alias(identifier)
        if server:
            return server

        # Fuzzy search (return top match if high score)
        results = self.search(identifier)
        if results and results[0][2] >= 80:
            return results[0][1]

        return None

    def get_by_index(self, index: int) -> Optional[Server]:
        """Get server by index (1-based)."""
        if 1 <= index <= len(self.servers):
            return self.servers[index - 1]
        return None

    def search(self, query: str) -> List[tuple]:
        """Fuzzy search servers. Returns list of (index, server, score)."""
        query_lower = query.lower()
        results = []

        for idx, server in enumerate(self.servers):
            score = 0
            name_lower = server.name.lower()

            if name_lower == query_lower:
                score = 100
            elif name_lower.startswith(query_lower):
                score = 80
            elif query_lower in name_lower:
                score = 60
            elif query_lower in server.host.lower():
                score = 50
            elif any(query_lower in tag.lower() for tag in server.tags):
                score = 40
            elif query_lower in server.group.lower():
                score = 30
            else:
                matches = sum(1 for c in query_lower if c in name_lower)
                if matches >= len(query_lower) * 0.6:
                    score = int(20 * matches / len(query_lower))

            if score > 0:
                results.append((idx, server, score))

        results.sort(key=lambda x: (-x[2], x[1].name))
        return results

    def get_groups(self) -> Dict[str, List[Server]]:
        """Get servers organized by group."""
        groups = {}
        for server in self.servers:
            group = server.group or "default"
            if group not in groups:
                groups[group] = []
            groups[group].append(server)
        return groups

    def get_favorites(self) -> List[Server]:
        """Get favorite servers."""
        return [s for s in self.servers if s.is_favorite]

    def get_recent(self, limit: int = 5) -> List[Server]:
        """Get recently connected servers."""
        connected = [s for s in self.servers if s.last_connected]
        connected.sort(key=lambda x: x.last_connected, reverse=True)
        return connected[:limit]

    def import_ssh_config(self, config_path: str = None) -> int:
        """Import servers from ~/.ssh/config."""
        if config_path is None:
            config_path = Path.home() / ".ssh" / "config"
        else:
            config_path = Path(config_path)

        if not config_path.exists():
            print(error(f"SSH config not found: {config_path}"))
            return 0

        imported = 0
        current_host = None
        current_config = {}

        with open(config_path) as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith('#'):
                    continue

                parts = line.split(None, 1)
                if len(parts) != 2:
                    continue

                key, value = parts[0].lower(), parts[1]

                if key == 'host':
                    if current_host and current_host != '*':
                        if not self.get_by_name(current_host):
                            server = Server(
                                name=current_host,
                                host=current_config.get('hostname', current_host),
                                user=current_config.get('user', os.getenv('USER', 'root')),
                                port=int(current_config.get('port', 22)),
                                key_path=current_config.get('identityfile', ''),
                                auth_type='key',
                                group='imported'
                            )
                            self.add(server)
                            imported += 1

                    current_host = value
                    current_config = {}
                else:
                    current_config[key] = value

        if current_host and current_host != '*':
            if not self.get_by_name(current_host):
                server = Server(
                    name=current_host,
                    host=current_config.get('hostname', current_host),
                    user=current_config.get('user', os.getenv('USER', 'root')),
                    port=int(current_config.get('port', 22)),
                    key_path=current_config.get('identityfile', ''),
                    auth_type='key',
                    group='imported'
                )
                self.add(server)
                imported += 1

        return imported

    def export_backup(self, path: str = None) -> str:
        """Export encrypted backup."""
        if path is None:
            timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
            path = BACKUP_DIR / f"ssher_backup_{timestamp}.enc"
        else:
            path = Path(path)

        shutil.copy(CONFIG_FILE, path)
        return str(path)

    def import_backup(self, path: str) -> bool:
        """Import from encrypted backup."""
        path = Path(path)
        if not path.exists():
            print(error(f"Backup file not found: {path}"))
            return False

        try:
            encrypted_data = path.read_bytes()
            decrypted = self.encryption.decrypt(encrypted_data)
            data = json.loads(decrypted)

            existing_names = {s.name.lower() for s in self.servers}
            imported = 0

            for server_data in data:
                server = Server.from_dict(server_data)
                if server.name.lower() not in existing_names:
                    self.servers.append(server)
                    existing_names.add(server.name.lower())
                    imported += 1

            self.save()
            print(success(f"Imported {imported} servers from backup."))
            return True

        except Exception as e:
            print(error(f"Failed to import backup: {e}"))
            return False

    # --- Alias Management ---

    def load_aliases(self):
        """Load global aliases from file."""
        if not ALIASES_FILE.exists():
            return
        try:
            self._aliases = json.loads(ALIASES_FILE.read_text())
        except Exception:
            self._aliases = {}

    def save_aliases(self):
        """Save global aliases to file."""
        ALIASES_FILE.write_text(json.dumps(self._aliases, indent=2))
        ALIASES_FILE.chmod(0o600)

    def add_alias(self, alias: str, server_name: str) -> bool:
        """Add a global alias for a server."""
        if not self.get_by_name(server_name):
            return False
        self.load_aliases()
        self._aliases[alias.lower()] = server_name
        self.save_aliases()
        return True

    def remove_alias(self, alias: str) -> bool:
        """Remove a global alias."""
        self.load_aliases()
        alias_lower = alias.lower()
        if alias_lower in self._aliases:
            del self._aliases[alias_lower]
            self.save_aliases()
            return True
        return False

    def list_aliases(self) -> Dict[str, str]:
        """List all aliases."""
        self.load_aliases()
        result = dict(self._aliases)
        # Also include server-level aliases
        for server in self.servers:
            for alias in server.aliases:
                result[alias] = server.name
        return result
