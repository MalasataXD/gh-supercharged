# gh-supercharged Usage Guide

A `gh` CLI extension that adds five productivity workflows for GitHub: viewing your plate, generating digests and standups, moving issues on project boards, and drafting new issues.

---

## Installation

```bash
gh extension install MalasataXD/gh-supercharged
```

**Optional alias** to shorten commands:

```bash
gh alias set sc supercharged
# Then use: gh sc plate
```

---

## Configuration

On first run, the extension creates a config file at:

```
<gh config dir>/extensions/supercharged/config.json
```

To find the exact path:

```bash
gh supercharged config path
```

To view the current config:

```bash
gh supercharged config show
```

### config.json fields

| Field | Default | Description |
|---|---|---|
| `github_handle` | `""` | Your GitHub username — required for plate/standup |
| `digest_window_days` | `7` | Default lookback window for digest |
| `standup_format` | see below | Template string for standup output |

Default standup format:

```
## Yesterday\n{closed}\n\n## Today\n{open}\n\n## Blockers\n{blockers}
```

Override config directory with the `GH_SC_CONFIG_DIR` environment variable.

---

## Commands

### `plate` — Your open issues

Shows all open issues assigned to you, grouped by repository.

```bash
gh supercharged plate                  # scoped to current repo
gh supercharged plate --full           # all repos
gh supercharged plate --owner myorg    # filter by org
gh supercharged plate --repo owner/repo
gh supercharged plate --json
```

---

### `digest` — Closed issues & merged PRs

Summarizes what shipped in a time window.

```bash
gh supercharged digest                 # default window (7 days)
gh supercharged digest 7d
gh supercharged digest 2w
gh supercharged digest last monday
gh supercharged digest 2026-04-10
gh supercharged digest --full          # all repos
gh supercharged digest --json
```

**Accepted date formats:**

| Format | Example | Meaning |
|---|---|---|
| N days | `7d` | Last N days |
| N weeks | `2w` | Last N weeks |
| Named day | `last monday` | Most recent occurrence |
| ISO date | `2026-04-10` | Exact start date |

---

### `standup` — Daily standup summary

Combines yesterday's closed work with today's open plate. Output format is controlled by `standup_format` in config.

```bash
gh supercharged standup                # current repo
gh supercharged standup --full
gh supercharged standup --json
```

---

### `move` — Move an issue to a project status

Updates an issue's status on a GitHub Projects v2 board.

```bash
gh supercharged move 42 "In Progress" --repo owner/repo
gh supercharged move "#42" "Done" --repo owner/repo
gh supercharged move 42 "In Progress" --repo owner/repo --json
```

- Status matching is case-insensitive.
- Project field metadata is cached locally; run `gh supercharged cache clear` if statuses seem stale.


## Cache management

The extension caches GitHub Projects v2 field metadata to avoid repeated API calls.

```bash
gh supercharged cache show    # inspect the cache
gh supercharged cache clear   # force refresh
```

Cache file: `<gh config dir>/extensions/supercharged/projects.json`

---

## Global flags

Available on every command:

| Flag | Description |
|---|---|
| `--json` | Structured JSON output instead of Markdown |
| `--repo owner/repo` | Target a specific repository |
| `--owner org` | Filter by organization or owner |
| `--verbose` | Show raw API errors for debugging |
