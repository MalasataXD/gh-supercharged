package render

import (
	"fmt"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/workflows"
)

func Draft(d *workflows.IssueDraft) string {
	var b strings.Builder
	fmt.Fprintf(&b, "**Title:** %s\n", d.Title)
	fmt.Fprintf(&b, "**Repo:** %s\n", d.Repo)
	if len(d.Labels) > 0 {
		fmt.Fprintf(&b, "**Labels:** %s\n", strings.Join(d.Labels, ", "))
	}
	if d.Template != "" {
		fmt.Fprintf(&b, "**Template:** %s\n", d.Template)
	}
	fmt.Fprintf(&b, "\n---\n\n%s\n", d.Body)
	fmt.Fprintln(&b, "\n---\nReply `ok` to create, or describe changes.")
	return b.String()
}

func NewIssue(r *workflows.NewIssueResult) string {
	return fmt.Sprintf("Created issue #%d: %s\n%s\n", r.Number, r.Title, r.URL)
}
