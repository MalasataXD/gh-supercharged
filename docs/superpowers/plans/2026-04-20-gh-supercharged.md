# gh-supercharged Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a `gh` extension (`gh supercharged`) in Go that implements five GitHub workflows (Plate, Digest, Standup, Move, New Issue) with dual Markdown/JSON output, replacing raw `gh` CLI calls in the `gh` skill.

**Architecture:** Cobra CLI at the top, `internal/workflows/` as pure business logic, `internal/ghclient/` wrapping `go-gh` REST+GraphQL, and `internal/render/` for dual output. Config and project-ID cache are owned by the binary and stored alongside `gh`'s own config directory.

**Tech Stack:** Go 1.26, `github.com/cli/go-gh/v2` (auth + HTTP clients), `github.com/spf13/cobra` (CLI), `encoding/json` (stdlib), `os`/`time`/`strings` (stdlib). No ORM, no DB.

---

## File Map

```
gh-supercharged/
  main.go
  cmd/
    root.go
    plate.go
    digest.go
    standup.go
    move.go
    newissue.go
    cache.go
    config.go
  internal/
    ghclient/
      client.go        # New() → Client struct, shared REST+GraphQL clients
      issues.go        # SearchIssues, ViewIssue, CreateIssue
      prs.go           # SearchPRs
      labels.go        # ListLabels
      templates.go     # GetIssueTemplates
      projects.go      # ProjectFields, UpdateProjectField
    workflows/
      types.go         # all result structs
      plate.go         # Plate(client, handle, opts) → PlateResult
      digest.go        # Digest(client, handle, since, opts) → DigestResult
      standup.go       # Standup(client, handle) → StandupResult
      move.go          # Move(client, resolver, req) → MoveResult
      newissue.go      # DraftIssue / CreateIssue
    projects/
      resolver.go      # Resolver struct + Resolve()
      types.go         # ProjectIDs, CacheEntry
    config/
      config.go        # Config struct, Load(), Save(), FirstRun()
      paths.go         # ConfigDir() — env override + go-gh default
      default.json     # embedded default config
    cache/
      cache.go         # Cache struct, Load(), Save(), Get(), Set(), Invalidate()
    render/
      json.go          # RenderJSON(any) → stdout
      plate.go         # RenderPlate(PlateResult) → string
      digest.go        # RenderDigest(DigestResult) → string
      standup.go       # RenderStandup(StandupResult) → string
      move.go          # RenderMove(MoveResult) → string
      newissue.go      # RenderDraft(DraftResult) → string
    dates/
      parse.go         # Parse(s string) → time.Time, error
      parse_test.go
  .github/workflows/
    ci.yml
    release.yml
  go.mod
  go.sum
  .gitignore
  README.md
  LICENSE
```

---

## M1 — Scaffold + Plate

### Task 1: Initialise the module and dependencies

**Files:**
- Create: `go.mod`
- Create: `.gitignore`

- [ ] **Step 1: Init the git repo and module**

```bash
cd C:/Users/MadsHyllebergNielsen/sandbox/gh-supercharged
git init
go mod init github.com/MalasataXD/gh-supercharged
```

- [ ] **Step 2: Add dependencies**

```bash
go get github.com/cli/go-gh/v2@latest
go get github.com/spf13/cobra@latest
```

- [ ] **Step 3: Create `.gitignore`**

```
gh-supercharged
gh-supercharged.exe
dist/
*.test
```

- [ ] **Step 4: Commit**

```bash
git init
git add go.mod go.sum .gitignore
git commit -m "chore: init module with go-gh and cobra"
```

---

### Task 2: Config package

**Files:**
- Create: `internal/config/paths.go`
- Create: `internal/config/config.go`
- Create: `internal/config/default.json`

- [ ] **Step 1: Write `internal/config/default.json`**

```json
{
  "github_handle": "your-github-username",
  "digest_window_days": 7,
  "standup_format": "## Yesterday\n{closed}\n\n## Today\n{open}\n\n## Blockers\n{blockers}"
}
```

- [ ] **Step 2: Write `internal/config/paths.go`**

```go
package config

import (
	"os"
	"path/filepath"

	"github.com/cli/go-gh/v2/pkg/config"
)

// ConfigDir returns the directory where config.json and projects.json live.
// Override with GH_SC_CONFIG_DIR.
func ConfigDir() (string, error) {
	if override := os.Getenv("GH_SC_CONFIG_DIR"); override != "" {
		return override, nil
	}
	ghDir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(ghDir, "extensions", "supercharged"), nil
}
```

- [ ] **Step 3: Write `internal/config/config.go`**

```go
package config

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed default.json
var defaultJSON []byte

type Config struct {
	GithubHandle     string `json:"github_handle"`
	DigestWindowDays int    `json:"digest_window_days"`
	StandupFormat    string `json:"standup_format"`
}

// Load reads config.json. On first run, creates it from the embedded default
// and returns ErrFirstRun so callers can print a setup message and exit.
var ErrFirstRun = errors.New("first run")

func Load() (*Config, string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, "", fmt.Errorf("config dir: %w", err)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, "", fmt.Errorf("mkdir: %w", err)
	}
	path := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if writeErr := os.WriteFile(path, defaultJSON, 0o600); writeErr != nil {
			return nil, path, fmt.Errorf("write default config: %w", writeErr)
		}
		return nil, path, ErrFirstRun
	}
	if err != nil {
		return nil, path, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, path, fmt.Errorf("parse config: %w", err)
	}
	if cfg.GithubHandle == "" || cfg.GithubHandle == "your-github-username" {
		return nil, path, fmt.Errorf("%w: set github_handle in %s", ErrFirstRun, path)
	}
	return &cfg, path, nil
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/config/
git commit -m "feat: config package with first-run detection"
```

---

### Task 3: Cache package

**Files:**
- Create: `internal/cache/cache.go`

- [ ] **Step 1: Write `internal/cache/cache.go`**

```go
package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/MalasataXD/gh-supercharged/internal/config"
)

type FieldOptions map[string]string // option name → option node ID

type StatusField struct {
	ID      string       `json:"id"`
	Options FieldOptions `json:"options"`
}

type ProjectFields struct {
	Status StatusField `json:"Status"`
}

type Entry struct {
	ProjectID string        `json:"project_id"`
	Fields    ProjectFields `json:"fields"`
	CachedAt  time.Time     `json:"cached_at"`
}

type Cache struct {
	Projects map[string]Entry `json:"projects"` // key: "owner/project-number"
	path     string
}

func Load() (*Cache, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "projects.json")
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Cache{Projects: map[string]Entry{}, path: path}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read cache: %w", err)
	}
	var c Cache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse cache: %w", err)
	}
	c.path = path
	return &c, nil
}

func (c *Cache) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o600)
}

func (c *Cache) Get(key string) (Entry, bool) {
	e, ok := c.Projects[key]
	return e, ok
}

func (c *Cache) Set(key string, e Entry) {
	if c.Projects == nil {
		c.Projects = map[string]Entry{}
	}
	e.CachedAt = time.Now().UTC()
	c.Projects[key] = e
}

func (c *Cache) Invalidate(key string) {
	delete(c.Projects, key)
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/cache/
git commit -m "feat: cache package for Projects v2 ID storage"
```

