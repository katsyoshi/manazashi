package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitUsesGitRootAndDefaultCache(t *testing.T) {
	if _, err := exec.LookPath("sqlite3"); err != nil {
		t.Skip("sqlite3 command not found")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git command not found")
	}
	root := t.TempDir()
	nested := filepath.Join(root, "nested", "deeper")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	initGitRepo(t, root)
	cache := t.TempDir()
	t.Setenv("CODE_INDEX_CACHE_DIR", cache)

	var result initJSONResult
	decodeRunJSON(t, []string{"init", "--format", "json", nested}, &result)
	if result.Root != root || result.DB != defaultDBPath(root) || !result.ConfigCreated || !result.DBCreated {
		t.Fatalf("init JSON = %#v", result)
	}
	config, err := os.ReadFile(filepath.Join(root, projectConfigName))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(config), "\ndb = ") || strings.HasPrefix(string(config), "db = ") {
		t.Fatalf("default config activates db: %q", config)
	}
	assertSQLiteValue(t, result.DB, "select count(*) from files;", "0")
}

func TestInitRejectsNonGitRootAndInvalidDBPaths(t *testing.T) {
	if _, err := exec.LookPath("sqlite3"); err != nil {
		t.Skip("sqlite3 command not found")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git command not found")
	}
	nonGit := t.TempDir()
	if err := run([]string{"init", nonGit}); err == nil || !strings.Contains(err.Error(), "not a Git work tree") {
		t.Fatalf("non-Git init error = %v", err)
	}

	for _, value := range []string{filepath.Join(string(filepath.Separator), "tmp", "index.sqlite"), "../index.sqlite"} {
		root := t.TempDir()
		initGitRepo(t, root)
		err := run([]string{"init", "--db", value, root})
		if err == nil || !strings.Contains(err.Error(), "invalid --db") {
			t.Fatalf("init --db %q error = %v", value, err)
		}
		if _, statErr := os.Stat(filepath.Join(root, projectConfigName)); !os.IsNotExist(statErr) {
			t.Fatalf("config exists after invalid --db %q: %v", value, statErr)
		}
	}
}

func TestInitRollsBackConfigWhenDatabaseCreationFails(t *testing.T) {
	if _, err := exec.LookPath("sqlite3"); err != nil {
		t.Skip("sqlite3 command not found")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git command not found")
	}
	root := t.TempDir()
	initGitRepo(t, root)
	blocked := filepath.Join(root, "blocked")
	if err := os.WriteFile(blocked, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, projectConfigName)
	if err := run([]string{"init", "--db", "blocked/index.sqlite", root}); err == nil {
		t.Fatal("init succeeded with unusable database parent")
	}
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatalf("new config was not rolled back: %v", err)
	}

	original := []byte("# personal settings\nmax_bytes = 42\n")
	if err := os.WriteFile(configPath, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"init", "--force", "--db", "blocked/index.sqlite", root}); err == nil {
		t.Fatal("forced init succeeded with unusable database parent")
	}
	restored, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored) != string(original) {
		t.Fatalf("restored config = %q, want %q", restored, original)
	}
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("restored config mode = %o, want 600", info.Mode().Perm())
	}
}
