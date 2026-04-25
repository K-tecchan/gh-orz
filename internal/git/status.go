package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// RepoStatus holds the status of a single repository.
type RepoStatus struct {
	Repo           string
	Branch         string
	Dirty          bool
	UnpushedCount  int
	AheadOfDefault int // commits ahead of default branch (0 if on default)
	Err            error
}

// Status inspects a repository and returns its status.
func Status(repoDir, repoName string) RepoStatus {
	branch := currentBranchName(repoDir)
	dirty := isDirty(repoDir)

	unpushed := countUnpushed(repoDir, branch)

	var aheadOfDefault int
	defaultBranch, err := detectDefaultBranch(repoDir)
	if err == nil && branch != defaultBranch {
		aheadOfDefault = countAhead(repoDir, defaultBranch, branch)
	}

	return RepoStatus{
		Repo:           repoName,
		Branch:         branch,
		Dirty:          dirty,
		UnpushedCount:  unpushed,
		AheadOfDefault: aheadOfDefault,
	}
}

// countUnpushed returns the number of commits not yet pushed to the tracking remote.
func countUnpushed(repoDir, branch string) int {
	remote := fmt.Sprintf("origin/%s", branch)
	out, err := exec.Command("git", "-C", repoDir, "rev-list", "--count", remote+"..HEAD").Output()
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return n
}

// countAhead returns the number of commits branch is ahead of base.
func countAhead(repoDir, base, branch string) int {
	out, err := exec.Command("git", "-C", repoDir, "rev-list", "--count", base+".."+branch).Output()
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return n
}
