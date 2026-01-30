"""
Argparse setup with all subcommands.
Developed by Inioluwa Adeyinka
"""

import argparse

from ssher import APP_NAME, DEVELOPER
from ssher.config import VERSION


def create_parser() -> argparse.ArgumentParser:
    """Create the argument parser."""
    parser = argparse.ArgumentParser(
        prog='ssher',
        description=f'{APP_NAME} - Ultimate SSH Configuration Manager by {DEVELOPER}',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  ssher                              Interactive mode
  ssher 1                            Connect to server #1
  ssher prod                         Connect to server named 'prod' (fuzzy)
  ssher list                         List all servers
  ssher add                          Add a new server
  ssher ping                         Check connectivity for all servers
  ssher exec "uptime"                Run 'uptime' on selected servers
  ssher upload f.txt /tmp            Upload f.txt to /tmp on server
  ssher download /log .              Download /log to current directory
  ssher wrap -e ssh user@host        sshpass-compatible password injection
  ssher vault status                 Show vault state
  ssher completion bash              Output bash completion script
  ssher generate-password            Generate a secure password
  ssher export-config                Export servers to ~/.ssh/config
  ssher copy <server> --field host   Copy server detail to clipboard
  ssher profile list                 List connection profiles
  ssher alias add prod pw            Create alias 'pw' for server 'prod'
  ssher record list                  List session recordings
  ssher import-csv servers.csv       Import from CSV file
  ssher export-json                  Export servers to JSON
        """
    )

    parser.add_argument('--version', '-v', action='version', version=f'{APP_NAME} v{VERSION}')
    parser.add_argument('--reconnect', action='store_true', help='Auto-reconnect on disconnect')
    parser.add_argument('--record', action='store_true', help='Record the SSH session')

    subparsers = parser.add_subparsers(dest='command', help='Commands')

    # List command
    subparsers.add_parser('list', aliases=['ls', 'l'], help='List all servers')

    # Add command
    subparsers.add_parser('add', aliases=['new'], help='Add a new server')

    # Ping command
    subparsers.add_parser('ping', aliases=['check'], help='Check connectivity for all servers')

    # History command
    subparsers.add_parser('history', aliases=['hist'], help='View connection history')

    # Import command
    import_parser = subparsers.add_parser('import', help='Import from SSH config')
    import_parser.add_argument('config_path', nargs='?', help='Path to SSH config file')

    # Exec command
    exec_parser = subparsers.add_parser('exec', aliases=['run'], help='Execute command on servers')
    exec_parser.add_argument('cmd', help='Command to execute')
    exec_parser.add_argument('-s', '--servers', help='Server numbers/names (comma-separated)')
    exec_parser.add_argument('-g', '--group', help='Server group')
    exec_parser.add_argument('--all', action='store_true', help='All servers')

    # Upload command
    upload_parser = subparsers.add_parser('upload', aliases=['up', 'put'], help='Upload file to server')
    upload_parser.add_argument('local', help='Local file path')
    upload_parser.add_argument('remote', help='Remote file path')
    upload_parser.add_argument('-s', '--server', help='Server number or name')
    upload_parser.add_argument('-r', '--recursive', action='store_true', help='Recursive upload')

    # Download command
    download_parser = subparsers.add_parser('download', aliases=['down', 'get'], help='Download file from server')
    download_parser.add_argument('remote', help='Remote file path')
    download_parser.add_argument('local', help='Local file path')
    download_parser.add_argument('-s', '--server', help='Server number or name')
    download_parser.add_argument('-r', '--recursive', action='store_true', help='Recursive download')

    # Backup command
    backup_parser = subparsers.add_parser('backup', help='Backup configurations')
    backup_parser.add_argument('path', nargs='?', help='Backup file path')

    # Groups command
    subparsers.add_parser('groups', help='List server groups')

    # --- New subcommands ---

    # Wrap (sshpass compat)
    wrap_parser = subparsers.add_parser('wrap', help='sshpass-compatible password wrapper')
    wrap_mode = wrap_parser.add_mutually_exclusive_group(required=True)
    wrap_mode.add_argument('-e', '--env', action='store_true',
                           help='Read password from $SSHPASS environment variable')
    wrap_mode.add_argument('-f', '--file', type=str, metavar='FILE',
                           help='Read password from file')
    wrap_mode.add_argument('-d', '--fd', type=int, metavar='FD',
                           help='Read password from file descriptor')
    wrap_parser.add_argument('-P', '--prompt', type=str, default=None,
                             help='Custom password prompt pattern')
    wrap_parser.add_argument('wrap_command', nargs=argparse.REMAINDER,
                             help='Command to execute (e.g., ssh user@host)')

    # Vault
    vault_parser = subparsers.add_parser('vault', help='Vault management')
    vault_parser.add_argument('vault_action', choices=['lock', 'unlock', 'change-password', 'status'],
                              help='Vault action')

    # Completion
    completion_parser = subparsers.add_parser('completion', help='Generate shell completion script')
    completion_parser.add_argument('shell', choices=['bash', 'zsh'], help='Shell type')

    # Copy to clipboard
    copy_parser = subparsers.add_parser('copy', help='Copy server details to clipboard')
    copy_parser.add_argument('copy_server', help='Server name or number')
    copy_parser.add_argument('--field', choices=['host', 'user', 'password', 'command', 'port'],
                             default='command', help='Field to copy (default: command)')

    # Export SSH config
    export_config_parser = subparsers.add_parser('export-config', help='Export servers to ~/.ssh/config')
    export_config_parser.add_argument('--append', action='store_true', help='Append to existing config')
    export_config_parser.add_argument('--output', type=str, default=None,
                                      help='Output file path (default: ~/.ssh/config)')

    # Profiles
    profile_parser = subparsers.add_parser('profile', help='Manage connection profiles')
    profile_sub = profile_parser.add_subparsers(dest='profile_action')

    profile_list = profile_sub.add_parser('list', help='List profiles')

    profile_add = profile_sub.add_parser('add', help='Add a profile')
    profile_add.add_argument('profile_name', help='Profile name')
    profile_add.add_argument('--timeout', type=int, default=30, help='Connection timeout')
    profile_add.add_argument('--keepalive', type=int, default=60, help='Keep-alive interval')
    profile_add.add_argument('--x11', action='store_true', help='Enable X11 forwarding')
    profile_add.add_argument('--reconnect', action='store_true', help='Enable auto-reconnect')

    profile_apply = profile_sub.add_parser('apply', help='Apply profile to server')
    profile_apply.add_argument('profile_name', help='Profile name')
    profile_apply.add_argument('--server', required=True, help='Server name')

    profile_remove = profile_sub.add_parser('remove', help='Remove a profile')
    profile_remove.add_argument('profile_name', help='Profile name')

    # Aliases
    alias_parser = subparsers.add_parser('alias', help='Manage server aliases')
    alias_sub = alias_parser.add_subparsers(dest='alias_action')

    alias_add = alias_sub.add_parser('add', help='Add an alias')
    alias_add.add_argument('alias_server', help='Server name')
    alias_add.add_argument('alias_name', help='Alias name')

    alias_remove = alias_sub.add_parser('remove', help='Remove an alias')
    alias_remove.add_argument('alias_name', help='Alias to remove')

    alias_sub.add_parser('list', help='List all aliases')

    # Session recording
    record_parser = subparsers.add_parser('record', help='Manage session recordings')
    record_sub = record_parser.add_subparsers(dest='record_action')
    record_sub.add_parser('list', help='List recordings')
    record_replay = record_sub.add_parser('replay', help='Replay a recording')
    record_replay.add_argument('record_file', help='Recording file to replay')

    # Password generator
    gen_parser = subparsers.add_parser('generate-password', aliases=['genpass'],
                                       help='Generate a secure password')
    gen_parser.add_argument('--length', type=int, default=20, help='Password length (default: 20)')
    gen_parser.add_argument('--no-symbols', action='store_true', help='Exclude symbols')
    gen_parser.add_argument('--no-numbers', action='store_true', help='Exclude numbers')
    gen_parser.add_argument('--count', type=int, default=1, help='Number of passwords to generate')

    # Batch import/export
    import_csv_parser = subparsers.add_parser('import-csv', help='Import servers from CSV')
    import_csv_parser.add_argument('csv_file', help='Path to CSV file')

    import_json_parser = subparsers.add_parser('import-json', help='Import servers from JSON')
    import_json_parser.add_argument('json_file', help='Path to JSON file')

    subparsers.add_parser('export-csv', help='Export servers to CSV')
    subparsers.add_parser('export-json', help='Export servers to JSON')

    return parser
