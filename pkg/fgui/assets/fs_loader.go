package assets

import (
	"context"
	"os"
	"path/filepath"
	"strings"
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
	path := filepath.Clean(key)
	root := filepath.Clean(l.Root)

	if !filepath.IsAbs(path) {
		if root != "" {
			if strings.HasPrefix(path, root) {
				// path already rooted relative to loader root
			} else if strings.HasPrefix(path, string(os.PathSeparator)) && strings.HasPrefix(root, string(os.PathSeparator)) {
				if strings.HasPrefix(path, root) {
					// already rooted
				} else {
					path = filepath.Join(root, path)
				}
			} else {
				path = filepath.Join(root, path)
			}
		}
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
