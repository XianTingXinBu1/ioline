package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServiceSetCurrentAndClear(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	svc := NewService()

	info, err := svc.Set(root)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if !info.IsSet {
		t.Fatalf("expected workspace to be set")
	}
	if info.RootPath != root {
		t.Fatalf("RootPath = %q, want %q", info.RootPath, root)
	}
	if info.Name != filepath.Base(root) {
		t.Fatalf("Name = %q, want %q", info.Name, filepath.Base(root))
	}

	current := svc.Current()
	if current.RootPath != root || !current.IsSet {
		t.Fatalf("Current() = %+v", current)
	}

	gotRoot, err := svc.RootPath()
	if err != nil {
		t.Fatalf("RootPath() error = %v", err)
	}
	if gotRoot != root {
		t.Fatalf("RootPath() = %q, want %q", gotRoot, root)
	}

	cleared := svc.Clear()
	if cleared.IsSet || cleared.RootPath != "" {
		t.Fatalf("Clear() = %+v, want empty info", cleared)
	}
	if _, err := svc.RootPath(); err != ErrNotConfigured {
		t.Fatalf("RootPath() after Clear error = %v, want %v", err, ErrNotConfigured)
	}
}

func TestServiceSetRejectsNonDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filePath := filepath.Join(root, "file.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	svc := NewService()
	if _, err := svc.Set(filePath); err == nil {
		t.Fatal("expected Set() to reject file path")
	}
}

func TestCandidatesIncludesCurrentWorkspaceWithoutDuplicates(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	svc := NewService()
	if _, err := svc.Set(root); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	items := svc.Candidates()
	if len(items) == 0 {
		t.Fatal("Candidates() returned no items")
	}

	seen := make(map[string]struct{})
	foundCurrent := false
	for _, item := range items {
		if _, ok := seen[item.Path]; ok {
			t.Fatalf("duplicate candidate path: %s", item.Path)
		}
		seen[item.Path] = struct{}{}
		if item.Path == root {
			foundCurrent = true
		}
	}
	if !foundCurrent {
		t.Fatalf("expected current workspace %q in candidates", root)
	}
}

func TestBrowseDirectoriesReturnsParentAndDirectoriesOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	alphaDir := filepath.Join(root, "alpha")
	betaDir := filepath.Join(root, "beta")
	if err := os.MkdirAll(alphaDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(alpha) error = %v", err)
	}
	if err := os.MkdirAll(betaDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(beta) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "note.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	svc := NewService()
	result, err := svc.BrowseDirectories(root)
	if err != nil {
		t.Fatalf("BrowseDirectories() error = %v", err)
	}
	if result.CurrentPath != root {
		t.Fatalf("CurrentPath = %q, want %q", result.CurrentPath, root)
	}
	if result.ParentPath != filepath.Dir(root) {
		t.Fatalf("ParentPath = %q, want %q", result.ParentPath, filepath.Dir(root))
	}
	if len(result.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(result.Items))
	}
	if result.Items[0].Name != "alpha" || result.Items[1].Name != "beta" {
		t.Fatalf("unexpected items order/content: %+v", result.Items)
	}
}
