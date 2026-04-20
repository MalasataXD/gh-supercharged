package workflows

import (
	"fmt"
	"sort"
	"time"

	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
)

type DigestOpts struct {
	Owner string
}

func Digest(c *ghclient.Client, handle string, since time.Time, opts DigestOpts) (*DigestResult, error) {
	sinceStr := since.Format("2006-01-02")

	issueQ := fmt.Sprintf("involves:%s state:closed closed:>=%s sort:updated", handle, sinceStr)
	if opts.Owner != "" {
		issueQ += " org:" + opts.Owner
	}
	issues, err := c.SearchIssues(issueQ)
	if err != nil {
		return nil, fmt.Errorf("search issues: %w", err)
	}

	prQ := fmt.Sprintf("author:%s state:merged merged:>=%s sort:updated", handle, sinceStr)
	if opts.Owner != "" {
		prQ += " org:" + opts.Owner
	}
	prs, err := c.SearchPRs(prQ)
	if err != nil {
		return nil, fmt.Errorf("search prs: %w", err)
	}

	byRepo := map[string]*DigestRepoGroup{}
	order := []string{}

	addRepo := func(repo string) {
		if byRepo[repo] == nil {
			byRepo[repo] = &DigestRepoGroup{Repo: repo}
			order = append(order, repo)
		}
	}

	for _, iss := range issues {
		repo := iss.RepoName()
		addRepo(repo)
		row := IssueRow{Number: iss.Number, Title: iss.Title, URL: iss.URL, Labels: labelsFrom(iss.Labels)}
		if iss.ClosedAt != nil {
			row.UpdatedAt = *iss.ClosedAt
		}
		byRepo[repo].Issues = append(byRepo[repo].Issues, row)
	}

	for _, pr := range prs {
		repo := pr.RepoName()
		addRepo(repo)
		row := PRRow{Number: pr.Number, Title: pr.Title, URL: pr.URL, Labels: labelsFrom(pr.Labels)}
		if pr.ClosedAt != nil {
			row.ClosedAt = *pr.ClosedAt
		}
		byRepo[repo].PRs = append(byRepo[repo].PRs, row)
	}

	sort.Strings(order)
	groups := make([]DigestRepoGroup, 0, len(order))
	for _, repo := range order {
		groups = append(groups, *byRepo[repo])
	}

	return &DigestResult{
		Since:       since,
		Until:       time.Now().UTC(),
		Groups:      groups,
		TotalIssues: len(issues),
		TotalPRs:    len(prs),
	}, nil
}
