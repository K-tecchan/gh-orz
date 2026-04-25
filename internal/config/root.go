package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RootDir resolves the root directory for cloning repos.
// Priority: ghq.root > gh-orz.root > ~/ghq
func RootDir() (string, error) {
	if root := gitConfigGet("ghq.root"); root != "" {
		return absPath(root)
	}
	if root := gitConfigGet("gh-orz.root"); root != "" {
		return absPath(root)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "ghq"), nil
}

func gitConfigGet(key string) string {
	out, err := exec.Command("git", "config", "--global", key).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func absPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}
	return filepath.Abs(path)
}
