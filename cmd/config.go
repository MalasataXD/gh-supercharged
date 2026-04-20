package cmd

import (
	"fmt"

	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gh-supercharged configuration",
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Print the config directory path",
	RunE: func(_ *cobra.Command, _ []string) error {
		dir, err := config.ConfigDir()
		if err != nil {
			return err
		}
		fmt.Println(dir)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Print current config as JSON",
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, _, err := config.Load()
		if err != nil {
			return err
		}
		return render.JSON(cfg)
	},
}

func init() {
	configCmd.AddCommand(configPathCmd, configShowCmd)
	rootCmd.AddCommand(configCmd)
}
