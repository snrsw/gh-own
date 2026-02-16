package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStore_PathSuffix(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() error: %v", err)
	}

	expected := filepath.Join("gh-own", "teams.json")
	if got := store.path; !filepath.IsAbs(got) {
		t.Errorf("NewStore().path = %q, want absolute path", got)
	}
	if got := store.path; len(got) < len(expected) || got[len(got)-len(expected):] != expected {
		t.Errorf("NewStore().path = %q, want suffix %q", got, expected)
	}
}

func TestIsExpired_WithinTTL(t *testing.T) {
	cachedAt := time.Now().Add(-1 * time.Hour)
	ttl := 6 * time.Hour

	if isExpired(cachedAt, ttl) {
		t.Errorf("isExpired(%v, %v) = true, want false (within TTL)", cachedAt, ttl)
	}
}

func TestIsExpired_BeyondTTL(t *testing.T) {
	cachedAt := time.Now().Add(-7 * time.Hour)
	ttl := 6 * time.Hour

	if !isExpired(cachedAt, ttl) {
		t.Errorf("isExpired(%v, %v) = false, want true (beyond TTL)", cachedAt, ttl)
	}
}

func TestWriteTeams_Success(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStoreWithPath(filepath.Join(tmpDir, "teams.json"))

	teams := []string{"org/team-a", "org/team-b"}
	err := store.WriteTeams(teams)
	if err != nil {
		t.Fatalf("WriteTeams() error: %v", err)
	}

	// Verify file exists and contains correct data
	data, err := os.ReadFile(filepath.Join(tmpDir, "teams.json"))
	if err != nil {
		t.Fatalf("failed to read cache file: %v", err)
	}

	var cache TeamCache
	if err := json.Unmarshal(data, &cache); err != nil {
		t.Fatalf("failed to unmarshal cache: %v", err)
	}

	if len(cache.Teams) != 2 {
		t.Errorf("got %d teams, want 2", len(cache.Teams))
	}
	if cache.Teams[0] != "org/team-a" || cache.Teams[1] != "org/team-b" {
		t.Errorf("got teams %v, want [org/team-a org/team-b]", cache.Teams)
	}
	if cache.CachedAt.IsZero() {
		t.Error("CachedAt is zero, want non-zero timestamp")
	}
}

func TestWriteTeams_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStoreWithPath(filepath.Join(tmpDir, "subdir", "teams.json"))

	teams := []string{"org/team-a"}
	err := store.WriteTeams(teams)
	if err != nil {
		t.Fatalf("WriteTeams() error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(filepath.Join(tmpDir, "subdir")); os.IsNotExist(err) {
		t.Error("directory was not created")
	}
}

func TestReadTeams_CacheHit(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStoreWithPath(filepath.Join(tmpDir, "teams.json"))

	// Write cache file
	expectedTeams := []string{"org/team-a", "org/team-b"}
	cache := TeamCache{
		Teams:    expectedTeams,
		CachedAt: time.Now(),
	}
	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatalf("failed to marshal cache: %v", err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "teams.json"), data, 0600)
	if err != nil {
		t.Fatalf("failed to write cache file: %v", err)
	}

	// Read teams with long TTL
	teams, err := store.ReadTeams(24 * time.Hour)
	if err != nil {
		t.Fatalf("ReadTeams() error: %v", err)
	}
	if teams == nil {
		t.Fatal("ReadTeams() returned nil, want teams")
	}
	if len(teams) != 2 {
		t.Errorf("got %d teams, want 2", len(teams))
	}
	if teams[0] != "org/team-a" || teams[1] != "org/team-b" {
		t.Errorf("got teams %v, want %v", teams, expectedTeams)
	}
}

func TestReadTeams_CacheMiss_FileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStoreWithPath(filepath.Join(tmpDir, "nonexistent.json"))

	teams, err := store.ReadTeams(6 * time.Hour)
	if err != nil {
		t.Fatalf("ReadTeams() error: %v (should treat as cache miss)", err)
	}
	if teams != nil {
		t.Errorf("ReadTeams() = %v, want nil (cache miss)", teams)
	}
}

func TestReadTeams_CacheExpired(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStoreWithPath(filepath.Join(tmpDir, "teams.json"))

	// Write cache file with old timestamp
	cache := TeamCache{
		Teams:    []string{"org/team-a"},
		CachedAt: time.Now().Add(-7 * time.Hour), // Older than 6 hour TTL
	}
	data, err := json.Marshal(cache)
	if err != nil {
		t.Fatalf("failed to marshal cache: %v", err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "teams.json"), data, 0600)
	if err != nil {
		t.Fatalf("failed to write cache file: %v", err)
	}

	// Read with 6 hour TTL - should return nil (expired)
	teams, err := store.ReadTeams(6 * time.Hour)
	if err != nil {
		t.Fatalf("ReadTeams() error: %v", err)
	}
	if teams != nil {
		t.Errorf("ReadTeams() = %v, want nil (expired)", teams)
	}
}

func TestReadTeams_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStoreWithPath(filepath.Join(tmpDir, "teams.json"))

	// Write invalid JSON
	err := os.WriteFile(filepath.Join(tmpDir, "teams.json"), []byte("invalid json"), 0600)
	if err != nil {
		t.Fatalf("failed to write invalid JSON: %v", err)
	}

	// Should treat as cache miss
	teams, err := store.ReadTeams(6 * time.Hour)
	if err != nil {
		t.Fatalf("ReadTeams() error: %v (should treat as cache miss)", err)
	}
	if teams != nil {
		t.Errorf("ReadTeams() = %v, want nil (cache miss due to invalid JSON)", teams)
	}
}
