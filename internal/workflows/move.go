package workflows

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/projects"
)

type MoveReq struct {
	Owner  string
	Repo   string
	Issue  int
	Status string
}

func Move(c *ghclient.Client, resolver *projects.Resolver, req MoveReq) (*MoveResult, error) {
	detail, err := c.ViewIssueProjects(req.Owner, req.Repo, req.Issue)
	if err != nil {
		return nil, fmt.Errorf("view issue: %w", err)
	}
	if len(detail.ProjectItems) == 0 {
		return nil, fmt.Errorf("issue #%d is not on any project board", req.Issue)
	}

	item := detail.ProjectItems[0]
	projectNumber := item.Project.Number
	projectNodeID := item.Project.ID
	itemNodeID := item.ID

	ids, err := resolver.Resolve(req.Owner, projectNumber, "Status", req.Status)
	if err != nil {
		return nil, err
	}
	ids.ItemNodeID = itemNodeID
	ids.ProjectNodeID = projectNodeID

	if err := c.UpdateProjectField(ghclient.UpdateFieldRequest{
		ItemNodeID:           ids.ItemNodeID,
		ProjectNodeID:        ids.ProjectNodeID,
		FieldID:              ids.FieldID,
		SingleSelectOptionID: ids.OptionID,
	}); err != nil {
		return nil, fmt.Errorf("update field: %w", err)
	}

	return &MoveResult{
		Number: req.Issue,
		Title:  detail.Title,
		Status: req.Status,
	}, nil
}

// ParseIssueArg accepts "#42" or "42" and returns the integer.
func ParseIssueArg(s string) (int, error) {
	s = strings.TrimPrefix(s, "#")
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid issue number %q", s)
	}
	return n, nil
}
