package render

import (
	"fmt"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/workflows"
)

func Standup(r *workflows.StandupResult) string {
	closed := formatRows(r.Closed)
	open := formatRows(r.Open)
	s := r.Format
	s = strings.ReplaceAll(s, "{closed}", closed)
	s = strings.ReplaceAll(s, "{open}", open)
	s = strings.ReplaceAll(s, "{blockers}", "")
	return s
}

func formatRows(rows []workflows.IssueRow) string {
	if len(rows) == 0 {
		return "- (none)"
	}
	var b strings.Builder
	for _, r := range rows {
		fmt.Fprintf(&b, "- [#%d](%s) %s\n", r.Number, r.URL, r.Title)
	}
	return strings.TrimRight(b.String(), "\n")
}
