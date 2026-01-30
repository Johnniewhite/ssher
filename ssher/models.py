"""
Data models for SSHer.
Developed by Inioluwa Adeyinka
"""

import os
from datetime import datetime
from pathlib import Path
from typing import List, Dict
from dataclasses import dataclass, asdict, field


@dataclass
class Server:
    """Server configuration data class."""
    name: str
    host: str
    user: str
    port: int = 22
    auth_type: str = "key"  # "key" or "password"
    password: str = ""
    key_path: str = ""
    group: str = "default"
    tags: List[str] = field(default_factory=list)
    notes: str = ""
    jump_host: str = ""  # Server name to use as jump host
    local_forwards: List[Dict[str, int]] = field(default_factory=list)
    remote_forwards: List[Dict[str, int]] = field(default_factory=list)
    x11_forward: bool = False
    keep_alive: int = 60  # seconds, 0 to disable
    connection_timeout: int = 30
    custom_options: Dict[str, str] = field(default_factory=dict)
    created_at: str = ""
    last_connected: str = ""
    connection_count: int = 0
    is_favorite: bool = False
    password_expires: str = ""  # ISO date string
    aliases: List[str] = field(default_factory=list)
    profile: str = ""  # Connection profile name
    auto_reconnect: bool = False
    max_reconnect_retries: int = 3

    def __post_init__(self):
        if not self.created_at:
            self.created_at = datetime.now().isoformat()
        if not self.key_path and self.auth_type == "key":
            self.key_path = str(Path.home() / ".ssh" / "id_rsa")
        if isinstance(self.tags, str):
            self.tags = [t.strip() for t in self.tags.split(",") if t.strip()]

    def to_dict(self) -> dict:
        return asdict(self)

    @classmethod
    def from_dict(cls, data: dict) -> 'Server':
        valid_fields = {f.name for f in cls.__dataclass_fields__.values()}
        filtered = {k: v for k, v in data.items() if k in valid_fields}
        return cls(**filtered)


@dataclass
class ConnectionHistory:
    """Connection history entry."""
    server_name: str
    host: str
    user: str
    timestamp: str
    duration: int = 0  # seconds
    success: bool = True
    error_message: str = ""

    def to_dict(self) -> dict:
        return asdict(self)

    @classmethod
    def from_dict(cls, data: dict) -> 'ConnectionHistory':
        valid_fields = {f.name for f in cls.__dataclass_fields__.values()}
        filtered = {k: v for k, v in data.items() if k in valid_fields}
        return cls(**filtered)


@dataclass
class ConnectionProfile:
    """Reusable connection profile with SSH options."""
    name: str
    connection_timeout: int = 30
    keep_alive: int = 60
    x11_forward: bool = False
    custom_options: Dict[str, str] = field(default_factory=dict)
    auto_reconnect: bool = False
    max_reconnect_retries: int = 3

    def to_dict(self) -> dict:
        return asdict(self)

    @classmethod
    def from_dict(cls, data: dict) -> 'ConnectionProfile':
        valid_fields = {f.name for f in cls.__dataclass_fields__.values()}
        filtered = {k: v for k, v in data.items() if k in valid_fields}
        return cls(**filtered)
