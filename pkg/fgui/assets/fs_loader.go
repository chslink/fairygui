package assets

import (
	"context"
	"os"
	"path/filepath"
)

// FileLoader loads resources from the local filesystem.
type FileLoader struct {
	Root string
}

// NewFileLoader constructs a loader using the provided root directory.
func NewFileLoader(root string) *FileLoader {
	return &FileLoader{Root: root}
}

// LoadOne reads a single resource from disk.
func (l *FileLoader) LoadOne(ctx context.Context, key string, typ ResourceType) ([]byte, error) {
	path := key
	if !filepath.IsAbs(path) {
		path = filepath.Join(l.Root, key)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Load reads a batch of resources sequentially.
func (l *FileLoader) Load(ctx context.Context, requests []ResourceRequest) (map[string][]byte, error) {
	result := make(map[string][]byte, len(requests))
	for _, req := range requests {
		data, err := l.LoadOne(ctx, req.Key, req.Type)
		if err != nil {
			return nil, err
		}
		result[req.Key] = data
	}
	return result, nil
}
