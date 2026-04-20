package dates_test

import (
	"testing"
	"time"

	"github.com/MalasataXD/gh-supercharged/internal/dates"
)

func TestParse(t *testing.T) {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	cases := []struct {
		input string
		want  time.Time
	}{
		{"7d", now.AddDate(0, 0, -7)},
		{"2w", now.AddDate(0, 0, -14)},
		{"2026-04-10", time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)},
	}
	for _, tc := range cases {
		got, err := dates.Parse(tc.input)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", tc.input, err)
			continue
		}
		if !got.Equal(tc.want) {
			t.Errorf("Parse(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	_, err := dates.Parse("foobar")
	if err == nil {
		t.Error("expected error for invalid input")
	}
}
