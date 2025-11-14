# API Migration Guide: builder.Factory → fgui.Factory

## Overview

The `builder.Factory` has been promoted to the top-level `fgui` package for a cleaner API surface. All functionality remains the same, but imports are simplified.

## Migration Steps

### 1. Update Imports

**Before:**
```go
import (
    "github.com/chslink/fairygui/pkg/fgui/assets"
    "github.com/chslink/fairygui/pkg/fgui/builder"
)
```

**After:**
```go
import (
    "github.com/chslink/fairygui/pkg/fgui"
)
```

### 2. Update Factory Creation

**Before:**
```go
loader := assets.NewFileLoader("./assets")
factory := builder.NewFactoryWithLoader(atlasManager, loader)
```

**After:**
```go
loader := fgui.NewFileLoader("./assets")
factory := fgui.NewFactoryWithLoader(atlasManager, loader)
```

### 3. Update Type References

**Before:**
```go
func NewEnvironment(loader assets.Loader, factory *builder.Factory, atlas *render.AtlasManager) *Environment
```

**After:**
```go
func NewEnvironment(loader fgui.Loader, factory *fgui.Factory, atlas *render.AtlasManager) *Environment
```

## New Unified API

The `fgui` package now provides a complete API surface:

```go
// Core types
fgui.GRoot
fgui.GComponent
fgui.GObject
fgui.Stage

// Asset types
fgui.Package
fgui.PackageItem
fgui.Loader
fgui.FileLoader

// Builder types
fgui.Factory
fgui.AtlasResolver
fgui.PackageResolver

// Resource constants
fgui.ResourceBinary
fgui.ResourceImage
fgui.ResourceSound
```

## Complete Example

**Before:**
```go
import (
    "context"
    "github.com/chslink/fairygui/pkg/fgui/assets"
    "github.com/chslink/fairygui/pkg/fgui/builder"
    "github.com/chslink/fairygui/pkg/fgui/core"
    "github.com/chslink/fairygui/pkg/fgui/render"
)

func main() {
    loader := assets.NewFileLoader("./assets")
    atlas := render.NewAtlasManager(loader)
    factory := builder.NewFactoryWithLoader(atlas, loader)

    data, _ := loader.LoadOne(ctx, "Main.fui", assets.ResourceBinary)
    pkg, _ := assets.ParsePackage(data, "assets/Main")
    factory.RegisterPackage(pkg)

    item := pkg.ItemByName("MainWindow")
    comp, _ := factory.BuildComponent(ctx, pkg, item)

    root := core.Root()
    root.AddChild(comp.GObject)
}
```

**After:**
```go
import (
    "context"
    "github.com/chslink/fairygui/pkg/fgui"
    "github.com/chslink/fairygui/pkg/fgui/render"
)

func main() {
    loader := fgui.NewFileLoader("./assets")
    atlas := render.NewAtlasManager(loader)
    factory := fgui.NewFactoryWithLoader(atlas, loader)

    data, _ := loader.LoadOne(ctx, "Main.fui", fgui.ResourceBinary)
    pkg, _ := fgui.ParsePackage(data, "assets/Main")
    factory.RegisterPackage(pkg)

    item := pkg.ItemByName("MainWindow")
    comp, _ := factory.BuildComponent(ctx, pkg, item)

    root := fgui.Root()
    root.AddChild(comp.GObject)
}
```

## Benefits

✅ **Simpler imports** - Only need `pkg/fgui` for most use cases
✅ **Cleaner API surface** - All FairyGUI types accessible from one package
✅ **Better discoverability** - IDE auto-completion shows all available types
✅ **Consistent naming** - Mirrors TypeScript's flat API structure

## Backward Compatibility

The `builder` package is still available for backward compatibility:
- `builder.Factory` remains functional
- Existing code will continue to work
- Migration is recommended but not required

## Implementation Details

The new API uses Go's type aliasing:
```go
// pkg/fgui/api.go
type Factory = builder.Factory
type AtlasResolver = builder.AtlasResolver
```

This means:
- Zero performance overhead (pure compile-time aliases)
- Both APIs point to the same underlying implementation
- Can mix old and new API in the same codebase during migration
