// Package atomicwrite test cho WriteFile: ghi mới, ghi đè, không để
// lại file tạm.
package atomicwrite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFile_New(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("content = %q", string(data))
	}
}

func TestWriteFile_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	os.WriteFile(path, []byte("old content longer"), 0o644)
	if err := WriteFile(path, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "new" {
		t.Errorf("content = %q, want new", string(data))
	}
}

func TestWriteFile_NoTempLeftover(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	WriteFile(path, []byte("x"), 0o644)
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".tmp_") {
			t.Errorf("còn file tạm: %s", e.Name())
		}
	}
}
