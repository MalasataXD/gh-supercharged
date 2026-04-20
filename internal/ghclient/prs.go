package ghclient

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

type PR struct {
	Number        int        `json:"number"`
	Title         string     `json:"title"`
	URL           string     `json:"html_url"`
	Labels        []Label    `json:"labels"`
	RepositoryURL string     `json:"repository_url"`
	ClosedAt      *time.Time `json:"closed_at"`
}

// RepoName extracts "owner/repo" from the repository_url field.
func (p PR) RepoName() string {
	const marker = "/repos/"
	if idx := strings.LastIndex(p.RepositoryURL, marker); idx >= 0 {
		return p.RepositoryURL[idx+len(marker):]
	}
	return p.RepositoryURL
}

// SearchPRs runs the GitHub search API with the given qualifier string.
// qualifiers example: "author:handle state:merged merged:>=2026-04-10"
func (c *Client) SearchPRs(qualifiers string) ([]PR, error) {
	path := fmt.Sprintf(
		"search/issues?q=%s&per_page=100",
		url.QueryEscape(qualifiers+" type:pr"),
	)
	var resp struct {
		Items []PR `json:"items"`
	}
	if err := c.REST.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
