"""
CLI dispatch logic.
Developed by Inioluwa Adeyinka
"""

import sys
import getpass

from ssher.config import VERSION, KEY_FILE
from ssher.formatting import colored, Colors, success, error, warning, info
from ssher.crypto import EncryptionManager
from ssher.servers import ServerManager
from ssher.connection import SSHConnection
from ssher.transfer import FileTransfer
from ssher.ui import InteractiveUI
from ssher.cli.parser import create_parser


def cli_main():
    """CLI entry point."""
    parser = create_parser()
    args, remaining = parser.parse_known_args()

    # If no subcommand was matched, treat remaining args as a target (server name/number)
    target = remaining[0] if remaining else None

    # Handle commands that don't need encryption
    if args.command in ['generate-password', 'genpass']:
        _handle_generate_password(args)
        return

    if args.command == 'completion':
        _handle_completion(args)
        return

    # Wrap command has its own auth flow
    if args.command == 'wrap':
        from ssher.cli.wrap import handle_wrap
        handle_wrap(args)
        return

    # Initialize encryption
    encryption = EncryptionManager()
    first_time = encryption.is_first_time()

    if first_time:
        print(f"\n{colored('[First Time Setup]', Colors.CYAN, bold=True)}")
        print("Create a master password to encrypt your SSH configurations.\n")

        while True:
            password = getpass.getpass(colored("Enter master password: ", Colors.YELLOW))
            if len(password) < 4:
                print(error("Password must be at least 4 characters."))
                continue
            confirm = getpass.getpass(colored("Confirm master password: ", Colors.YELLOW))
            if password != confirm:
                print(error("Passwords don't match. Try again."))
                continue
            break

        if not encryption.initialize_first_time(password):
            sys.exit(1)
        print(success("Master password set successfully!"))
    else:
        # Vault unlock or prompt
        if args.command == 'vault' and args.vault_action == 'unlock':
            password = getpass.getpass(colored("Enter master password: ", Colors.YELLOW))
            if not encryption.initialize_with_password(password):
                print(error("Incorrect master password!"))
                sys.exit(1)
            print(success("Vault unlocked."))
        elif args.command == 'vault' and args.vault_action == 'status':
            from ssher.vault import VaultManager
            vault = VaultManager(encryption)
            vault.status()
            return
        elif args.command == 'vault' and args.vault_action == 'lock':
            from ssher.vault import VaultManager
            vault = VaultManager(encryption)
            vault.lock()
            return
        elif args.command == 'vault' and args.vault_action == 'change-password':
            from ssher.vault import VaultManager
            vault = VaultManager(encryption)
            vault.interactive_change_password()
            return
        else:
            password = getpass.getpass(colored("Enter master password: ", Colors.YELLOW))
            if not encryption.initialize_with_password(password):
                print(error("Incorrect master password!"))
                sys.exit(1)

    # Initialize managers
    manager = ServerManager(encryption)
    manager.load()
    manager.load_history()

    ssh = SSHConnection(manager)
    transfer = FileTransfer(manager)
    ui = InteractiveUI(manager, ssh, transfer)

    # Handle commands
    if args.command in ['list', 'ls', 'l']:
        ui.print_header()
        ui.display_servers()

    elif args.command in ['add', 'new']:
        ui.print_header()
        server = ui.prompt_add_server()
        if server:
            manager.add(server)
            print(success(f"Server '{server.name}' added!"))

    elif args.command in ['ping', 'check']:
        ui.print_header()
        ui.prompt_ping_all()

    elif args.command in ['history', 'hist']:
        ui.print_header()
        ui.prompt_view_history()

    elif args.command == 'import':
        ui.print_header()
        count = manager.import_ssh_config(getattr(args, 'config_path', None))
        print(success(f"Imported {count} servers.") if count else info("No new servers to import."))

    elif args.command in ['exec', 'run']:
        _handle_exec(args, manager, ssh)

    elif args.command in ['upload', 'up', 'put']:
        _handle_upload(args, manager, transfer, ui)

    elif args.command in ['download', 'down', 'get']:
        _handle_download(args, manager, transfer, ui)

    elif args.command == 'backup':
        path = manager.export_backup(getattr(args, 'path', None))
        print(success(f"Backup created: {path}"))

    elif args.command == 'groups':
        groups = manager.get_groups()
        print(f"\n{colored('Server Groups:', Colors.CYAN, bold=True)}")
        for group, servers in sorted(groups.items()):
            print(f"  \u2022 {colored(group, Colors.GREEN)}: {len(servers)} servers")
            for s in servers:
                print(f"    - {s.name} ({s.host})")

    elif args.command == 'copy':
        _handle_copy(args, manager)

    elif args.command == 'export-config':
        _handle_export_config(args, manager)

    elif args.command == 'profile':
        _handle_profile(args, manager)

    elif args.command == 'alias':
        _handle_alias(args, manager)

    elif args.command == 'record':
        _handle_record(args)

    elif args.command == 'import-csv':
        _handle_import_csv(args, manager)

    elif args.command == 'import-json':
        _handle_import_json(args, manager)

    elif args.command == 'export-csv':
        _handle_export_csv(manager)

    elif args.command == 'export-json':
        _handle_export_json(manager)

    elif target:
        reconnect = getattr(args, 'reconnect', False)
        record = getattr(args, 'record', False)
        _handle_direct_connect(target, manager, ssh, reconnect, record)

    else:
        ui.run()


