package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/cli/go-gh/v2/pkg/config"
)

type TeamCache struct {
	Teams    []string  `json:"teams"`
	CachedAt time.Time `json:"cached_at"`
}

type Store struct {
	path string
}

func NewStore() (*Store, error) {
	cacheDir := config.CacheDir()
	path := filepath.Join(cacheDir, "gh-own", "teams.json")
	return &Store{path: path}, nil
}

func NewStoreWithPath(path string) *Store {
	return &Store{path: path}
}

func (s *Store) ReadTeams(ttl time.Duration) ([]string, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, nil // Treat as cache miss (file doesn't exist or can't be read)
	}

	var cache TeamCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, nil // Treat as cache miss (corrupted JSON)
	}

	if isExpired(cache.CachedAt, ttl) {
		return nil, nil
	}

	return cache.Teams, nil
}

func (s *Store) WriteTeams(teams []string) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	cache := TeamCache{
		Teams:    teams,
		CachedAt: time.Now(),
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return err
	}

	return os.Rename(tempPath, s.path)
}

func isExpired(cachedAt time.Time, ttl time.Duration) bool {
	return time.Since(cachedAt) > ttl
}
