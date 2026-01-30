"""
Shared pexpect password injection logic.
Consolidated from duplicated code in SSHConnection and FileTransfer.
Developed by Inioluwa Adeyinka
"""

from typing import List, Optional

try:
    import pexpect
    HAS_PEXPECT = True
except ImportError:
    HAS_PEXPECT = False


PASSWORD_PATTERNS = [
    'password:',
    'Password:',
    'passphrase',
    pexpect.EOF if HAS_PEXPECT else None,
    pexpect.TIMEOUT if HAS_PEXPECT else None,
    'Permission denied',
    'Connection refused',
    'No route to host',
    'Connection timed out',
]


def spawn_with_password(cmd: List[str], password: str, timeout: int = 30,
                        interact: bool = False, custom_prompt: str = None) -> tuple:
    """Spawn a command and inject password via pexpect.

    Args:
        cmd: Command as list of strings.
        password: Password to inject.
        timeout: Connection timeout.
        interact: If True, hand off to user after auth (for interactive SSH).
        custom_prompt: Custom password prompt pattern to match.

    Returns:
        (success: bool, output: str)
    """
    if not HAS_PEXPECT:
        return (False, "pexpect not installed")

    cmd_str = ' '.join(cmd)

    try:
        child = pexpect.spawn(cmd_str, encoding='utf-8', timeout=timeout)

        prompt_patterns = [custom_prompt] if custom_prompt else ['password:', 'Password:', 'passphrase']
        patterns = prompt_patterns + [
            pexpect.EOF,
            pexpect.TIMEOUT,
            'Permission denied',
            'Connection refused',
            'No route to host',
            'Connection timed out',
        ]

        idx = child.expect(patterns)
        num_prompts = len(prompt_patterns)

        if idx < num_prompts:  # Password prompt
            child.sendline(password)
            if interact:
                child.interact()
                return (True, "")
            else:
                child.expect(pexpect.EOF, timeout=max(timeout, 60))
                output = child.before if child.before else ""
                return (True, output)
        elif idx == num_prompts:  # EOF
            output = child.before if child.before else ""
            return (True, output)
        elif idx == num_prompts + 1:  # Timeout
            return (False, "Connection timed out.")
        else:  # Error
            error_msg = patterns[idx] if isinstance(patterns[idx], str) else "Unknown error"
            return (False, f"Connection failed: {error_msg}")

    except Exception as e:
        return (False, str(e))


def run_with_password_or_key(cmd: List[str], server_password: str,
                             auth_type: str, timeout: int = 30,
                             interact: bool = False,
                             transfer_timeout: int = 300) -> tuple:
    """Run a command with password auth (if needed) or directly.

    Returns:
        (success: bool, output: str)
    """
    import subprocess

    if auth_type == 'password' and server_password and HAS_PEXPECT:
        effective_timeout = transfer_timeout if not interact else timeout
        return spawn_with_password(cmd, server_password, effective_timeout, interact)
    else:
        if interact:
            subprocess.run(cmd)
            return (True, "")
        else:
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
            return (result.returncode == 0, result.stdout + result.stderr)
