"""
Session recording and replay.
Developed by Inioluwa Adeyinka
"""

import os
import subprocess
import shutil
from datetime import datetime
from pathlib import Path

from ssher.config import RECORDINGS_DIR
from ssher.formatting import colored, Colors, success, error, warning, info


class SessionRecorder:
    """Manages session recordings."""

    def __init__(self):
        RECORDINGS_DIR.mkdir(mode=0o700, exist_ok=True)

    def create_recording_path(self, server_name: str) -> str:
        """Create a recording file path for a new session."""
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        safe_name = server_name.replace('/', '_').replace(' ', '_')
        filename = f"{safe_name}_{timestamp}.log"
        return str(RECORDINGS_DIR / filename)

    def list_recordings(self):
        """List all session recordings."""
        recordings = sorted(RECORDINGS_DIR.glob("*.log"), reverse=True)

        if not recordings:
            print(warning("No session recordings found."))
            return

        print(f"\n{colored('[Session Recordings]', Colors.CYAN, bold=True)}")
        print("-" * 60)

        for rec in recordings:
            size = rec.stat().st_size / 1024
            mtime = datetime.fromtimestamp(rec.stat().st_mtime)
            time_str = mtime.strftime("%Y-%m-%d %H:%M")
            print(f"  {colored(rec.name, Colors.GREEN)}  ({size:.1f} KB)  {time_str}")

        print("-" * 60)
        print(f"\n{info('Replay with:')} ssher record replay <filename>")

    def replay(self, filename: str):
        """Replay a session recording."""
        # Check if it's a full path or just a filename
        path = Path(filename)
        if not path.exists():
            path = RECORDINGS_DIR / filename

        if not path.exists():
            print(error(f"Recording not found: {filename}"))
            return

        print(f"{info('Replaying recording:')} {path.name}")
        print(colored("Press Ctrl+C to stop.", Colors.BRIGHT_BLACK))
        print("-" * 60)

        try:
            # Use cat for simple log files
            with open(path, 'r', errors='replace') as f:
                print(f.read())
        except KeyboardInterrupt:
            print(f"\n{info('Replay stopped.')}")
        except Exception as e:
            print(error(f"Error replaying: {e}"))
