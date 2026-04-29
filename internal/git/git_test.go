package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const testHost = "github.com"

func initRepoWithCommit(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	cmds := [][]string{
		{"git", "init", dir},
		{"git", "-C", dir, "config", "user.email", "test@test.com"},
		{"git", "-C", dir, "config", "user.name", "test"},
		{"git", "-C", dir, "commit", "--allow-empty", "-m", "init"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %v\n%s", args, err, out)
		}
	}
}

func cloneRepo(t *testing.T, bare, dest string) {
	t.Helper()
	cmd := exec.Command("git", "clone", bare, dest)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone failed: %v\n%s", err, out)
	}
}

func TestClone_SkipsExisting(t *testing.T) {
	root := t.TempDir()
	owner := "test-owner"
	repo := "test-repo"

	// Create existing directory
	targetDir := filepath.Join(root, testHost, owner, repo)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Should skip without error
	result := Clone(owner, repo, root, testHost)
	if result.Status != CloneSkipped {
		t.Errorf("Clone() status = %v, want CloneSkipped", result.Status)
	}
}

func TestPullAll_EmptyOwnerDir(t *testing.T) {
	root := t.TempDir()
	owner := "empty-owner"

	ownerDir := filepath.Join(root, testHost, owner)
	if err := os.MkdirAll(ownerDir, 0755); err != nil {
		t.Fatal(err)
	}

	results := PullAll(owner, root, testHost, true)
	if results != nil {
		t.Errorf("PullAll() = %v, want nil for empty dir", results)
	}
}

func TestPullAll_NonExistentDir(t *testing.T) {
	root := t.TempDir()

	results := PullAll("nonexistent", root, testHost, true)
	if len(results) != 1 {
		t.Fatalf("PullAll() returned %d results, want 1", len(results))
	}
	if results[0].Err == nil {
		t.Error("PullAll() error = nil, want error for nonexistent dir")
	}
}

func TestPullAll_WithRepos(t *testing.T) {
	root := t.TempDir()
	owner := "test-owner"
	ownerDir := filepath.Join(root, testHost, owner)

	// Create an upstream repo with a commit, then clone it
	upstreamDir := filepath.Join(t.TempDir(), "upstream")
	initRepoWithCommit(t, upstreamDir)

	repoDir := filepath.Join(ownerDir, "my-repo")
	cloneRepo(t, upstreamDir, repoDir)

	results := PullAll(owner, root, testHost, true)
	if len(results) != 1 {
		t.Fatalf("PullAll() returned %d results, want 1", len(results))
	}
	if results[0].Repo != "my-repo" {
		t.Errorf("result repo = %q, want %q", results[0].Repo, "my-repo")
	}
	if results[0].Err != nil {
		t.Errorf("PullAll() error = %v, want nil", results[0].Err)
	}
}

func TestPullAll_DefaultBranch(t *testing.T) {
	root := t.TempDir()
	owner := "test-owner"
	ownerDir := filepath.Join(root, testHost, owner)

	// Create upstream with a commit
	upstreamDir := filepath.Join(t.TempDir(), "upstream")
	initRepoWithCommit(t, upstreamDir)

	// Clone it
	repoDir := filepath.Join(ownerDir, "my-repo")
	cloneRepo(t, upstreamDir, repoDir)

	// Create and checkout a feature branch
	runGit(t, repoDir, "checkout", "-b", "feature-branch")

	// Pull with currentBranch=false should checkout default branch
	results := PullAll(owner, root, testHost, false)
	if len(results) != 1 {
		t.Fatalf("PullAll() returned %d results, want 1", len(results))
	}
	if results[0].Err != nil {
		t.Fatalf("PullAll() error = %v, want nil", results[0].Err)
	}

	// Verify we're back on the default branch
	out, err := exec.Command("git", "-C", repoDir, "branch", "--show-current").Output()
	if err != nil {
		t.Fatal(err)
	}
	branch := strings.TrimSpace(string(out))
	if branch == "feature-branch" {
		t.Error("expected checkout to default branch, but still on feature-branch")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	fullArgs := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", fullArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

func TestExecAll_RunsCommandInAllRepos(t *testing.T) {
	root := t.TempDir()
	owner := "test-owner"
	ownerDir := filepath.Join(root, testHost, owner)

	// Create two repos
	initRepoWithCommit(t, filepath.Join(ownerDir, "repo-a"))
	initRepoWithCommit(t, filepath.Join(ownerDir, "repo-b"))

	results := ExecAll(owner, root, testHost, []string{"git", "rev-parse", "--is-inside-work-tree"})
	if len(results) != 2 {
		t.Fatalf("ExecAll() returned %d results, want 2", len(results))
	}
	for _, r := range results {
		if r.Err != nil {
			t.Errorf("ExecAll() repo %s error = %v", r.Repo, r.Err)
		}
		if strings.TrimSpace(r.Output) != "true" {
			t.Errorf("ExecAll() repo %s output = %q, want \"true\"", r.Repo, r.Output)
		}
	}
}

func TestExecAll_NonExistentDir(t *testing.T) {
	root := t.TempDir()

	results := ExecAll("nonexistent", root, testHost, []string{"echo", "hello"})
	if len(results) != 1 {
		t.Fatalf("ExecAll() returned %d results, want 1", len(results))
	}
	if results[0].Err == nil {
		t.Error("ExecAll() error = nil, want error for nonexistent dir")
	}
}

func TestExecAll_EmptyOwnerDir(t *testing.T) {
	root := t.TempDir()
	owner := "empty-owner"
	os.MkdirAll(filepath.Join(root, testHost, owner), 0755)

	results := ExecAll(owner, root, testHost, []string{"echo", "hello"})
	if results != nil {
		t.Errorf("ExecAll() = %v, want nil for empty dir", results)
	}
}

func TestExecAll_CommandFailure(t *testing.T) {
	root := t.TempDir()
	owner := "test-owner"
	initRepoWithCommit(t, filepath.Join(root, testHost, owner, "repo-a"))

	results := ExecAll(owner, root, testHost, []string{"false"})
	if len(results) != 1 {
		t.Fatalf("ExecAll() returned %d results, want 1", len(results))
	}
	if results[0].Err == nil {
		t.Error("ExecAll() error = nil, want error for failed command")
	}
}

func TestBuildCloneURL_SSH(t *testing.T) {
	url := buildCloneURL("owner", "repo", "ghe.example.com")
	// gitProtocol() may return "https" or "ssh" depending on environment,
	// but the host should always be ghe.example.com
	if url != "git@ghe.example.com:owner/repo.git" && url != "https://ghe.example.com/owner/repo.git" {
		t.Errorf("buildCloneURL() = %q, want URL with host ghe.example.com", url)
	}
}

func TestBuildCloneURL_HTTPS(t *testing.T) {
	url := buildCloneURL("myorg", "myrepo", "github.com")
	if url != "git@github.com:myorg/myrepo.git" && url != "https://github.com/myorg/myrepo.git" {
		t.Errorf("buildCloneURL() = %q, unexpected URL", url)
	}
}
