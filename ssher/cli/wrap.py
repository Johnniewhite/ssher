"""
sshpass-compatible wrap subcommand handler.
Developed by Inioluwa Adeyinka
"""

import os
import sys

from ssher.formatting import error, success, info
from ssher.sshpass_compat import get_password_from_env, get_password_from_file, get_password_from_fd
from ssher.password_auth import spawn_with_password, HAS_PEXPECT


def handle_wrap(args):
    """Handle the 'ssher wrap' subcommand (sshpass-compatible)."""
    if not HAS_PEXPECT:
        print(error("pexpect is required for wrap mode. Install with: pip install pexpect"))
        sys.exit(1)

    # Determine password source
    password = None

    if args.env:
        password = get_password_from_env()
    elif args.file:
        password = get_password_from_file(args.file)
    elif args.fd is not None:
        password = get_password_from_fd(args.fd)

    if password is None:
        print(error("Could not read password from specified source."))
        sys.exit(1)

    # Get the command to wrap
    command = args.wrap_command
    if not command:
        print(error("No command specified to wrap."))
        sys.exit(1)

    # Remove leading '--' if present
    if command and command[0] == '--':
        command = command[1:]

    custom_prompt = args.prompt

    ok, output = spawn_with_password(
        command, password, timeout=30,
        interact=True, custom_prompt=custom_prompt
    )

    if not ok:
        print(error(output))
        sys.exit(1)
