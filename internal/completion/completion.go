// Package completion provides shell autocompletion script generation for Bash, Zsh, and Fish.
package completion

import (
	"fmt"
	"io"
	"strings"
)

const bashCompletionScript = `# bash completion for tskr
_tskr_completion() {
    local cur prev subcommands
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    subcommands="init config install start switch sync list go completion help"

    if [ "$COMP_CWORD" -eq 1 ]; then
        COMPREPLY=( $(compgen -W "$subcommands" -- "$cur") )
        return 0
    fi

    case "${COMP_WORDS[1]}" in
        config)
            if [ "$COMP_CWORD" -eq 2 ]; then
                COMPREPLY=( $(compgen -W "show set convert" -- "$cur") )
            elif [ "${COMP_WORDS[2]}" = "set" ] && [ "$COMP_CWORD" -eq 3 ]; then
                COMPREPLY=( $(compgen -W "storage_dir files.naming" -- "$cur") )
            elif [ "${COMP_WORDS[2]}" = "convert" ] && [ "$COMP_CWORD" -eq 3 ]; then
                COMPREPLY=( $(compgen -W "json yaml toml" -- "$cur") )
            fi
            ;;
        install)
            if [ "$COMP_CWORD" -eq 2 ]; then
                COMPREPLY=( $(compgen -W "--dir -d" -- "$cur") )
            fi
            ;;
        completion)
            if [ "$COMP_CWORD" -eq 2 ]; then
                COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
            fi
            ;;
    esac
}

complete -F _tskr_completion tskr
`

const zshCompletionScript = `#compdef tskr

_tskr() {
    local -a commands
    commands=(
        'init:Initialize configuration and storage directory'
        'config:Show, set, or convert configuration'
        'install:Install tskr binary'
        'start:Start task tracking in current repository'
        'switch:Switch symlinks to current branch task context'
        'sync:Synchronize symlinks for active branch'
        'list:List all tracked task references'
        'go:Print command to navigate to repository and branch'
        'completion:Generate shell autocompletion script'
        'help:Show help information'
    )

    _arguments -C \
        '1: :->command' \
        '*:: :->args'

    case $state in
        command)
            _describe -t commands 'tskr command' commands
            ;;
        args)
            case $words[1] in
                config)
                    local -a config_cmds
                    config_cmds=(
                        'show:Show active configuration'
                        'set:Set a configuration value'
                        'convert:Convert configuration format'
                    )
                    _arguments '1: :->config_cmd' '*:: :->config_args'
                    case $state in
                        config_cmd)
                            _describe -t config_cmds 'config command' config_cmds
                            ;;
                        config_args)
                            case $words[1] in
                                set)
                                    local -a keys
                                    keys=('storage_dir:Storage directory path' 'files.naming:Comma-separated task file names')
                                    _describe -t keys 'configuration key' keys
                                    ;;
                                convert)
                                    local -a fmts
                                    fmts=('json:JSON format' 'yaml:YAML format' 'toml:TOML format')
                                    _describe -t fmts 'target format' fmts
                                    ;;
                            esac
                            ;;
                    esac
                    ;;
                install)
                    _arguments '--dir[Custom installation directory]:directory:_files -/' '-d[Custom installation directory]:directory:_files -/'
                    ;;
                completion)
                    local -a shells
                    shells=('bash:Bash autocompletion' 'zsh:Zsh autocompletion' 'fish:Fish autocompletion')
                    _describe -t shells 'shell' shells
                    ;;
            esac
            ;;
    esac
}

_tskr "$@"
`

const fishCompletionScript = `# fish completion for tskr
complete -c tskr -f

# Main subcommands
complete -c tskr -n "__fish_use_subcommand" -a init -d "Initialize configuration and storage directory"
complete -c tskr -n "__fish_use_subcommand" -a config -d "Show, set, or convert configuration"
complete -c tskr -n "__fish_use_subcommand" -a install -d "Install tskr binary"
complete -c tskr -n "__fish_use_subcommand" -a start -d "Start task tracking in current repository"
complete -c tskr -n "__fish_use_subcommand" -a switch -d "Switch symlinks to current branch task context"
complete -c tskr -n "__fish_use_subcommand" -a sync -d "Synchronize symlinks for active branch"
complete -c tskr -n "__fish_use_subcommand" -a list -d "List all tracked task references"
complete -c tskr -n "__fish_use_subcommand" -a go -d "Print navigation command string"
complete -c tskr -n "__fish_use_subcommand" -a completion -d "Generate shell completion script"

# config subcommands
complete -c tskr -n "__fish_seen_subcommand_from config" -a show -d "Show active configuration"
complete -c tskr -n "__fish_seen_subcommand_from config" -a set -d "Set a configuration key"
complete -c tskr -n "__fish_seen_subcommand_from config" -a convert -d "Convert configuration format"

# config set keys
complete -c tskr -n "__fish_seen_subcommand_from config; and __fish_seen_subcommand_from set" -a storage_dir -d "Storage directory path"
complete -c tskr -n "__fish_seen_subcommand_from config; and __fish_seen_subcommand_from set" -a files.naming -d "Comma-separated task file names"

# config convert formats
complete -c tskr -n "__fish_seen_subcommand_from config; and __fish_seen_subcommand_from convert" -a "json yaml toml"

# install flags
complete -c tskr -n "__fish_seen_subcommand_from install" -l dir -r -d "Custom installation directory"
complete -c tskr -n "__fish_seen_subcommand_from install" -s d -r -d "Custom installation directory"

# completion arguments
complete -c tskr -n "__fish_seen_subcommand_from completion" -a "bash zsh fish"
`

// Generate outputs the autocompletion script for bash, zsh, or fish.
func Generate(args []string, w io.Writer) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: tskr completion <bash|zsh|fish>")
	}

	sh := strings.ToLower(args[0])
	switch sh {
	case "bash":
		fmt.Fprint(w, bashCompletionScript)
	case "zsh":
		fmt.Fprint(w, zshCompletionScript)
	case "fish":
		fmt.Fprint(w, fishCompletionScript)
	default:
		return fmt.Errorf("unsupported shell %q (must be bash, zsh, or fish)", args[0])
	}
	return nil
}
