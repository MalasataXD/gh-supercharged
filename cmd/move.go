package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/cache"
	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/projects"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/MalasataXD/gh-supercharged/internal/workflows"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <issue> <status>",
	Short: "Move an issue to a new project status",
	Args:  cobra.ExactArgs(2),
	RunE:  runMove,
}

func init() {
	rootCmd.AddCommand(moveCmd)
}

func runMove(_ *cobra.Command, args []string) error {
	_, cfgPath, err := config.Load()
	if errors.Is(err, config.ErrFirstRun) {
		fmt.Fprintf(os.Stderr, "First-time setup — created %s\nSet github_handle, then re-run.\n", cfgPath)
		os.Exit(3)
	}
	if err != nil {
		return err
	}

	issueNum, err := workflows.ParseIssueArg(args[0])
	if err != nil {
		return err
	}

	parts := strings.SplitN(flagRepo, "/", 2)
	if len(parts) != 2 || parts[0] == "" {
		return fmt.Errorf("--repo owner/repo is required for move")
	}
	owner, repo := parts[0], parts[1]

	client, err := ghclient.New()
	if err != nil {
		return err
	}
	ch, err := cache.Load()
	if err != nil {
		return err
	}
	resolver := projects.NewResolver(client, ch)

	result, err := workflows.Move(client, resolver, workflows.MoveReq{
		Owner:  owner,
		Repo:   repo,
		Issue:  issueNum,
		Status: args[1],
	})
	if err != nil {
		return err
	}

	if flagJSON {
		return render.JSON(result)
	}
	fmt.Print(render.Move(result))
	return nil
}
