"""
SSH connection handling.
Developed by Inioluwa Adeyinka
"""

import os
import subprocess
import socket
import time
from datetime import datetime
from typing import Optional, List, Dict
from concurrent.futures import ThreadPoolExecutor, as_completed

from ssher.formatting import colored, Colors, success, error, warning, info
from ssher.models import Server, ConnectionHistory
from ssher.password_auth import HAS_PEXPECT, run_with_password_or_key, spawn_with_password


class SSHConnection:
    """Handles SSH connections."""

    def __init__(self, server_manager):
        self.manager = server_manager

    def _build_ssh_command(self, server: Server, command: str = None) -> List[str]:
        """Build SSH command with all options."""
        cmd = ['ssh']

        cmd.extend(['-o', f'ConnectTimeout={server.connection_timeout}'])

        if server.keep_alive > 0:
            cmd.extend(['-o', f'ServerAliveInterval={server.keep_alive}'])
            cmd.extend(['-o', 'ServerAliveCountMax=3'])

        cmd.extend(['-o', 'StrictHostKeyChecking=accept-new'])

        if server.auth_type == 'key' and server.key_path:
            key_path = os.path.expanduser(server.key_path)
            if os.path.exists(key_path):
                cmd.extend(['-i', key_path])

        if server.port != 22:
            cmd.extend(['-p', str(server.port)])

        if server.x11_forward:
            cmd.append('-X')

        for fwd in server.local_forwards:
            cmd.extend(['-L', f"{fwd['local']}:localhost:{fwd['remote']}"])

        for fwd in server.remote_forwards:
            cmd.extend(['-R', f"{fwd['remote']}:localhost:{fwd['local']}"])

        if server.jump_host:
            jump_server = self.manager.get_by_name(server.jump_host)
            if jump_server:
                jump_str = f"{jump_server.user}@{jump_server.host}"
                if jump_server.port != 22:
                    jump_str += f":{jump_server.port}"
                cmd.extend(['-J', jump_str])

        for key, value in server.custom_options.items():
            cmd.extend(['-o', f'{key}={value}'])

        cmd.append(f'{server.user}@{server.host}')

        if command:
            cmd.append(command)

        return cmd

    def connect(self, server: Server, command: str = None,
                reconnect: bool = False, record_file: str = None) -> bool:
        """Connect to a server."""
        print(f"\n{colored('Connecting to', Colors.CYAN)} {colored(server.name, Colors.GREEN, bold=True)} "
              f"({server.user}@{server.host}:{server.port})...")

        server.last_connected = datetime.now().isoformat()
        server.connection_count += 1

        for idx, s in enumerate(self.manager.servers):
            if s.name == server.name:
                self.manager.update(idx, server)
                break

        should_reconnect = reconnect or server.auto_reconnect
        max_retries = server.max_reconnect_retries if should_reconnect else 0
        attempt = 0

        while True:
            start_time = time.time()
            success_flag = True
            error_msg = ""

            try:
                cmd = self._build_ssh_command(server, command)

                if record_file:
                    # Wrap with script command for recording
                    cmd = ['script', '-q', record_file] + cmd

                ok, output = run_with_password_or_key(
                    cmd, server.password, server.auth_type,
                    timeout=server.connection_timeout, interact=True
                )
                success_flag = ok
                if not ok:
                    error_msg = output

            except KeyboardInterrupt:
                print(f"\n{info('Connection closed.')}")
                break
            except Exception as e:
                success_flag = False
                error_msg = str(e)
                print(error(f"Connection error: {e}"))

            duration = int(time.time() - start_time)

            history_entry = ConnectionHistory(
                server_name=server.name,
                host=server.host,
                user=server.user,
                timestamp=datetime.now().isoformat(),
                duration=duration,
                success=success_flag,
                error_message=error_msg
            )
            self.manager.add_history(history_entry)

            if not success_flag and should_reconnect and attempt < max_retries:
                attempt += 1
                backoff = min(2 ** attempt, 30)
                print(warning(f"Connection failed. Reconnecting in {backoff}s "
                              f"(attempt {attempt}/{max_retries})..."))
                try:
                    time.sleep(backoff)
                except KeyboardInterrupt:
                    print(f"\n{info('Reconnection cancelled.')}")
                    break
                continue

            break

        return success_flag

    def execute_on_multiple(self, servers: List[Server], command: str,
                            parallel: bool = True) -> Dict[str, str]:
        """Execute command on multiple servers."""
        results = {}

        def execute_single(server: Server) -> tuple:
            try:
                cmd = self._build_ssh_command(server, command)
                ok, output = run_with_password_or_key(
                    cmd, server.password, server.auth_type,
                    timeout=server.connection_timeout, interact=False
                )
                return (server.name, output)
            except Exception as e:
                return (server.name, f"Error: {e}")

        if parallel:
            with ThreadPoolExecutor(max_workers=10) as executor:
                futures = {executor.submit(execute_single, s): s for s in servers}
                for future in as_completed(futures):
                    name, output = future.result()
                    results[name] = output
        else:
            for server in servers:
                name, output = execute_single(server)
                results[name] = output

        return results

    def check_connectivity(self, server: Server) -> tuple:
        """Check if server is reachable. Returns (reachable, latency_ms)."""
        try:
            start = time.time()
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(5)
            result = sock.connect_ex((server.host, server.port))
            latency = int((time.time() - start) * 1000)
            sock.close()
            return (result == 0, latency)
        except Exception:
            return (False, -1)

    def check_all_connectivity(self, servers: List[Server] = None) -> Dict[str, tuple]:
        """Check connectivity for all servers in parallel."""
        if servers is None:
            servers = self.manager.servers

        results = {}

        with ThreadPoolExecutor(max_workers=20) as executor:
            futures = {executor.submit(self.check_connectivity, s): s for s in servers}
            for future in as_completed(futures):
                server = futures[future]
                results[server.name] = future.result()

        return results
