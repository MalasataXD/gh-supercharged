package render

import (
	"fmt"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/workflows"
)

func Plate(r *workflows.PlateResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "**%d open issues assigned to you**\n\n", r.Total)
	for _, g := range r.Groups {
		fmt.Fprintf(&b, "### %s\n", g.Repo)
		for _, iss := range g.Issues {
			labels := ""
			if len(iss.Labels) > 0 {
				labels = " · " + strings.Join(iss.Labels, ", ")
			}
			milestone := ""
			if iss.Milestone != "" {
				milestone = " · " + iss.Milestone
			}
			fmt.Fprintf(&b, "- [#%d](%s) %s%s%s\n",
				iss.Number, iss.URL, iss.Title, labels, milestone)
		}
		fmt.Fprintln(&b)
	}
	return b.String()
}
