package builder

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/gears"
	"github.com/chslink/fairygui/pkg/fgui/render"
	"github.com/chslink/fairygui/pkg/fgui/utils"
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
	root.SetResourceID(item.ID)
	if item.Component.InitWidth > 0 || item.Component.InitHeight > 0 {
		root.SetSize(float64(item.Component.InitWidth), float64(item.Component.InitHeight))
	}
	if item.Component.PivotAnchor {
		root.SetPivotWithAnchor(float64(item.Component.PivotX), float64(item.Component.PivotY), true)
	} else if item.Component.PivotX != 0 || item.Component.PivotY != 0 {
		root.SetPivot(float64(item.Component.PivotX), float64(item.Component.PivotY))
	}
	rootSourceW := float64(item.Component.SourceWidth)
	rootSourceH := float64(item.Component.SourceHeight)
	rootInitW := float64(item.Component.InitWidth)
	rootInitH := float64(item.Component.InitHeight)
	if rootSourceW == 0 {
		rootSourceW = root.Width()
	}
	if rootSourceH == 0 {
		rootSourceH = root.Height()
	}
	if rootInitW == 0 {
		rootInitW = rootSourceW
	}
	if rootInitH == 0 {
		rootInitH = rootSourceH
	}
	root.SetSourceSize(rootSourceW, rootSourceH)
	root.SetInitSize(rootInitW, rootInitH)

	for _, ctrlData := range item.Component.Controllers {
		ctrl := core.NewController(ctrlData.Name)
		ctrl.AutoRadio = ctrlData.AutoRadio
		ctrl.SetPages(ctrlData.PageIDs, ctrlData.PageNames)
		root.AddController(ctrl)
	}

	for idx := range item.Component.Children {
		child := &item.Component.Children[idx]
		childObj := f.buildChild(ctx, pkg, item, root, child)
		root.AddChild(childObj)
	}

	f.setupRelations(item, root)
	f.setupGears(item, root)
	root.GObject.SetupAfterAdd(root, item.RawData, 0)
	setupResolver := componentSetupResolver{component: root, item: item}
	root.SetupAfterAdd(item.RawData, 0, setupResolver, setupResolver)

	f.finalizeComponentSize(root)
	return root, nil
}

func childBuffer(owner *assets.PackageItem, child *assets.ComponentChild) *utils.ByteBuffer {
	if owner == nil || owner.RawData == nil || child == nil || child.RawDataLength <= 0 {
		return nil
	}
	sub, err := owner.RawData.SubBuffer(child.RawDataOffset, child.RawDataLength)
	if err != nil {
		return nil
	}
	return sub
}

func (f *Factory) newSetupContext(ctx context.Context, pkg *assets.Package, owner *assets.PackageItem, parent *core.GComponent, child *assets.ComponentChild, resolved *assets.PackageItem) *widgets.SetupContext {
	return &widgets.SetupContext{
		Owner:        owner,
		Child:        child,
		Parent:       parent,
		Package:      pkg,
		ResolvedItem: resolved,
		ResolveIcon: func(icon string) *assets.PackageItem {
			return f.resolveIcon(ctx, pkg, icon)
		},
	}
}

