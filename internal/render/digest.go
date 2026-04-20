package render

import (
	"fmt"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/workflows"
)

func Digest(r *workflows.DigestResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Digest: %s → %s\n\n",
		r.Since.Format("2006-01-02"),
		r.Until.Format("2006-01-02"))
	for _, g := range r.Groups {
		fmt.Fprintf(&b, "### %s\n", g.Repo)
		for _, iss := range g.Issues {
			fmt.Fprintf(&b, "- [#%d](%s) %s\n", iss.Number, iss.URL, iss.Title)
		}
		for _, pr := range g.PRs {
			fmt.Fprintf(&b, "- PR [#%d](%s) %s\n", pr.Number, pr.URL, pr.Title)
		}
		fmt.Fprintln(&b)
	}
	fmt.Fprintf(&b, "**Total:** %d issues closed · %d PRs merged\n",
		r.TotalIssues, r.TotalPRs)
	return b.String()
}
