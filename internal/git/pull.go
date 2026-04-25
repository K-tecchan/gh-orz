package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// PullResult holds the result of a git pull operation.
type PullResult struct {
	Repo    string
	Branch  string
	Skipped bool
	Err     error
}

// PullAll runs git pull on all repositories under <root>/<host>/<owner>/ in parallel.
// If currentBranch is false, it checks out the default branch before pulling.
func PullAll(owner, root, host string, currentBranch bool) []PullResult {
	ownerDir := filepath.Join(root, host, owner)

	entries, err := os.ReadDir(ownerDir)
	if err != nil {
		return []PullResult{{Repo: owner, Err: fmt.Errorf("failed to read directory %s: %w", ownerDir, err)}}
	}

	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}

	if len(dirs) == 0 {
		fmt.Printf("No cloned repositories found for %s\n", owner)
		return nil
	}

	results := make([]PullResult, len(dirs))
	var wg sync.WaitGroup

	for i, dir := range dirs {
		wg.Add(1)
		go func(idx int, repoName string) {
			defer wg.Done()
			repoDir := filepath.Join(ownerDir, repoName)
			results[idx] = pullRepo(repoDir, repoName, currentBranch)
		}(i, dir)
	}

	wg.Wait()
	return results
}

func pullRepo(repoDir, repoName string, currentBranch bool) PullResult {
	if isDirty(repoDir) {
		branch := currentBranchName(repoDir)
		return PullResult{Repo: repoName, Branch: branch, Skipped: true}
	}

	if !currentBranch {
		defaultBranch, err := detectDefaultBranch(repoDir)
		if err != nil {
			return PullResult{Repo: repoName, Err: fmt.Errorf("failed to detect default branch: %w", err)}
		}

		if out, err := exec.Command("git", "-C", repoDir, "checkout", defaultBranch).CombinedOutput(); err != nil {
			return PullResult{Repo: repoName, Err: fmt.Errorf("failed to checkout %s: %v: %s", defaultBranch, err, string(out))}
		}
	}

	if out, err := exec.Command("git", "-C", repoDir, "pull", "--ff-only").CombinedOutput(); err != nil {
		return PullResult{Repo: repoName, Err: fmt.Errorf("%v: %s", err, string(out))}
	}

	branch := currentBranchName(repoDir)
	return PullResult{Repo: repoName, Branch: branch}
}

// isDirty returns true if the working tree has uncommitted changes.
func isDirty(repoDir string) bool {
	out, err := exec.Command("git", "-C", repoDir, "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}

func currentBranchName(repoDir string) string {
	out, err := exec.Command("git", "-C", repoDir, "branch", "--show-current").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// detectDefaultBranch returns the default branch name by checking origin/HEAD.
func detectDefaultBranch(repoDir string) (string, error) {
	out, err := exec.Command("git", "-C", repoDir, "symbolic-ref", "refs/remotes/origin/HEAD").Output()
	if err != nil {
		// origin/HEAD may not be set; try to set it
		if _, setErr := exec.Command("git", "-C", repoDir, "remote", "set-head", "origin", "--auto").CombinedOutput(); setErr != nil {
			return "", fmt.Errorf("could not determine default branch: %w", err)
		}
		out, err = exec.Command("git", "-C", repoDir, "symbolic-ref", "refs/remotes/origin/HEAD").Output()
		if err != nil {
			return "", fmt.Errorf("could not determine default branch: %w", err)
		}
	}

	// output is like "refs/remotes/origin/main"
	ref := strings.TrimSpace(string(out))
	parts := strings.SplitN(ref, "refs/remotes/origin/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected ref format: %s", ref)
	}
	return parts[1], nil
}
