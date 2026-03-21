package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTarGzRoundTrip(t *testing.T) {
	// Create source directory with test files
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "bin", "tool"), "#!/bin/sh\necho hello", 0755)
	writeFile(t, filepath.Join(src, "config.txt"), "key=value", 0644)
	writeFile(t, filepath.Join(src, "sub", "deep.txt"), "nested", 0644)

	// Create tar.gz
	archive := filepath.Join(t.TempDir(), "test.tar.gz")
	if err := CreateTarGz(archive, src); err != nil {
		t.Fatalf("CreateTarGz: %v", err)
	}

	// Extract to new location
	dst := t.TempDir()
	if err := ExtractTarGz(archive, dst); err != nil {
		t.Fatalf("ExtractTarGz: %v", err)
	}

	// Verify files exist with correct content
	tests := []struct {
		path    string
		content string
	}{
		{"bin/tool", "#!/bin/sh\necho hello"},
		{"config.txt", "key=value"},
		{"sub/deep.txt", "nested"},
	}

	for _, tt := range tests {
		got := readFile(t, filepath.Join(dst, tt.path))
		if got != tt.content {
			t.Errorf("%s: got %q, want %q", tt.path, got, tt.content)
		}
	}
}

func TestTarGzSkipsAppleDoubleFiles(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "real.txt"), "real", 0644)
	writeFile(t, filepath.Join(src, "._real.txt"), "appleddouble", 0644)
	writeFile(t, filepath.Join(src, "sub", "file.txt"), "nested", 0644)
	writeFile(t, filepath.Join(src, "sub", "._file.txt"), "appleddouble2", 0644)

	archive := filepath.Join(t.TempDir(), "test.tar.gz")
	if err := CreateTarGz(archive, src); err != nil {
		t.Fatalf("CreateTarGz: %v", err)
	}

	dst := t.TempDir()
	if err := ExtractTarGz(archive, dst); err != nil {
		t.Fatalf("ExtractTarGz: %v", err)
	}

	// Real files should exist
	if _, err := os.Stat(filepath.Join(dst, "real.txt")); err != nil {
		t.Error("real.txt should exist")
	}
	if _, err := os.Stat(filepath.Join(dst, "sub", "file.txt")); err != nil {
		t.Error("sub/file.txt should exist")
	}

	// AppleDouble files should not exist
	if _, err := os.Stat(filepath.Join(dst, "._real.txt")); err == nil {
		t.Error("._real.txt should not exist")
	}
	if _, err := os.Stat(filepath.Join(dst, "sub", "._file.txt")); err == nil {
		t.Error("sub/._file.txt should not exist")
	}
}

func TestZipRoundTrip(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "bin", "tool.exe"), "MZ binary", 0755)
	writeFile(t, filepath.Join(src, "readme.txt"), "hello", 0644)

	archive := filepath.Join(t.TempDir(), "test.zip")
	if err := CreateZip(archive, src); err != nil {
		t.Fatalf("CreateZip: %v", err)
	}

	dst := t.TempDir()
	if err := ExtractZip(archive, dst); err != nil {
		t.Fatalf("ExtractZip: %v", err)
	}

	tests := []struct {
		path    string
		content string
	}{
		{"bin/tool.exe", "MZ binary"},
		{"readme.txt", "hello"},
	}

	for _, tt := range tests {
		got := readFile(t, filepath.Join(dst, tt.path))
		if got != tt.content {
			t.Errorf("%s: got %q, want %q", tt.path, got, tt.content)
		}
	}
}

func TestZipSkipsAppleDoubleFiles(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "real.txt"), "real", 0644)
	writeFile(t, filepath.Join(src, "._real.txt"), "appledouble", 0644)

	archive := filepath.Join(t.TempDir(), "test.zip")
	if err := CreateZip(archive, src); err != nil {
		t.Fatalf("CreateZip: %v", err)
	}

	dst := t.TempDir()
	if err := ExtractZip(archive, dst); err != nil {
		t.Fatalf("ExtractZip: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dst, "real.txt")); err != nil {
		t.Error("real.txt should exist")
	}
	if _, err := os.Stat(filepath.Join(dst, "._real.txt")); err == nil {
		t.Error("._real.txt should not exist")
	}
}

func TestCreateTarGzEmptyDir(t *testing.T) {
	src := t.TempDir()
	archive := filepath.Join(t.TempDir(), "empty.tar.gz")

	if err := CreateTarGz(archive, src); err != nil {
		t.Fatalf("CreateTarGz on empty dir: %v", err)
	}

	// Should extract without error
	dst := t.TempDir()
	if err := ExtractTarGz(archive, dst); err != nil {
		t.Fatalf("ExtractTarGz on empty archive: %v", err)
	}
}

func TestExtractTarGzMissingFile(t *testing.T) {
	err := ExtractTarGz("/nonexistent/file.tar.gz", t.TempDir())
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestExtractZipMissingFile(t *testing.T) {
	err := ExtractZip("/nonexistent/file.zip", t.TempDir())
	if err == nil {
		t.Error("expected error for missing file")
	}
}

// helpers

func writeFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