---

### Task 4: ghclient — Client and SearchIssues

**Files:**
- Create: `internal/ghclient/client.go`
- Create: `internal/ghclient/issues.go`

- [ ] **Step 1: Write `internal/ghclient/client.go`**

```go
package ghclient

import (
	"github.com/cli/go-gh/v2/pkg/api"
)

type Client struct {
	REST    *api.RESTClient
	GQL     *api.GraphQLClient
}

func New() (*Client, error) {
	rest, err := api.DefaultRESTClient()
	if err != nil {
		return nil, err
	}
	gql, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, err
	}
	return &Client{REST: rest, GQL: gql}, nil
}
```

- [ ] **Step 2: Write `internal/ghclient/issues.go`**

```go
package ghclient

import (
	"fmt"
	"net/url"
	"time"
)

type Label struct {
	Name string `json:"name"`
}

type Milestone struct {
	Title string `json:"title"`
}

type Repository struct {
	NameWithOwner string `json:"nameWithOwner"`
}

type Issue struct {
	Number     int        `json:"number"`
	Title      string     `json:"title"`
	URL        string     `json:"url"`
	Labels     []Label    `json:"labels"`
	Milestone  *Milestone `json:"milestone"`
	Repository Repository `json:"repository"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	ClosedAt   *time.Time `json:"closedAt"`
	State      string     `json:"state"`
}

type ProjectItem struct {
	ID      string `json:"id"`
	Project struct {
		ID     string `json:"id"`
		Number int    `json:"number"`
		Title  string `json:"title"`
	} `json:"project"`
}

type IssueDetail struct {
	Issue
	Body         string        `json:"body"`
	Assignees    []struct{ Login string `json:"login"` } `json:"assignees"`
	ProjectItems []ProjectItem `json:"projectItems"`
}

