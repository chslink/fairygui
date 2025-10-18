package assets

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFileLoaderLoadOne(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.bin")
	content := []byte{0x01, 0x02, 0x03}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	loader := NewFileLoader(dir)
	data, err := loader.LoadOne(context.Background(), "test.bin", ResourceBinary)
	if err != nil {
		t.Fatalf("LoadOne failed: %v", err)
	}
	if len(data) != len(content) {
		t.Fatalf("unexpected length %d", len(data))
	}
}

func TestFileLoaderBatch(t *testing.T) {
	dir := t.TempDir()
	files := map[string][]byte{
		"a.txt": []byte("hello"),
		"b.bin": {0x10, 0x20},
	}
	for name, data := range files {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatalf("write temp file: %v", err)
		}
	}

	loader := NewFileLoader(dir)
	reqs := []ResourceRequest{
		{Key: "a.txt", Type: ResourceBinary},
		{Key: "b.bin", Type: ResourceBinary},
	}
	data, err := loader.Load(context.Background(), reqs)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(data) != len(reqs) {
		t.Fatalf("expected %d results, got %d", len(reqs), len(data))
	}
}
