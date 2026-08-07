# bash completion for tskr
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
