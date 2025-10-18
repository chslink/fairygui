package assets

import "context"

// ResourceType defines the type of asset requested from the loader.
type ResourceType string

const (
	// ResourceBinary represents raw binary data (e.g. .fui descriptor files).
	ResourceBinary ResourceType = "binary"
	// ResourceImage represents an image resource (atlas textures, etc.).
	ResourceImage ResourceType = "image"
	// ResourceSound represents an audio file.
	ResourceSound ResourceType = "sound"
)

// ResourceRequest describes a single asset load operation.
type ResourceRequest struct {
	Key  string
	Type ResourceType
}

// Loader abstracts the asset loading backend (filesystem, embedded data, network, etc.).
type Loader interface {
	// Load retrieves multiple resources in one batch. The returned map keys match the
	// request keys. Implementation-specific streaming or caching behaviours are allowed.
	Load(ctx context.Context, requests []ResourceRequest) (map[string][]byte, error)

	// LoadOne retrieves a single resource and returns its raw bytes.
	LoadOne(ctx context.Context, key string, typ ResourceType) ([]byte, error)
}
