package scenes

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/builder"
	"github.com/chslink/fairygui/pkg/fgui/render"
)

// Environment gathers shared services needed by demo scenes.
type Environment struct {
	Loader  assets.Loader
	Factory *builder.Factory
	Atlas   *render.AtlasManager

	mu       sync.Mutex
	packages map[string]*assets.Package
}

// NewEnvironment wires the shared loader/factory/atlas trio used by scenes.
func NewEnvironment(loader assets.Loader, factory *builder.Factory, atlas *render.AtlasManager) *Environment {
	return &Environment{
		Loader:   loader,
		Factory:  factory,
		Atlas:    atlas,
		packages: make(map[string]*assets.Package),
	}
}

// Package loads (or retrieves from cache) the FairyGUI package identified by name.
// The loader is expected to expose "<name>.fui" and accompanying asset folders
// relative to its configured root.
func (e *Environment) Package(ctx context.Context, name string) (*assets.Package, error) {
	e.mu.Lock()
	if pkg, ok := e.packages[name]; ok {
		e.mu.Unlock()
		return pkg, nil
	}
	e.mu.Unlock()

	data, err := e.Loader.LoadOne(ctx, name+".fui", assets.ResourceBinary)
	if err != nil {
		return nil, err
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(name))
	if err != nil {
		return nil, err
	}

	e.Factory.RegisterPackage(pkg)

	e.mu.Lock()
	e.packages[name] = pkg
	e.mu.Unlock()

	return pkg, nil
}
