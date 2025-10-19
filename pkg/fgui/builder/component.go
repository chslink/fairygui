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

	for _, ctrlData := range item.Component.Controllers {
		ctrl := core.NewController(ctrlData.Name)
		ctrl.AutoRadio = ctrlData.AutoRadio
		ctrl.PageNames = append(ctrl.PageNames, ctrlData.PageNames...)
		ctrl.PageIDs = append(ctrl.PageIDs, ctrlData.PageIDs...)
		root.AddController(ctrl)
	}
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
		if pi := f.resolvePackageItem(pkg, owner, child); pi != nil {
			obj.SetData(pi)
		}
	case *widgets.GLoader:
		obj = widget.GObject
		if child.Src != "" {
			if pi := f.resolvePackageItem(pkg, owner, child); pi != nil {
				obj.SetData(pi)
				if (child.Width < 0 || child.Height < 0) && pi.Sprite != nil {
					obj.SetSize(float64(pi.Sprite.Rect.Width), float64(pi.Sprite.Rect.Height))
				}
			} else {
				obj.SetData(child.Src)
			}
		}
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

func (f *Factory) resolvePackageItem(pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild) *assets.PackageItem {
	if pkg == nil || child.Src == "" {
		return nil
	}
	target := pkg
	if child.PackageID != "" && child.PackageID != pkg.ID {
		// TODO: resolve cross-package references (dependencies)
	}
	return target.ItemByID(child.Src)
}

func (f *Factory) resolveImageSprite(ctx context.Context, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild) *assets.PackageItem {
	pi := f.resolvePackageItem(pkg, owner, child)
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
	nestedItem := f.resolvePackageItem(pkg, owner, child)
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
