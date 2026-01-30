"""
CLI entry point.
Developed by Inioluwa Adeyinka
"""

from ssher.cli.commands import cli_main


def main():
    """Main entry point."""
    try:
        cli_main()
    except KeyboardInterrupt:
        from ssher.formatting import colored, Colors
        print(f"\n{colored('Interrupted.', Colors.YELLOW)}")
        import sys
        sys.exit(0)
