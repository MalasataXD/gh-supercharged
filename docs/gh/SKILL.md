---
name: gh
description: >
  GitHub CLI workflows for issues and projects. Triggers automatically when
  the user asks about their open issues, upcoming work, what's on their plate,
  moving an issue to a new status, summarizing work done in a period, standup
  prep, drafting a new issue, tagging an issue, writing a well-formed issue,
  or anything phrased as "look at my issues", "what did we ship", "new issue",
  "move #<n> to <status>", "standup", "digest", "triage", or similar.
---

# gh — GitHub CLI Skill

Orchestrate GitHub issue and project workflows through `gh` CLI. Every workflow
reads config and cache, then executes the appropriate `gh` commands and formats
output as clean Markdown.

## 1. Bootstrap

### Load config

Read `${CLAUDE_PLUGIN_ROOT}/skills/gh/config.json`.

If the file does not exist:
1. Copy `${CLAUDE_PLUGIN_ROOT}/skills/gh/config.example.json` to `config.json`.
2. Tell the user: "First time setup — I created `config.json` inside the `gh` skill. Please set `github_handle` to your GitHub username, then re-run."
3. Stop.

### Load cache

Read `${CLAUDE_PLUGIN_ROOT}/skills/gh/cache.json`.

If the file does not exist, treat the cache as `{"projects": {}}`.

---

## 2. Intent Detection

Map the user's request to exactly one workflow:

| User says | Workflow |
|---|---|
| "what's on my plate", "my open issues", "upcoming work" | **Plate** |
| "digest", "what did we ship", "summarize last week / since <date>" | **Digest** |
| "standup", "what to report" | **Standup** |
| "move #<n> to <status>", "set status to", "mark as done/in progress" | **Move** |
| "new issue", "draft issue", "create issue", "open an issue", "write issue" | **New Issue** |

If the intent is ambiguous, state your interpretation and proceed. Do not ask for confirmation unless there is a genuine fork (e.g., the user mentions both "digest" and "move" — ask which one to do first).

---

## 3. Workflows

### Plate — open work assigned to me

Show open issues assigned to the current user, sorted by last update.

```bash
gh search issues \
  --assignee @me \
  --state open \
  --sort updated \
  --order desc \
  --limit 50 \
  --json number,title,repository,labels,milestone,updatedAt,url
```

Group results by `repository.nameWithOwner`. For each repo, list issues with
their labels and milestone (if set). Summarize total count at the top.

If the user specifies a repo (`--repo owner/name`) or an owner, add `--repo`
or `--owner` flags accordingly.

---

### Digest — period summary

Defaults to `config.digest_window_days` days if no date is given. Accept
relative input like `7d`, `2w`, `last monday`, or ISO dates like `2026-04-10`.

Convert to an ISO date for the `--closed` / `--merged` filters.

**Closed issues involving me:**
```bash
gh search issues \
  --involves <config.github_handle> \
  --state closed \
  --closed ">=<since-date>" \
  --sort updated \
  --limit 100 \
  --json number,title,repository,labels,closedAt,url
```

**Merged PRs authored by me:**
```bash
gh search prs \
  --author <config.github_handle> \
  --state merged \
  --merged ">=<since-date>" \
  --sort updated \
  --limit 100 \
  --json number,title,repository,labels,closedAt,url
```

Format as Markdown:

```
## Digest: <since-date> → today

### <owner/repo>
- Closed issues: #<n> <title> · <labels>
- Merged PRs: #<n> <title>

### <owner/repo>
...

**Total:** <X> issues closed · <Y> PRs merged
```

If results are zero, say so and suggest widening the window.

---

### Standup — yesterday / today / blockers

Closed issues and PRs since yesterday (1-day digest), plus today's open plate.

**Yesterday's closes:** run Digest with `--closed ">=<yesterday-ISO>"`.

**Today's open plate:** run Plate limited to 20 items.

Format using `config.standup_format`. Replace tokens:
- `{closed}` — bulleted list of closed items (title + repo)
- `{open}` — bulleted list of top open items (title + repo)
- `{blockers}` — leave blank; prompt the user to fill it in

