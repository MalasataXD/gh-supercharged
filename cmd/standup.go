package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/MalasataXD/gh-supercharged/internal/workflows"
	"github.com/spf13/cobra"
)

var standupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Yesterday's closes + today's open plate",
	RunE:  runStandup,
}

func init() { rootCmd.AddCommand(standupCmd) }

func runStandup(_ *cobra.Command, _ []string) error {
	cfg, cfgPath, err := config.Load()
	if errors.Is(err, config.ErrFirstRun) {
		fmt.Fprintf(os.Stderr, "First-time setup — created %s\nSet github_handle, then re-run.\n", cfgPath)
		os.Exit(3)
	}
	if err != nil {
		return err
	}

	client, err := ghclient.New()
	if err != nil {
		return err
	}

	result, err := workflows.Standup(client, cfg.GithubHandle, cfg.StandupFormat)
	if err != nil {
		return err
	}

	if flagJSON {
		return render.JSON(result)
	}
	fmt.Print(render.Standup(result))
	return nil
}