func (f *Factory) buildChild(ctx context.Context, pkg *assets.Package, owner *assets.PackageItem, parent *core.GComponent, child *assets.ComponentChild) *core.GObject {
	resolvedItem := f.resolvePackageItem(ctx, pkg, owner, child)
	w := widgets.CreateWidget(child)
	if resolvedItem != nil {
		if alt := widgets.CreateWidgetFromPackage(resolvedItem); alt != nil {
			if w == nil || resolvedItem.ObjectType != child.Type {
				w = alt
			}
		}
	}
	sub := childBuffer(owner, child)
	var setupCtx *widgets.SetupContext
	ensureCtx := func() *widgets.SetupContext {
		if setupCtx == nil {
			setupCtx = f.newSetupContext(ctx, pkg, owner, parent, child, resolvedItem)
		} else {
			setupCtx.ResolvedItem = resolvedItem
		}
		return setupCtx
	}
	var obj *core.GObject
	switch widget := w.(type) {
	case *widgets.GImage:
		obj = widget.GObject
		if spriteItem := f.resolveImageSprite(ctx, pkg, owner, child); spriteItem != nil {
			resolvedItem = spriteItem
		}
		widget.SetPackageItem(resolvedItem)
		obj.SetData(widget)
		if sub != nil {
			if before, ok := interface{}(widget).(widgets.BeforeAdder); ok {
				before.SetupBeforeAdd(ensureCtx(), sub)
			}
		}
	case *widgets.GMovieClip:
		obj = widget.GObject
		if resolvedItem != nil && resolvedItem.Type == assets.PackageItemTypeMovieClip {
			widget.SetPackageItem(resolvedItem)
		} else if child.Src != "" {
			if clipItem := f.resolveIcon(ctx, pkg, child.Src); clipItem != nil {
				resolvedItem = clipItem
				widget.SetPackageItem(clipItem)
			}
		} else {
			widget.SetPackageItem(resolvedItem)
		}
		obj.SetData(widget)
		if sub != nil {
			if before, ok := interface{}(widget).(widgets.BeforeAdder); ok {
				before.SetupBeforeAdd(ensureCtx(), sub)
			}
		}
	case *widgets.GRichTextField:
		// 富文本必须在 GTextField 之前处理，因为它嵌入了 GTextField
		obj = widget.GObject()
		widget.SetText(child.Text)
		if sub != nil {
			if before, ok := interface{}(widget).(widgets.BeforeAdder); ok {
				before.SetupBeforeAdd(ensureCtx(), sub)
			}
		}
		// 关键：保持 Data 指向 GRichTextField 实例，不要覆盖
		obj.SetData(widget)
	case *widgets.GTextField:
		obj = widget.GObject
		widget.SetText(child.Text)
		if sub != nil {
			if before, ok := interface{}(widget).(widgets.BeforeAdder); ok {
				before.SetupBeforeAdd(ensureCtx(), sub)
			}
		}
		obj.SetData(widget)
	case *widgets.GTextInput:
		obj = widget.GObject
		widget.SetText(child.Text)
		if sub != nil {
			if before, ok := interface{}(widget).(widgets.BeforeAdder); ok {
				before.SetupBeforeAdd(ensureCtx(), sub)
			}
		}
		obj.SetData(widget)
	case *widgets.GButton:
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
			f.applyButtonTemplate(ctx, widget, pkg, owner, resolvedItem)
		}
		obj = widget.GComponent.GObject
		if obj != nil && (obj.Width() == 0 || obj.Height() == 0) {
			if tpl := widget.TemplateComponent(); tpl != nil {
				width := tpl.Width()
				height := tpl.Height()
				if width <= 0 {
					width = tpl.GObject.Width()
				}
				if height <= 0 {
					height = tpl.GObject.Height()
				}
				if width <= 0 {
					width = obj.Width()
				}
				if height <= 0 {
					height = obj.Height()
				}
				if width > 0 || height > 0 {
					obj.SetSize(width, height)
				}
			} else if resolvedItem != nil && resolvedItem.Component != nil {
				width := float64(resolvedItem.Component.InitWidth)
				height := float64(resolvedItem.Component.InitHeight)
				if width <= 0 {
					width = float64(resolvedItem.Component.SourceWidth)
				}
				if height <= 0 {
					height = float64(resolvedItem.Component.SourceHeight)
				}
				if width <= 0 {
					width = obj.Width()
				}
				if height <= 0 {
					height = obj.Height()
				}
				if width > 0 || height > 0 {
					obj.SetSize(width, height)
				}
			}
		}
		resource := child.Src
		if resource == "" {
			resource = child.Data
		}
		widget.SetResource(resource)
		if child.Text != "" {
			widget.SetTitle(child.Text)
		}
		if child.Icon != "" {
			widget.SetIcon(child.Icon)
			if iconItem := f.resolveIcon(ctx, pkg, child.Icon); iconItem != nil {
				widget.SetIconItem(iconItem)
			}
		}
		if sub != nil {
			if after, ok := interface{}(widget).(widgets.AfterAdder); ok {
				after.SetupAfterAdd(ensureCtx(), sub)
			}
		}
		obj.SetData(widget)
	case *widgets.GLoader:
		obj = widget.GObject
		obj.SetData(widget)
		if child.Src != "" {
			widget.SetURL(child.Src)
		}
		autoSize := child.Width < 0 || child.Height < 0
		widget.SetAutoSize(autoSize)
		if sub != nil {
			if before, ok := interface{}(widget).(widgets.BeforeAdder); ok {
				before.SetupBeforeAdd(ensureCtx(), sub)
			}
		}
		if resolvedItem == nil {
			resolvedItem = f.resolveIcon(ctx, pkg, widget.URL())
		}
		if setupCtx != nil {
			setupCtx.ResolvedItem = resolvedItem
		}
		f.assignLoaderPackage(ctx, widget, pkg, resolvedItem)
	case *widgets.GLabel:
		obj = widget.GComponent.GObject
		widget.SetTitle(child.Text)
		widget.SetIcon(child.Icon)
		if iconItem := f.resolveIcon(ctx, pkg, child.Icon); iconItem != nil {
			widget.SetIconItem(iconItem)
		}
		resource := child.Src
		if resource == "" {
			resource = child.Data
		}
		widget.SetResource(resource)
		obj.SetData(widget)
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
		}
		f.applyLabelTemplate(ctx, widget, pkg, owner, resolvedItem)
		if sub != nil {
			if after, ok := interface{}(widget).(widgets.AfterAdder); ok {
				after.SetupAfterAdd(ensureCtx(), sub)
			}
		}
	case *widgets.GList:
		obj = widget.GComponent.GObject
		widget.SetResource(child.Data)
		widget.SetDefaultItem(child.Data)
		widget.SetPackageItem(resolvedItem)
		obj.SetData(widget)
		if sub != nil {
			if before, ok := interface{}(widget).(widgets.BeforeAdder); ok {
				before.SetupBeforeAdd(ensureCtx(), sub)
			}
		}
		if sub != nil {
			if after, ok := interface{}(widget).(widgets.AfterAdder); ok {
				after.SetupAfterAdd(ensureCtx(), sub)
			}
		}
	case *widgets.GProgressBar:
		obj = widget.GComponent.GObject
		obj.SetData(widget)
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
		}
		f.applyProgressBarTemplate(ctx, widget, pkg, owner, resolvedItem)
		if sub != nil {
			if after, ok := interface{}(widget).(widgets.AfterAdder); ok {
				after.SetupAfterAdd(ensureCtx(), sub)
			}
		}
	case *widgets.GSlider:
		obj = widget.GComponent.GObject
		obj.SetData(widget)
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
		}
		f.applySliderTemplate(ctx, widget, pkg, owner, resolvedItem)
		if sub != nil {
			if after, ok := interface{}(widget).(widgets.AfterAdder); ok {
				after.SetupAfterAdd(ensureCtx(), sub)
			}
		}
	case *widgets.GScrollBar:
		obj = widget.GComponent.GObject
		obj.SetData(widget)
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
		}
		f.applyScrollBarTemplate(ctx, widget, pkg, owner, resolvedItem)
	case *widgets.GTree:
		obj = widget.GComponent.GObject
		obj.SetData(widget)
		resource := child.Src
		if resource == "" {
			resource = child.Data
		}
		widget.SetResource(resource)
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
		}
		if sub != nil {
			if before, ok := interface{}(widget).(widgets.BeforeAdder); ok {
				before.SetupBeforeAdd(ensureCtx(), sub)
			}
			f.populateTree(ctx, widget, pkg, owner, child, sub)
			if after, ok := interface{}(widget).(widgets.AfterAdder); ok {
				after.SetupAfterAdd(ensureCtx(), sub)
			}
		}
	case *widgets.GComboBox:
		obj = widget.GComponent.GObject
		obj.SetData(widget)
		resource := child.Src
		if resource == "" {
			resource = child.Data
		}
		widget.SetResource(resource)
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
		}
		f.applyComboBoxTemplate(ctx, widget, pkg, owner, resolvedItem)
		if sub != nil {
			if after, ok := interface{}(widget).(widgets.AfterAdder); ok {
				after.SetupAfterAdd(ensureCtx(), sub)
			}
		}
	case *widgets.GGroup:
		obj = widget.GObject
		if sub != nil {
			if before, ok := interface{}(widget).(widgets.BeforeAdder); ok {
				before.SetupBeforeAdd(ensureCtx(), sub)
			}
		}
		if sub != nil {
			if after, ok := interface{}(widget).(widgets.AfterAdder); ok {
				after.SetupAfterAdd(ensureCtx(), sub)
			}
		}
	case *widgets.GGraph:
		obj = widget.GObject
		obj.SetData(widget)
		if sub != nil {
			if before, ok := interface{}(widget).(widgets.BeforeAdder); ok {
				before.SetupBeforeAdd(ensureCtx(), sub)
			}
		}
	default:
		obj = core.NewGObject()
	}
	if obj != nil {
		obj.ApplyComponentChild(child)
	}
	rid := child.ID
	if resolvedItem != nil && resolvedItem.ID != "" {
		rid = resolvedItem.ID
	}
	obj.SetResourceID(rid)
	if button, ok := obj.Data().(*widgets.GButton); ok {
		button.UpdateTemplateBounds(obj.Width(), obj.Height())
	}

	switch child.Type {
	case assets.ObjectTypeImage:
		if resolvedItem != nil {
			if (child.Width < 0 || child.Height < 0) && resolvedItem.Sprite != nil {
				w := float64(resolvedItem.Sprite.OriginalSize.X)
				h := float64(resolvedItem.Sprite.OriginalSize.Y)
				if w <= 0 {
					w = float64(resolvedItem.Sprite.Rect.Width)
				}
				if h <= 0 {
					h = float64(resolvedItem.Sprite.Rect.Height)
				}
				if w > 0 || h > 0 {
					obj.SetSize(w, h)
				}
			}
			if resolvedItem.PixelHitTest != nil {
				render.ApplyPixelHitTest(obj.DisplayObject(), resolvedItem.PixelHitTest)
			}
		}
	case assets.ObjectTypeComponent:
		switch data := obj.Data().(type) {
		case *widgets.GButton:
			// already handled by applyButtonTemplate
		case *widgets.GLabel:
			f.applyLabelTemplate(ctx, data, pkg, owner, resolvedItem)
		case *widgets.GComboBox:
			f.applyComboBoxTemplate(ctx, data, pkg, owner, resolvedItem)
		case *widgets.GList:
			// list is a composite widget; no additional handling yet.
		case *core.GComponent, nil:
			if nested, nestedItem := f.buildNestedComponent(ctx, pkg, owner, child); nested != nil {
				obj.SetData(nested)
				if nestedItem != nil {
					resolvedItem = nestedItem
				}
				if (child.Width < 0 || child.Height < 0) && nested.Width() > 0 && nested.Height() > 0 {
					obj.SetSize(nested.Width(), nested.Height())
				}
			}
		default:
			// unsupported widget types fall back to existing behaviour.
		}
	}

	if child.Type == assets.ObjectTypeGroup {
		obj.SetTouchable(false)
	}

	if loader, ok := obj.Data().(*widgets.GLoader); ok && loader != nil {
		loader.RefreshLayout()
	}

	initW := obj.Width()
	initH := obj.Height()
	if child.Width >= 0 {
		initW = float64(child.Width)
	}
	if child.Height >= 0 {
		initH = float64(child.Height)
	}
	if initW < 0 {
		initW = 0
	}
	if initH < 0 {
		initH = 0
	}
	obj.SetSourceSize(initW, initH)
	obj.SetInitSize(initW, initH)

	return obj
}

