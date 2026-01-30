"""
File transfer operations (SCP, SFTP, rsync).
Developed by Inioluwa Adeyinka
"""

import os
import subprocess
import shutil
from typing import List

from ssher.formatting import success, error, info
from ssher.models import Server
from ssher.password_auth import HAS_PEXPECT, run_with_password_or_key


class FileTransfer:
    """Handles SCP and SFTP operations."""

    def __init__(self, server_manager):
        self.manager = server_manager

    def _build_scp_base(self, server: Server) -> List[str]:
        """Build base SCP command."""
        cmd = ['scp']

        if server.port != 22:
            cmd.extend(['-P', str(server.port)])

        if server.auth_type == 'key' and server.key_path:
            key_path = os.path.expanduser(server.key_path)
            if os.path.exists(key_path):
                cmd.extend(['-i', key_path])

        if server.jump_host:
            jump_server = self.manager.get_by_name(server.jump_host)
            if jump_server:
                jump_str = f"{jump_server.user}@{jump_server.host}"
                if jump_server.port != 22:
                    jump_str += f":{jump_server.port}"
                cmd.extend(['-J', jump_str])

        return cmd

    def upload(self, server: Server, local_path: str, remote_path: str,
               recursive: bool = False) -> bool:
        """Upload file(s) to server."""
        cmd = self._build_scp_base(server)

        if recursive:
            cmd.append('-r')

        cmd.append(local_path)
        cmd.append(f'{server.user}@{server.host}:{remote_path}')

        print(f"{info('Uploading')} {local_path} \u2192 {server.name}:{remote_path}")

        try:
            ok, output = run_with_password_or_key(
                cmd, server.password, server.auth_type,
                timeout=server.connection_timeout,
                interact=False, transfer_timeout=300
            )
            if ok:
                print(success("Upload complete."))
                return True
            else:
                print(error(f"Upload failed: {output}"))
                return False
        except Exception as e:
            print(error(f"Upload error: {e}"))
            return False

    def download(self, server: Server, remote_path: str, local_path: str,
                 recursive: bool = False) -> bool:
        """Download file(s) from server."""
        cmd = self._build_scp_base(server)

        if recursive:
            cmd.append('-r')

        cmd.append(f'{server.user}@{server.host}:{remote_path}')
        cmd.append(local_path)

        print(f"{info('Downloading')} {server.name}:{remote_path} \u2192 {local_path}")

        try:
            ok, output = run_with_password_or_key(
                cmd, server.password, server.auth_type,
                timeout=server.connection_timeout,
                interact=False, transfer_timeout=300
            )
            if ok:
                print(success("Download complete."))
                return True
            else:
                print(error(f"Download failed: {output}"))
                return False
        except Exception as e:
            print(error(f"Download error: {e}"))
            return False

    def rsync(self, server: Server, source: str, dest: str, to_remote: bool = True,
              delete: bool = False, dry_run: bool = False) -> bool:
        """Rsync files to/from server."""
        if not shutil.which('rsync'):
            print(error("rsync is not installed."))
            return False

        cmd = ['rsync', '-avz', '--progress']

        if delete:
            cmd.append('--delete')

        if dry_run:
            cmd.append('--dry-run')

        ssh_cmd = f'ssh -p {server.port}'
        if server.auth_type == 'key' and server.key_path:
            key_path = os.path.expanduser(server.key_path)
            if os.path.exists(key_path):
                ssh_cmd += f' -i {key_path}'

        cmd.extend(['-e', ssh_cmd])

        if to_remote:
            cmd.append(source)
            cmd.append(f'{server.user}@{server.host}:{dest}')
            print(f"{info('Syncing')} {source} \u2192 {server.name}:{dest}")
        else:
            cmd.append(f'{server.user}@{server.host}:{source}')
            cmd.append(dest)
            print(f"{info('Syncing')} {server.name}:{source} \u2192 {dest}")

        try:
            result = subprocess.run(cmd)
            return result.returncode == 0
        except Exception as e:
            print(error(f"Rsync error: {e}"))
            return False

    def sftp_interactive(self, server: Server):
        """Open interactive SFTP session."""
        cmd = ['sftp']

        if server.port != 22:
            cmd.extend(['-P', str(server.port)])

        if server.auth_type == 'key' and server.key_path:
            key_path = os.path.expanduser(server.key_path)
            if os.path.exists(key_path):
                cmd.extend(['-i', key_path])

        cmd.append(f'{server.user}@{server.host}')

        print(f"{info('Opening SFTP session to')} {server.name}...")

        try:
            subprocess.run(cmd)
        except KeyboardInterrupt:
            print(f"\n{info('SFTP session closed.')}")
