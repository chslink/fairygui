package builder

import (
	"context"
	"fmt"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// Factory builds runtime components from parsed package metadata.
type Factory struct {
	atlasManager AtlasResolver
}

// AtlasResolver fetches renderable sprites (build-tagged implementation lives under render package).
type AtlasResolver interface {
	LoadPackage(ctx context.Context, pkg *assets.Package) error
	ResolveSprite(item *assets.PackageItem) (any, error)
}

// NewFactory creates a builder factory. Atlas manager can be nil for logic-only builds.
func NewFactory(resolver AtlasResolver) *Factory {
	return &Factory{atlasManager: resolver}
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

	root := core.NewGComponent()
	if item.Component.InitWidth > 0 || item.Component.InitHeight > 0 {
		root.SetSize(float64(item.Component.InitWidth), float64(item.Component.InitHeight))
	}

	for _, child := range item.Component.Children {
		childObj := f.buildChild(ctx, pkg, item, &child)
		root.AddChild(childObj)
	}
	return root, nil
}

func (f *Factory) buildChild(ctx context.Context, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild) *core.GObject {
	w := widgets.CreateWidget(child)
	var obj *core.GObject
	switch widget := w.(type) {
	case *widgets.GImage:
		obj = widget.GObject
	default:
		obj = core.NewGObject()
	}
	obj.SetName(child.Name)
	obj.SetPosition(float64(child.X), float64(child.Y))
	if child.Width >= 0 && child.Height >= 0 {
		obj.SetSize(float64(child.Width), float64(child.Height))
	}
	if !child.Visible {
		obj.SetVisible(false)
	}
	obj.SetAlpha(float64(child.Alpha))

	if child.Type == assets.ObjectTypeImage && child.Src != "" {
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
	}
	return obj
}

func (f *Factory) resolveImageSprite(ctx context.Context, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild) *assets.PackageItem {
	if pkg == nil {
		return nil
	}
	targetPkg := pkg
	if child.PackageID != "" && child.PackageID != pkg.ID {
		targetPkg = pkg
	}
	pi := targetPkg.ItemByID(child.Src)
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
