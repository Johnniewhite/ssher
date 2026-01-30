"""
sshpass-compatible wrapper functions.
Reads passwords from env, file, or file descriptor.
Developed by Inioluwa Adeyinka
"""

import os
from typing import Optional

from ssher.formatting import error


def get_password_from_env() -> Optional[str]:
    """Read password from $SSHPASS environment variable."""
    password = os.environ.get('SSHPASS')
    if password is None:
        print(error("$SSHPASS environment variable not set."))
        return None
    return password


def get_password_from_file(path: str) -> Optional[str]:
    """Read password from a file (first line)."""
    try:
        with open(path, 'r') as f:
            password = f.readline().rstrip('\n')
        return password
    except FileNotFoundError:
        print(error(f"Password file not found: {path}"))
        return None
    except PermissionError:
        print(error(f"Permission denied reading: {path}"))
        return None


def get_password_from_fd(fd: int) -> Optional[str]:
    """Read password from a file descriptor."""
    try:
        with os.fdopen(fd, 'r', closefd=False) as f:
            password = f.readline().rstrip('\n')
        return password
    except OSError as e:
        print(error(f"Failed to read from file descriptor {fd}: {e}"))
        return None
