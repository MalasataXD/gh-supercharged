# Project V2 ID Resolution

GitHub Projects v2 uses opaque GraphQL node IDs for everything. This reference
documents the exact commands to resolve them so `gh project item-edit` can run.

## Full discovery chain (cold cache)

### 1. Get item node ID + project membership

```bash
gh issue view <issue-number> -R <owner/repo> \
  --json number,title,projectItems,url
```

Example output:
```json
{
  "number": 42,
  "title": "Fix login button misalignment",
  "projectItems": [
    {
      "id": "PVTI_lADOB...",
      "project": {
        "id": "PVT_kwDOB...",
        "number": 3,
        "title": "My Roadmap"
      }
    }
  ]
}
```

- `projectItems[0].id` → item node ID (used in `--id` for `item-edit`)
- `projectItems[0].project.id` → project node ID (used in `--project-id`)
- `projectItems[0].project.number` → project number (used in `field-list`)

If `projectItems` is empty the issue is not on any project board — add it first
with `gh project item-add <project-number> --owner <owner> --url <issue-url>`,
then re-run step 1.

### 2. Get field and option IDs

```bash
gh project field-list <project-number> \
  --owner <owner> \
  --format json \
  --limit 50
```

Example output:
```json
{
  "fields": [
    {
      "id": "PVTSSF_lADOB...",
      "name": "Status",
      "type": "ProjectV2SingleSelectField",
      "options": [
        {"id": "abc123", "name": "Todo"},
        {"id": "def456", "name": "In Progress"},
        {"id": "ghi789", "name": "Done"}
      ]
    },
    {
      "id": "PVTF_lADOB...",
      "name": "Priority",
      "type": "ProjectV2SingleSelectField",
      "options": [
        {"id": "ppp111", "name": "High"},
        {"id": "ppp222", "name": "Medium"},
        {"id": "ppp333", "name": "Low"}
      ]
    }
  ]
}
```

- Find the field by `name` (e.g. `"Status"`).
- Field `id` → used in `--field-id`.
- Match the desired status by `options[].name` → `options[].id` used in `--single-select-option-id`.

Do a **case-insensitive, trimmed** match when resolving the user's status string
to an option name. Raise an error if no option matches; list available options.

### 3. Update the field

```bash
gh project item-edit \
  --id <item-node-id> \
  --project-id <project-node-id> \
  --field-id <status-field-id> \
  --single-select-option-id <option-id>
```

One flag combination per field type:

| Field type | Flag |
|---|---|
| Single-select (Status, Priority) | `--single-select-option-id` |
| Text | `--text` |
| Number | `--number` |
| Date | `--date YYYY-MM-DD` |
| Iteration | `--iteration-id` |
| Clear any field | `--clear` |

### 4. Write to cache

After successful resolution, write the entry:
```json
{
  "projects": {
    "<owner>/<project-number>": {
      "project_id": "<PVT_...>",
      "fields": {
        "Status": {
          "id": "<PVTSSF_...>",
          "options": {
            "Todo": "<option-id>",
            "In Progress": "<option-id>",
            "Done": "<option-id>"
          }
        }
      },
      "cached_at": "<ISO-8601-timestamp>"
    }
  }
}
```

## Cache invalidation

Remove (or overwrite) an entry when:
- `item-edit` returns an error containing "not found" or "does not exist"
- `field-list` returns an option set that differs from what's cached
- The user explicitly says "clear cache" or "refresh project IDs"

After invalidation, re-run the full discovery chain before retrying.

## GraphQL fallback

When typed commands fail or you need bulk operations:

```bash
gh api graphql -f query='
  mutation UpdateStatus($item: ID!, $project: ID!, $field: ID!, $option: ID!) {
    updateProjectV2ItemFieldValue(input: {
      projectId: $project
      itemId: $item
      fieldId: $field
      value: { singleSelectOptionId: $option }
    }) {
      projectV2Item { id }
    }
  }
' \
  -F item="<item-node-id>" \
  -F project="<project-node-id>" \
  -F field="<field-id>" \
  -F option="<option-id>"
```
