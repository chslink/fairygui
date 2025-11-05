package scenes

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/chslink/fairygui/pkg/fgui"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/render"
)

// Environment gathers shared services needed by demo scenes.
type Environment struct {
	Loader  assets.Loader
	Factory *fgui.Factory
	Atlas   *render.AtlasManager

	mu       sync.Mutex
	packages map[string]*assets.Package
}

// NewEnvironment wires the shared loader/factory/atlas trio used by scenes.
func NewEnvironment(loader assets.Loader, factory *fgui.Factory, atlas *render.AtlasManager) *Environment {
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

// SetupDefaultScrollBars 设置默认滚动条资源
// 从指定包中查找滚动条组件并设置为全局默认值
func (e *Environment) SetupDefaultScrollBars(ctx context.Context, packageName, verticalName, horizontalName string) error {
	pkg, err := e.Package(ctx, packageName)
	if err != nil {
		return err
	}

	// 查找垂直滚动条
	if verticalName != "" {
		if item := pkg.ItemByName(verticalName); item != nil {
			vtURL := "ui://" + pkg.ID + "/" + item.ID
			fmt.Printf("[DefaultScrollBars] Setting vertical scrollbar: %s (from %s.%s)\n", vtURL, packageName, verticalName)
			fgui.SetDefaultScrollBars(vtURL, "")
		} else {
			fmt.Printf("[DefaultScrollBars] WARNING: Vertical scrollbar '%s' not found in package %s\n", verticalName, packageName)
		}
	}

	// 查找水平滚动条
	if horizontalName != "" {
		if item := pkg.ItemByName(horizontalName); item != nil {
			hzURL := "ui://" + pkg.ID + "/" + item.ID
			fmt.Printf("[DefaultScrollBars] Setting horizontal scrollbar: %s (from %s.%s)\n", hzURL, packageName, horizontalName)
			fgui.SetDefaultScrollBars("", hzURL)
		} else {
			fmt.Printf("[DefaultScrollBars] WARNING: Horizontal scrollbar '%s' not found in package %s\n", horizontalName, packageName)
		}
	}

	return nil
}
