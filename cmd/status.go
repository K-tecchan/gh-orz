package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/K-tecchan/gh-orz/internal/config"
	gitops "github.com/K-tecchan/gh-orz/internal/git"
	"github.com/K-tecchan/gh-orz/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [owner]",
	Short: "Show status summary of all cloned repositories",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := config.RootDir()
		if err != nil {
			return fmt.Errorf("failed to resolve root directory: %w", err)
		}

		hostDir := filepath.Join(root, hostFlag)

		var owners []string
		if len(args) > 0 {
			owners = []string{args[0]}
		} else {
			entries, err := os.ReadDir(hostDir)
			if err != nil {
				return fmt.Errorf("no repositories found under %s", hostDir)
			}
			for _, e := range entries {
				if e.IsDir() {
					owners = append(owners, e.Name())
				}
			}
		}

		type repoEntry struct {
			owner string
			name  string
			dir   string
		}

		var repos []repoEntry
		for _, owner := range owners {
			ownerDir := filepath.Join(hostDir, owner)
			entries, err := os.ReadDir(ownerDir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if e.IsDir() {
					repos = append(repos, repoEntry{
						owner: owner,
						name:  e.Name(),
						dir:   filepath.Join(ownerDir, e.Name()),
					})
				}
			}
		}

		if len(repos) == 0 {
			fmt.Println("No cloned repositories found")
			return nil
		}

		// Collect status in parallel
		statuses := make([]gitops.RepoStatus, len(repos))
		var wg sync.WaitGroup
		for i, r := range repos {
			wg.Add(1)
			go func(idx int, entry repoEntry) {
				defer wg.Done()
				statuses[idx] = gitops.Status(entry.dir, entry.owner+"/"+entry.name)
			}(i, r)
		}
		wg.Wait()

		// Display
		var clean, withIssues []gitops.RepoStatus
		for _, s := range statuses {
			if s.Dirty || s.UnpushedCount > 0 || s.AheadOfDefault > 0 {
				withIssues = append(withIssues, s)
			} else {
				clean = append(clean, s)
			}
		}

		if len(withIssues) > 0 {
			fmt.Println(ui.Bold(ui.Warnf("Needs attention (%d):", len(withIssues))))
			for _, s := range withIssues {
				tags := formatTags(s)
				fmt.Printf("  %s %s (%s) %s\n", ui.Warn("!"), s.Repo, s.Branch, tags)
			}
		}
		if len(clean) > 0 {
			fmt.Println(ui.Bold(ui.Successf("Clean (%d):", len(clean))))
			for _, s := range clean {
				fmt.Printf("  %s %s (%s)\n", ui.Success("✓"), s.Repo, s.Branch)
			}
		}

		return nil
	},
}

func formatTags(s gitops.RepoStatus) string {
	var tags []string
	if s.Dirty {
		tags = append(tags, ui.Error("uncommitted changes"))
	}
	if s.UnpushedCount > 0 {
		tags = append(tags, ui.Error(fmt.Sprintf("%d unpushed", s.UnpushedCount)))
	}
	if s.AheadOfDefault > 0 {
		tags = append(tags, ui.Warn(fmt.Sprintf("%d ahead of default", s.AheadOfDefault)))
	}
	return fmt.Sprintf("[%s]", joinTags(tags))
}

func joinTags(tags []string) string {
	result := ""
	for i, t := range tags {
		if i > 0 {
			result += ", "
		}
		result += t
	}
	return result
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
