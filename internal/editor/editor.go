package editor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// OpenEditor opens the specified file in the user's preferred editor
func OpenEditor(filePath string) error {
	editor := getEditor()
	if editor == "" {
		return fmt.Errorf("no editor found. Set $EDITOR environment variable or install a default editor")
	}

	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// getEditor returns the user's preferred editor
func getEditor() string {
	// Check environment variable first
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// Platform-specific defaults
	switch runtime.GOOS {
	case "windows":
		// Try common Windows editors
		editors := []string{"code", "notepad++", "notepad"}
		for _, editor := range editors {
			if _, err := exec.LookPath(editor); err == nil {
				return editor
			}
		}
	case "darwin":
		// Try common macOS editors
		editors := []string{"code", "vim", "nano", "emacs"}
		for _, editor := range editors {
			if _, err := exec.LookPath(editor); err == nil {
				return editor
			}
		}
	default:
		// Try common Linux editors
		editors := []string{"code", "vim", "nano", "emacs", "gedit"}
		for _, editor := range editors {
			if _, err := exec.LookPath(editor); err == nil {
				return editor
			}
		}
	}

	return ""
}

// IsEditorAvailable checks if an editor is available
func IsEditorAvailable() bool {
	return getEditor() != ""
}
