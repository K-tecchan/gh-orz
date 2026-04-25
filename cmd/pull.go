package cmd

import (
	"fmt"
	"os"

	"github.com/K-tecchan/gh-orz/internal/config"
	gitops "github.com/K-tecchan/gh-orz/internal/git"
	"github.com/K-tecchan/gh-orz/internal/ui"
	"github.com/spf13/cobra"
)

var currentBranchFlag bool

var pullCmd = &cobra.Command{
	Use:   "pull <owner>",
	Short: "Pull (update) all cloned repositories under an org or user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := args[0]

		root, err := config.RootDir()
		if err != nil {
			return fmt.Errorf("failed to resolve root directory: %w", err)
		}

		results := gitops.PullAll(owner, root, hostFlag, currentBranchFlag)

		var succeeded, failed []gitops.PullResult
		for _, r := range results {
			if r.Err != nil {
				failed = append(failed, r)
			} else {
				succeeded = append(succeeded, r)
			}
		}

		fmt.Println()
		if len(succeeded) > 0 {
			fmt.Println(ui.Bold(ui.Successf("Updated (%d):", len(succeeded))))
			for _, r := range succeeded {
				fmt.Printf("  %s %s (%s)\n", ui.Success("✓"), r.Repo, ui.Bold(r.Branch))
			}
		}
		if len(failed) > 0 {
			fmt.Fprintln(os.Stderr, ui.Bold(ui.Errorf("Failed (%d):", len(failed))))
			for _, r := range failed {
				fmt.Fprintf(os.Stderr, "  %s %s: %v\n", ui.Error("✗"), r.Repo, r.Err)
			}
			fmt.Fprintf(os.Stderr, "\n%s\n", ui.Warn("Hint: these repos may have diverged from the remote."))
			fmt.Fprintf(os.Stderr, "      Resolve manually with: cd <repo> && git pull --rebase\n")
			return fmt.Errorf("%d repo(s) failed to update", len(failed))
		}

		return nil
	},
}

func init() {
	pullCmd.Flags().BoolVar(&currentBranchFlag, "current-branch", false, "pull the current branch instead of the default branch")
	rootCmd.AddCommand(pullCmd)
}
