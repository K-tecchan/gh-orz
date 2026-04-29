package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/K-tecchan/gh-orz/internal/config"
	gitops "github.com/K-tecchan/gh-orz/internal/git"
	"github.com/K-tecchan/gh-orz/internal/ui"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec <owner> -- <command> [args...]",
	Short: "Run a command in all cloned repositories under an org or user",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		owner := args[0]
		command := args[1:]

		root, err := config.RootDir()
		if err != nil {
			return fmt.Errorf("failed to resolve root directory: %w", err)
		}

		results := gitops.ExecAll(owner, root, hostFlag, command)

		if len(results) == 0 {
			fmt.Printf("No cloned repositories found for %s\n", owner)
			return nil
		}

		var succeeded, failed []gitops.ExecResult
		for _, r := range results {
			if r.Err != nil {
				failed = append(failed, r)
			} else {
				succeeded = append(succeeded, r)
			}
		}

		for _, r := range succeeded {
			fmt.Printf("%s %s\n", ui.Success("✓"), ui.Bold(r.Repo))
			if r.Output != "" {
				for _, line := range strings.Split(r.Output, "\n") {
					fmt.Printf(" %s\n", line)
				}
			}
		}
		for _, r := range failed {
			fmt.Fprintf(os.Stderr, "%s %s\n", ui.Error("✗"), ui.Bold(r.Repo))
			if r.Output != "" {
				for _, line := range strings.Split(r.Output, "\n") {
					fmt.Fprintf(os.Stderr, " %s\n", line)
				}
			}
		}

		if len(failed) > 0 {
			return fmt.Errorf("%d repo(s) failed", len(failed))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
