package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/K-tecchan/gh-orz/internal/config"
	"github.com/spf13/cobra"
)

var fullPathFlag bool

var listCmd = &cobra.Command{
	Use:   "list [owner]",
	Short: "List cloned repositories",
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

		for _, owner := range owners {
			ownerDir := filepath.Join(hostDir, owner)
			repos, err := os.ReadDir(ownerDir)
			if err != nil {
				continue
			}
			for _, repo := range repos {
				if !repo.IsDir() {
					continue
				}
				if fullPathFlag {
					fmt.Println(filepath.Join(ownerDir, repo.Name()))
				} else {
					fmt.Printf("%s/%s/%s\n", hostFlag, owner, repo.Name())
				}
			}
		}

		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&fullPathFlag, "full-path", false, "show full path instead of owner/repo")
	rootCmd.AddCommand(listCmd)
}
