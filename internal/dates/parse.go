package dates

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Parse converts user input to a time.Time (start of that day, UTC).
// Accepted formats:
//   - "7d" / "14d"   → N days ago
//   - "2w"           → N weeks ago
//   - "last monday"  → last occurrence of that weekday
//   - "YYYY-MM-DD"   → exact date
func Parse(s string) (time.Time, error) {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	s = strings.ToLower(strings.TrimSpace(s))

	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err == nil {
			return now.AddDate(0, 0, -n), nil
		}
	}
	if strings.HasSuffix(s, "w") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "w"))
		if err == nil {
			return now.AddDate(0, 0, -n*7), nil
		}
	}
	if strings.HasPrefix(s, "last ") {
		day := strings.TrimPrefix(s, "last ")
		target, err := parseWeekday(day)
		if err != nil {
			return time.Time{}, err
		}
		d := now
		for d.Weekday() != target {
			d = d.AddDate(0, 0, -1)
		}
		return d, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		return t.UTC(), nil
	}
	return time.Time{}, fmt.Errorf("unrecognised date %q — use 7d, 2w, last monday, or YYYY-MM-DD", s)
}

func parseWeekday(s string) (time.Weekday, error) {
	days := map[string]time.Weekday{
		"sunday": time.Sunday, "monday": time.Monday,
		"tuesday": time.Tuesday, "wednesday": time.Wednesday,
		"thursday": time.Thursday, "friday": time.Friday,
		"saturday": time.Saturday,
	}
	if w, ok := days[s]; ok {
		return w, nil
	}
	return 0, fmt.Errorf("unknown weekday %q", s)
}