// SearchIssues runs gh search issues with the given qualifier string.
// qualifiers example: "assignee:@me state:open sort:updated-desc"
func (c *Client) SearchIssues(qualifiers string) ([]Issue, error) {
	path := fmt.Sprintf(
		"search/issues?q=%s&per_page=50",
		url.QueryEscape(qualifiers+" type:issue"),
	)
	var resp struct {
		Items []Issue `json:"items"`
	}
	if err := c.REST.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// ViewIssue fetches a single issue with project membership.
func (c *Client) ViewIssue(owner, repo string, number int) (*IssueDetail, error) {
	var detail IssueDetail
	path := fmt.Sprintf("repos/%s/%s/issues/%d", owner, repo, number)
	if err := c.REST.Get(path, &detail); err != nil {
		return nil, err
	}
	// projectItems requires GraphQL — fetched separately in projects.go
	return &detail, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/ghclient/
git commit -m "feat: ghclient package with REST+GQL client and SearchIssues"
```

---

### Task 5: workflows.Plate

**Files:**
- Create: `internal/workflows/types.go`
- Create: `internal/workflows/plate.go`

- [ ] **Step 1: Write `internal/workflows/types.go`**

```go
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
	Since  time.Time
	Until  time.Time
	Groups []DigestRepoGroup
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

// --- New Issue ---

type IssueDraft struct {
	Repo     string
	Title    string
	Labels   []string
	Template string
	Body     string
}

type NewIssueResult struct {
	URL    string
	Number int
	Title  string
}

// helpers
func labelsFrom(ls []ghclient.Label) []string {
	out := make([]string, len(ls))
	for i, l := range ls {
		out[i] = l.Name
	}
	return out
}
```

- [ ] **Step 2: Write `internal/workflows/plate.go`**

```go
package workflows

import (
	"sort"

	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
)

type PlateOpts struct {
	Repo  string // optional "owner/repo"
	Owner string // optional owner filter
}

func Plate(c *ghclient.Client, handle string, opts PlateOpts) (*PlateResult, error) {
	q := "assignee:@me state:open sort:updated-desc"
	if opts.Repo != "" {
		q += " repo:" + opts.Repo
	} else if opts.Owner != "" {
		q += " org:" + opts.Owner
	}

	issues, err := c.SearchIssues(q)
	if err != nil {
		return nil, err
	}

	byRepo := map[string]*RepoGroup{}
	order := []string{}

	for _, iss := range issues {
		repo := iss.Repository.NameWithOwner
		if byRepo[repo] == nil {
			byRepo[repo] = &RepoGroup{Repo: repo}
			order = append(order, repo)
		}
		row := IssueRow{
			Number:    iss.Number,
			Title:     iss.Title,
			URL:       iss.URL,
			Labels:    labelsFrom(iss.Labels),
			UpdatedAt: iss.UpdatedAt,
		}
		if iss.Milestone != nil {
			row.Milestone = iss.Milestone.Title
		}
		byRepo[repo].Issues = append(byRepo[repo].Issues, row)
	}

	sort.Strings(order)
	groups := make([]RepoGroup, 0, len(order))
	for _, repo := range order {
		groups = append(groups, *byRepo[repo])
	}

	return &PlateResult{Groups: groups, Total: len(issues)}, nil
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/workflows/
git commit -m "feat: workflows.Plate — group open issues by repo"
```

---

### Task 6: render — Plate (Markdown + JSON)

**Files:**
- Create: `internal/render/json.go`
- Create: `internal/render/plate.go`
- Create: `internal/render/testdata/plate_golden.md`

- [ ] **Step 1: Write `internal/render/json.go`**

```go
package render

import (
	"encoding/json"
	"fmt"
	"os"
)

func JSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("render json: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: Write `internal/render/plate.go`**

```go
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
```

- [ ] **Step 3: Write golden test file `internal/render/testdata/plate_golden.md`**

```markdown
**2 open issues assigned to you**

### owner/repo-a
- [#1](https://github.com/owner/repo-a/issues/1) Fix login · bug

### owner/repo-b
- [#2](https://github.com/owner/repo-b/issues/2) Add dark mode · enhancement · v2.0

```

- [ ] **Step 4: Write `internal/render/plate_test.go`**

```go
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
```

- [ ] **Step 5: Run test**

```bash
go test ./internal/render/ -v -run TestPlateGolden
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/render/
git commit -m "feat: render.Plate with golden-file test"
```

---

### Task 7: Cobra root + cmd/plate.go

**Files:**
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `cmd/plate.go`

- [ ] **Step 1: Write `main.go`**

```go
package main

import (
	"os"

	"github.com/MalasataXD/gh-supercharged/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Write `cmd/root.go`**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gh-supercharged",
	Short: "Supercharged GitHub CLI workflows",
}

// global flags wired to each subcommand via PersistentFlags
var (
	flagJSON    bool
	flagRepo    string
	flagOwner   string
	flagVerbose bool
)

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().StringVar(&flagRepo, "repo", "", "Repository (owner/repo)")
	rootCmd.PersistentFlags().StringVar(&flagOwner, "owner", "", "Owner filter")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Verbose error output")
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}
```

- [ ] **Step 3: Write `cmd/plate.go`**

```go
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/MalasataXD/gh-supercharged/internal/workflows"
	"github.com/spf13/cobra"
)

var plateCmd = &cobra.Command{
	Use:   "plate",
	Short: "Show open issues assigned to you",
	RunE:  runPlate,
}

func init() {
	rootCmd.AddCommand(plateCmd)
}

func runPlate(cmd *cobra.Command, _ []string) error {
	cfg, cfgPath, err := config.Load()
	if errors.Is(err, config.ErrFirstRun) {
		fmt.Fprintf(os.Stderr, "First-time setup — created %s\nSet github_handle, then re-run.\n", cfgPath)
		os.Exit(3)
	}
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	client, err := ghclient.New()
	if err != nil {
		return fmt.Errorf("gh client: %w", err)
	}

	result, err := workflows.Plate(client, cfg.GithubHandle, workflows.PlateOpts{
		Repo:  flagRepo,
		Owner: flagOwner,
	})
	if err != nil {
		return err
	}

	if flagJSON {
		return render.JSON(result)
	}
	fmt.Print(render.Plate(result))
	return nil
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
go build -o gh-supercharged.exe .
./gh-supercharged.exe plate
./gh-supercharged.exe plate --json
```

Expected: Markdown list of your open issues, then the same as JSON.

- [ ] **Step 5: Commit**

```bash
git add main.go cmd/
git commit -m "feat: cobra root and plate command"
```

---

## M2 — Digest + Standup

### Task 8: dates package

**Files:**
- Create: `internal/dates/parse.go`
- Create: `internal/dates/parse_test.go`

- [ ] **Step 1: Write `internal/dates/parse.go`**

```go
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

	// Nd
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err == nil {
			return now.AddDate(0, 0, -n), nil
		}
	}
	// Nw
	if strings.HasSuffix(s, "w") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "w"))
		if err == nil {
			return now.AddDate(0, 0, -n*7), nil
		}
	}
	// "last <weekday>"
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
	// ISO date
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
```

- [ ] **Step 2: Write `internal/dates/parse_test.go`**

```go
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
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/dates/ -v
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/dates/
git commit -m "feat: dates.Parse supporting Nd, Nw, last <weekday>, ISO"
```

---

### Task 9: ghclient.SearchPRs

**Files:**
- Modify: `internal/ghclient/prs.go`

- [ ] **Step 1: Create `internal/ghclient/prs.go`**

```go
package ghclient

import (
	"fmt"
	"net/url"
	"time"
)

type PR struct {
	Number     int        `json:"number"`
	Title      string     `json:"title"`
	URL        string     `json:"url"`
	Labels     []Label    `json:"labels"`
	Repository Repository `json:"repository"`
	ClosedAt   *time.Time `json:"closed_at"`
}

// SearchPRs runs gh search prs with the given qualifier string.
// qualifiers example: "author:handle state:merged merged:>=2026-04-10"
func (c *Client) SearchPRs(qualifiers string) ([]PR, error) {
	path := fmt.Sprintf(
		"search/issues?q=%s&per_page=100",
		url.QueryEscape(qualifiers+" type:pr"),
	)
	var resp struct {
		Items []PR `json:"items"`
	}
	if err := c.REST.Get(path, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/ghclient/prs.go
git commit -m "feat: ghclient.SearchPRs"
```

---

### Task 10: workflows.Digest

**Files:**
- Create: `internal/workflows/digest.go`

- [ ] **Step 1: Write `internal/workflows/digest.go`**

```go
package workflows

import (
	"fmt"
	"sort"
	"time"

	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
)

type DigestOpts struct {
	Owner string
}

func Digest(c *ghclient.Client, handle string, since time.Time, opts DigestOpts) (*DigestResult, error) {
	sinceStr := since.Format("2006-01-02")

	issueQ := fmt.Sprintf("involves:%s state:closed closed:>=%s sort:updated", handle, sinceStr)
	if opts.Owner != "" {
		issueQ += " org:" + opts.Owner
	}
	issues, err := c.SearchIssues(issueQ)
	if err != nil {
		return nil, fmt.Errorf("search issues: %w", err)
	}

	prQ := fmt.Sprintf("author:%s state:merged merged:>=%s sort:updated", handle, sinceStr)
	if opts.Owner != "" {
		prQ += " org:" + opts.Owner
	}
	prs, err := c.SearchPRs(prQ)
	if err != nil {
		return nil, fmt.Errorf("search prs: %w", err)
	}

	byRepo := map[string]*DigestRepoGroup{}
	order := []string{}

	addRepo := func(repo string) {
		if byRepo[repo] == nil {
			byRepo[repo] = &DigestRepoGroup{Repo: repo}
			order = append(order, repo)
		}
	}

	for _, iss := range issues {
		repo := iss.Repository.NameWithOwner
		addRepo(repo)
		row := IssueRow{Number: iss.Number, Title: iss.Title, URL: iss.URL, Labels: labelsFrom(iss.Labels)}
		if iss.ClosedAt != nil {
			row.UpdatedAt = *iss.ClosedAt
		}
		byRepo[repo].Issues = append(byRepo[repo].Issues, row)
	}

	for _, pr := range prs {
		repo := pr.Repository.NameWithOwner
		addRepo(repo)
		row := PRRow{Number: pr.Number, Title: pr.Title, URL: pr.URL, Labels: labelsFrom(pr.Labels)}
		if pr.ClosedAt != nil {
			row.ClosedAt = *pr.ClosedAt
		}
		byRepo[repo].PRs = append(byRepo[repo].PRs, row)
	}

	sort.Strings(order)
	groups := make([]DigestRepoGroup, 0, len(order))
	for _, repo := range order {
		groups = append(groups, *byRepo[repo])
	}

	return &DigestResult{
		Since:       since,
		Until:       time.Now().UTC(),
		Groups:      groups,
		TotalIssues: len(issues),
		TotalPRs:    len(prs),
	}, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/workflows/digest.go
git commit -m "feat: workflows.Digest — closed issues + merged PRs in window"
```

---

### Task 11: workflows.Standup

**Files:**
- Create: `internal/workflows/standup.go`

- [ ] **Step 1: Write `internal/workflows/standup.go`**

```go
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

	open := plate.Groups
	var openRows []IssueRow
	limit := 20
	for _, g := range open {
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
```

- [ ] **Step 2: Commit**

```bash
git add internal/workflows/standup.go
git commit -m "feat: workflows.Standup composing Digest(yesterday) + Plate"
```

---

### Task 12: render.Digest + render.Standup + cmd wiring

**Files:**
- Create: `internal/render/digest.go`
- Create: `internal/render/standup.go`
- Create: `cmd/digest.go`
- Create: `cmd/standup.go`

- [ ] **Step 1: Write `internal/render/digest.go`**

```go
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
```

- [ ] **Step 2: Write `internal/render/standup.go`**

```go
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
```

- [ ] **Step 3: Write `cmd/digest.go`**

```go
package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/MalasataXD/gh-supercharged/internal/dates"
	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/MalasataXD/gh-supercharged/internal/workflows"
	"github.com/spf13/cobra"
)

var digestCmd = &cobra.Command{
	Use:   "digest [since]",
	Short: "Summarise closed issues and merged PRs",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runDigest,
}

func init() { rootCmd.AddCommand(digestCmd) }

func runDigest(_ *cobra.Command, args []string) error {
	cfg, cfgPath, err := config.Load()
	if errors.Is(err, config.ErrFirstRun) {
		fmt.Fprintf(os.Stderr, "First-time setup — created %s\nSet github_handle, then re-run.\n", cfgPath)
		os.Exit(3)
	}
	if err != nil {
		return err
	}

	since := time.Now().UTC().Truncate(24 * time.Hour).AddDate(0, 0, -cfg.DigestWindowDays)
	if len(args) == 1 {
		since, err = dates.Parse(args[0])
		if err != nil {
			return err
		}
	}

	client, err := ghclient.New()
	if err != nil {
		return err
	}

	result, err := workflows.Digest(client, cfg.GithubHandle, since, workflows.DigestOpts{Owner: flagOwner})
	if err != nil {
		return err
	}

	if flagJSON {
		return render.JSON(result)
	}
	fmt.Print(render.Digest(result))
	return nil
}
```

- [ ] **Step 4: Write `cmd/standup.go`**

```go
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/MalasataXD/gh-supercharged/internal/workflows"
	"github.com/spf13/cobra"
)

var standupCmd = &cobra.Command{
	Use:   "standup",
	Short: "Yesterday's closes + today's open plate",
	RunE:  runStandup,
}

func init() { rootCmd.AddCommand(standupCmd) }

func runStandup(_ *cobra.Command, _ []string) error {
	cfg, cfgPath, err := config.Load()
	if errors.Is(err, config.ErrFirstRun) {
		fmt.Fprintf(os.Stderr, "First-time setup — created %s\nSet github_handle, then re-run.\n", cfgPath)
		os.Exit(3)
	}
	if err != nil {
		return err
	}

	client, err := ghclient.New()
	if err != nil {
		return err
	}

	result, err := workflows.Standup(client, cfg.GithubHandle, cfg.StandupFormat)
	if err != nil {
		return err
	}

	if flagJSON {
		return render.JSON(result)
	}
	fmt.Print(render.Standup(result))
	return nil
}
```

- [ ] **Step 5: Build and smoke-test**

```bash
go build -o gh-supercharged.exe .
./gh-supercharged.exe digest 7d
./gh-supercharged.exe standup --json
```

- [ ] **Step 6: Commit**

```bash
git add internal/render/digest.go internal/render/standup.go cmd/digest.go cmd/standup.go
git commit -m "feat: digest and standup commands with dual output"
```

---

## M3 — Move

### Task 13: Projects v2 GraphQL types + ghclient.ProjectFields

**Files:**
- Create: `internal/ghclient/projects.go`
- Create: `internal/projects/types.go`

- [ ] **Step 1: Write `internal/projects/types.go`**

```go
package projects

type FieldOption struct {
	ID   string
	Name string
}

type Field struct {
	ID      string
	Name    string
	Type    string // ProjectV2SingleSelectField, ProjectV2Field, etc.
	Options []FieldOption
}

type IDs struct {
	ItemNodeID    string
	ProjectNodeID string
	FieldID       string
	OptionID      string // empty for non-single-select fields
}
```

- [ ] **Step 2: Write `internal/ghclient/projects.go`**

```go
package ghclient

import "fmt"

type ProjectItemsResponse struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	ProjectItems []struct {
		ID      string `json:"id"`
		Project struct {
			ID     string `json:"id"`
			Number int    `json:"number"`
			Title  string `json:"title"`
		} `json:"project"`
	} `json:"projectItems"`
}

// ViewIssueProjects fetches an issue including its Projects v2 membership via GraphQL.
func (c *Client) ViewIssueProjects(owner, repo string, number int) (*ProjectItemsResponse, error) {
	var resp struct {
		Data struct {
			Repository struct {
				Issue ProjectItemsResponse `json:"issue"`
			} `json:"repository"`
		} `json:"data"`
	}
	query := `query($owner:String!,$repo:String!,$number:Int!) {
		repository(owner:$owner,name:$repo) {
			issue(number:$number) {
				number title url
				projectItems(first:10) {
					nodes {
						id
						project { id number title }
					}
				}
			}
		}
	}`
	// go-gh GraphQL client
	vars := map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}
	if err := c.GQL.Do(query, vars, &resp.Data); err != nil {
		return nil, fmt.Errorf("ViewIssueProjects: %w", err)
	}
	return &resp.Data.Repository.Issue, nil
}

type FieldListResponse struct {
	Fields []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Type    string `json:"dataType"`
		Options []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"options"`
	} `json:"fields"`
}

// ProjectFields returns all fields for a project via GraphQL.
func (c *Client) ProjectFields(owner string, projectNumber int) (*FieldListResponse, error) {
	var data struct {
		User struct {
			ProjectV2 struct {
				Fields struct {
					Nodes []struct {
						ID      string `json:"id"`
						Name    string `json:"name"`
						Options []struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"options"`
					} `json:"nodes"`
				} `json:"fields"`
			} `json:"projectV2"`
		} `json:"user"`
	}
	query := `query($owner:String!,$number:Int!) {
		user(login:$owner) {
			projectV2(number:$number) {
				fields(first:50) {
					nodes {
						... on ProjectV2SingleSelectField { id name options { id name } }
						... on ProjectV2Field { id name }
						... on ProjectV2IterationField { id name }
					}
				}
			}
		}
	}`
	vars := map[string]interface{}{"owner": owner, "number": projectNumber}
	if err := c.GQL.Do(query, vars, &data); err != nil {
		return nil, fmt.Errorf("ProjectFields: %w", err)
	}
	resp := &FieldListResponse{}
	for _, n := range data.User.ProjectV2.Fields.Nodes {
		f := struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Type    string `json:"dataType"`
			Options []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"options"`
		}{ID: n.ID, Name: n.Name, Options: n.Options}
		resp.Fields = append(resp.Fields, f)
	}
	return resp, nil
}

type UpdateFieldRequest struct {
	ItemNodeID    string
	ProjectNodeID string
	FieldID       string
	// Exactly one of the following should be set:
	SingleSelectOptionID string
	Text                 string
	Number               *float64
	Date                 string // YYYY-MM-DD
	IterationID          string
	Clear                bool
}

// UpdateProjectField mutates a project item field via GraphQL.
func (c *Client) UpdateProjectField(req UpdateFieldRequest) error {
	var mutation string
	vars := map[string]interface{}{
		"project": req.ProjectNodeID,
		"item":    req.ItemNodeID,
		"field":   req.FieldID,
	}

	switch {
	case req.Clear:
		mutation = `mutation($project:ID!,$item:ID!,$field:ID!) {
			clearProjectV2ItemFieldValue(input:{projectId:$project,itemId:$item,fieldId:$field}) { projectV2Item { id } }
		}`
	case req.SingleSelectOptionID != "":
		vars["option"] = req.SingleSelectOptionID
		mutation = `mutation($project:ID!,$item:ID!,$field:ID!,$option:String!) {
			updateProjectV2ItemFieldValue(input:{projectId:$project,itemId:$item,fieldId:$field,value:{singleSelectOptionId:$option}}) { projectV2Item { id } }
		}`
	case req.Text != "":
		vars["text"] = req.Text
		mutation = `mutation($project:ID!,$item:ID!,$field:ID!,$text:String!) {
			updateProjectV2ItemFieldValue(input:{projectId:$project,itemId:$item,fieldId:$field,value:{text:$text}}) { projectV2Item { id } }
		}`
	case req.Number != nil:
		vars["number"] = *req.Number
		mutation = `mutation($project:ID!,$item:ID!,$field:ID!,$number:Float!) {
			updateProjectV2ItemFieldValue(input:{projectId:$project,itemId:$item,fieldId:$field,value:{number:$number}}) { projectV2Item { id } }
		}`
	case req.Date != "":
		vars["date"] = req.Date
		mutation = `mutation($project:ID!,$item:ID!,$field:ID!,$date:Date!) {
			updateProjectV2ItemFieldValue(input:{projectId:$project,itemId:$item,fieldId:$field,value:{date:$date}}) { projectV2Item { id } }
		}`
	case req.IterationID != "":
		vars["iteration"] = req.IterationID
		mutation = `mutation($project:ID!,$item:ID!,$field:ID!,$iteration:String!) {
			updateProjectV2ItemFieldValue(input:{projectId:$project,itemId:$item,fieldId:$field,value:{iterationId:$iteration}}) { projectV2Item { id } }
		}`
	default:
		return fmt.Errorf("UpdateProjectField: no value provided")
	}

	var result interface{}
	return c.GQL.Do(mutation, vars, &result)
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/ghclient/projects.go internal/projects/types.go
git commit -m "feat: ghclient ProjectFields + UpdateProjectField via GraphQL"
```

---

### Task 14: projects.Resolver

**Files:**
- Create: `internal/projects/resolver.go`

- [ ] **Step 1: Write `internal/projects/resolver.go`**

```go
package projects

import (
	"fmt"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/cache"
	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
)

type Resolver struct {
	client *ghclient.Client
	cache  *cache.Cache
}

func NewResolver(c *ghclient.Client, ca *cache.Cache) *Resolver {
	return &Resolver{client: c, cache: ca}
}

// Resolve returns the node IDs needed to call UpdateProjectField.
// owner/projectNumber identify the project; fieldName is e.g. "Status";
// optionName is the desired value (case-insensitive). For non-single-select
// fields pass optionName="".
func (r *Resolver) Resolve(owner string, projectNumber int, fieldName, optionName string) (*IDs, error) {
	ids, err := r.resolve(owner, projectNumber, fieldName, optionName)
	if err != nil && isStaleError(err) {
		// invalidate and retry once
		key := fmt.Sprintf("%s/%d", owner, projectNumber)
		r.cache.Invalidate(key)
		_ = r.cache.Save()
		ids, err = r.resolve(owner, projectNumber, fieldName, optionName)
	}
	return ids, err
}

func (r *Resolver) resolve(owner string, projectNumber int, fieldName, optionName string) (*IDs, error) {
	key := fmt.Sprintf("%s/%d", owner, projectNumber)
	entry, hit := r.cache.Get(key)

	if !hit {
		if err := r.populate(owner, projectNumber, key); err != nil {
			return nil, err
		}
		entry, _ = r.cache.Get(key)
	}

	ids := &IDs{ProjectNodeID: entry.ProjectID}

	// Resolve field
	statusField := entry.Fields.Status
	if !strings.EqualFold(fieldName, "Status") {
		return nil, fmt.Errorf("field %q not yet supported — only Status is cached", fieldName)
	}
	ids.FieldID = statusField.ID

	// Resolve option (case-insensitive)
	if optionName != "" {
		for name, optID := range statusField.Options {
			if strings.EqualFold(name, optionName) {
				ids.OptionID = optID
				break
			}
		}
		if ids.OptionID == "" {
			available := make([]string, 0, len(statusField.Options))
			for name := range statusField.Options {
				available = append(available, name)
			}
			return nil, fmt.Errorf("status %q not found — available: %s", optionName, strings.Join(available, ", "))
		}
	}

	return ids, nil
}

func (r *Resolver) populate(owner string, projectNumber int, key string) error {
	fields, err := r.client.ProjectFields(owner, projectNumber)
	if err != nil {
		return fmt.Errorf("discover project fields: %w", err)
	}

	entry := cache.Entry{}
	for _, f := range fields.Fields {
		if strings.EqualFold(f.Name, "Status") {
			entry.Fields.Status.ID = f.ID
			entry.Fields.Status.Options = make(cache.FieldOptions)
			for _, opt := range f.Options {
				entry.Fields.Status.Options[opt.Name] = opt.ID
			}
		}
	}

	// project_id needs to come from the issue view — we set it during Move
	r.cache.Set(key, entry)
	return r.cache.Save()
}

func isStaleError(err error) bool {
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "not found") || strings.Contains(s, "does not exist")
}
```

- [ ] **Step 2: Commit**

```bash
git add internal/projects/resolver.go
git commit -m "feat: projects.Resolver with cold discovery and stale-retry"
```

---

### Task 15: workflows.Move + render + cmd

**Files:**
- Create: `internal/workflows/move.go`
- Create: `internal/render/move.go`
- Create: `cmd/move.go`

- [ ] **Step 1: Write `internal/workflows/move.go`**

```go
package workflows

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/projects"
)

type MoveReq struct {
	Owner  string
	Repo   string
	Issue  int
	Status string
}

func Move(c *ghclient.Client, resolver *projects.Resolver, req MoveReq) (*MoveResult, error) {
	detail, err := c.ViewIssueProjects(req.Owner, req.Repo, req.Issue)
	if err != nil {
		return nil, fmt.Errorf("view issue: %w", err)
	}
	if len(detail.ProjectItems) == 0 {
		return nil, fmt.Errorf("issue #%d is not on any project board", req.Issue)
	}

	item := detail.ProjectItems[0]
	projectNumber := item.Project.Number
	projectNodeID := item.Project.ID
	itemNodeID := item.ID

	ids, err := resolver.Resolve(req.Owner, projectNumber, "Status", req.Status)
	if err != nil {
		return nil, err
	}
	ids.ItemNodeID = itemNodeID
	ids.ProjectNodeID = projectNodeID

	if err := c.UpdateProjectField(ghclient.UpdateFieldRequest{
		ItemNodeID:           ids.ItemNodeID,
		ProjectNodeID:        ids.ProjectNodeID,
		FieldID:              ids.FieldID,
		SingleSelectOptionID: ids.OptionID,
	}); err != nil {
		return nil, fmt.Errorf("update field: %w", err)
	}

	return &MoveResult{
		Number: req.Issue,
		Title:  detail.Title,
		Status: req.Status,
	}, nil
}

// ParseIssueArg accepts "#42" or "42" and returns the integer.
func ParseIssueArg(s string) (int, error) {
	s = strings.TrimPrefix(s, "#")
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid issue number %q", s)
	}
	return n, nil
}
```

- [ ] **Step 2: Write `internal/render/move.go`**

```go
package render

import (
	"fmt"

	"github.com/MalasataXD/gh-supercharged/internal/workflows"
)

func Move(r *workflows.MoveResult) string {
	return fmt.Sprintf("Moved #%d `%s` → **%s**\n", r.Number, r.Title, r.Status)
}
```

- [ ] **Step 3: Write `cmd/move.go`**

```go
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/cache"
	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/projects"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/MalasataXD/gh-supercharged/internal/workflows"
	"github.com/spf13/cobra"
)

var moveCmd = &cobra.Command{
	Use:   "move <issue> <status>",
	Short: "Move an issue to a new project status",
	Args:  cobra.ExactArgs(2),
	RunE:  runMove,
}

var moveProject int

func init() {
	moveCmd.Flags().IntVar(&moveProject, "project", 0, "Project number (if issue is in multiple projects)")
	rootCmd.AddCommand(moveCmd)
}

func runMove(_ *cobra.Command, args []string) error {
	_, cfgPath, err := config.Load()
	if errors.Is(err, config.ErrFirstRun) {
		fmt.Fprintf(os.Stderr, "First-time setup — created %s\nSet github_handle, then re-run.\n", cfgPath)
		os.Exit(3)
	}
	if err != nil {
		return err
	}

	issueNum, err := workflows.ParseIssueArg(args[0])
	if err != nil {
		return err
	}

	parts := strings.SplitN(flagRepo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("--repo owner/repo is required for move")
	}
	owner, repo := parts[0], parts[1]

	client, err := ghclient.New()
	if err != nil {
		return err
	}
	ch, err := cache.Load()
	if err != nil {
		return err
	}
	resolver := projects.NewResolver(client, ch)

	result, err := workflows.Move(client, resolver, workflows.MoveReq{
		Owner:  owner,
		Repo:   repo,
		Issue:  issueNum,
		Status: args[1],
	})
	if err != nil {
		return err
	}

	if flagJSON {
		return render.JSON(result)
	}
	fmt.Print(render.Move(result))
	return nil
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
go build -o gh-supercharged.exe .
./gh-supercharged.exe move 42 "In Progress" --repo owner/repo
```

Expected: `Moved #42 \`title\` → **In Progress**`

- [ ] **Step 5: Commit**

```bash
git add internal/workflows/move.go internal/render/move.go cmd/move.go
git commit -m "feat: move command with Projects v2 resolver and cache"
```

---

## M4 — New Issue

### Task 16: ghclient.ListLabels + GetIssueTemplates + CreateIssue

**Files:**
- Create: `internal/ghclient/labels.go`
- Create: `internal/ghclient/templates.go`
- Modify: `internal/ghclient/issues.go`

- [ ] **Step 1: Write `internal/ghclient/labels.go`**

```go
package ghclient

import "fmt"

type GHLabel struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func (c *Client) ListLabels(owner, repo string) ([]GHLabel, error) {
	var labels []GHLabel
	path := fmt.Sprintf("repos/%s/%s/labels?per_page=100", owner, repo)
	if err := c.REST.Get(path, &labels); err != nil {
		return nil, err
	}
	return labels, nil
}
```

- [ ] **Step 2: Write `internal/ghclient/templates.go`**

```go
package ghclient

import (
	"fmt"
)

type IssueTemplate struct {
	Name string `json:"name"`
}

func (c *Client) GetIssueTemplates(owner, repo string) ([]IssueTemplate, error) {
	var files []struct {
		Name string `json:"name"`
	}
	path := fmt.Sprintf("repos/%s/%s/contents/.github/ISSUE_TEMPLATE", owner, repo)
	err := c.REST.Get(path, &files)
	if err != nil {
		// No templates dir is a valid state — return empty slice
		return nil, nil
	}
	out := make([]IssueTemplate, len(files))
	for i, f := range files {
		out[i] = IssueTemplate{Name: f.Name}
	}
	return out, nil
}
```

- [ ] **Step 3: Add `CreateIssue` to `internal/ghclient/issues.go`**

```go
// Add at the end of issues.go

type CreateIssueRequest struct {
	Owner  string
	Repo   string
	Title  string
	Body   string
	Labels []string
}

type CreatedIssue struct {
	Number int    `json:"number"`
	URL    string `json:"html_url"`
	Title  string `json:"title"`
}

func (c *Client) CreateIssue(req CreateIssueRequest) (*CreatedIssue, error) {
	body := map[string]interface{}{
		"title":  req.Title,
		"body":   req.Body,
		"labels": req.Labels,
	}
	var created CreatedIssue
	path := fmt.Sprintf("repos/%s/%s/issues", req.Owner, req.Repo)
	if err := c.REST.Post(path, body, &created); err != nil {
		return nil, err
	}
	return &created, nil
}
```

- [ ] **Step 4: Commit**

```bash
git add internal/ghclient/labels.go internal/ghclient/templates.go internal/ghclient/issues.go
git commit -m "feat: ghclient ListLabels, GetIssueTemplates, CreateIssue"
```

---

### Task 17: workflows.NewIssue + render + cmd

**Files:**
- Create: `internal/workflows/newissue.go`
- Create: `internal/render/newissue.go`
- Create: `cmd/newissue.go`

- [ ] **Step 1: Write `internal/workflows/newissue.go`**

```go
package workflows

import (
	"fmt"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
)

type NewIssueOpts struct {
	Owner string
	Repo  string
}

// DraftIssue builds an IssueDraft from the user's description + repo context.
// It picks labels by matching description keywords against repo label names/descriptions.
func DraftIssue(c *ghclient.Client, description string, opts NewIssueOpts) (*IssueDraft, error) {
	labels, err := c.ListLabels(opts.Owner, opts.Repo)
	if err != nil {
		return nil, fmt.Errorf("list labels: %w", err)
	}
	templates, err := c.GetIssueTemplates(opts.Owner, opts.Repo)
	if err != nil {
		return nil, fmt.Errorf("get templates: %w", err)
	}

	picked := pickLabels(description, labels)
	template := pickTemplate(description, templates)

	title := buildTitle(description)
	body := buildBody(description, template)

	return &IssueDraft{
		Repo:     opts.Owner + "/" + opts.Repo,
		Title:    title,
		Labels:   picked,
		Template: template,
		Body:     body,
	}, nil
}

// CreateFromDraft submits the draft to GitHub.
func CreateFromDraft(c *ghclient.Client, draft *IssueDraft, owner, repo string) (*NewIssueResult, error) {
	created, err := c.CreateIssue(ghclient.CreateIssueRequest{
		Owner:  owner,
		Repo:   repo,
		Title:  draft.Title,
		Body:   draft.Body,
		Labels: draft.Labels,
	})
	if err != nil {
		return nil, err
	}
	return &NewIssueResult{URL: created.URL, Number: created.Number, Title: created.Title}, nil
}

func pickLabels(desc string, labels []ghclient.GHLabel) []string {
	desc = strings.ToLower(desc)
	var picked []string
	for _, l := range labels {
		name := strings.ToLower(l.Name)
		descText := strings.ToLower(l.Description)
		if strings.Contains(desc, name) || (descText != "" && strings.Contains(desc, descText)) {
			picked = append(picked, l.Name)
		}
	}
	return picked
}

func pickTemplate(desc string, templates []ghclient.IssueTemplate) string {
	desc = strings.ToLower(desc)
	for _, t := range templates {
		if strings.Contains(desc, strings.ToLower(strings.TrimSuffix(t.Name, ".md"))) {
			return t.Name
		}
	}
	if len(templates) > 0 {
		return templates[0].Name
	}
	return ""
}

func buildTitle(desc string) string {
	// Capitalise first letter, truncate at 72 chars
	if desc == "" {
		return "Untitled"
	}
	title := strings.ToUpper(desc[:1]) + desc[1:]
	if len(title) > 72 {
		title = title[:69] + "..."
	}
	return title
}

func buildBody(desc, template string) string {
	return fmt.Sprintf("## Problem / Goal\n\n%s\n\n## Expected behaviour\n\n<!-- describe expected behaviour -->\n\n## Context\n\n<!-- screenshots, logs, links -->\n\n## Acceptance criteria\n\n- [ ] <!-- testable criterion -->", desc)
}
```

- [ ] **Step 2: Write `internal/render/newissue.go`**

```go
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
```

- [ ] **Step 3: Write `cmd/newissue.go`**

```go
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/MalasataXD/gh-supercharged/internal/ghclient"
	"github.com/MalasataXD/gh-supercharged/internal/render"
	"github.com/MalasataXD/gh-supercharged/internal/workflows"
	"github.com/spf13/cobra"
)

var newIssueCmd = &cobra.Command{
	Use:   "new-issue <description>",
	Short: "Draft and create a well-formed GitHub issue",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runNewIssue,
}

var newIssueConfirm bool

func init() {
	newIssueCmd.Flags().BoolVar(&newIssueConfirm, "confirm", false, "Create without approval prompt (use after reviewing --json draft)")
	rootCmd.AddCommand(newIssueCmd)
}

func runNewIssue(_ *cobra.Command, args []string) error {
	_, cfgPath, err := config.Load()
	if errors.Is(err, config.ErrFirstRun) {
		fmt.Fprintf(os.Stderr, "First-time setup — created %s\nSet github_handle, then re-run.\n", cfgPath)
		os.Exit(3)
	}
	if err != nil {
		return err
	}

	parts := strings.SplitN(flagRepo, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("--repo owner/repo is required for new-issue")
	}
	owner, repo := parts[0], parts[1]

	description := strings.Join(args, " ")

	client, err := ghclient.New()
	if err != nil {
		return err
	}

	draft, err := workflows.DraftIssue(client, description, workflows.NewIssueOpts{Owner: owner, Repo: repo})
	if err != nil {
		return err
	}

	if !newIssueConfirm {
		// Show draft — skill uses --json here to get structured draft back
		if flagJSON {
			return render.JSON(draft)
		}
		fmt.Print(render.Draft(draft))
		return nil
	}

	// --confirm: create
	result, err := workflows.CreateFromDraft(client, draft, owner, repo)
	if err != nil {
		return err
	}
	if flagJSON {
		return render.JSON(result)
	}
	fmt.Print(render.NewIssue(result))
	return nil
}
```

- [ ] **Step 4: Build and smoke-test**

```bash
go build -o gh-supercharged.exe .
./gh-supercharged.exe new-issue "fix login button misalignment" --repo owner/repo --json
# review draft, then:
./gh-supercharged.exe new-issue "fix login button misalignment" --repo owner/repo --confirm
```

- [ ] **Step 5: Commit**

```bash
git add internal/workflows/newissue.go internal/render/newissue.go cmd/newissue.go
git commit -m "feat: new-issue command with two-phase draft→confirm flow"
```

---

## M5 — Ship

### Task 18: Utility subcommands (cache + config)

**Files:**
- Create: `cmd/cache.go`
- Create: `cmd/config.go`

- [ ] **Step 1: Write `cmd/cache.go`**

```go
package cmd

import (
	"fmt"

	"github.com/MalasataXD/gh-supercharged/internal/cache"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the projects cache",
}

var cacheClearCmd = &cobra.Command{
	Use:  "clear",
	RunE: func(_ *cobra.Command, _ []string) error {
		ch, err := cache.Load()
		if err != nil {
			return err
		}
		ch.Projects = map[string]cache.Entry{}
		if err := ch.Save(); err != nil {
			return err
		}
		fmt.Println("Cache cleared.")
		return nil
	},
}

var cacheShowCmd = &cobra.Command{
	Use:  "show",
	RunE: func(_ *cobra.Command, _ []string) error {
		ch, err := cache.Load()
		if err != nil {
			return err
		}
		return render.JSON(ch)
	},
}

func init() {
	cacheCmd.AddCommand(cacheClearCmd, cacheShowCmd)
	rootCmd.AddCommand(cacheCmd)
}
```

- [ ] **Step 2: Add `render` import to `cmd/cache.go`** (add `"github.com/MalasataXD/gh-supercharged/internal/render"` to imports)

- [ ] **Step 3: Write `cmd/config.go`**

```go
package cmd

import (
	"fmt"

	"github.com/MalasataXD/gh-supercharged/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gh-supercharged configuration",
}

var configPathCmd = &cobra.Command{
	Use:  "path",
	RunE: func(_ *cobra.Command, _ []string) error {
		dir, err := config.ConfigDir()
		if err != nil {
			return err
		}
		fmt.Println(dir)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:  "show",
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, _, err := config.Load()
		if err != nil {
			return err
		}
		return render.JSON(cfg)
	},
}

func init() {
	configCmd.AddCommand(configPathCmd, configShowCmd)
	rootCmd.AddCommand(configCmd)
}
```

- [ ] **Step 4: Add render import to `cmd/config.go`**

- [ ] **Step 5: Commit**

```bash
git add cmd/cache.go cmd/config.go
git commit -m "feat: cache and config utility subcommands"
```

---

### Task 19: CI + release workflows

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Write `.github/workflows/ci.yml`**

```yaml
name: CI
on:
  push:
    branches: ["**"]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - run: go test ./...
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
```

- [ ] **Step 2: Write `.github/workflows/release.yml`**

```yaml
name: Release
on:
  push:
    tags: ["v*"]

jobs:
  release:
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - name: Build binaries
        run: |
          TARGETS=(
            "linux/amd64"
            "linux/arm64"
            "darwin/amd64"
            "darwin/arm64"
            "windows/amd64"
          )
          mkdir dist
          for TARGET in "${TARGETS[@]}"; do
            OS="${TARGET%%/*}"
            ARCH="${TARGET##*/}"
            EXT=""
            [ "$OS" = "windows" ] && EXT=".exe"
            GOOS=$OS GOARCH=$ARCH go build \
              -o "dist/gh-supercharged_${OS}_${ARCH}${EXT}" .
          done
      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          files: dist/*
```

- [ ] **Step 3: Commit**

```bash
git add .github/
git commit -m "ci: add test and release workflows"
```

---

### Task 20: Rewrite docs/gh/SKILL.md

**Files:**
- Modify: `docs/gh/SKILL.md`

- [ ] **Step 1: Replace `docs/gh/SKILL.md` content**

Keep sections 1 (Bootstrap — now shortened), 2 (Intent Detection — unchanged), and replace section 3 with command invocations. Remove sections 4 (Cache) and 5 (Reference Files) since the binary owns those.

New section 3:

````markdown
## 3. Workflows

All workflows are executed by `gh supercharged`. Output is pre-formatted Markdown
unless `--json` is passed for interactive steps.

### Plate

```bash
gh supercharged plate [--repo <owner/repo>] [--owner <org>]
```
Output is Markdown — show verbatim.

---

### Digest

```bash
gh supercharged digest [<since>] [--owner <org>]
```
`<since>` accepts `7d`, `2w`, `last monday`, or `YYYY-MM-DD`.
Output is Markdown — show verbatim.

---

### Standup

```bash
gh supercharged standup
```
Output is Markdown — show verbatim. Prompt user to fill in blockers.

---

### Move

```bash
gh supercharged move <issue> "<status>" --repo <owner/repo>
```
Output: `Moved #N \`title\` → **status**`

---

### New Issue

**Step 1 — get draft:**
```bash
gh supercharged new-issue "<description>" --repo <owner/repo> --json
```
Show the returned `title`, `labels`, `template`, and `body` to the user.
Say: "Here's the draft — reply `ok` to create it, or tell me what to change."

**Step 2 — create (after user approval):**
```bash
gh supercharged new-issue "<description>" --repo <owner/repo> --confirm
```
Confirm with the returned issue URL.
````

- [ ] **Step 2: Verify the new skill file reads cleanly**

Open `docs/gh/SKILL.md` and confirm it's under ~100 lines, all commands match the implemented binary.

- [ ] **Step 3: Commit**

```bash
git add docs/gh/SKILL.md
git commit -m "docs: rewrite gh skill to shell out to gh supercharged"
```

---

### Task 21: End-to-end install test + README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Write `README.md`**

```markdown
# gh-supercharged

A `gh` extension implementing five GitHub workflows with dual Markdown/JSON output.

## Install

```bash
gh extension install MalasataXD/gh-supercharged
```

## First-time setup

```bash
gh supercharged config path   # shows config directory
# edit config.json, set github_handle
```

## Usage

```bash
gh supercharged plate                         # open issues assigned to you
gh supercharged digest 7d                     # last 7 days of closes + PRs
gh supercharged standup                       # yesterday/today/blockers
gh supercharged move 42 "In Progress" --repo owner/repo
gh supercharged new-issue "fix login" --repo owner/repo --json   # draft
gh supercharged new-issue "fix login" --repo owner/repo --confirm # create
```

Add `--json` to any command for structured output. Add `--verbose` for raw errors.

## Config

File: `<gh config dir>/extensions/supercharged/config.json`
Override directory: `GH_SC_CONFIG_DIR`

## Local alias

```bash
gh alias set sc supercharged
# now: gh sc plate
```
```

- [ ] **Step 2: Tag and push to trigger release**

```bash
git tag v0.1.0
git push origin main --tags
```

- [ ] **Step 3: Test install from fresh extension**

```bash
gh extension remove supercharged 2>/dev/null || true
gh extension install MalasataXD/gh-supercharged
gh supercharged plate
gh supercharged config show
gh supercharged digest 7d --json
```

Expected: all three commands return valid output without errors.

- [ ] **Step 4: Final commit**

```bash
git add README.md
git commit -m "docs: add README with install and usage"
```

---

## Verification Checklist

At end of each milestone:
- `go test ./...` — PASS
- `go build -o gh-supercharged.exe .` — no errors
- Smoke-test the new command(s) with and without `--json`

After M5 (full acceptance):
1. Fresh install via `gh extension install MalasataXD/gh-supercharged` succeeds on Windows.
2. `gh supercharged config show` reports the correct path.
3. `gh supercharged plate` output matches what the old skill produced.
4. `gh supercharged digest 7d` — Markdown and `--json` agree on content.
5. `gh supercharged move <issue> "In Progress" --repo owner/repo` succeeds on cold cache, then warm cache.
6. `gh supercharged new-issue "..." --repo owner/repo --json` → review draft → `--confirm` creates the issue.
7. Claude Code triggers the `gh` skill on "what's on my plate" → skill shells out to `gh supercharged plate` successfully.
