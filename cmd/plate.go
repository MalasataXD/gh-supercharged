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

var plateCmd = &cobra.Command{
	Use:   "plate",
	Short: "Show open issues assigned to you",
	RunE:  runPlate,
}

func init() {
	rootCmd.AddCommand(plateCmd)
}

func runPlate(_ *cobra.Command, _ []string) error {
	cfg, cfgPath, err := config.Load()
	if errors.Is(err, config.ErrFirstRun) {
		fmt.Fprintf(os.Stderr, "First-time setup — created %s\nSet github_handle, then re-run.\n", cfgPath)
		os.Exit(3)
	}
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	client, err := ghclient.New()
	if err != nil {
		return fmt.Errorf("gh client: %w", err)
	}

	result, err := workflows.Plate(client, cfg.GithubHandle, workflows.PlateOpts{
		Repo:  flagRepo,
		Owner: flagOwner,
	})
	if err != nil {
		return err
	}

	if flagJSON {
		return render.JSON(result)
	}
	fmt.Print(render.Plate(result))
	return nil
}
