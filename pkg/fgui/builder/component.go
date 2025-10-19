package builder

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// Factory builds runtime components from parsed package metadata.
type Factory struct {
	atlasManager    AtlasResolver
	packageResolver PackageResolver
	packagesByID    map[string]*assets.Package
	packagesByName  map[string]*assets.Package
	packageDirs     map[string]string
	loader          assets.Loader
	loaderRoot      string
}

// AtlasResolver fetches renderable sprites (build-tagged implementation lives under render package).
type AtlasResolver interface {
	LoadPackage(ctx context.Context, pkg *assets.Package) error
	ResolveSprite(item *assets.PackageItem) (any, error)
}

// PackageResolver returns a package by id (or name) when cross-package references are encountered.
type PackageResolver func(ctx context.Context, owner *assets.Package, id string) (*assets.Package, error)

// NewFactory creates a builder factory. Atlas manager can be nil for logic-only builds.
func NewFactory(resolver AtlasResolver, pkgResolver PackageResolver) *Factory {
	return &Factory{
		atlasManager:    resolver,
		packageResolver: pkgResolver,
		packagesByID:    make(map[string]*assets.Package),
		packagesByName:  make(map[string]*assets.Package),
		packageDirs:     make(map[string]string),
	}
}

// NewFactoryWithLoader creates a factory that automatically resolves cross-package
// dependencies using the provided asset loader.
func NewFactoryWithLoader(resolver AtlasResolver, loader assets.Loader) *Factory {
	factory := NewFactory(resolver, nil)
	factory.loader = loader
	if fl, ok := loader.(*assets.FileLoader); ok && fl != nil {
		factory.loaderRoot = filepath.Clean(fl.Root)
	}
	factory.packageResolver = factory.defaultPackageResolver()
	return factory
}

// RegisterPackage makes the given package discoverable for cross-package lookups.
func (f *Factory) RegisterPackage(pkg *assets.Package) {
	if pkg == nil {
		return
	}
	if f.packagesByID == nil {
		f.packagesByID = make(map[string]*assets.Package)
	}
	if f.packagesByName == nil {
		f.packagesByName = make(map[string]*assets.Package)
	}
	if pkg.ID != "" {
		f.packagesByID[pkg.ID] = pkg
	}
	if pkg.Name != "" {
		f.packagesByName[pkg.Name] = pkg
	}
	dir := filepath.Dir(pkg.ResKey)
	f.registerPackageDir(pkg.ID, dir)
	f.registerPackageDir(pkg.Name, dir)
}

// BuildComponent instantiates a component hierarchy for the given package item.
func (f *Factory) BuildComponent(ctx context.Context, pkg *assets.Package, item *assets.PackageItem) (*core.GComponent, error) {
	if item == nil || item.Type != assets.PackageItemTypeComponent {
		return nil, fmt.Errorf("builder: package item must be a component")
	}
	if item.Component == nil {
		return nil, fmt.Errorf("builder: component data missing for %s", item.Name)
	}
	if f.atlasManager != nil {
		if err := f.atlasManager.LoadPackage(ctx, pkg); err != nil {
			return nil, err
		}
	}
	f.RegisterPackage(pkg)

	root := core.NewGComponent()
	if item.Component.InitWidth > 0 || item.Component.InitHeight > 0 {
		root.SetSize(float64(item.Component.InitWidth), float64(item.Component.InitHeight))
	}
	if item.Component.PivotAnchor {
		root.SetPivotWithAnchor(float64(item.Component.PivotX), float64(item.Component.PivotY), true)
	} else if item.Component.PivotX != 0 || item.Component.PivotY != 0 {
		root.SetPivot(float64(item.Component.PivotX), float64(item.Component.PivotY))
	}

	for _, child := range item.Component.Children {
		childObj := f.buildChild(ctx, pkg, item, &child)
		root.AddChild(childObj)
	}

	for _, ctrlData := range item.Component.Controllers {
		ctrl := core.NewController(ctrlData.Name)
		ctrl.AutoRadio = ctrlData.AutoRadio
		ctrl.PageNames = append(ctrl.PageNames, ctrlData.PageNames...)
		ctrl.PageIDs = append(ctrl.PageIDs, ctrlData.PageIDs...)
		root.AddController(ctrl)
	}

	f.finalizeComponentSize(root)
	return root, nil
}

