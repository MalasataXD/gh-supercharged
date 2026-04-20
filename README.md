# gh-supercharged

A `gh` extension implementing five GitHub workflows with dual Markdown/JSON output.

## Install

```bash
gh extension install MalasataXD/gh-supercharged
```

## First-time setup

```bash
gh supercharged config path   # shows config directory
# edit config.json in that directory, set github_handle
```

## Usage

```bash
gh supercharged plate                                          # open issues assigned to you
gh supercharged digest 7d                                      # last 7 days of closes + merged PRs
gh supercharged standup                                        # yesterday/today/blockers
gh supercharged move 42 "In Progress" --repo owner/repo       # move issue to status
gh supercharged new-issue "fix login" --repo owner/repo --json    # draft issue
gh supercharged new-issue "fix login" --repo owner/repo --confirm # create issue
```

Add `--json` to any command for structured output. Add `--verbose` for raw errors.

## Config

File: `<gh config dir>/extensions/supercharged/config.json`

Override directory: `GH_SC_CONFIG_DIR`

```json
{
  "github_handle": "your-github-username",
  "digest_window_days": 7,
  "standup_format": "## Yesterday\n{closed}\n\n## Today\n{open}\n\n## Blockers\n{blockers}"
}
```

## Cache

Project ID cache: `<gh config dir>/extensions/supercharged/projects.json`

```bash
gh supercharged cache show    # inspect cache
gh supercharged cache clear   # invalidate all entries
```

## Local alias

```bash
gh alias set sc supercharged
# now: gh sc plate
```
