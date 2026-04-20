package ghclient

import "fmt"

type GHLabel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func (c *Client) ListLabels(owner, repo string) ([]GHLabel, error) {
	var labels []GHLabel
	path := fmt.Sprintf("repos/%s/%s/labels?per_page=100", owner, repo)
	if err := c.REST.Get(path, &labels); err != nil {
		return nil, err
	}
	return labels, nil
}
