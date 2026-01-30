"""
CSV/JSON batch import and export.
Developed by Inioluwa Adeyinka
"""

import csv
import json
from datetime import datetime
from pathlib import Path

from ssher.formatting import error, warning, info
from ssher.models import Server


def import_csv(file_path: str, manager) -> int:
    """Import servers from a CSV file.

    Expected columns: name, host, user, port, auth_type, group, tags, notes
    Password columns are intentionally not supported in plaintext CSV import.

    Returns:
        Number of servers imported.
    """
    path = Path(file_path)
    if not path.exists():
        print(error(f"File not found: {file_path}"))
        return 0

    imported = 0
    try:
        with open(path, 'r', newline='') as f:
            reader = csv.DictReader(f)

            for row in reader:
                name = row.get('name', '').strip()
                host = row.get('host', '').strip()
                user = row.get('user', '').strip()

                if not name or not host:
                    continue

                if manager.get_by_name(name):
                    print(warning(f"Server '{name}' already exists, skipping."))
                    continue

                server = Server(
                    name=name,
                    host=host,
                    user=user or 'root',
                    port=int(row.get('port', 22) or 22),
                    auth_type=row.get('auth_type', 'key').strip() or 'key',
                    group=row.get('group', 'imported').strip() or 'imported',
                    tags=[t.strip() for t in row.get('tags', '').split(',') if t.strip()],
                    notes=row.get('notes', '').strip(),
                    key_path=row.get('key_path', '').strip(),
                )

                manager.add(server)
                imported += 1

    except Exception as e:
        print(error(f"Failed to import CSV: {e}"))

    return imported


def import_json(file_path: str, manager) -> int:
    """Import servers from a JSON file.

    Expects a JSON array of server objects.

    Returns:
        Number of servers imported.
    """
    path = Path(file_path)
    if not path.exists():
        print(error(f"File not found: {file_path}"))
        return 0

    imported = 0
    try:
        data = json.loads(path.read_text())

        if not isinstance(data, list):
            print(error("JSON file must contain an array of server objects."))
            return 0

        for item in data:
            name = item.get('name', '').strip()
            if not name:
                continue

            if manager.get_by_name(name):
                print(warning(f"Server '{name}' already exists, skipping."))
                continue

            server = Server.from_dict(item)
            manager.add(server)
            imported += 1

    except json.JSONDecodeError as e:
        print(error(f"Invalid JSON: {e}"))
    except Exception as e:
        print(error(f"Failed to import JSON: {e}"))

    return imported


def export_csv(manager) -> str:
    """Export servers to a CSV file.

    Returns:
        Path to the exported file.
    """
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    path = Path(f"ssher_export_{timestamp}.csv")

    fieldnames = ['name', 'host', 'user', 'port', 'auth_type', 'group', 'tags',
                  'notes', 'key_path', 'created_at', 'last_connected', 'connection_count',
                  'is_favorite']

    with open(path, 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()

        for server in manager.servers:
            row = {
                'name': server.name,
                'host': server.host,
                'user': server.user,
                'port': server.port,
                'auth_type': server.auth_type,
                'group': server.group,
                'tags': ','.join(server.tags),
                'notes': server.notes,
                'key_path': server.key_path,
                'created_at': server.created_at,
                'last_connected': server.last_connected,
                'connection_count': server.connection_count,
                'is_favorite': server.is_favorite,
            }
            writer.writerow(row)

    return str(path)


def export_json(manager) -> str:
    """Export servers to a JSON file (without passwords for security).

    Returns:
        Path to the exported file.
    """
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    path = Path(f"ssher_export_{timestamp}.json")

    data = []
    for server in manager.servers:
        d = server.to_dict()
        # Remove sensitive fields from plaintext export
        d.pop('password', None)
        data.append(d)

    path.write_text(json.dumps(data, indent=2))
    return str(path)
