//go:build !ebiten

package render

import (
	"context"
	"errors"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// AtlasManager is a placeholder when the ebiten build tag is not enabled.
type AtlasManager struct{}

// NewAtlasManager returns a manager that produces errors when used without the ebiten tag.
func NewAtlasManager(loader assets.Loader) *AtlasManager {
	return &AtlasManager{}
}

// LoadPackage is a no-op without the ebiten tag.
func (m *AtlasManager) LoadPackage(ctx context.Context, pkg *assets.Package) error {
	return errors.New("render: atlas loading requires ebiten build tag")
}

// ResolveSprite always returns an error without the ebiten tag.
func (m *AtlasManager) ResolveSprite(item *assets.PackageItem) (any, error) {
	return nil, errors.New("render: sprite resolution requires ebiten build tag")
}
