package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/K-tecchan/gh-orz/internal/config"
	"github.com/K-tecchan/gh-orz/internal/ui"
	"github.com/spf13/cobra"
)

var rmRepoFlag string

var rmCmd = &cobra.Command{
	Use:   "rm <owner>",
	Short: "Remove cloned repositories under an org or user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := args[0]

		root, err := config.RootDir()
		if err != nil {
			return fmt.Errorf("failed to resolve root directory: %w", err)
		}

		ownerDir := filepath.Join(root, hostFlag, owner)
		entries, err := os.ReadDir(ownerDir)
		if err != nil {
			return fmt.Errorf("no cloned repositories found for %s", owner)
		}

		var repos []string
		for _, e := range entries {
			if e.IsDir() {
				repos = append(repos, e.Name())
			}
		}

		if len(repos) == 0 {
			fmt.Printf("No cloned repositories found for %s\n", owner)
			return nil
		}

		var selected []string
		if rmRepoFlag != "" {
			selected = strings.Split(rmRepoFlag, ",")
		} else {
			selected, err = ui.SelectItems("Select repositories to remove:", repos)
			if err != nil {
				return fmt.Errorf("selection failed: %w", err)
			}
		}

		if len(selected) == 0 {
			fmt.Println("No repositories selected")
			return nil
		}

		var removed, failed []string
		for _, repo := range selected {
			repoDir := filepath.Join(ownerDir, repo)
			if _, err := os.Stat(repoDir); os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "  %s %s: not found\n", ui.Warn("-"), repo)
				continue
			}
			if err := os.RemoveAll(repoDir); err != nil {
				fmt.Fprintf(os.Stderr, "  %s %s: %v\n", ui.Error("✗"), repo, err)
				failed = append(failed, repo)
			} else {
				removed = append(removed, repo)
			}
		}

		// Clean up owner directory if empty
		remaining, err := os.ReadDir(ownerDir)
		if err == nil {
			hasDir := false
			for _, e := range remaining {
				if e.IsDir() {
					hasDir = true
					break
				}
			}
			if !hasDir {
				os.Remove(ownerDir)
			}
		}

		fmt.Println()
		if len(removed) > 0 {
			fmt.Println(ui.Bold(ui.Successf("Removed (%d):", len(removed))))
			for _, repo := range removed {
				fmt.Printf("  %s %s\n", ui.Success("✓"), repo)
			}
		}
		if len(failed) > 0 {
			fmt.Fprintln(os.Stderr, ui.Bold(ui.Errorf("Failed (%d):", len(failed))))
			for _, repo := range failed {
				fmt.Fprintf(os.Stderr, "  %s %s\n", ui.Error("✗"), repo)
			}
			return fmt.Errorf("%d repo(s) failed to remove", len(failed))
		}

		return nil
	},
}

func init() {
	rmCmd.Flags().StringVar(&rmRepoFlag, "repo", "", "comma-separated list of repos to remove (skips interactive selection)")
	rootCmd.AddCommand(rmCmd)
}
