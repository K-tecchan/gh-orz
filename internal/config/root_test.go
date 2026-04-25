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
