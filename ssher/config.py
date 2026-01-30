"""
Configuration constants and paths.
Developed by Inioluwa Adeyinka
"""

from pathlib import Path

from ssher import __version__, APP_NAME, DEVELOPER

VERSION = __version__

CONFIG_DIR = Path.home() / ".ssher"
CONFIG_FILE = CONFIG_DIR / "servers.enc"
KEY_FILE = CONFIG_DIR / ".key"
SALT_FILE = CONFIG_DIR / ".salt"
HISTORY_FILE = CONFIG_DIR / "history.json"
SESSION_FILE = CONFIG_DIR / ".session"
BACKUP_DIR = CONFIG_DIR / "backups"
PROFILES_FILE = CONFIG_DIR / "profiles.json"
ALIASES_FILE = CONFIG_DIR / "aliases.json"
RECORDINGS_DIR = CONFIG_DIR / "recordings"

# Session timeout in seconds (30 minutes)
SESSION_TIMEOUT = 1800
