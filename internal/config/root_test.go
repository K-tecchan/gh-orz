package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAbsPath_Tilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	got, err := absPath("~/projects")
	if err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(home, "projects")
	if got != want {
		t.Errorf("absPath(~/projects) = %q, want %q", got, want)
	}
}

func TestAbsPath_Absolute(t *testing.T) {
	got, err := absPath("/tmp/test")
	if err != nil {
		t.Fatal(err)
	}

	if got != "/tmp/test" {
		t.Errorf("absPath(/tmp/test) = %q, want /tmp/test", got)
	}
}

func TestAbsPath_Relative(t *testing.T) {
	got, err := absPath("relative/path")
	if err != nil {
		t.Fatal(err)
	}

	if !filepath.IsAbs(got) {
		t.Errorf("absPath(relative/path) = %q, expected absolute path", got)
	}
}

func TestRootDir_Default(t *testing.T) {
	// When neither ghq.root nor gh-orz.root is set, should fall back to ~/ghq.
	// This test works as long as the test runner doesn't have those git config keys set globally.
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	root, err := RootDir()
	if err != nil {
		t.Fatal(err)
	}

	// If gh-orz.root is set in the tester's environment, accept that too
	ghOrzRoot := gitConfigGet("gh-orz.root")
	if ghOrzRoot != "" {
		expected, _ := absPath(ghOrzRoot)
		if root != expected {
			t.Errorf("RootDir() = %q, want %q (from gh-orz.root)", root, expected)
		}
		return
	}

	ghqRoot := gitConfigGet("ghq.root")
	if ghqRoot != "" {
		expected, _ := absPath(ghqRoot)
		if root != expected {
			t.Errorf("RootDir() = %q, want %q (from ghq.root)", root, expected)
		}
		return
	}

	want := filepath.Join(home, "gh-orz")
	if root != want {
		t.Errorf("RootDir() = %q, want %q", root, want)
	}
}

func TestIconsDisabled_Default(t *testing.T) {
	// Mirrors TestRootDir_Default: honors the tester's actual global
	// gh-orz.iconDisabled if set, otherwise expects the default (enabled, i.e. not disabled).
	want := gitConfigGet("gh-orz.iconDisabled") == "true"
	if got := IconsDisabled(); got != want {
		t.Errorf("IconsDisabled() = %v, want %v", got, want)
	}
}

func TestClonedOwnerCounts(t *testing.T) {
	root := t.TempDir()
	host := "github.com"

	mkRepoDir := func(owner, repo string) {
		if err := os.MkdirAll(filepath.Join(root, host, owner, repo), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mkRepoDir("acme-corp", "repo-a")
	mkRepoDir("acme-corp", "repo-b")
	mkRepoDir("solo-user", "dotfiles")

	// A non-directory entry under an owner dir shouldn't be counted as a repo.
	if err := os.WriteFile(filepath.Join(root, host, "acme-corp", "README.md"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	counts := ClonedOwnerCounts(root, host)

	if counts["acme-corp"] != 2 {
		t.Errorf(`counts["acme-corp"] = %d, want 2`, counts["acme-corp"])
	}
	if counts["solo-user"] != 1 {
		t.Errorf(`counts["solo-user"] = %d, want 1`, counts["solo-user"])
	}
	if len(counts) != 2 {
		t.Errorf("got %d owners, want 2: %v", len(counts), counts)
	}
}

func TestClonedOwnerCounts_EmptyOwnerDirOmitted(t *testing.T) {
	root := t.TempDir()
	host := "github.com"

	if err := os.MkdirAll(filepath.Join(root, host, "empty-org"), 0o755); err != nil {
		t.Fatal(err)
	}

	counts := ClonedOwnerCounts(root, host)
	if len(counts) != 0 {
		t.Errorf("got %d owners, want 0: %v", len(counts), counts)
	}
}

func TestClonedOwnerCounts_MissingHostDir(t *testing.T) {
	root := t.TempDir()

	counts := ClonedOwnerCounts(root, "github.com")
	if len(counts) != 0 {
		t.Errorf("got %d owners, want 0: %v", len(counts), counts)
	}
}
