package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	Use:   "clone [owner]",
	Short: "Clone repositories under an org or user",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := config.RootDir()
		if err != nil {
			return fmt.Errorf("failed to resolve root directory: %w", err)
		}

		var owner string
		if len(args) > 0 {
			owner = args[0]
		} else {
			selectedOwner, err := selectOwner(hostFlag, root)
			if err != nil {
				return err
			}
			if selectedOwner == "" {
				fmt.Println("No org or user selected")
				return nil
			}
			owner = selectedOwner
		}

		repos, err := gh.ListRepos(owner, hostFlag, includeArchivedFlag)
		if err != nil {
			return fmt.Errorf("failed to list repos for %s: %w", owner, err)
		}

		if len(repos) == 0 {
			fmt.Printf("No repositories found for %s\n", owner)
			return nil
		}

		var selected []string
		if repoFlag != "" {
			selected = strings.Split(repoFlag, ",")
		} else {
			options := make([]ui.RepoOption, len(repos))
			for i, r := range repos {
				targetDir := filepath.Join(root, hostFlag, owner, r.Name)
				_, existsErr := os.Stat(targetDir)
				options[i] = ui.RepoOption{Name: r.Name, Fork: r.Fork, Private: r.Private, Cloned: existsErr == nil}
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

// selectOwner prompts the user to pick an org or their own account via a
// fuzzy finder. Orgs the user belongs to are tagged with an org icon, and
// the user's own account is always offered too, tagged with a user icon.
// Owners with repos already cloned under root/host are also included as
// options and tagged with how many repos are cloned locally.
func selectOwner(host, root string) (string, error) {
	orgs, err := gh.ListUserOrgs(host)
	if err != nil {
		return "", fmt.Errorf("failed to list organizations: %w", err)
	}

	user, err := gh.CurrentUser(host)
	if err != nil {
		return "", fmt.Errorf("failed to get authenticated user: %w", err)
	}

	clonedCounts := config.ClonedOwnerCounts(root, host)

	seen := make(map[string]bool, len(orgs)+1)
	options := make([]ui.OwnerOption, 0, len(orgs)+1)
	for _, o := range orgs {
		seen[o] = true
		options = append(options, ui.OwnerOption{Name: o, ClonedCount: clonedCounts[o], IsOrg: true})
	}

	seen[user] = true
	options = append(options, ui.OwnerOption{Name: user, ClonedCount: clonedCounts[user], IsUser: true})

	var extra []string
	for name := range clonedCounts {
		if !seen[name] {
			extra = append(extra, name)
		}
	}
	sort.Strings(extra)
	for _, name := range extra {
		options = append(options, ui.OwnerOption{Name: name, ClonedCount: clonedCounts[name]})
	}

	return ui.SelectOwner(options)
}
