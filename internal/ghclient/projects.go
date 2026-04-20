package ghclient

import (
	"fmt"
)

// ViewIssueProjects fetches an issue's project membership via GraphQL.
func (c *Client) ViewIssueProjects(owner, repo string, number int) (*IssueProjectsResponse, error) {
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
	vars := map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}

	var data struct {
		Repository struct {
			Issue struct {
				Number int    `json:"number"`
				Title  string `json:"title"`
				URL    string `json:"url"`
				ProjectItems struct {
					Nodes []ProjectItem `json:"nodes"`
				} `json:"projectItems"`
			} `json:"issue"`
		} `json:"repository"`
	}

	if err := c.GQL.Do(query, vars, &data); err != nil {
		return nil, fmt.Errorf("ViewIssueProjects: %w", err)
	}

	iss := data.Repository.Issue
	resp := &IssueProjectsResponse{
		Number:       iss.Number,
		Title:        iss.Title,
		URL:          iss.URL,
		ProjectItems: iss.ProjectItems.Nodes,
	}
	return resp, nil
}

type ProjectItem struct {
	ID      string `json:"id"`
	Project struct {
		ID     string `json:"id"`
		Number int    `json:"number"`
		Title  string `json:"title"`
	} `json:"project"`
}

type IssueProjectsResponse struct {
	Number       int
	Title        string
	URL          string
	ProjectItems []ProjectItem
}

// FieldListResponse holds project fields returned by ProjectFields.
type FieldListResponse struct {
	Fields []ProjectField
}

type ProjectField struct {
	ID      string
	Name    string
	Options []struct {
		ID   string
		Name string
	}
}

// ProjectFields returns all fields for a project via GraphQL, resolved by node ID.
// Accepts both user- and org-owned projects.
func (c *Client) ProjectFields(projectNodeID string) (*FieldListResponse, error) {
	query := `query($id:ID!) {
		node(id:$id) {
			... on ProjectV2 {
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
	vars := map[string]interface{}{"id": projectNodeID}

	var data struct {
		Node struct {
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
		} `json:"node"`
	}

	if err := c.GQL.Do(query, vars, &data); err != nil {
		return nil, fmt.Errorf("ProjectFields: %w", err)
	}

	resp := &FieldListResponse{}
	for _, n := range data.Node.Fields.Nodes {
		f := ProjectField{ID: n.ID, Name: n.Name}
		for _, opt := range n.Options {
			f.Options = append(f.Options, struct {
				ID   string
				Name string
			}{ID: opt.ID, Name: opt.Name})
		}
		resp.Fields = append(resp.Fields, f)
	}
	return resp, nil
}

// UpdateFieldRequest describes a Projects v2 field mutation.
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
	type gqlRequest struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}

	vars := map[string]interface{}{
		"project": req.ProjectNodeID,
		"item":    req.ItemNodeID,
		"field":   req.FieldID,
	}

	var mutation string
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
