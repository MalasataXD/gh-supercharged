package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-supercharged",
	Short: "Supercharged GitHub CLI workflows",
}

var (
	flagJSON    bool
	flagRepo    string
	flagOwner   string
	flagVerbose bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().StringVar(&flagRepo, "repo", "", "Repository (owner/repo)")
	rootCmd.PersistentFlags().StringVar(&flagOwner, "owner", "", "Owner filter")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Verbose error output")
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
