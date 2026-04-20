package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/MalasataXD/gh-supercharged/internal/repoctx"
	"github.com/MalasataXD/gh-supercharged/internal/workflows"
	"github.com/spf13/cobra"
)

var plateCmd = &cobra.Command{
	Use:   "plate",
	Short: "Show open issues assigned to you",
	RunE:  runPlate,
}

var plateFull bool

func init() {
	rootCmd.AddCommand(plateCmd)
	plateCmd.Flags().BoolVar(&plateFull, "full", false, "Search across all repos instead of the current one")
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

	repo := flagRepo
	owner := flagOwner
	if !plateFull && repo == "" && owner == "" {
		repo = repoctx.CurrentRepo()
		if repo != "" && flagVerbose {
			fmt.Fprintf(os.Stderr, "scope: %s (auto)\n", repo)
		}
	}

	result, err := workflows.Plate(client, cfg.GithubHandle, workflows.PlateOpts{
		Repo:  repo,
		Owner: owner,
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
