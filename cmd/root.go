package cmd

import (
	"github.com/cli/go-gh/v2/pkg/auth"
	"github.com/spf13/cobra"
)

var hostFlag string

var rootCmd = &cobra.Command{
	Use:   "gh-orz",
	Short: "Bulk clone and update GitHub repositories",
	Long:  "gh-orz is a GitHub CLI extension to bulk clone and update repositories under an org or user, organized in a ghq-compatible directory structure.",
}

func init() {
	defaultHost, _ := auth.DefaultHost()
	rootCmd.PersistentFlags().StringVar(&hostFlag, "host", defaultHost, "GitHub host (e.g. github.example.com for GHE)")
}

func Execute() error {
	return rootCmd.Execute()
}
