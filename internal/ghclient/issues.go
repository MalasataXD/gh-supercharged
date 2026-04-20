package ghclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type Label struct {
	Name string `json:"name"`
}

type Milestone struct {
	Title string `json:"title"`
}

// Issue represents a GitHub issue as returned by the search API.
// RepositoryURL is "https://api.github.com/repos/owner/repo"; use RepoName() for "owner/repo".
type Issue struct {
	Number        int        `json:"number"`
	Title         string     `json:"title"`
	URL           string     `json:"html_url"`
	Labels        []Label    `json:"labels"`
	Milestone     *Milestone `json:"milestone"`
	RepositoryURL string     `json:"repository_url"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ClosedAt      *time.Time `json:"closed_at"`
	State         string     `json:"state"`
}

// RepoName extracts "owner/repo" from the repository_url field.
// repository_url format: https://api.github.com/repos/owner/repo
func (i Issue) RepoName() string {
	const marker = "/repos/"
	if idx := strings.LastIndex(i.RepositoryURL, marker); idx >= 0 {
		return i.RepositoryURL[idx+len(marker):]
	}
	return i.RepositoryURL
}

// SearchIssues runs the GitHub search API with the given qualifier string.
// qualifiers example: "assignee:@me state:open sort:updated-desc"
func (c *Client) SearchIssues(qualifiers string) ([]Issue, error) {
	path := fmt.Sprintf(
		"search/issues?q=%s&per_page=50",
		url.QueryEscape(qualifiers+" type:issue"),
	)
	var resp struct {
		Items []Issue `json:"items"`
	}
	if err := c.REST.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

type CreateIssueRequest struct {
	Owner  string
	Repo   string
	Title  string
	Body   string
	Labels []string
}

type CreatedIssue struct {
	Number int    `json:"number"`
	URL    string `json:"html_url"`
	Title  string `json:"title"`
}

func (c *Client) CreateIssue(req CreateIssueRequest) (*CreatedIssue, error) {
	body := map[string]interface{}{
		"title":  req.Title,
		"body":   req.Body,
		"labels": req.Labels,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	var created CreatedIssue
	path := fmt.Sprintf("repos/%s/%s/issues", req.Owner, req.Repo)
	if err := c.REST.Post(path, bytes.NewReader(data), &created); err != nil {
		return nil, err
	}
	return &created, nil
}
