# gh CLI Cheatsheet

Quick reference for every `gh` command used in this skill. All JSON output
fields and useful `--jq` expressions included.

## Issues

### List open issues in a repo
```bash
gh issue list -R <owner/repo> \
  --state open \
  --json number,title,labels,assignees,milestone,createdAt,updatedAt,url \
  --limit 50
```

### View a single issue (with project membership)
```bash
gh issue view <number> -R <owner/repo> \
  --json number,title,body,labels,assignees,milestone,projectItems,state,url
```

### Create an issue (body from stdin)
```bash
gh issue create -R <owner/repo> \
  -t "<title>" \
  -l "<label1>" -l "<label2>" \
  -a @me \
  -F -   <<< "<body-text>"
```

### Create an issue using a template
```bash
gh issue create -R <owner/repo> \
  -T "<template-name>" \
  -t "<title>" \
  -l "<label>" \
  --web
```
Use `--web` to open the browser pre-filled when the template needs interactive
fields that the CLI can't fill automatically.

### Edit an issue
```bash
gh issue edit <number> -R <owner/repo> \
  --add-label "<label>" \
  --remove-label "<old-label>" \
  --add-assignee @me \
  --milestone "<milestone-name>" \
  --title "<new-title>"
```

### Close / reopen
```bash
gh issue close <number> -R <owner/repo>
gh issue reopen <number> -R <owner/repo>
```

---

## Labels

### List all labels in a repo
```bash
gh label list -R <owner/repo> \
  --json name,description,color \
  --limit 200
```

### Create a label
```bash
gh label create "<name>" \
  -R <owner/repo> \
  --description "<desc>" \
  --color "<hex-without-hash>"
```

### Clone labels from another repo
```bash
gh label clone <source-owner/repo> -R <target-owner/repo>
```

---

## Search

### Open issues assigned to me across all repos
```bash
gh search issues \
  --assignee @me \
  --state open \
  --sort updated \
  --order desc \
  --limit 50 \
  --json number,title,repository,labels,updatedAt,url
```

### Issues closed in a period (involving me)
```bash
gh search issues \
  --involves <handle> \
  --state closed \
  --closed ">=<YYYY-MM-DD>" \
  --sort updated \
  --limit 100 \
  --json number,title,repository,labels,closedAt,url
```

Date qualifiers: `>=`, `<=`, `YYYY-MM-DD..YYYY-MM-DD` (range).

### PRs merged by me in a period
```bash
gh search prs \
  --author <handle> \
  --state merged \
  --merged ">=<YYYY-MM-DD>" \
  --sort updated \
  --limit 100 \
  --json number,title,repository,labels,closedAt,url
```

### Issues missing a label or assignee
```bash
gh issue list -R <owner/repo> \
  --search "no:label no:assignee" \
  --state open \
  --json number,title,createdAt,url
```

### Cross-repo search with owner scope
```bash
gh search issues \
  --owner <org-or-user> \
  --state open \
  --sort created \
  --limit 100 \
  --json number,title,repository,labels,url
```

---

## Projects (v2)

### List projects for an owner
```bash
gh project list --owner <owner> --format json --limit 50
```

### List fields in a project
```bash
gh project field-list <project-number> \
  --owner <owner> \
  --format json \
  --limit 50
```

### List items in a project
```bash
gh project item-list <project-number> \
  --owner <owner> \
  --format json \
  --limit 100
```

### Update a single-select field (e.g., Status)
```bash
gh project item-edit \
  --id <item-node-id> \
  --project-id <project-node-id> \
  --field-id <field-id> \
  --single-select-option-id <option-id>
```

### Add an issue to a project
```bash
gh project item-add <project-number> \
  --owner <owner> \
  --url <issue-url>
```

### Archive an item
```bash
gh project item-archive <project-number> \
  --owner <owner> \
  --id <item-node-id>
```

---

## API / GraphQL

### REST — read issue templates
```bash
gh api repos/<owner>/<repo>/contents/.github/ISSUE_TEMPLATE \
  --jq '.[].name'
```

### REST — get repo default branch / metadata
```bash
gh api repos/<owner>/<repo> \
  --jq '{default_branch,description,visibility}'
```

### GraphQL — paginate all project items
```bash
gh api graphql --paginate --slurp -f query='
  query($owner: String!, $number: Int!, $endCursor: String) {
    user(login: $owner) {
      projectV2(number: $number) {
        items(first: 100, after: $endCursor) {
          nodes { id content { ... on Issue { number title } } }
          pageInfo { hasNextPage endCursor }
        }
      }
    }
  }
' -F owner=<owner> -F number=<project-number>
```

---

## Useful jq snippets

```bash
# Group search results by repo
--jq 'group_by(.repository.nameWithOwner) | map({repo: .[0].repository.nameWithOwner, issues: map({number,title})})'

# Count items per repo
--jq '[group_by(.repository.nameWithOwner)[] | {repo: .[0].repository.nameWithOwner, count: length}]'

# Extract label names
--jq '.labels | map(.name) | join(", ")'

# Filter issues updated in the last 24h
--jq '[.[] | select(.updatedAt >= (now - 86400 | todate))]'
```
