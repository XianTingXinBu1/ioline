package files

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestServiceListAndStat(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "dir"), 0o755); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	svc := NewService()
	items, err := svc.List(root, ".")
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	if items[0].Type != "directory" || items[0].Name != "dir" {
		t.Fatalf("first item = %+v, want directory first", items[0])
	}
	if items[1].Type != "file" || items[1].Name != "file.txt" {
		t.Fatalf("second item = %+v, want file second", items[1])
	}

	entry, err := svc.Stat(root, "file.txt")
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if entry.Path != "file.txt" || entry.Type != "file" {
		t.Fatalf("Stat() = %+v", entry)
	}
}

func TestServiceReadAndSaveText(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filePath := filepath.Join(root, "hello.txt")
	if err := os.WriteFile(filePath, []byte("line1\r\nline2"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	svc := NewService()
	content, err := svc.ReadText(root, "hello.txt")
	if err != nil {
		t.Fatalf("ReadText() error = %v", err)
	}
	if content.Content != "line1\r\nline2" {
		t.Fatalf("Content = %q", content.Content)
	}
	if content.LineEnding != "crlf" {
		t.Fatalf("LineEnding = %q, want crlf", content.LineEnding)
	}

	result, err := svc.SaveText(root, SaveRequest{Path: "nested/out.txt", Content: "saved"})
	if err != nil {
		t.Fatalf("SaveText() error = %v", err)
	}
	if result.Path != "nested/out.txt" {
		t.Fatalf("SaveText path = %q", result.Path)
	}
	data, err := os.ReadFile(filepath.Join(root, "nested", "out.txt"))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "saved" {
		t.Fatalf("saved file content = %q", string(data))
	}
}

func TestServiceCreateMoveAndDelete(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	svc := NewService()

	if _, err := svc.CreateDirectory(root, CreateDirectoryRequest{Path: "tmp/demo"}); err != nil {
		t.Fatalf("CreateDirectory() error = %v", err)
	}
	if _, err := svc.CreateFile(root, CreateFileRequest{Path: "tmp/demo/a.txt", Content: "a"}); err != nil {
		t.Fatalf("CreateFile() error = %v", err)
	}
	moveResult, err := svc.Move(root, MoveRequest{FromPath: "tmp/demo/a.txt", ToPath: "tmp/demo/b.txt"})
	if err != nil {
		t.Fatalf("Move() error = %v", err)
	}
	if moveResult.ToPath != "tmp/demo/b.txt" {
		t.Fatalf("Move result = %+v", moveResult)
	}
	if _, err := os.Stat(filepath.Join(root, "tmp", "demo", "b.txt")); err != nil {
		t.Fatalf("moved file missing: %v", err)
	}

	if _, err := svc.Delete(root, DeleteRequest{Path: "tmp/demo", Recursive: true}); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "tmp", "demo")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected deleted directory, got err=%v", err)
	}
}

func TestServiceRejectsInvalidPathsAndBinaryFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	binaryPath := filepath.Join(root, "bin.dat")
	if err := os.WriteFile(binaryPath, []byte{0x00, 0x01, 0x02}, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	svc := NewService()
	if _, err := svc.List(root, "../outside"); err != ErrInvalidPath {
		t.Fatalf("List invalid path error = %v, want %v", err, ErrInvalidPath)
	}
	if _, err := svc.ReadText(root, "bin.dat"); err != ErrBinaryFile {
		t.Fatalf("ReadText binary error = %v, want %v", err, ErrBinaryFile)
	}
}
