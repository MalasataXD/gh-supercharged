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
	if err := os.MkdirAll(dir, 0o700); err != nil {
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
