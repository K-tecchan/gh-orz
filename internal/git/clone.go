package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CloneStatus represents the outcome of a clone operation.
type CloneStatus int

const (
	CloneSuccess CloneStatus = iota
	CloneSkipped
	CloneFailed
)

// CloneResult holds the result of a clone operation.
type CloneResult struct {
	Repo   string
	Status CloneStatus
	Path   string
	Err    error
}

// Clone clones a repository into the ghq-compatible directory structure.
func Clone(owner, repo, root, host string) CloneResult {
	targetDir := filepath.Join(root, host, owner, repo)

	if _, err := os.Stat(targetDir); err == nil {
		return CloneResult{Repo: repo, Status: CloneSkipped, Path: targetDir}
	}

	if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
		return CloneResult{Repo: repo, Status: CloneFailed, Err: fmt.Errorf("failed to create directory: %w", err)}
	}

	cloneURL := buildCloneURL(owner, repo, host)
	fmt.Printf("  cloning %s ...\n", repo)

	cmd := exec.Command("git", "clone", cloneURL, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return CloneResult{Repo: repo, Status: CloneFailed, Err: fmt.Errorf("git clone failed: %w", err)}
	}

	return CloneResult{Repo: repo, Status: CloneSuccess, Path: targetDir}
}

// buildCloneURL returns the clone URL based on gh's git_protocol config.
func buildCloneURL(owner, repo, host string) string {
	if gitProtocol() == "ssh" {
		return fmt.Sprintf("git@%s:%s/%s.git", host, owner, repo)
	}
	return fmt.Sprintf("https://%s/%s/%s.git", host, owner, repo)
}

func gitProtocol() string {
	out, err := exec.Command("gh", "config", "get", "git_protocol").Output()
	if err != nil {
		return "https"
	}
	return strings.TrimSpace(string(out))
}