---

### Move — change a project status field

See `${CLAUDE_PLUGIN_ROOT}/skills/gh/references/project-ids.md` for the full
ID resolution walkthrough. Summary:

**Step 1 — resolve the issue and its project membership:**
```bash
gh issue view <number> -R <owner/repo> \
  --json number,title,projectItems,url
```
`projectItems` contains `id` (the item node ID) and nested project info.
If the issue is in multiple projects, ask the user which one.

**Step 2 — resolve field and option IDs (use cache when available):**

Cache key: `<owner>/<project-number>` (e.g. `"MalasataXD/3"`).

If key is in cache: read `project_id`, `fields.Status.id`, and
`fields.Status.options.<status-name>` directly.

If key is missing or stale: run the discovery chain in `references/project-ids.md`,
then write the result to cache before proceeding.

**Step 3 — update the field:**
```bash
gh project item-edit \
  --id <item-node-id> \
  --project-id <project-node-id> \
  --field-id <status-field-id> \
  --single-select-option-id <option-id>
```

On success: confirm with "Moved #<n> `<title>` → **<status>**."

On error containing "not found" or "does not exist": invalidate the cache
entry, re-run Step 2, and retry once. If it fails again, report the raw error.

**Date, iteration, text, and number fields** are supported too — use
`--date`, `--iteration-id`, `--text`, `--number` respectively. Detect the
field type from cache or `field-list` output before choosing the flag.

---

### New Issue — draft and create a well-formed issue

**Step 1 — establish target repo:**
- If the user is inside a git repo with a GitHub remote, run `gh repo view --json nameWithOwner` to get it.
- Otherwise ask: "Which repo? (owner/repo format)"

**Step 2 — read repo label taxonomy:**
```bash
gh label list -R <owner/repo> \
  --json name,description,color \
  --limit 100
```

**Step 3 — read available issue templates:**
```bash
gh api repos/<owner>/<repo>/contents/.github/ISSUE_TEMPLATE \
  --jq '.[].name' 2>/dev/null || echo "(no templates)"
```

**Step 4 — draft the issue:**
Based on the user's description, the label list, and available templates:
1. Pick the best matching template (if any).
2. Pick labels that fit the type and scope (bug, feature, enhancement, etc.).
3. Draft a title (imperative, ≤ 72 chars).
4. Draft a body following the template structure, or a sensible default:
   - **Problem / Goal** — what is broken or what needs to exist
   - **Expected behaviour** — what should happen
   - **Context** — screenshots, logs, links (prompt user if needed)
   - **Acceptance criteria** — bulleted, testable

**Step 5 — show draft and get approval:**
Present the title, labels, and body to the user. Say:
"Here's the draft — reply `ok` to create it, or tell me what to change."

**Step 6 — create:**
```bash
gh issue create \
  -R <owner/repo> \
  -t "<title>" \
  -l "<label1>" -l "<label2>" \
  -F -   <<< "<body>"
```

Or, if a template was selected and it matches a `--template` name exactly,
use `-T <template-name>` instead of `-F`.

Confirm with the created issue URL.

---

## 4. Cache Format

The cache file lives at `${CLAUDE_PLUGIN_ROOT}/skills/gh/cache.json` (gitignored).
Schema:

```json
{
  "projects": {
    "<owner>/<project-number>": {
      "project_id": "PVT_...",
      "fields": {
        "Status": {
          "id": "PVTSSF_...",
          "options": {
            "Todo": "option-node-id",
            "In Progress": "option-node-id",
            "Done": "option-node-id"
          }
        }
      },
      "cached_at": "2026-04-17T12:00:00Z"
    }
  }
}
```

Write the full cache file back whenever an entry is added or invalidated.

---

## 5. Reference Files

Load these on demand — do not preload both:

- `references/gh-cheatsheet.md` — full flag reference for each workflow command
- `references/project-ids.md` — step-by-step ID discovery chain for Projects v2

Load `project-ids.md` when executing a Move and the cache is cold.
Load `gh-cheatsheet.md` when the user asks about flags or when a command errors
unexpectedly.
