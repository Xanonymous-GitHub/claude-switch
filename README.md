# claude-switch

A modern Go CLI tool to manage multiple Claude Code settings.json configurations and switch between them easily.

## Features

- 🎯 **Add configurations** - Create new configs using your preferred editor (supports Neovim)
- 📋 **List configurations** - View all saved configs in a beautiful table
- 🔄 **Apply configurations** - Switch to any saved configuration safely with JSON validation
- 🔁 **Sync configurations** - Save live settings changes back to stored configs
- 🗑️ **Remove configurations** - Delete configs you no longer need
- ✅ **Validate configurations** - Check JSON syntax and structure before applying
- 💾 **Safe operations** - Automatic backups and atomic file operations
- 🎨 **Beautiful output** - Colored tables and clear status messages
- 🔒 **JSON validation** - Ensures all configurations are valid JSON
- ⚠️ **Conflict detection** - Detects and resolves sync conflicts safely

## Installation

### Using go install

- Please [install Go](https://go.dev/doc/install) first.

```bash
go install github.com/Xanonymous-GitHub/claude-switch@latest
export PATH="$PATH:$(go env GOPATH)/bin" # Add this to your .zshrc or .bashrc or others
```

### From source

```bash
git clone git@github.com:Xanonymous-GitHub/claude-switch.git
cd claude-switch
go build -o claude-switch
```

## Usage

### Add a new configuration

```bash
claude-switch add
```

This will:
1. Copy your current `~/.claude/settings.json` to a temporary file (or create a default if none exists)
2. Open the file in your default editor (`$EDITOR` or system default)
3. After saving and closing the editor, prompt for a name and description
4. Save the configuration for future use

### List all configurations

```bash
claude-switch list
```

View configurations in different formats:

```bash
claude-switch list --detailed    # Show full IDs and descriptions
claude-switch list --json        # Output in JSON format
```

### Apply a configuration

```bash
claude-switch apply my-config-name
```

Switch to a configuration safely:

```bash
claude-switch apply my-config --confirm  # Prompt for confirmation
claude-switch apply my-config --dry-run  # Preview changes only
```

### Remove a configuration

```bash
claude-switch remove my-old-config
```

Remove with options:

```bash
claude-switch remove my-config --force    # Skip confirmation
claude-switch remove my-config --dry-run  # Preview what would be removed
```

### Validate configurations

```bash
claude-switch validate                    # Validate all configurations
claude-switch validate my-config         # Validate specific configuration
claude-switch validate --verbose --all   # Detailed validation output
```

### Sync changes back to stored config

Save changes from your live `~/.claude/settings.json` back to the stored configuration:

```bash
claude-switch sync                        # Sync current configuration
claude-switch sync my-config              # Sync specific configuration
claude-switch sync --dry-run              # Preview changes without saving
claude-switch sync --force                # Skip confirmation prompt
```

The sync command will:
1. Compare live settings.json with the stored configuration
2. Display detected changes in a diff format
3. Check for conflicts (if stored config was modified externally)
4. Prompt for confirmation before saving

For detailed sync documentation, see [docs/SYNC_GUIDE.md](docs/SYNC_GUIDE.md).

### Auto-sync when switching configurations

Automatically sync your current config before switching to another:

```bash
claude-switch apply new-config --auto-sync
```

### Check sync status

View current configuration and detect unsaved changes:

```bash
claude-switch status                      # Show current config and sync status
claude-switch status --diff               # Show diff of unsaved changes
```

### Help

```bash
claude-switch help              # General help
claude-switch add --help        # Command-specific help
```

## Configuration Storage

- **Tool data**: `~/.claude-switch/`
- **Configuration files**: `~/.claude-switch/configs/`
- **Metadata**: `~/.claude-switch/config.json`
- **State tracking**: `~/.claude-switch/state.json`
- **Target file**: `~/.claude/settings.json`
- **Backups**: `~/.claude/settings.json.backup`

## Requirements

- Go 1.25 or later
- Claude Code installed (with `~/.claude` directory)
- Default editor configured (`$EDITOR` environment variable) or system default available

### Supported Editors

The tool automatically detects available editors with Neovim support:

- **Windows**: VS Code, Neovim, Notepad++, Notepad
- **macOS**: VS Code, Neovim, vim, nano, emacs
- **Linux**: VS Code, Neovim, vim, nano, emacs, gedit

Set your preferred editor:

```bash
export EDITOR=code    # VS Code
export EDITOR=nvim    # Neovim
export EDITOR=vim     # Vim
export EDITOR=nano    # Nano
```

## Examples

### Basic workflow

```bash
# Create a new configuration
claude-switch add

# List all configurations
claude-switch list

# Apply a configuration
claude-switch apply work-setup

# Sync changes back after editing settings
claude-switch sync

# Validate configurations
claude-switch validate

# Remove an old configuration
claude-switch remove old-config
```

### Advanced usage

```bash
# Add with predefined name and description
claude-switch add --name "work-setup" --description "My work environment settings"

# View detailed information
claude-switch list --detailed

# Apply with confirmation
claude-switch apply work-setup --confirm

# Auto-sync before switching configs
claude-switch apply new-config --auto-sync

# Sync with verbose diff output
claude-switch sync --verbose

# Preview sync changes without saving
claude-switch sync --dry-run

# Validate with verbose output
claude-switch validate --verbose --all

# Preview removal
claude-switch remove old-config --dry-run
```

## Safety Features

- **Automatic backups** - Current settings are backed up before applying new ones
- **JSON validation** - All configurations are validated before saving/applying
- **Atomic operations** - File operations are atomic to prevent corruption
- **Confirmation prompts** - Important operations require confirmation
- **Rollback support** - Easy rollback instructions provided
- **Conflict detection** - Detects when stored configs are modified externally
- **Sync safety** - Preview changes with `--dry-run` before syncing

## Error Handling

The tool includes comprehensive error handling for:

- Missing Claude Code installation
- Invalid JSON configurations
- File permission issues
- Missing editor configuration
- Network and filesystem errors

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Changelog

### v1.1.0

- New sync command to save live settings back to stored configs
- Diff detection with colored output
- Conflict detection and resolution
- Auto-sync flag for apply command (`--auto-sync`)
- State tracking for current configuration
- Comprehensive sync documentation

### v1.0.0

- Initial release
- Add, list, apply, remove commands
- JSON validation and error handling
- Automatic backups and safety features
- Beautiful table output with colors
- Cross-platform editor support
- Comprehensive help and examples