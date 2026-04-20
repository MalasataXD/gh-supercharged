package workflows

import (
	"sort"

	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
)

type PlateOpts struct {
	Repo  string // optional "owner/repo"
	Owner string // optional owner filter
}

func Plate(c *ghclient.Client, handle string, opts PlateOpts) (*PlateResult, error) {
	q := "assignee:@me state:open sort:updated-desc"
	if opts.Repo != "" {
		q += " repo:" + opts.Repo
	} else if opts.Owner != "" {
		q += " org:" + opts.Owner
	}

	issues, err := c.SearchIssues(q)
	if err != nil {
		return nil, err
	}

	byRepo := map[string]*RepoGroup{}
	order := []string{}

	for _, iss := range issues {
		repo := iss.RepoName()
		if byRepo[repo] == nil {
			byRepo[repo] = &RepoGroup{Repo: repo}
			order = append(order, repo)
		}
		row := IssueRow{
			Number:    iss.Number,
			Title:     iss.Title,
			URL:       iss.URL,
			Labels:    labelsFrom(iss.Labels),
			UpdatedAt: iss.UpdatedAt,
		}
		if iss.Milestone != nil {
			row.Milestone = iss.Milestone.Title
		}
		byRepo[repo].Issues = append(byRepo[repo].Issues, row)
	}

	sort.Strings(order)
	groups := make([]RepoGroup, 0, len(order))
	for _, repo := range order {
		groups = append(groups, *byRepo[repo])
	}

	return &PlateResult{Groups: groups, Total: len(issues)}, nil
}