def _handle_exec(args, manager, ssh):
    """Handle exec command."""
    servers = []
    if getattr(args, 'all', False):
        servers = manager.servers
    elif args.group:
        servers = [s for s in manager.servers if s.group.lower() == args.group.lower()]
    elif args.servers:
        for part in args.servers.split(','):
            part = part.strip()
            if part.isdigit():
                s = manager.get_by_index(int(part))
                if s:
                    servers.append(s)
            else:
                s = manager.get_by_name(part)
                if s:
                    servers.append(s)

    if not servers:
        print(error("No servers specified. Use -s, -g, or --all"))
        sys.exit(1)

    print(f"{colored('Executing on:', Colors.CYAN)} {', '.join(s.name for s in servers)}")
    results = ssh.execute_on_multiple(servers, args.cmd)

    for name, output in results.items():
        banner = colored("\u2550\u2550\u2550 " + name + " \u2550\u2550\u2550", Colors.CYAN, bold=True)
        print(f"\n{banner}")
        print(output if output.strip() else colored("(no output)", Colors.BRIGHT_BLACK))


def _handle_upload(args, manager, transfer, ui):
    """Handle upload command."""
    server = _resolve_server(args.server, manager, ui)
    if server:
        transfer.upload(server, args.local, args.remote, args.recursive)
    else:
        print(error("Invalid server."))


def _handle_download(args, manager, transfer, ui):
    """Handle download command."""
    server = _resolve_server(args.server, manager, ui)
    if server:
        transfer.download(server, args.remote, args.local, args.recursive)
    else:
        print(error("Invalid server."))


def _resolve_server(server_arg, manager, ui):
    """Resolve a server from CLI argument or interactive prompt."""
    if server_arg:
        if server_arg.isdigit():
            return manager.get_by_index(int(server_arg))
        else:
            return manager.resolve_server(server_arg)

    ui.display_servers(show_groups=False)
    choice = input(f"\n{colored('Server #:', Colors.CYAN)} ").strip()
    if choice.isdigit():
        return manager.get_by_index(int(choice))
    return None


def _handle_direct_connect(target, manager, ssh, reconnect=False, record=False):
    """Handle direct connection by number or name."""
    record_file = None
    if record:
        from ssher.recording import SessionRecorder
        recorder = SessionRecorder()
        record_file = recorder.create_recording_path(target)

    if target.isdigit():
        server = manager.get_by_index(int(target))
        if server:
            ssh.connect(server, reconnect=reconnect, record_file=record_file)
        else:
            print(error(f"Invalid server number: {target}"))
    else:
        server = manager.resolve_server(target)
        if server:
            ssh.connect(server, reconnect=reconnect, record_file=record_file)
        else:
            # Show fuzzy results
            results = manager.search(target)
            if results:
                print(f"{colored('Multiple matches:', Colors.YELLOW)}")
                for i, (_, s, _) in enumerate(results[:5], 1):
                    print(f"  [{i}] {s.name} ({s.host})")
                choice = input(f"\n{colored('Select:', Colors.CYAN)} ").strip()
                if choice.isdigit() and 1 <= int(choice) <= len(results):
                    _, server, _ = results[int(choice) - 1]
                    ssh.connect(server, reconnect=reconnect, record_file=record_file)
            else:
                print(error(f"Server not found: {target}"))


def _handle_generate_password(args):
    """Handle generate-password command."""
    from ssher.password_gen import generate_password
    length = getattr(args, 'length', 20)
    no_symbols = getattr(args, 'no_symbols', False)
    no_numbers = getattr(args, 'no_numbers', False)
    count = getattr(args, 'count', 1)

    for _ in range(count):
        pw = generate_password(length=length, symbols=not no_symbols, numbers=not no_numbers)
        print(pw)


def _handle_completion(args):
    """Handle completion command."""
    from ssher.completion import generate_completion
    print(generate_completion(args.shell))


