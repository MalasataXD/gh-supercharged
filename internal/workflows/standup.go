package workflows

import (
	"time"

	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
)

func Standup(c *ghclient.Client, handle string, format string) (*StandupResult, error) {
	yesterday := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, -1)

	digest, err := Digest(c, handle, yesterday, DigestOpts{})
	if err != nil {
		return nil, err
	}

	var closed []IssueRow
	for _, g := range digest.Groups {
		closed = append(closed, g.Issues...)
	}

	plate, err := Plate(c, handle, PlateOpts{})
	if err != nil {
		return nil, err
	}

	var openRows []IssueRow
	limit := 20
	for _, g := range plate.Groups {
		for _, iss := range g.Issues {
			if limit == 0 {
				break
			}
			openRows = append(openRows, iss)
			limit--
		}
	}

	return &StandupResult{
		Closed: closed,
		Open:   openRows,
		Format: format,
	}, nil
}
