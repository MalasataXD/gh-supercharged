package ghclient

import "fmt"

type IssueTemplate struct {
	Name string `json:"name"`
}

func (c *Client) GetIssueTemplates(owner, repo string) ([]IssueTemplate, error) {
	var files []struct {
		Name string `json:"name"`
	}
	path := fmt.Sprintf("repos/%s/%s/contents/.github/ISSUE_TEMPLATE", owner, repo)
	if err := c.REST.Get(path, &files); err != nil {
		// No templates directory is a valid state
		return nil, nil
	}
	out := make([]IssueTemplate, len(files))
	for i, f := range files {
		out[i] = IssueTemplate{Name: f.Name}
	}
	return out, nil
}
