# fish completion for tskr
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
