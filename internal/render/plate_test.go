package render_test

import (
	"os"
	"testing"
	"time"

	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/MalasataXD/gh-supercharged/internal/workflows"
)

func TestPlateGolden(t *testing.T) {
	result := &workflows.PlateResult{
		Total: 2,
		Groups: []workflows.RepoGroup{
			{
				Repo: "owner/repo-a",
				Issues: []workflows.IssueRow{
					{Number: 1, Title: "Fix login", URL: "https://github.com/owner/repo-a/issues/1", Labels: []string{"bug"}, UpdatedAt: time.Now()},
				},
			},
			{
				Repo: "owner/repo-b",
				Issues: []workflows.IssueRow{
					{Number: 2, Title: "Add dark mode", URL: "https://github.com/owner/repo-b/issues/2", Labels: []string{"enhancement"}, Milestone: "v2.0", UpdatedAt: time.Now()},
				},
			},
		},
	}

	got := render.Plate(result)

	golden, err := os.ReadFile("testdata/plate_golden.md")
	if err != nil {
		t.Fatal(err)
	}
	if got != string(golden) {
		t.Errorf("render.Plate output mismatch\ngot:\n%s\nwant:\n%s", got, string(golden))
	}
}
