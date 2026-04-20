package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/MalasataXD/gh-supercharged/internal/workflows"
	"github.com/spf13/cobra"
)

var newIssueCmd = &cobra.Command{
	Use:   "new-issue <description>",
	Short: "Draft and create a well-formed GitHub issue",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runNewIssue,
}

var newIssueConfirm bool

func init() {
	newIssueCmd.Flags().BoolVar(&newIssueConfirm, "confirm", false, "Create without approval prompt (use after reviewing --json draft)")
	rootCmd.AddCommand(newIssueCmd)
}

func runNewIssue(_ *cobra.Command, args []string) error {
	_, cfgPath, err := config.Load()
	if errors.Is(err, config.ErrFirstRun) {
		fmt.Fprintf(os.Stderr, "First-time setup — created %s\nSet github_handle, then re-run.\n", cfgPath)
		os.Exit(3)
	}
	if err != nil {
		return err
	}

	parts := strings.SplitN(flagRepo, "/", 2)
	if len(parts) != 2 || parts[0] == "" {
		return fmt.Errorf("--repo owner/repo is required for new-issue")
	}
	owner, repo := parts[0], parts[1]

	description := strings.Join(args, " ")

	client, err := ghclient.New()
	if err != nil {
		return err
	}

	draft, err := workflows.DraftIssue(client, description, workflows.NewIssueOpts{Owner: owner, Repo: repo})
	if err != nil {
		return err
	}

	if !newIssueConfirm {
		if flagJSON {
			return render.JSON(draft)
		}
		fmt.Print(render.Draft(draft))
		return nil
	}

	result, err := workflows.CreateFromDraft(client, draft, owner, repo)
	if err != nil {
		return err
	}
	if flagJSON {
		return render.JSON(result)
	}
	fmt.Print(render.NewIssue(result))
	return nil
}
