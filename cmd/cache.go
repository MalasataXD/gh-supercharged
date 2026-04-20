package cmd

import (
	"fmt"

	"github.com/MalasataXD/gh-supercharged/internal/cache"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the projects cache",
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear the projects cache",
	RunE: func(_ *cobra.Command, _ []string) error {
		ch, err := cache.Load()
		if err != nil {
			return err
		}
		ch.Projects = map[string]cache.Entry{}
		if err := ch.Save(); err != nil {
			return err
		}
		fmt.Println("Cache cleared.")
		return nil
	},
}

var cacheShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print the projects cache as JSON",
	RunE: func(_ *cobra.Command, _ []string) error {
		ch, err := cache.Load()
		if err != nil {
			return err
		}
		return render.JSON(ch)
	},
}

func init() {
	cacheCmd.AddCommand(cacheClearCmd, cacheShowCmd)
	rootCmd.AddCommand(cacheCmd)
}
