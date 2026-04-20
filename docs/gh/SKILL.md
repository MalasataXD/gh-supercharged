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

Orchestrate GitHub issue and project workflows through the `gh supercharged`
extension. Every workflow is a single command invocation with pre-formatted
Markdown output (or `--json` for structured data in interactive steps).

## 1. Bootstrap

Verify the extension is installed:

```bash
gh extension list | grep supercharged
```

If missing, tell the user:
> "Run `gh extension install MalasataXD/gh-supercharged` to install the extension, then re-run."

Stop.

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

If the intent is ambiguous, state your interpretation and proceed.

---

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

`<since>` accepts `7d`, `2w`, `last monday`, or `YYYY-MM-DD`. Defaults to
`digest_window_days` from config when omitted.

Output is Markdown — show verbatim. If results are zero, suggest widening the window.

---

### Standup

```bash
gh supercharged standup
```

Output is Markdown — show verbatim. Prompt the user to fill in the blockers section.

---

### Move

```bash
gh supercharged move <issue> "<status>" --repo <owner/repo>
```

`<issue>` accepts `#42` or `42`. `<status>` is case-insensitive and matched
against the project's Status field options.

Output: `Moved #N \`title\` → **status**`

If the issue is in multiple projects, add `--project <number>` to disambiguate.

---

### New Issue

**Step 1 — get draft (structured output for review):**

```bash
gh supercharged new-issue "<description>" --repo <owner/repo> --json
```

Show the returned `title`, `labels`, `template`, and `body` to the user. Say:
"Here's the draft — reply `ok` to create it, or tell me what to change."

**Step 2 — create (after user approval):**

```bash
gh supercharged new-issue "<description>" --repo <owner/repo> --confirm
```

Confirm with the returned issue URL.

---

## 4. Reference Files

Load these on demand — do not preload:

- `references/gh-cheatsheet.md` — full `gh` flag reference (consult when a command errors unexpectedly)
- `references/project-ids.md` — Projects v2 ID discovery chain (background context; the extension handles this automatically)
