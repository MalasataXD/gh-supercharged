package render

import (
	"fmt"

	"github.com/MalasataXD/gh-supercharged/internal/workflows"
)

func Move(r *workflows.MoveResult) string {
	return fmt.Sprintf("Moved #%d `%s` → **%s**\n", r.Number, r.Title, r.Status)
}
