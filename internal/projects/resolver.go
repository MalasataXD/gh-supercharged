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
// projectNodeID is the project's global node ID; fieldName is e.g. "Status";
// optionName is the desired value (case-insensitive). For non-single-select
// fields pass optionName="".
func (r *Resolver) Resolve(projectNodeID, fieldName, optionName string) (*IDs, error) {
	ids, err := r.resolve(projectNodeID, fieldName, optionName)
	if err != nil && isStaleError(err) {
		r.cache.Invalidate(projectNodeID)
		_ = r.cache.Save()
		ids, err = r.resolve(projectNodeID, fieldName, optionName)
	}
	return ids, err
}

func (r *Resolver) resolve(projectNodeID, fieldName, optionName string) (*IDs, error) {
	entry, hit := r.cache.Get(projectNodeID)

	if !hit {
		if err := r.populate(projectNodeID); err != nil {
			return nil, err
		}
		entry, _ = r.cache.Get(projectNodeID)
	}

	ids := &IDs{ProjectNodeID: entry.ProjectID}

	if !strings.EqualFold(fieldName, "Status") {
		return nil, fmt.Errorf("field %q not yet supported — only Status is cached", fieldName)
	}
	ids.FieldID = entry.Fields.Status.ID

	if optionName != "" {
		for name, optID := range entry.Fields.Status.Options {
			if strings.EqualFold(name, optionName) {
				ids.OptionID = optID
				break
			}
		}
		if ids.OptionID == "" {
			available := make([]string, 0, len(entry.Fields.Status.Options))
			for name := range entry.Fields.Status.Options {
				available = append(available, name)
			}
			return nil, fmt.Errorf("status %q not found — available: %s", optionName, strings.Join(available, ", "))
		}
	}

	return ids, nil
}

func (r *Resolver) populate(projectNodeID string) error {
	fields, err := r.client.ProjectFields(projectNodeID)
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

	r.cache.Set(projectNodeID, entry)
	return r.cache.Save()
}

func isStaleError(err error) bool {
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "not found") || strings.Contains(s, "does not exist")
}