func (f *Factory) buildChild(ctx context.Context, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild) *core.GObject {
	w := widgets.CreateWidget(child)
	var obj *core.GObject
	switch widget := w.(type) {
	case *widgets.GImage:
		obj = widget.GObject
	case *widgets.GTextField:
		obj = widget.GObject
		widget.SetText(child.Text)
		obj.SetData(child.Text)
	case *widgets.GButton:
		obj = widget.GComponent.GObject
		if pi := f.resolvePackageItem(ctx, pkg, owner, child); pi != nil {
			obj.SetData(pi)
		}
	case *widgets.GLoader:
		obj = widget.GObject
		if child.Src != "" {
			if pi := f.resolvePackageItem(ctx, pkg, owner, child); pi != nil {
				obj.SetData(pi)
				if (child.Width < 0 || child.Height < 0) && pi.Sprite != nil {
					obj.SetSize(float64(pi.Sprite.Rect.Width), float64(pi.Sprite.Rect.Height))
				}
			} else {
				obj.SetData(child.Src)
			}
		}
	case *widgets.GGroup:
		obj = widget.GComponent.GObject
	case *widgets.GGraph:
		obj = widget.GObject
	default:
		obj = core.NewGObject()
	}
	obj.SetName(child.Name)
	obj.SetPosition(float64(child.X), float64(child.Y))
	if child.Width >= 0 && child.Height >= 0 {
		obj.SetSize(float64(child.Width), float64(child.Height))
	}
	if child.ScaleX != 1 || child.ScaleY != 1 {
		obj.SetScale(float64(child.ScaleX), float64(child.ScaleY))
	}
	if child.Rotation != 0 {
		obj.SetRotation(float64(child.Rotation) * math.Pi / 180)
	}
	if child.SkewX != 0 || child.SkewY != 0 {
		obj.SetSkew(float64(child.SkewX)*math.Pi/180, float64(child.SkewY)*math.Pi/180)
	}
	if child.PivotAnchor || child.PivotX != 0 || child.PivotY != 0 {
		obj.SetPivotWithAnchor(float64(child.PivotX), float64(child.PivotY), child.PivotAnchor)
	}
	if !child.Visible {
		obj.SetVisible(false)
	}
	obj.SetAlpha(float64(child.Alpha))

	switch child.Type {
	case assets.ObjectTypeImage:
		pi := f.resolveImageSprite(ctx, pkg, owner, child)
		if pi != nil {
			obj.SetData(pi)
			if (child.Width < 0 || child.Height < 0) && pi.Sprite != nil {
				w := float64(pi.Sprite.Rect.Width)
				h := float64(pi.Sprite.Rect.Height)
				obj.SetSize(w, h)
			}
			if pi.PixelHitTest != nil {
				render.ApplyPixelHitTest(obj.DisplayObject(), pi.PixelHitTest)
			}
		}
	case assets.ObjectTypeComponent:
		if nested := f.buildNestedComponent(ctx, pkg, owner, child); nested != nil {
			obj.SetData(nested)
			if (child.Width < 0 || child.Height < 0) && nested.Width() > 0 && nested.Height() > 0 {
				obj.SetSize(nested.Width(), nested.Height())
			}
		}
	}
	return obj
}

func (f *Factory) resolvePackageItem(ctx context.Context, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild) *assets.PackageItem {
	if pkg == nil || child.Src == "" {
		return nil
	}
	target := pkg
	if child.PackageID != "" && child.PackageID != pkg.ID {
		candidates := f.dependencyCandidates(pkg, child.PackageID)
		var resolvedPkg *assets.Package
		for _, key := range candidates {
			if dep := f.lookupRegisteredPackage(key); dep != nil {
				resolvedPkg = dep
				break
			}
		}
		if resolvedPkg == nil && f.packageResolver != nil {
			for _, key := range candidates {
				resolved, err := f.packageResolver(ctx, pkg, key)
				if err != nil {
					fmt.Printf("builder: resolve package %s failed: %v\n", key, err)
					continue
				}
				if resolved == nil {
					continue
				}
				f.RegisterPackage(resolved)
				if f.atlasManager != nil {
					if err := f.atlasManager.LoadPackage(ctx, resolved); err != nil {
						fmt.Printf("builder: load dependent package failed: %v\n", err)
					}
				}
				resolvedPkg = resolved
				break
			}
		}
		if resolvedPkg == nil {
			return nil
		}
		target = resolvedPkg
	}
	return target.ItemByID(child.Src)
}

func (f *Factory) resolveImageSprite(ctx context.Context, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild) *assets.PackageItem {
	pi := f.resolvePackageItem(ctx, pkg, owner, child)
	if pi == nil {
		return nil
	}
	if f.atlasManager != nil {
		if _, err := f.atlasManager.ResolveSprite(pi); err != nil {
			fmt.Printf("builder: resolve sprite failed: %v\n", err)
		}
	}
	return pi
}

