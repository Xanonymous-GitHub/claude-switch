# Claude-Switch Configuration Management Tool

**A modern Go CLI tool for managing multiple Claude Code settings.json configurations with intelligent switching and safety features.**

## Project Context

### Architecture Overview
- **Framework**: Cobra CLI with Go 1.25
- **Structure**: Clean architecture with internal packages
  - `cmd/`: CLI commands (add, apply, list, remove)
  - `internal/config/`: Configuration management
  - `internal/editor/`: Cross-platform editor integration
  - `internal/storage/`: Safe file operations
- **Dependencies**: Modern Go stack with UUID, TableWriter, and color libraries

### Core Features
- ✅ Safe JSON configuration management with validation
- ✅ Cross-platform editor support (VS Code, vim, nano, etc.)
- ✅ Atomic file operations with automatic backups
- ✅ Beautiful CLI output with colored tables
- ✅ Configuration storage in `~/.claude-switch/`

## Development Standards

### Go Best Practices
```go
// Always follow these patterns:
// 1. Proper error handling with context
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// 2. Use Go 1.25 features and idioms
// 3. Follow effective Go guidelines
// 4. Implement comprehensive testing
```

### Code Quality Requirements
- **Formatting**: Always run `go fmt ./...` before commits
- **Linting**: Always run `go vet ./...` to catch issues early
- **Testing**: Maintain >80% code coverage
- **Documentation**: Clear package and function documentation
- **Error Handling**: Comprehensive error wrapping and context

### Git Workflow
```bash
# Standard development workflow
git checkout -b feature/new-feature
# Make changes
go fmt ./...
go vet ./...
go test ./...
git add .
git commit -m "feat: implement new feature"
```

## Command References

### Primary Commands
- `/build` - Build and test the CLI tool
- `/test` - Run comprehensive test suite
- `/fmt` - Format Go code and run quality checks
- `/debug` - Debug CLI functionality and configuration issues
- `/validate` - Validate JSON configuration files

### Development Workflow
```bash
# Development cycle
go mod tidy                    # Clean dependencies
go fmt ./...                   # Format code
go vet ./...                   # Static analysis
go test -v ./...              # Run tests
go build -o claude-switch     # Build binary

# Test new features
./claude-switch validate       # Validate configurations
./claude-switch --help        # Check help output
```

## Configuration Management

### Tool Configuration
- **Data Directory**: `~/.claude-switch/`
- **Configurations**: `~/.claude-switch/configs/`
- **Metadata**: `~/.claude-switch/config.json`
- **Target**: `~/.claude/settings.json`
- **Backups**: `~/.claude/settings.json.backup`

### Editor Integration
Automatically detects and uses available editors with Neovim support:
- **Windows**: VS Code, Neovim, Notepad++, Notepad
- **macOS**: VS Code, Neovim, vim, nano, emacs
- **Linux**: VS Code, Neovim, vim, nano, emacs, gedit

Set preferred editor:
```bash
export EDITOR=code    # VS Code
export EDITOR=nvim    # Neovim
export EDITOR=vim     # Vim
export EDITOR=nano    # Nano
```

## JSON Validation System

### Automatic Validation
**All configuration operations include automatic JSON validation:**
- ✅ **Adding configurations**: Validates JSON before saving
- ✅ **Applying configurations**: Validates JSON before applying to Claude
- ✅ **Manual validation**: Dedicated `validate` command for checking stored configs

### Validation Features
```bash
# Validate all configurations
claude-switch validate

# Validate specific configuration
claude-switch validate my-config

# Verbose validation with details
claude-switch validate --verbose --all
```

### What Gets Validated
- **JSON Syntax**: Proper JSON formatting and structure
- **File Accessibility**: Ensures files are readable
- **Claude Settings Structure**: Basic structure validation for Claude Code settings
- **Error Reporting**: Clear error messages with specific issues

## Quality Assurance

### Mandatory Quality Checks
**ALWAYS run these before any commit or release:**
```bash
# 1. Code formatting (required)
go fmt ./...

# 2. Static analysis (required)
go vet ./...

# 3. Security scanning
go mod audit

# 4. Test execution
go test -race -cover ./...

# 5. Build verification
go build -ldflags="-s -w" -o claude-switch
```

### Testing Strategy
- **Unit Tests**: Each package thoroughly tested
- **Integration Tests**: CLI command workflows
- **Edge Cases**: Error conditions and recovery
- **Cross-Platform**: Windows, macOS, Linux compatibility

## Security Guidelines

### Data Safety
- **Never commit secrets** to version control
- **Validate all JSON** before file operations
- **Use atomic operations** to prevent corruption
- **Implement proper backups** before modifications
- **Sanitize file paths** to prevent directory traversal

### Configuration Security
```go
// Example: Safe configuration handling
func (m *Manager) SaveConfig(config Config) error {
    // 1. Validate JSON structure
    if err := validateJSON(config); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }
    
    // 2. Create backup before modification
    if err := m.createBackup(); err != nil {
        return fmt.Errorf("backup failed: %w", err)
    }
    
    // 3. Atomic write operation
    return m.atomicWrite(config)
}
```

## Advanced Features

### MCP Server Integration
This project leverages modern Claude Code capabilities:

#### Sequential Thinking
- Use `--seq` for complex analysis tasks
- Systematic debugging and problem solving
- Architecture review and optimization

#### Context7 Documentation
- Use `--c7` for Go framework patterns
- Official documentation lookup
- Best practice implementation

#### Serena Memory Management
- Persistent project context across sessions
- Automatic learning from development patterns
- Cross-session knowledge retention

### Auto-Memory Configuration
The project includes intelligent memory management:
```markdown
# Session Lifecycle
1. /sc:load - Initialize project context
2. Development work with persistent memory
3. /sc:save - Checkpoint progress
4. Auto-memory updates for learned patterns
```

## Error Handling & Debugging

### Common Issues
1. **Permission Errors**: Check file permissions for `~/.claude/`
2. **JSON Validation**: Ensure valid JSON in configurations
3. **Editor Not Found**: Set `EDITOR` environment variable
4. **Backup Failures**: Verify disk space and permissions

### Debugging Commands
```bash
# Debug configuration
claude-switch list --detailed

# Dry run operations
claude-switch apply config-name --dry-run

# Verbose error output
claude-switch --verbose add
```

## Project Maintenance

### Regular Tasks
- **Dependencies**: `go mod tidy && go mod audit`
- **Security**: Regular dependency updates
- **Documentation**: Keep README.md synchronized
- **Testing**: Expand test coverage for new features

### Release Process
1. Run full quality check suite
2. Update version in relevant files
3. Create release notes
4. Build cross-platform binaries
5. Verify installation procedures

## Best Practices Summary

### Development Workflow
1. **Think First**: Plan before implementing
2. **Format Always**: `go fmt ./...` before every commit
3. **Analyze Early**: `go vet ./...` catches issues quickly
4. **Test Thoroughly**: Comprehensive coverage including edge cases
5. **Document Clearly**: Code should be self-documenting

### Code Organization
- Keep functions focused and testable
- Use meaningful variable and function names
- Implement proper error handling with context
- Follow Go naming conventions consistently
- Structure packages for clear separation of concerns

### Quality Gates
- All code must pass `go fmt` and `go vet`
- Tests must pass with race detection enabled
- Documentation must be current and accurate
- Security considerations must be addressed
- Cross-platform compatibility must be maintained

---

*This CLAUDE.md file is continuously updated to reflect current best practices and project evolution. Always refer to the latest version for development guidance.*