package search

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFindFilesRequiresQuery(t *testing.T) {
	t.Parallel()

	svc := NewService()
	if _, err := svc.FindFiles(t.TempDir(), "   "); !errors.Is(err, ErrQueryRequired) {
		t.Fatalf("FindFiles() error = %v, want %v", err, ErrQueryRequired)
	}
}

func TestFindFilesMatchesNamesAndIgnoresConfiguredDirectories(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "internal", "server"))
	mustMkdirAll(t, filepath.Join(root, ".git", "objects"))
	mustWriteFile(t, filepath.Join(root, "internal", "server", "server.go"), "package server")
	mustWriteFile(t, filepath.Join(root, ".git", "server.txt"), "hidden")

	svc := NewService()
	result, err := svc.FindFiles(root, "server")
	if err != nil {
		t.Fatalf("FindFiles() error = %v", err)
	}
	if result.Query != "server" {
		t.Fatalf("Query = %q", result.Query)
	}
	if len(result.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2; items=%+v", len(result.Items), result.Items)
	}
	if result.Items[0].Type != "directory" || result.Items[0].Path != "internal/server" {
		t.Fatalf("unexpected first item: %+v", result.Items[0])
	}
	if result.Items[1].Type != "file" || result.Items[1].Path != "internal/server/server.go" {
		t.Fatalf("unexpected second item: %+v", result.Items[1])
	}
}

func TestFindTextMatchesCaseInsensitiveContentAndSkipsIgnoredDirectories(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "pkg"))
	mustMkdirAll(t, filepath.Join(root, "node_modules", "demo"))
	mustWriteFile(t, filepath.Join(root, "pkg", "a.txt"), "Workspace line\nsecond workspace hit\n")
	mustWriteFile(t, filepath.Join(root, "node_modules", "demo", "skip.txt"), "workspace should be ignored")

	svc := NewService()
	result, err := svc.FindText(root, TextSearchRequest{Query: "workspace"})
	if err != nil {
		t.Fatalf("FindText() error = %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("Total = %d, want 2", result.Total)
	}
	if len(result.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(result.Items))
	}
	if result.Items[0].Path != "pkg/a.txt" || result.Items[0].Line != 1 || result.Items[0].Column != 1 {
		t.Fatalf("unexpected first match: %+v", result.Items[0])
	}
	if result.Items[0].MatchText != "Workspace" {
		t.Fatalf("MatchText = %q, want original case preserved", result.Items[0].MatchText)
	}
}

func TestFindTextRequiresQuery(t *testing.T) {
	t.Parallel()

	svc := NewService()
	if _, err := svc.FindText(t.TempDir(), TextSearchRequest{Query: ""}); !errors.Is(err, ErrQueryRequired) {
		t.Fatalf("FindText() error = %v, want %v", err, ErrQueryRequired)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q) error = %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