func (f *Factory) buildNestedComponent(ctx context.Context, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild) *core.GComponent {
	nestedItem := f.resolvePackageItem(ctx, pkg, owner, child)
	if nestedItem == nil {
		return nil
	}
	nested, err := f.BuildComponent(ctx, pkg, nestedItem)
	if err != nil {
		fmt.Println("builder: nested component error", err)
		return nil
	}
	return nested
}

func (f *Factory) finalizeComponentSize(comp *core.GComponent) {
	if comp == nil {
		return
	}
	width := comp.Width()
	height := comp.Height()
	maxX := width
	maxY := height
	for _, child := range comp.Children() {
		if child == nil || !child.Visible() {
			continue
		}
		cw := child.Width()
		ch := child.Height()
		if nested, ok := child.Data().(*core.GComponent); ok && nested != nil {
			if cw <= 0 {
				cw = nested.Width()
			}
			if ch <= 0 {
				ch = nested.Height()
			}
		}
		maxX = math.Max(maxX, child.X()+cw)
		maxY = math.Max(maxY, child.Y()+ch)
	}
	if width <= 0 && maxX > 0 {
		comp.SetSize(maxX, comp.Height())
	}
	if height <= 0 && maxY > 0 {
		comp.SetSize(comp.Width(), maxY)
	}
}

func (f *Factory) lookupRegisteredPackage(key string) *assets.Package {
	if key == "" {
		return nil
	}
	if f.packagesByID != nil {
		if pkg := f.packagesByID[key]; pkg != nil {
			return pkg
		}
	}
	if f.packagesByName != nil {
		if pkg := f.packagesByName[key]; pkg != nil {
			return pkg
		}
	}
	return nil
}

func (f *Factory) dependencyCandidates(pkg *assets.Package, id string) []string {
	candidates := make([]string, 0, 3)
	add := func(value string) {
		if value == "" {
			return
		}
		for _, existing := range candidates {
			if existing == value {
				return
			}
		}
		candidates = append(candidates, value)
	}
	add(id)
	if pkg != nil {
		for _, dep := range pkg.Dependencies {
			if dep.ID == id || dep.Name == id {
				add(dep.ID)
				add(dep.Name)
				break
			}
		}
	}
	return candidates
}

func (f *Factory) registerPackageDir(key string, absPath string) {
	if key == "" || absPath == "" {
		return
	}
	normalized := filepath.Clean(absPath)
	rel := normalized
	if f.loaderRoot != "" {
		if r, err := filepath.Rel(f.loaderRoot, normalized); err == nil && !strings.HasPrefix(r, "..") {
			if r == "." {
				rel = ""
			} else {
				rel = r
			}
		}
	}
	f.packageDirs[key] = rel
}

func (f *Factory) candidateDirectories(owner *assets.Package, id string) []string {
	var dirs []string
	seen := map[string]struct{}{}
	add := func(dir string) {
		dir = filepath.Clean(dir)
		if dir == "." {
			dir = ""
		}
		if _, ok := seen[dir]; ok {
			return
		}
		seen[dir] = struct{}{}
		dirs = append(dirs, dir)
	}
	if dir, ok := f.packageDirs[id]; ok {
		add(dir)
	}
	if owner != nil {
		if dir, ok := f.packageDirs[owner.ID]; ok {
			add(dir)
		}
		if dir, ok := f.packageDirs[owner.Name]; ok {
			add(dir)
		}
		add(filepath.Dir(owner.ResKey))
	}
	add("")
	return dirs
}

func (f *Factory) defaultPackageResolver() PackageResolver {
	return func(ctx context.Context, owner *assets.Package, id string) (*assets.Package, error) {
		if f.loader == nil {
			return nil, fmt.Errorf("builder: no loader configured for dependency %s", id)
		}
		candidates := f.dependencyCandidates(owner, id)
		dirs := f.candidateDirectories(owner, id)
		for _, candidate := range candidates {
			for _, dir := range dirs {
				file := candidate + ".fui"
				key := file
				if dir != "" && !strings.HasPrefix(dir, ".") {
					key = filepath.Join(dir, file)
				}
				data, err := f.loader.LoadOne(ctx, key, assets.ResourceBinary)
				if err != nil {
					continue
				}
				resKey := key
				if ext := filepath.Ext(resKey); ext != "" {
					resKey = resKey[:len(resKey)-len(ext)]
				}
				if f.loaderRoot != "" && !filepath.IsAbs(resKey) {
					resKey = filepath.Join(f.loaderRoot, resKey)
				}
				pkg, err := assets.ParsePackage(data, filepath.Clean(resKey))
				if err != nil {
					return nil, err
				}
				return pkg, nil
			}
		}
		return nil, fmt.Errorf("builder: dependency %s not found", id)
	}
}
