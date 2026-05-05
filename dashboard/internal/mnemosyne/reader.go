package mnemosyne

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Reader struct {
	dir      string
	mu       sync.RWMutex
	skills   []Skill
	loadedAt time.Time
	ttl      time.Duration
}

func NewReader(dir string) *Reader {
	return &Reader{dir: dir, ttl: 5 * time.Minute}
}

func (r *Reader) Skills() ([]Skill, error) {
	r.mu.RLock()
	if !r.loadedAt.IsZero() && time.Since(r.loadedAt) < r.ttl {
		skills := make([]Skill, len(r.skills))
		copy(skills, r.skills)
		r.mu.RUnlock()
		return skills, nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	// double-check after acquiring write lock
	if !r.loadedAt.IsZero() && time.Since(r.loadedAt) < r.ttl {
		skills := make([]Skill, len(r.skills))
		copy(skills, r.skills)
		return skills, nil
	}
	loaded, err := loadFromDir(r.dir)
	if err != nil {
		return nil, err
	}
	r.skills = loaded
	r.loadedAt = time.Now()
	skills := make([]Skill, len(loaded))
	copy(skills, loaded)
	return skills, nil
}

func loadFromDir(dir string) ([]Skill, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil // missing dir = empty, not error
	}
	if err != nil {
		return nil, err
	}
	var skills []Skill
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		s, err := parseSkill(string(data), path)
		if err != nil {
			continue
		}
		skills = append(skills, s)
	}
	return skills, nil
}

// parseSkill splits --- frontmatter --- from body and parses YAML.
func parseSkill(content, path string) (Skill, error) {
	var s Skill
	s.FilePath = path
	// strip frontmatter
	if strings.HasPrefix(content, "---") {
		rest := content[3:]
		end := strings.Index(rest, "\n---")
		if end != -1 {
			fm := rest[:end]
			if err := yaml.Unmarshal([]byte(fm), &s); err != nil {
				return s, err
			}
			s.Body = strings.TrimPrefix(rest[end+4:], "\n")
		} else {
			s.Body = content
		}
	} else {
		s.Body = content
	}
	if s.Name == "" {
		// fallback: use filename without extension
		base := filepath.Base(path)
		s.Name = strings.TrimSuffix(base, ".md")
	}
	return s, nil
}
