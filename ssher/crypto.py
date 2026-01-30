"""
Pure encryption logic (no interactive prompts).
Developed by Inioluwa Adeyinka
"""

import os
import json
import base64
from datetime import datetime, timedelta
from pathlib import Path
from typing import Optional

from cryptography.fernet import Fernet
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.kdf.pbkdf2 import PBKDF2HMAC

from ssher.config import (
    CONFIG_DIR, KEY_FILE, SALT_FILE, SESSION_FILE,
    BACKUP_DIR, SESSION_TIMEOUT, RECORDINGS_DIR,
)


class EncryptionManager:
    """Handles all encryption operations (pure crypto, no interactive prompts)."""

    def __init__(self):
        self.fernet: Optional[Fernet] = None
        self._ensure_config_dir()

    def _ensure_config_dir(self):
        """Ensure configuration directory exists."""
        CONFIG_DIR.mkdir(mode=0o700, exist_ok=True)
        BACKUP_DIR.mkdir(mode=0o700, exist_ok=True)
        RECORDINGS_DIR.mkdir(mode=0o700, exist_ok=True)

    def _generate_key(self, password: str, salt: bytes) -> bytes:
        """Generate encryption key from password."""
        kdf = PBKDF2HMAC(
            algorithm=hashes.SHA256(),
            length=32,
            salt=salt,
            iterations=480000,
        )
        return base64.urlsafe_b64encode(kdf.derive(password.encode()))

    def check_session(self) -> Optional[str]:
        """Check if there's a valid session."""
        if not SESSION_FILE.exists():
            return None

        try:
            data = json.loads(SESSION_FILE.read_text())
            timestamp = datetime.fromisoformat(data['timestamp'])
            if datetime.now() - timestamp < timedelta(seconds=SESSION_TIMEOUT):
                return data.get('key_hash')
        except (json.JSONDecodeError, KeyError, ValueError):
            pass

        return None

    def save_session(self, key_hash: str):
        """Save session for timeout feature."""
        SESSION_FILE.write_text(json.dumps({
            'timestamp': datetime.now().isoformat(),
            'key_hash': key_hash
        }))
        SESSION_FILE.chmod(0o600)

    def clear_session(self):
        """Clear the current session."""
        if SESSION_FILE.exists():
            SESSION_FILE.unlink()

    def get_session_info(self) -> Optional[dict]:
        """Get session information (for vault status)."""
        if not SESSION_FILE.exists():
            return None

        try:
            data = json.loads(SESSION_FILE.read_text())
            timestamp = datetime.fromisoformat(data['timestamp'])
            elapsed = datetime.now() - timestamp
            remaining = timedelta(seconds=SESSION_TIMEOUT) - elapsed
            if remaining.total_seconds() > 0:
                return {
                    'started': timestamp.isoformat(),
                    'remaining_seconds': int(remaining.total_seconds()),
                    'timeout': SESSION_TIMEOUT,
                }
        except (json.JSONDecodeError, KeyError, ValueError):
            pass

        return None

    def initialize_first_time(self, password: str) -> bool:
        """Initialize encryption for first time with a given password."""
        if not SALT_FILE.exists():
            salt = os.urandom(16)
            SALT_FILE.write_bytes(salt)
            SALT_FILE.chmod(0o600)
        else:
            salt = SALT_FILE.read_bytes()

        key = self._generate_key(password, salt)
        self.fernet = Fernet(key)

        # Store verification token
        verification = self.fernet.encrypt(b"SSHER_VERIFIED")
        KEY_FILE.write_bytes(verification)
        KEY_FILE.chmod(0o600)

        # Save session
        self.save_session(base64.b64encode(key).decode()[:32])
        return True

    def initialize_with_password(self, password: str) -> bool:
        """Initialize encryption with existing password."""
        if not SALT_FILE.exists():
            return False

        salt = SALT_FILE.read_bytes()
        key = self._generate_key(password, salt)
        self.fernet = Fernet(key)

        try:
            verification = KEY_FILE.read_bytes()
            decrypted = self.fernet.decrypt(verification)
            if decrypted != b"SSHER_VERIFIED":
                raise ValueError("Invalid verification")

            self.save_session(base64.b64encode(key).decode()[:32])
            return True
        except Exception:
            self.fernet = None
            return False

    def change_password(self, old_password: str, new_password: str) -> bool:
        """Change master password: re-encrypt everything."""
        if not self.initialize_with_password(old_password):
            return False

        # Read all encrypted data before re-keying
        from ssher.config import CONFIG_FILE
        decrypted_data = None
        if CONFIG_FILE.exists():
            try:
                decrypted_data = self.decrypt(CONFIG_FILE.read_bytes())
            except Exception:
                return False

        # Generate new key
        salt = os.urandom(16)
        SALT_FILE.write_bytes(salt)
        SALT_FILE.chmod(0o600)

        key = self._generate_key(new_password, salt)
        self.fernet = Fernet(key)

        # Re-encrypt verification token
        verification = self.fernet.encrypt(b"SSHER_VERIFIED")
        KEY_FILE.write_bytes(verification)
        KEY_FILE.chmod(0o600)

        # Re-encrypt server data
        if decrypted_data is not None:
            encrypted = self.encrypt(decrypted_data)
            CONFIG_FILE.write_bytes(encrypted)
            CONFIG_FILE.chmod(0o600)

        self.save_session(base64.b64encode(key).decode()[:32])
        return True

    def encrypt(self, data: str) -> bytes:
        """Encrypt string data."""
        if not self.fernet:
            raise RuntimeError("Encryption not initialized")
        return self.fernet.encrypt(data.encode())

    def decrypt(self, data: bytes) -> str:
        """Decrypt data to string."""
        if not self.fernet:
            raise RuntimeError("Encryption not initialized")
        return self.fernet.decrypt(data).decode()

    @staticmethod
    def is_first_time() -> bool:
        """Check if this is a first-time setup."""
        return not KEY_FILE.exists()
