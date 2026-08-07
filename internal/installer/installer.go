// Package installer handles copying the binary executable into target directories and validating PATH settings.
package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FilepathInPATH returns true if targetDir matches any directory entry in pathEnv.
func FilepathInPATH(targetDir, pathEnv string) bool {
	dirs := filepath.SplitList(pathEnv)
	targetClean := filepath.Clean(targetDir)
	for _, d := range dirs {
		if filepath.Clean(d) == targetClean {
			return true
		}
	}
	return false
}

// Install copies the running binary executable to target directory (default $HOME/.local/bin) and checks PATH.
func Install(args []string, w io.Writer) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	installDir := ""
	for i := 0; i < len(args); i++ {
		if (args[i] == "--dir" || args[i] == "-d") && i+1 < len(args) {
			installDir = args[i+1]
			break
		}
	}

	if installDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get user home dir: %w", err)
		}
		installDir = filepath.Join(home, ".local", "bin")
	}

	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("create install dir: %w", err)
	}

	destPath := filepath.Join(installDir, "tskr")

	srcFile, err := os.Open(exePath)
	if err != nil {
		return fmt.Errorf("open binary for reading: %w", err)
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("create destination binary: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("copy binary: %w", err)
	}

	fmt.Fprintf(w, "tskr installed to %s\n", destPath)

	pathEnv := os.Getenv("PATH")
	if !FilepathInPATH(installDir, pathEnv) {
		fmt.Fprintf(w, "warning: %s is not in your PATH. Add it to your shell profile.\n", installDir)
	} else {
		fmt.Fprintln(w, "tskr is in your PATH and ready to use.")
	}

	return nil
}
