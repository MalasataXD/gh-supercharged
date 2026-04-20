package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/MalasataXD/gh-supercharged/internal/dates"
	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/MalasataXD/gh-supercharged/internal/workflows"
	"github.com/spf13/cobra"
)

var digestCmd = &cobra.Command{
	Use:   "digest [since]",
	Short: "Summarise closed issues and merged PRs",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDigest,
}

func init() { rootCmd.AddCommand(digestCmd) }

func runDigest(_ *cobra.Command, args []string) error {
	cfg, cfgPath, err := config.Load()
	if errors.Is(err, config.ErrFirstRun) {
		fmt.Fprintf(os.Stderr, "First-time setup — created %s\nSet github_handle, then re-run.\n", cfgPath)
		os.Exit(3)
	}
	if err != nil {
		return err
	}

	since := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, -cfg.DigestWindowDays)
	if len(args) == 1 {
		since, err = dates.Parse(args[0])
		if err != nil {
			return err
		}
	}

	client, err := ghclient.New()
	if err != nil {
		return err
	}

	result, err := workflows.Digest(client, cfg.GithubHandle, since, workflows.DigestOpts{Owner: flagOwner})
	if err != nil {
		return err
	}

	if flagJSON {
		return render.JSON(result)
	}
	fmt.Print(render.Digest(result))
	return nil
}
