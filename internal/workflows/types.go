package workflows

import (
	"time"

	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
)

// --- Plate ---

type IssueRow struct {
	Number    int
	Title     string
	URL       string
	Labels    []string
	Milestone string
	UpdatedAt time.Time
}

type RepoGroup struct {
	Repo   string
	Issues []IssueRow
}

type PlateResult struct {
	Groups []RepoGroup
	Total  int
}

// --- Digest ---

type PRRow struct {
	Number   int
	Title    string
	URL      string
	Labels   []string
	ClosedAt time.Time
}

type DigestRepoGroup struct {
	Repo   string
	Issues []IssueRow
	PRs    []PRRow
}

type DigestResult struct {
	Since       time.Time
	Until       time.Time
	Groups      []DigestRepoGroup
	TotalIssues int
	TotalPRs    int
}

// --- Standup ---

type StandupResult struct {
	Closed []IssueRow
	Open   []IssueRow
	Format string
}

// --- Move ---

type MoveResult struct {
	Number int
	Title  string
	Status string
}

// helpers

func labelsFrom(ls []ghclient.Label) []string {
	out := make([]string, len(ls))
	for i, l := range ls {
		out[i] = l.Name
	}
	return out
}
