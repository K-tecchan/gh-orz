package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RootDir resolves the root directory for cloning repos.
// Priority: gh-orz.root > ghq.root > ~/gh-orz
func RootDir() (string, error) {
	if root := gitConfigGet("gh-orz.root"); root != "" {
		return absPath(root)
	}
	if root := gitConfigGet("ghq.root"); root != "" {
		return absPath(root)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "gh-orz"), nil
}

// ClonedOwnerCounts scans root/host for owner directories that already have
// cloned repos, returning a map of owner name to number of cloned repos.
// Owners with no cloned repos are omitted. A missing root/host directory
// yields an empty map rather than an error.
func ClonedOwnerCounts(root, host string) map[string]int {
	hostDir := filepath.Join(root, host)
	ownerEntries, err := os.ReadDir(hostDir)
	if err != nil {
		return nil
	}

	counts := make(map[string]int)
	for _, owner := range ownerEntries {
		if !owner.IsDir() {
			continue
		}
		repoEntries, err := os.ReadDir(filepath.Join(hostDir, owner.Name()))
		if err != nil {
			continue
		}
		n := 0
		for _, repo := range repoEntries {
			if repo.IsDir() {
				n++
			}
		}
		if n > 0 {
			counts[owner.Name()] = n
		}
	}
	return counts
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