def _handle_copy(args, manager):
    """Handle copy command."""
    from ssher.clipboard import copy_to_clipboard

    server_id = args.copy_server
    if server_id.isdigit():
        server = manager.get_by_index(int(server_id))
    else:
        server = manager.resolve_server(server_id)

    if not server:
        print(error(f"Server not found: {server_id}"))
        return

    field = args.field
    if field == 'host':
        value = server.host
    elif field == 'user':
        value = server.user
    elif field == 'password':
        if not server.password:
            print(error("No password stored for this server."))
            return
        value = server.password
    elif field == 'command':
        value = f"ssh {server.user}@{server.host}"
        if server.port != 22:
            value += f" -p {server.port}"
    elif field == 'port':
        value = str(server.port)
    else:
        print(error(f"Unknown field: {field}"))
        return

    if copy_to_clipboard(value):
        print(success(f"Copied {field} to clipboard."))
    else:
        print(error("Failed to copy to clipboard."))
        print(info(f"Value: {value}"))


def _handle_export_config(args, manager):
    """Handle export-config command."""
    from ssher.ssh_config_export import export_ssh_config
    output = getattr(args, 'output', None)
    append = getattr(args, 'append', False)
    export_ssh_config(manager.servers, output_path=output, append=append)


def _handle_profile(args, manager):
    """Handle profile subcommands."""
    from ssher.profiles import ProfileManager
    pm = ProfileManager()

    action = getattr(args, 'profile_action', None)

    if action == 'list' or action is None:
        profiles = pm.list_profiles()
        if not profiles:
            print(info("No profiles defined."))
        else:
            print(f"\n{colored('Connection Profiles:', Colors.CYAN, bold=True)}")
            for p in profiles:
                print(f"  \u2022 {colored(p.name, Colors.GREEN)}: "
                      f"timeout={p.connection_timeout}s, keepalive={p.keep_alive}s, "
                      f"x11={p.x11_forward}, reconnect={p.auto_reconnect}")

    elif action == 'add':
        from ssher.models import ConnectionProfile
        profile = ConnectionProfile(
            name=args.profile_name,
            connection_timeout=args.timeout,
            keep_alive=args.keepalive,
            x11_forward=getattr(args, 'x11', False),
            auto_reconnect=getattr(args, 'reconnect', False),
        )
        pm.add_profile(profile)
        print(success(f"Profile '{profile.name}' created."))

    elif action == 'apply':
        ok = pm.apply_to_server(args.profile_name, args.server, manager)
        if ok:
            print(success(f"Profile '{args.profile_name}' applied to server '{args.server}'."))
        else:
            print(error("Failed to apply profile."))

    elif action == 'remove':
        if pm.remove_profile(args.profile_name):
            print(success(f"Profile '{args.profile_name}' removed."))
        else:
            print(error(f"Profile '{args.profile_name}' not found."))


def _handle_alias(args, manager):
    """Handle alias subcommands."""
    action = getattr(args, 'alias_action', None)

    if action == 'add':
        if manager.add_alias(args.alias_name, args.alias_server):
            print(success(f"Alias '{args.alias_name}' -> '{args.alias_server}' created."))
        else:
            print(error(f"Server '{args.alias_server}' not found."))

    elif action == 'remove':
        if manager.remove_alias(args.alias_name):
            print(success(f"Alias '{args.alias_name}' removed."))
        else:
            print(error(f"Alias '{args.alias_name}' not found."))

    elif action == 'list' or action is None:
        aliases = manager.list_aliases()
        if not aliases:
            print(info("No aliases defined."))
        else:
            print(f"\n{colored('Server Aliases:', Colors.CYAN, bold=True)}")
            for alias, server_name in sorted(aliases.items()):
                print(f"  {colored(alias, Colors.GREEN)} -> {server_name}")


def _handle_record(args):
    """Handle record subcommands."""
    from ssher.recording import SessionRecorder
    recorder = SessionRecorder()

    action = getattr(args, 'record_action', None)

    if action == 'list' or action is None:
        recorder.list_recordings()

    elif action == 'replay':
        recorder.replay(args.record_file)


def _handle_import_csv(args, manager):
    """Handle CSV import."""
    from ssher.batch_import import import_csv
    count = import_csv(args.csv_file, manager)
    print(success(f"Imported {count} servers from CSV.") if count else info("No servers imported."))


def _handle_import_json(args, manager):
    """Handle JSON import."""
    from ssher.batch_import import import_json
    count = import_json(args.json_file, manager)
    print(success(f"Imported {count} servers from JSON.") if count else info("No servers imported."))


def _handle_export_csv(manager):
    """Handle CSV export."""
    from ssher.batch_import import export_csv
    path = export_csv(manager)
    print(success(f"Exported to {path}"))


def _handle_export_json(manager):
    """Handle JSON export."""
    from ssher.batch_import import export_json
    path = export_json(manager)
    print(success(f"Exported to {path}"))
