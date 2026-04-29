package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ExecResult holds the result of running a command in a repository.
type ExecResult struct {
	Repo   string
	Output string
	Err    error
}

// ExecAll runs the given command in all repositories under <root>/<host>/<owner>/ in parallel.
func ExecAll(owner, root, host string, command []string) []ExecResult {
	ownerDir := filepath.Join(root, host, owner)

	entries, err := os.ReadDir(ownerDir)
	if err != nil {
		return []ExecResult{{Repo: owner, Err: fmt.Errorf("failed to read directory %s: %w", ownerDir, err)}}
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}

	if len(dirs) == 0 {
		return nil
	}

	results := make([]ExecResult, len(dirs))
	var wg sync.WaitGroup

	for i, dir := range dirs {
		wg.Add(1)
		go func(idx int, repoName string) {
			defer wg.Done()
			repoDir := filepath.Join(ownerDir, repoName)
			cmd := exec.Command(command[0], command[1:]...)
			cmd.Dir = repoDir
			out, err := cmd.CombinedOutput()
			results[idx] = ExecResult{
				Repo:   repoName,
				Output: strings.TrimRight(string(out), "\n"),
				Err:    err,
			}
		}(i, dir)
	}

	wg.Wait()
	return results
}
