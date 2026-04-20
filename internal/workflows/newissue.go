package workflows

import (
	"fmt"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
)

type NewIssueOpts struct {
	Owner string
	Repo  string
}

// DraftIssue builds an IssueDraft from the user's description and repo context.
func DraftIssue(c *ghclient.Client, description string, opts NewIssueOpts) (*IssueDraft, error) {
	labels, err := c.ListLabels(opts.Owner, opts.Repo)
	if err != nil {
		return nil, fmt.Errorf("list labels: %w", err)
	}
	templates, err := c.GetIssueTemplates(opts.Owner, opts.Repo)
	if err != nil {
		return nil, fmt.Errorf("get templates: %w", err)
	}

	return &IssueDraft{
		Repo:     opts.Owner + "/" + opts.Repo,
		Title:    buildTitle(description),
		Labels:   pickLabels(description, labels),
		Template: pickTemplate(description, templates),
		Body:     buildBody(description),
	}, nil
}

// CreateFromDraft submits the draft to GitHub and returns the created issue.
func CreateFromDraft(c *ghclient.Client, draft *IssueDraft, owner, repo string) (*NewIssueResult, error) {
	created, err := c.CreateIssue(ghclient.CreateIssueRequest{
		Owner:  owner,
		Repo:   repo,
		Title:  draft.Title,
		Body:   draft.Body,
		Labels: draft.Labels,
	})
	if err != nil {
		return nil, err
	}
	return &NewIssueResult{URL: created.URL, Number: created.Number, Title: created.Title}, nil
}

func pickLabels(desc string, labels []ghclient.GHLabel) []string {
	desc = strings.ToLower(desc)
	var picked []string
	for _, l := range labels {
		name := strings.ToLower(l.Name)
		descText := strings.ToLower(l.Description)
		if strings.Contains(desc, name) || (descText != "" && strings.Contains(desc, descText)) {
			picked = append(picked, l.Name)
		}
	}
	return picked
}

func pickTemplate(desc string, templates []ghclient.IssueTemplate) string {
	desc = strings.ToLower(desc)
	for _, t := range templates {
		stem := strings.ToLower(strings.TrimSuffix(t.Name, ".md"))
		if strings.Contains(desc, stem) {
			return t.Name
		}
	}
	if len(templates) > 0 {
		return templates[0].Name
	}
	return ""
}

func buildTitle(desc string) string {
	if desc == "" {
		return "Untitled"
	}
	title := strings.ToUpper(desc[:1]) + desc[1:]
	if len(title) > 72 {
		title = title[:69] + "..."
	}
	return title
}

func buildBody(desc string) string {
	return fmt.Sprintf(
		"## Problem / Goal\n\n%s\n\n## Expected behaviour\n\n<!-- describe expected behaviour -->\n\n## Context\n\n<!-- screenshots, logs, links -->\n\n## Acceptance criteria\n\n- [ ] <!-- testable criterion -->",
		desc,
	)
}
