"""
Clipboard support with platform detection.
Developed by Inioluwa Adeyinka
"""

import subprocess
import shutil
import sys


def copy_to_clipboard(text: str) -> bool:
    """Copy text to system clipboard.

    Detects platform and uses appropriate clipboard command:
    - macOS: pbcopy
    - Linux: xclip or xsel
    - WSL: clip.exe

    Returns:
        True if successful, False otherwise.
    """
    try:
        if sys.platform == 'darwin':
            # macOS
            proc = subprocess.Popen(['pbcopy'], stdin=subprocess.PIPE)
            proc.communicate(text.encode('utf-8'))
            return proc.returncode == 0

        elif sys.platform.startswith('linux'):
            # Check for xclip first
            if shutil.which('xclip'):
                proc = subprocess.Popen(
                    ['xclip', '-selection', 'clipboard'],
                    stdin=subprocess.PIPE
                )
                proc.communicate(text.encode('utf-8'))
                return proc.returncode == 0

            # Try xsel
            if shutil.which('xsel'):
                proc = subprocess.Popen(
                    ['xsel', '--clipboard', '--input'],
                    stdin=subprocess.PIPE
                )
                proc.communicate(text.encode('utf-8'))
                return proc.returncode == 0

            # Try clip.exe (WSL)
            if shutil.which('clip.exe'):
                proc = subprocess.Popen(['clip.exe'], stdin=subprocess.PIPE)
                proc.communicate(text.encode('utf-8'))
                return proc.returncode == 0

            return False

        else:
            return False

    except Exception:
        return False