func (f *Factory) resolvePackageItem(ctx context.Context, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild) *assets.PackageItem {
	if pkg == nil || child.Src == "" {
		// fall back to child.Data for components referencing packaged resources (e.g., buttons/lists)
		if child.Data == "" {
			return nil
		}
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
	if child.Src != "" {
		if item := target.ItemByID(child.Src); item != nil {
			return item
		}
		if item := target.ItemByName(child.Src); item != nil {
			return item
		}
	}
	if child.Data != "" {
		if item := target.ItemByID(child.Data); item != nil {
			return item
		}
		if item := target.ItemByName(child.Data); item != nil {
			return item
		}
	}
	return nil
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

func (f *Factory) buildNestedComponent(ctx context.Context, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild) (*core.GComponent, *assets.PackageItem) {
	nestedItem := f.resolvePackageItem(ctx, pkg, owner, child)
	if nestedItem == nil {
		return nil, nil
	}
	nested, err := f.BuildComponent(ctx, pkg, nestedItem)
	if err != nil {
		fmt.Println("builder: nested component error", err)
		return nil, nil
	}
	nested.SetResourceID(nestedItem.ID)
	return nested, nestedItem
}

func (f *Factory) resolveIcon(ctx context.Context, owner *assets.Package, icon string) *assets.PackageItem {
	if icon == "" {
		return nil
	}
	pkgCandidates, resourceCandidates := f.iconCandidates(owner, icon)
	for _, pkgKey := range pkgCandidates {
		pkg := f.lookupRegisteredPackage(pkgKey)
		if pkg == nil && f.packageResolver != nil {
			resolved, err := f.packageResolver(ctx, owner, pkgKey)
			if err != nil {
				fmt.Printf("builder: resolve icon package %s failed: %v\n", pkgKey, err)
			}
			if resolved != nil {
				f.RegisterPackage(resolved)
				if f.atlasManager != nil {
					if err := f.atlasManager.LoadPackage(ctx, resolved); err != nil {
						fmt.Printf("builder: load icon package failed: %v\n", err)
					}
				}
				pkg = resolved
			}
		}
		if pkg == nil {
			continue
		}
		for _, key := range resourceCandidates {
			if key == "" {
				continue
			}
			if item := pkg.ItemByID(key); item != nil {
				return item
			}
			if item := pkg.ItemByName(key); item != nil {
				return item
			}
		}
	}
	return nil
}

func (f *Factory) applyButtonTemplate(ctx context.Context, widget *widgets.GButton, pkg *assets.Package, owner *assets.PackageItem, item *assets.PackageItem) {
	if widget == nil || item == nil {
		return
	}
	if widget.TemplateComponent() == nil {
		targetPkg := item.Owner
		if targetPkg == nil {
			targetPkg = pkg
		}
		if targetPkg == nil && owner != nil {
			targetPkg = owner.Owner
		}
		if targetPkg != nil {
			if tmpl, err := f.BuildComponent(ctx, targetPkg, item); err == nil && tmpl != nil {
				widget.SetTemplateComponent(tmpl)
			}
		}
	}
	if template := widget.TemplateComponent(); template != nil {
		titleObj := template.ChildByName("title")
		if titleObj != nil {
			widget.SetTitleObject(titleObj)
		}
		iconObj := template.ChildByName("icon")
		if iconObj != nil {
			widget.SetIconObject(iconObj)
		}
		if ctrl := template.ControllerByName("button"); ctrl != nil {
			widget.SetButtonController(ctrl)
		} else if ctrl := template.ControllerByName("Button"); ctrl != nil {
			widget.SetButtonController(ctrl)
		}
	} else if item.Component != nil && len(widget.Controllers()) == 0 {
		for _, ctrlData := range item.Component.Controllers {
			ctrl := core.NewController(ctrlData.Name)
			ctrl.AutoRadio = ctrlData.AutoRadio
			ctrl.SetPages(ctrlData.PageIDs, ctrlData.PageNames)
			widget.AddController(ctrl)
			if strings.EqualFold(ctrl.Name, "button") {
				widget.SetButtonController(ctrl)
			}
		}
		if widget.ButtonController() == nil {
			for _, ctrl := range widget.Controllers() {
				widget.SetButtonController(ctrl)
				break
			}
		}
	}
	buf := item.RawData
	if buf == nil {
		return
	}
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(0, 6) {
		return
	}
	mode := widgets.ButtonMode(buf.ReadByte())
	widget.SetMode(mode)
	if sound := buf.ReadS(); sound != nil && *sound != "" {
		widget.SetSound(*sound)
	}
	widget.SetSoundVolumeScale(float64(buf.ReadFloat32()))
	widget.SetDownEffect(int(buf.ReadByte()))
	widget.SetDownEffectValue(float64(buf.ReadFloat32()))
	if widget.DownEffect() == 2 {
		widget.SetPivotWithAnchor(0.5, 0.5, widget.PivotAsAnchor())
	}
}

func (f *Factory) applyLabelTemplate(ctx context.Context, widget *widgets.GLabel, pkg *assets.Package, owner *assets.PackageItem, item *assets.PackageItem) {
	if widget == nil {
		return
	}
	if item != nil {
		widget.SetPackageItem(item)
	}
	if item == nil || item.Type != assets.PackageItemTypeComponent || item.Component == nil {
		return
	}
	if widget.TemplateComponent() == nil {
		targetPkg := item.Owner
		if targetPkg == nil {
			targetPkg = pkg
		}
		if targetPkg == nil && owner != nil {
			targetPkg = owner.Owner
		}
		if targetPkg != nil {
			if tmpl, err := f.BuildComponent(ctx, targetPkg, item); err == nil && tmpl != nil {
				widget.SetTemplateComponent(tmpl)
			}
		}
	}
	if template := widget.TemplateComponent(); template != nil {
		if titleObj := template.ChildByName("title"); titleObj != nil {
			widget.SetTitleObject(titleObj)
		}
		if iconObj := template.ChildByName("icon"); iconObj != nil {
			widget.SetIconObject(iconObj)
		}
	}
}

func (f *Factory) applyProgressBarTemplate(ctx context.Context, widget *widgets.GProgressBar, pkg *assets.Package, owner *assets.PackageItem, item *assets.PackageItem) {
	if widget == nil {
		return
	}
	if item != nil {
		widget.SetPackageItem(item)
	}
	if item == nil {
		item = widget.PackageItem()
	}
	if item == nil || item.Type != assets.PackageItemTypeComponent || item.RawData == nil {
		return
	}
	if widget.TemplateComponent() == nil {
		targetPkg := item.Owner
		if targetPkg == nil {
			targetPkg = pkg
		}
		if targetPkg == nil && owner != nil {
			targetPkg = owner.Owner
		}
		if targetPkg != nil {
			if tmpl, err := f.BuildComponent(ctx, targetPkg, item); err == nil && tmpl != nil {
				widget.SetTemplateComponent(tmpl)
			}
		}
	}
	buf := item.RawData
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(0, 6) || buf.Remaining() == 0 {
		return
	}
	widget.SetTitleType(widgets.ProgressTitleType(buf.ReadByte()))
	if buf.Remaining() > 0 {
		widget.SetReverse(buf.ReadBool())
	}
}

func (f *Factory) applySliderTemplate(ctx context.Context, widget *widgets.GSlider, pkg *assets.Package, owner *assets.PackageItem, item *assets.PackageItem) {
	if widget == nil {
		return
	}
	if item != nil {
		widget.SetPackageItem(item)
	}
	if item == nil {
		item = widget.PackageItem()
	}
	if item == nil || item.Type != assets.PackageItemTypeComponent || item.RawData == nil {
		return
	}
	if widget.TemplateComponent() == nil {
		targetPkg := item.Owner
		if targetPkg == nil {
			targetPkg = pkg
		}
		if targetPkg == nil && owner != nil {
			targetPkg = owner.Owner
		}
		if targetPkg != nil {
			if tmpl, err := f.BuildComponent(ctx, targetPkg, item); err == nil && tmpl != nil {
				widget.SetTemplateComponent(tmpl)
			}
		}
	}
	buf := item.RawData
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(0, 6) || buf.Remaining() == 0 {
		return
	}
	widget.SetTitleType(widgets.ProgressTitleType(buf.ReadByte()))
	if buf.Remaining() > 0 {
		widget.SetReverse(buf.ReadBool())
	}
	if buf.Version >= 2 {
		if buf.Remaining() > 0 {
			widget.SetWholeNumbers(buf.ReadBool())
		}
		if buf.Remaining() > 0 {
			widget.SetChangeOnClick(buf.ReadBool())
		}
	}
}

func (f *Factory) applyScrollBarTemplate(ctx context.Context, widget *widgets.GScrollBar, pkg *assets.Package, owner *assets.PackageItem, item *assets.PackageItem) {
	if widget == nil {
		return
	}
	if item != nil {
		widget.SetPackageItem(item)
	}
	if item == nil {
		item = widget.PackageItem()
	}
	if item == nil || item.Type != assets.PackageItemTypeComponent || item.RawData == nil {
		return
	}
	if widget.TemplateComponent() == nil {
		targetPkg := item.Owner
		if targetPkg == nil {
			targetPkg = pkg
		}
		if targetPkg == nil && owner != nil {
			targetPkg = owner.Owner
		}
		if targetPkg != nil {
			if tmpl, err := f.BuildComponent(ctx, targetPkg, item); err == nil && tmpl != nil {
				widget.SetTemplateComponent(tmpl)
			}
		}
	}
	buf := item.RawData
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(0, 6) || buf.Remaining() == 0 {
		return
	}
	widget.SetFixedGrip(buf.ReadBool())
}

func (f *Factory) applyComboBoxTemplate(ctx context.Context, widget *widgets.GComboBox, pkg *assets.Package, owner *assets.PackageItem, item *assets.PackageItem) {
	if widget == nil {
		return
	}
	if item != nil {
		widget.SetPackageItem(item)
	}
	if item == nil {
		item = widget.PackageItem()
	}
	if item == nil || item.Type != assets.PackageItemTypeComponent || item.RawData == nil {
		return
	}
	if widget.TemplateComponent() == nil {
		targetPkg := item.Owner
		if targetPkg == nil {
			targetPkg = pkg
		}
		if targetPkg == nil && owner != nil {
			targetPkg = owner.Owner
		}
		if targetPkg != nil {
			if tmpl, err := f.BuildComponent(ctx, targetPkg, item); err == nil && tmpl != nil {
				widget.SetTemplateComponent(tmpl)
			}
		}
	}

	if template := widget.TemplateComponent(); template != nil {
		if ctrl := template.ControllerByName("button"); ctrl != nil {
			widget.SetButtonController(ctrl)
		} else if ctrl := template.ControllerByName("Button"); ctrl != nil {
			widget.SetButtonController(ctrl)
		}
		if title := template.ChildByName("title"); title != nil {
			widget.SetTitleObject(title)
		}
		if icon := template.ChildByName("icon"); icon != nil {
			widget.SetIconObject(icon)
		}
	}

	buf := item.RawData
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(0, 6) {
		return
	}
	dropdownURL := buf.ReadS()
	if dropdownURL == nil || *dropdownURL == "" {
		return
	}
	url := *dropdownURL
	widget.SetDropdownURL(url)

	dropdownItem := f.resolveIcon(ctx, pkg, url)
	if dropdownItem == nil {
		return
	}
	widget.SetDropdownItem(dropdownItem)

	targetPkg := dropdownItem.Owner
	if targetPkg == nil {
		targetPkg = pkg
	}
	if targetPkg == nil && owner != nil {
		targetPkg = owner.Owner
	}
	if targetPkg == nil {
		return
	}
	dropdownComp, err := f.BuildComponent(ctx, targetPkg, dropdownItem)
	if err != nil || dropdownComp == nil {
		return
	}
	widget.SetDropdownComponent(dropdownComp)

	if listObj := dropdownComp.ChildByName("list"); listObj != nil {
		switch data := listObj.Data().(type) {
		case *widgets.GList:
			widget.SetList(data)
		case *core.GComponent:
			if embedded := data.ChildByName("list"); embedded != nil {
				if list, ok := embedded.Data().(*widgets.GList); ok {
					widget.SetList(list)
				}
			}
		}
	}
}

func (f *Factory) populateTree(ctx context.Context, tree *widgets.GTree, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild, buf *utils.ByteBuffer) {
	if tree == nil || buf == nil {
		return
	}
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(0, 8) {
		return
	}
	if def := buf.ReadS(); def != nil && *def != "" {
		tree.SetDefaultItem(*def)
	}
	if buf.Remaining() < 2 {
		return
	}
	count := int(buf.ReadInt16())
	var lastNode *widgets.GTreeNode
	prevLevel := 0
	for i := 0; i < count; i++ {
		if buf.Remaining() < 2 {
			break
		}
		nextPos := int(buf.ReadInt16()) + buf.Pos()
		resPtr := buf.ReadS()
		resource := stringFromPtr(resPtr)
		if resource == "" {
			resource = tree.DefaultItem()
		}
		if resource == "" {
			if nextPos >= 0 && nextPos <= buf.Len() {
				buf.SetPos(nextPos)
			}
			continue
		}
		isFolder := false
		if buf.Remaining() > 0 {
			isFolder = buf.ReadBool()
		}
		level := 0
		if buf.Remaining() > 0 {
			level = int(buf.ReadUint8())
		}
		node := widgets.NewTreeNode(isFolder, resource)
		if cell := f.buildTreeCell(ctx, tree, pkg, owner, resource); cell != nil {
			node.SetCell(cell)
		}
		if isFolder {
			node.SetExpanded(true)
		}
		var parentNode *widgets.GTreeNode
		if i == 0 {
			parentNode = tree.RootNode()
		} else if level > prevLevel {
			parentNode = lastNode
		} else if level == prevLevel {
			parentNode = lastNode.Parent()
		} else {
			cursor := lastNode
			for j := level; j <= prevLevel; j++ {
				if cursor != nil {
					cursor = cursor.Parent()
				}
			}
			parentNode = cursor
		}
		if parentNode == nil {
			parentNode = tree.RootNode()
		}
		parentNode.AddChild(node)
		f.setupTreeNodeItem(buf, node)
		lastNode = node
		prevLevel = level
		if nextPos >= 0 && nextPos <= buf.Len() {
			buf.SetPos(nextPos)
		} else {
			break
		}
	}
}

func (f *Factory) buildTreeCell(ctx context.Context, tree *widgets.GTree, pkg *assets.Package, owner *assets.PackageItem, resource string) *core.GComponent {
	if tree == nil {
		return nil
	}
	res := resource
	if res == "" {
		res = tree.DefaultItem()
	}
	if res == "" {
		return nil
	}
	item := f.resolveIcon(ctx, pkg, res)
	if item == nil && owner != nil && owner.Owner != nil {
		if alt := owner.Owner.ItemByID(res); alt != nil {
			item = alt
		} else if alt := owner.Owner.ItemByName(res); alt != nil {
			item = alt
		}
	}
	if item == nil {
		return nil
	}
	if item.Type != assets.PackageItemTypeComponent || item.Component == nil {
		return nil
	}
	targetPkg := item.Owner
	if targetPkg == nil {
		targetPkg = pkg
	}
	if targetPkg == nil {
		return nil
	}
	comp, err := f.BuildComponent(ctx, targetPkg, item)
	if err != nil || comp == nil {
		return nil
	}
	return comp
}

func (f *Factory) setupTreeNodeItem(buf *utils.ByteBuffer, node *widgets.GTreeNode) {
	if buf == nil || node == nil {
		return
	}
	if text := stringFromPtr(buf.ReadS()); text != "" {
		node.SetText(text)
	}
	_ = buf.ReadS() // selected title
	if icon := stringFromPtr(buf.ReadS()); icon != "" {
		node.SetIcon(icon)
	}
	_ = buf.ReadS() // selected icon
	if name := stringFromPtr(buf.ReadS()); name != "" {
		if cell := node.Cell(); cell != nil {
			cell.GObject.SetName(name)
		}
	}
	cell := node.Cell()
	if buf.Remaining() >= 2 {
		cnt := int(buf.ReadInt16())
		for i := 0; i < cnt; i++ {
			ctrlName := stringFromPtr(buf.ReadS())
			pageID := stringFromPtr(buf.ReadS())
			if cell != nil && ctrlName != "" && pageID != "" {
				if ctrl := cell.ControllerByName(ctrlName); ctrl != nil {
					ctrl.SetSelectedPageID(pageID)
				}
			}
		}
	}
	if buf.Version >= 2 && buf.Remaining() >= 2 {
		assigns := int(buf.ReadInt16())
		for i := 0; i < assigns; i++ {
			targetPath := stringFromPtr(buf.ReadS())
			propID := gears.ObjectPropID(buf.ReadInt16())
			value := stringFromPtr(buf.ReadS())
			if cell == nil || targetPath == "" || value == "" {
				continue
			}
			if target := core.FindChildByPath(cell, targetPath); target != nil {
				target.SetProp(propID, value)
			}
		}
	}
}

func stringFromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (f *Factory) assignLoaderPackage(ctx context.Context, loader *widgets.GLoader, currentPkg *assets.Package, item *assets.PackageItem) {
	if loader == nil {
		return
	}
	loader.SetComponent(nil)
	loader.SetPackageItem(nil)
	if item == nil {
		if resolved := f.resolveIcon(ctx, currentPkg, loader.URL()); resolved != nil {
			item = resolved
		} else {
			return
		}
	}
	loader.SetPackageItem(item)
	loader.SetScale9Grid(nil)
	if item.Scale9Grid != nil {
		loader.SetScale9Grid(item.Scale9Grid)
	}
	loader.SetScaleByTile(item.ScaleByTile)
	loader.SetTileGridIndice(item.TileGridIndice)
	if item.Type == assets.PackageItemTypeComponent {
		targetPkg := item.Owner
		if targetPkg == nil {
			targetPkg = currentPkg
		}
		nested, err := f.BuildComponent(ctx, targetPkg, item)
		if err == nil && nested != nil {
			loader.SetComponent(nested)
		}
	}
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

func (f *Factory) setupRelations(item *assets.PackageItem, comp *core.GComponent) {
	if item == nil || item.RawData == nil || comp == nil {
		return
	}
	buf := item.RawData
	saved := buf.Pos()
	defer buf.SetPos(saved)

	if buf.Seek(0, 3) {
		parseRelations(buf, comp.GObject, comp, true)
	}

	if !buf.Seek(0, 2) {
		return
	}
	buf.Skip(2)
	childCount := len(comp.Children())
	for i := 0; i < childCount; i++ {
		nextPos := int(buf.ReadInt16()) + buf.Pos()
		current := buf.Pos()
		if buf.Seek(current, 3) {
			child := comp.ChildAt(i)
			parseRelations(buf, child, comp, false)
		}
		buf.SetPos(nextPos)
	}
}

func parseRelations(buf *utils.ByteBuffer, owner *core.GObject, parent *core.GComponent, parentToChild bool) {
	if buf == nil || owner == nil {
		return
	}
	cnt := int(buf.ReadByte())
	for i := 0; i < cnt; i++ {
		targetIndex := int(buf.ReadInt16())
		var target *core.GObject
		if targetIndex == -1 {
			if owner.Parent() != nil {
				target = owner.Parent().GObject
			}
		} else {
			var container *core.GComponent
			if parentToChild {
				container = parent
			} else {
				container = owner.Parent()
			}
			if container != nil {
				target = container.ChildAt(targetIndex)
			}
		}
		relCount := int(buf.ReadByte())
		for j := 0; j < relCount; j++ {
			relation := core.RelationType(buf.ReadByte())
			usePercent := buf.ReadBool()
			if target != nil {
				owner.AddRelation(target, relation, usePercent)
			}
		}
	}
}

func (f *Factory) setupGears(item *assets.PackageItem, comp *core.GComponent) {
	if item == nil || item.RawData == nil || comp == nil {
		return
	}
	buf := item.RawData
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(0, 2) {
		return
	}
	if err := buf.Skip(2); err != nil {
		return
	}
	childCount := len(comp.Children())
	resolver := newComponentControllerResolver(comp)
	for i := 0; i < childCount; i++ {
		nextPos := int(buf.ReadInt16()) + buf.Pos()
		start := buf.Pos()
		child := comp.ChildAt(i)
		length := nextPos - start
		if child != nil && length > 0 {
			if sub, err := item.RawData.SubBuffer(start, length); err == nil {
				setupObjectGears(child, resolver, sub)
			}
			child.SetupAfterAdd(comp, item.RawData, start)
		}
		buf.SetPos(nextPos)
	}
}

func setupObjectGears(obj *core.GObject, resolver componentControllerResolver, buf *utils.ByteBuffer) {
	if obj == nil || buf == nil {
		return
	}
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(0, 2) {
		return
	}
	gearCount := int(buf.ReadInt16())
	for i := 0; i < gearCount; i++ {
		nextPos := int(buf.ReadInt16()) + buf.Pos()
		index := int(buf.ReadByte())
		if gear := obj.GetGear(index); gear != nil {
			gear.Setup(buf, resolver)
			gear.Apply()
		}
		buf.SetPos(nextPos)
	}
}

type componentControllerResolver struct {
	controllers []gears.Controller
}

type componentSetupResolver struct {
	component *core.GComponent
	item      *assets.PackageItem
}

func (r componentSetupResolver) MaskChild(index int) *core.GObject {
	if r.component == nil || index < 0 {
		return nil
	}
	return r.component.ChildAt(index)
}

func (r componentSetupResolver) PixelData(itemID string) *assets.PixelHitTestData {
	if itemID == "" || r.item == nil || r.item.Owner == nil {
		return nil
	}
	if pi := r.item.Owner.ItemByID(itemID); pi != nil {
		return pi.PixelHitTest
	}
	return nil
}

func (r componentSetupResolver) Configure(comp *core.GComponent, hit core.HitTest, data *assets.PixelHitTestData) {
	render.ConfigureComponentHitArea(comp, hit, data)
}

func newComponentControllerResolver(comp *core.GComponent) componentControllerResolver {
	resolver := componentControllerResolver{}
	if comp == nil {
		return resolver
	}
	ctrls := comp.Controllers()
	if len(ctrls) == 0 {
		return resolver
	}
	resolver.controllers = make([]gears.Controller, len(ctrls))
	for i, ctrl := range ctrls {
		resolver.controllers[i] = ctrl
	}
	return resolver
}

func (r componentControllerResolver) ControllerAt(index int) gears.Controller {
	if index < 0 || index >= len(r.controllers) {
		return nil
	}
	return r.controllers[index]
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

func (f *Factory) iconCandidates(owner *assets.Package, icon string) ([]string, []string) {
	icon = strings.TrimSpace(icon)
	if icon == "" {
		return nil, nil
	}
	body := icon
	if strings.HasPrefix(body, "ui://") {
		body = body[len("ui://"):]
	}
	pkgCandidates := make([]string, 0, 4)
	resourceCandidates := make([]string, 0, 3)
	if idx := strings.Index(body, "/"); idx >= 0 {
		pkgKey := body[:idx]
		res := body[idx+1:]
		if pkgKey != "" {
			pkgCandidates = append(pkgCandidates, pkgKey)
		}
		if res != "" {
			resourceCandidates = append(resourceCandidates, res)
		}
	} else {
		if len(body) > 8 {
			pkgCandidates = append(pkgCandidates, body[:8])
			resourceCandidates = append(resourceCandidates, body[8:])
		}
		resourceCandidates = append(resourceCandidates, body)
	}
	if owner != nil {
		if owner.ID != "" {
			pkgCandidates = append(pkgCandidates, owner.ID)
		}
		if owner.Name != "" {
			pkgCandidates = append(pkgCandidates, owner.Name)
		}
	}
	return uniqueStrings(pkgCandidates), uniqueStrings(resourceCandidates)
}

func uniqueStrings(values []string) []string {
	var out []string
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
