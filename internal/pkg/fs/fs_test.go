package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	t.Run("returns nil for existing file", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(f, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := FileExists(f); err != nil {
			t.Errorf("expected nil error for existing file, got: %v", err)
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		err := FileExists("/nonexistent/path/file.txt")
		if err == nil {
			t.Error("expected error for missing file, got nil")
		}
	})
}

func TestMustDir(t *testing.T) {
	t.Run("creates nested directories", func(t *testing.T) {
		dir := t.TempDir()
		nested := filepath.Join(dir, "a", "b", "c")
		MustDir(nested)
		if _, err := os.Stat(nested); err != nil {
			t.Errorf("expected directory to exist, got: %v", err)
		}
	})
}

func TestDeleteFile(t *testing.T) {
	t.Run("deletes an existing file", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "del.txt")
		if err := os.WriteFile(f, []byte("bye"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := DeleteFile(f); err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Error("expected file to be deleted")
		}
	})

	t.Run("returns error for nonexistent file", func(t *testing.T) {
		err := DeleteFile("/nonexistent/file.txt")
		if err == nil {
			t.Error("expected error for nonexistent file, got nil")
		}
	})
}

func TestDeleteDirIfExists(t *testing.T) {
	t.Run("deletes existing directory", func(t *testing.T) {
		dir := t.TempDir()
		sub := filepath.Join(dir, "subdir")
		if err := os.MkdirAll(sub, 0755); err != nil {
			t.Fatal(err)
		}
		if err := DeleteDirIfExists(sub); err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}
		if _, err := os.Stat(sub); !os.IsNotExist(err) {
			t.Error("expected directory to be deleted")
		}
	})

	t.Run("no-op for nonexistent path", func(t *testing.T) {
		err := DeleteDirIfExists("/nonexistent/dir")
		if err != nil {
			t.Errorf("expected nil error for nonexistent path, got: %v", err)
		}
	})
}

func TestRenameFile(t *testing.T) {
	t.Run("renames a file", func(t *testing.T) {
		dir := t.TempDir()
		old := filepath.Join(dir, "old.txt")
		new_ := filepath.Join(dir, "new.txt")
		if err := os.WriteFile(old, []byte("data"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := RenameFile(old, new_); err != nil {
			t.Errorf("expected nil error, got: %v", err)
		}
		if _, err := os.Stat(new_); err != nil {
			t.Errorf("expected new file to exist, got: %v", err)
		}
		if _, err := os.Stat(old); !os.IsNotExist(err) {
			t.Error("expected old file to be gone")
		}
	})

	t.Run("returns error if source does not exist", func(t *testing.T) {
		err := RenameFile("/nonexistent/old.txt", "/nonexistent/new.txt")
		if err == nil {
			t.Error("expected error for nonexistent source, got nil")
		}
	})
}

func TestListFilesFromDir(t *testing.T) {
	setup := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		for _, name := range []string{"a.mp3", "b.mp3", "c.aac", "d.wav", "notes.txt"} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0644); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}

	t.Run("lists files with extension filter", func(t *testing.T) {
		dir := setup(t)
		files, err := ListFilesFromDir(dir, ".mp3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 2 {
			t.Errorf("expected 2 mp3 files, got %d: %v", len(files), files)
		}
	})

	t.Run("lists all files with empty extension", func(t *testing.T) {
		dir := setup(t)
		files, err := ListFilesFromDir(dir, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 5 {
			t.Errorf("expected 5 files, got %d: %v", len(files), files)
		}
	})

	t.Run("returns error for nonexistent directory", func(t *testing.T) {
		_, err := ListFilesFromDir("/nonexistent/dir", "")
		if err == nil {
			t.Error("expected error for nonexistent directory, got nil")
		}
	})

	t.Run("returns error if path is a file not a directory", func(t *testing.T) {
		dir := t.TempDir()
		f := filepath.Join(dir, "file.txt")
		if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := ListFilesFromDir(f, "")
		if err == nil {
			t.Error("expected error when path is a file, got nil")
		}
	})

	t.Run("returns empty slice for empty directory", func(t *testing.T) {
		dir := t.TempDir()
		files, err := ListFilesFromDir(dir, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %d", len(files))
		}
	})
}