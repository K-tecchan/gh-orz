package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/K-tecchan/gh-orz/internal/config"
	gitops "github.com/K-tecchan/gh-orz/internal/git"
	gh "github.com/K-tecchan/gh-orz/internal/github"
	"github.com/K-tecchan/gh-orz/internal/ui"
	"github.com/spf13/cobra"
)

var (
	repoFlag            string
	includeArchivedFlag bool
)

var cloneCmd = &cobra.Command{
	Use:   "clone <owner>",
	Short: "Clone repositories under an org or user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := args[0]

		repos, err := gh.ListRepos(owner, hostFlag, includeArchivedFlag)
		if err != nil {
			return fmt.Errorf("failed to list repos for %s: %w", owner, err)
		}

		if len(repos) == 0 {
			fmt.Printf("No repositories found for %s\n", owner)
			return nil
		}

		root, err := config.RootDir()
		if err != nil {
			return fmt.Errorf("failed to resolve root directory: %w", err)
		}

		var selected []string
		if repoFlag != "" {
			selected = strings.Split(repoFlag, ",")
		} else {
			options := make([]ui.RepoOption, len(repos))
			for i, r := range repos {
				targetDir := filepath.Join(root, hostFlag, owner, r.Name)
				_, existsErr := os.Stat(targetDir)
				options[i] = ui.RepoOption{Name: r.Name, Fork: r.Fork, Cloned: existsErr == nil}
			}
			selected, err = ui.SelectRepos(options)
			if err != nil {
				return fmt.Errorf("selection failed: %w", err)
			}
		}

		if len(selected) == 0 {
			fmt.Println("No repositories selected")
			return nil
		}

		var cloned, skipped, failed []gitops.CloneResult
		for _, repo := range selected {
			r := gitops.Clone(owner, repo, root, hostFlag)
			switch r.Status {
			case gitops.CloneSuccess:
				cloned = append(cloned, r)
			case gitops.CloneSkipped:
				skipped = append(skipped, r)
			case gitops.CloneFailed:
				failed = append(failed, r)
			}
		}

		fmt.Println()
		if len(cloned) > 0 {
			fmt.Println(ui.Bold(ui.Successf("Cloned (%d):", len(cloned))))
			for _, r := range cloned {
				fmt.Printf("  %s %s -> %s\n", ui.Success("✓"), r.Repo, r.Path)
			}
		}
		if len(skipped) > 0 {
			fmt.Println(ui.Bold(ui.Warnf("Skipped (%d, already exists):", len(skipped))))
			for _, r := range skipped {
				fmt.Printf("  %s %s\n", ui.Warn("-"), r.Repo)
			}
		}
		if len(failed) > 0 {
			fmt.Fprintln(os.Stderr, ui.Bold(ui.Errorf("Failed (%d):", len(failed))))
			for _, r := range failed {
				fmt.Fprintf(os.Stderr, "  %s %s: %v\n", ui.Error("✗"), r.Repo, r.Err)
			}
			return fmt.Errorf("%d repo(s) failed to clone", len(failed))
		}

		return nil
	},
}

func init() {
	cloneCmd.Flags().StringVar(&repoFlag, "repo", "", "comma-separated list of repos to clone (skips interactive selection)")
	cloneCmd.Flags().BoolVar(&includeArchivedFlag, "include-archived", false, "include archived repositories")
	rootCmd.AddCommand(cloneCmd)
}
