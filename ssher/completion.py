"""
Bash/Zsh tab completion script generator.
Developed by Inioluwa Adeyinka
"""


def generate_completion(shell: str) -> str:
    """Generate shell completion script.

    Args:
        shell: 'bash' or 'zsh'

    Returns:
        Completion script as string.
    """
    if shell == 'bash':
        return _bash_completion()
    elif shell == 'zsh':
        return _zsh_completion()
    return ""


def _bash_completion() -> str:
    return r'''# SSHer bash completion
# Add to ~/.bashrc: eval "$(ssher completion bash)"

_ssher_completions() {
    local cur prev opts subcommands
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    subcommands="list ls l add new ping check history hist import exec run upload up put download down get backup groups wrap vault completion copy export-config profile alias record generate-password genpass import-csv import-json export-csv export-json"

    case "${prev}" in
        ssher)
            COMPREPLY=( $(compgen -W "${subcommands}" -- "${cur}") )
            return 0
            ;;
        vault)
            COMPREPLY=( $(compgen -W "lock unlock change-password status" -- "${cur}") )
            return 0
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh" -- "${cur}") )
            return 0
            ;;
        profile)
            COMPREPLY=( $(compgen -W "list add apply remove" -- "${cur}") )
            return 0
            ;;
        alias)
            COMPREPLY=( $(compgen -W "add remove list" -- "${cur}") )
            return 0
            ;;
        record)
            COMPREPLY=( $(compgen -W "list replay" -- "${cur}") )
            return 0
            ;;
        wrap)
            COMPREPLY=( $(compgen -W "-e -f -d -P" -- "${cur}") )
            return 0
            ;;
        copy)
            COMPREPLY=( $(compgen -W "--field" -- "${cur}") )
            return 0
            ;;
        --field)
            COMPREPLY=( $(compgen -W "host user password command port" -- "${cur}") )
            return 0
            ;;
        export-config)
            COMPREPLY=( $(compgen -W "--append --output" -- "${cur}") )
            return 0
            ;;
        -s|--server)
            # Server names would ideally be dynamically loaded
            return 0
            ;;
    esac

    if [[ "${cur}" == -* ]]; then
        opts="--version --help --reconnect --record"
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    COMPREPLY=( $(compgen -W "${subcommands}" -- "${cur}") )
    return 0
}

complete -F _ssher_completions ssher
'''


def _zsh_completion() -> str:
    return r'''#compdef ssher
# SSHer zsh completion
# Add to ~/.zshrc: eval "$(ssher completion zsh)"

_ssher() {
    local -a commands
    commands=(
        'list:List all servers'
        'add:Add a new server'
        'ping:Check connectivity for all servers'
        'history:View connection history'
        'import:Import from SSH config'
        'exec:Execute command on servers'
        'upload:Upload file to server'
        'download:Download file from server'
        'backup:Backup configurations'
        'groups:List server groups'
        'wrap:sshpass-compatible password wrapper'
        'vault:Vault management'
        'completion:Generate shell completion script'
        'copy:Copy server details to clipboard'
        'export-config:Export servers to SSH config'
        'profile:Manage connection profiles'
        'alias:Manage server aliases'
        'record:Manage session recordings'
        'generate-password:Generate a secure password'
        'import-csv:Import servers from CSV'
        'import-json:Import servers from JSON'
        'export-csv:Export servers to CSV'
        'export-json:Export servers to JSON'
    )

    _arguments \
        '(-v --version)'{-v,--version}'[Show version]' \
        '--reconnect[Auto-reconnect on disconnect]' \
        '--record[Record the SSH session]' \
        '1:command:->command' \
        '*::arg:->args'

    case "$state" in
        command)
            _describe 'command' commands
            ;;
        args)
            case "${words[1]}" in
                vault)
                    _values 'action' lock unlock change-password status
                    ;;
                completion)
                    _values 'shell' bash zsh
                    ;;
                profile)
                    _values 'action' list add apply remove
                    ;;
                alias)
                    _values 'action' add remove list
                    ;;
                record)
                    _values 'action' list replay
                    ;;
                copy)
                    _arguments \
                        '--field[Field to copy]:field:(host user password command port)'
                    ;;
            esac
            ;;
    esac
}

_ssher "$@"
'''
