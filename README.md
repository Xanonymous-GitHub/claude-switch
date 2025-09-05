# claude-switch

A modern Go CLI tool to manage multiple Claude Code settings.json configurations and switch between them easily.

## Features

- 🎯 **Add configurations** - Create new configs using your preferred editor
- 📋 **List configurations** - View all saved configs in a beautiful table
- 🔄 **Apply configurations** - Switch to any saved configuration safely
- 🗑️ **Remove configurations** - Delete configs you no longer need
- 💾 **Safe operations** - Automatic backups and atomic file operations
- 🎨 **Beautiful output** - Colored tables and clear status messages
- 🔒 **JSON validation** - Ensures all configurations are valid JSON

## Installation

### Using go install

```bash
go install github.com/username/claude-switch@latest
```

### From source

```bash
git clone https://github.com/username/claude-switch.git
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

### Help

```bash
claude-switch help              # General help
claude-switch add --help        # Command-specific help
```

## Configuration Storage

- **Tool data**: `~/.claude-switch/`
- **Configuration files**: `~/.claude-switch/configs/`
- **Metadata**: `~/.claude-switch/config.json`
- **Target file**: `~/.claude/settings.json`
- **Backups**: `~/.claude/settings.json.backup`

## Requirements

- Go 1.25 or later
- Claude Code installed (with `~/.claude` directory)
- Default editor configured (`$EDITOR` environment variable) or system default available

### Supported Editors

The tool automatically detects available editors:

- **Windows**: VS Code, Notepad++, Notepad
- **macOS**: VS Code, vim, nano, emacs
- **Linux**: VS Code, vim, nano, emacs, gedit

Set your preferred editor:

```bash
export EDITOR=code    # VS Code
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

# Preview removal
claude-switch remove old-config --dry-run
```

## Safety Features

- **Automatic backups** - Current settings are backed up before applying new ones
- **JSON validation** - All configurations are validated before saving/applying
- **Atomic operations** - File operations are atomic to prevent corruption
- **Confirmation prompts** - Important operations require confirmation
- **Rollback support** - Easy rollback instructions provided

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

### v1.0.0

- Initial release
- Add, list, apply, remove commands
- JSON validation and error handling
- Automatic backups and safety features
- Beautiful table output with colors
- Cross-platform editor support
- Comprehensive help and examples