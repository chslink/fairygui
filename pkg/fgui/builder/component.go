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

	for _, child := range item.Component.Children {
		childObj := f.buildChild(ctx, pkg, item, root, &child)
		root.AddChild(childObj)
	}

	f.setupRelations(item, root)
	f.setupGears(item, root)
	setupObjectAfterAdd(root.GObject, root, item.RawData, 0)
	f.applyComponentMaskAndHitTest(root, item)
	f.applyComponentTransitions(root, item)

	f.finalizeComponentSize(root)
	return root, nil
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
	var obj *core.GObject
	switch widget := w.(type) {
	case *widgets.GImage:
		obj = widget.GObject
		if spriteItem := f.resolveImageSprite(ctx, pkg, owner, child); spriteItem != nil {
			resolvedItem = spriteItem
		}
		widget.SetPackageItem(resolvedItem)
		obj.SetData(widget)
	case *widgets.GTextField:
		obj = widget.GObject
		widget.SetText(child.Text)
		f.applyTextFieldSettings(widget, owner, child)
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
		f.applyButtonInstanceSettings(widget, parent, owner, child, resolvedItem)
		obj.SetData(widget)
	case *widgets.GLoader:
		obj = widget.GObject
		obj.SetData(widget)
		if child.Src != "" {
			widget.SetURL(child.Src)
		}
		autoSize := child.Width < 0 || child.Height < 0
		widget.SetAutoSize(autoSize)
		f.applyLoaderSettings(ctx, widget, pkg, owner, child)
		if resolvedItem == nil {
			resolvedItem = f.resolveIcon(ctx, pkg, widget.URL())
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
		f.applyLabelInstanceSettings(ctx, widget, pkg, owner, child, resolvedItem)
	case *widgets.GList:
		obj = widget.GComponent.GObject
		widget.SetResource(child.Data)
		widget.SetDefaultItem(child.Data)
		widget.SetPackageItem(resolvedItem)
		obj.SetData(widget)
	case *widgets.GGroup:
		obj = widget.GObject
	case *widgets.GGraph:
		obj = widget.GObject
		obj.SetData(widget)
		f.applyGraphSettings(widget, owner, child)
	default:
		obj = core.NewGObject()
	}
	obj.SetName(child.Name)
	rid := child.ID
	if resolvedItem != nil && resolvedItem.ID != "" {
		rid = resolvedItem.ID
	}
	obj.SetResourceID(rid)
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
	obj.SetTouchable(child.Touchable)
	obj.SetGrayed(child.Grayed)
	obj.SetAlpha(float64(child.Alpha))
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
			f.applyLabelInstanceSettings(ctx, data, pkg, owner, child, resolvedItem)
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

func (f *Factory) applyLoaderSettings(ctx context.Context, loader *widgets.GLoader, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild) {
	if loader == nil || owner == nil || owner.RawData == nil || child == nil || child.RawDataLength <= 0 {
		return
	}
	sub, err := owner.RawData.SubBuffer(child.RawDataOffset, child.RawDataLength)
	if err != nil {
		return
	}
	if !sub.Seek(0, 5) {
		return
	}
	if url := sub.ReadS(); url != nil && *url != "" {
		loader.SetURL(*url)
	}

	mapAlign := func(code int8, horizontal bool) widgets.LoaderAlign {
		switch code {
		case 1:
			if horizontal {
				return widgets.LoaderAlignCenter
			}
			return widgets.LoaderAlignMiddle
		case 2:
			if horizontal {
				return widgets.LoaderAlignRight
			}
			return widgets.LoaderAlignBottom
		default:
			if horizontal {
				return widgets.LoaderAlignLeft
			}
			return widgets.LoaderAlignTop
		}
	}

	loader.SetAlign(mapAlign(sub.ReadByte(), true))
	loader.SetVerticalAlign(mapAlign(sub.ReadByte(), false))

	loader.SetFill(widgets.LoaderFillType(sub.ReadByte()))
	loader.SetShrinkOnly(sub.ReadBool())
	loader.SetAutoSize(sub.ReadBool())

	_ = sub.ReadBool() // showErrorSign
	loader.SetPlaying(sub.ReadBool())
	loader.SetFrame(int(sub.ReadInt32()))

	if sub.ReadBool() {
		loader.SetColor(sub.ReadColorString(true))
	}

	fillMethod := sub.ReadByte()
	loader.SetFillMethod(int(fillMethod))
	if fillMethod != 0 {
		loader.SetFillOrigin(int(sub.ReadByte()))
		loader.SetFillClockwise(sub.ReadBool())
		loader.SetFillAmount(float64(sub.ReadFloat32()))
	}

	if sub.Version >= 7 {
		loader.SetUseResize(sub.ReadBool())
	}

	if pkg == nil && owner != nil {
		pkg = owner.Owner
	}
	if resolved := f.resolveIcon(ctx, pkg, loader.URL()); resolved != nil {
		f.assignLoaderPackage(ctx, loader, pkg, resolved)
	}

	loader.RefreshLayout()
}

func (f *Factory) applyTextFieldSettings(widget *widgets.GTextField, owner *assets.PackageItem, child *assets.ComponentChild) {
	if widget == nil || owner == nil || owner.RawData == nil || child == nil || child.RawDataLength <= 0 {
		return
	}
	sub, err := owner.RawData.SubBuffer(child.RawDataOffset, child.RawDataLength)
	if err != nil {
		return
	}
	if !sub.Seek(0, 5) {
		return
	}
	if font := sub.ReadS(); font != nil && *font != "" {
		widget.SetFont(*font)
	}
	if size := int(sub.ReadInt16()); size > 0 {
		widget.SetFontSize(size)
	}
	if color := sub.ReadColorString(true); color != "" {
		widget.SetColor(color)
	}
	mapAlign := func(code int8) widgets.TextAlign {
		switch code {
		case 1:
			return widgets.TextAlignCenter
		case 2:
			return widgets.TextAlignRight
		default:
			return widgets.TextAlignLeft
		}
	}
	mapVAlign := func(code int8) widgets.TextVerticalAlign {
		switch code {
		case 1:
			return widgets.TextVerticalAlignMiddle
		case 2:
			return widgets.TextVerticalAlignBottom
		default:
			return widgets.TextVerticalAlignTop
		}
	}
	widget.SetAlign(mapAlign(sub.ReadByte()))
	widget.SetVerticalAlign(mapVAlign(sub.ReadByte()))
	widget.SetLeading(int(sub.ReadInt16()))
	widget.SetLetterSpacing(int(sub.ReadInt16()))
	widget.SetUBBEnabled(sub.ReadBool())
	widget.SetAutoSize(widgets.TextAutoSize(sub.ReadByte()))
	widget.SetUnderline(sub.ReadBool())
	widget.SetItalic(sub.ReadBool())
	widget.SetBold(sub.ReadBool())
	widget.SetSingleLine(sub.ReadBool())
	if sub.ReadBool() {
		if strokeColor := sub.ReadColorString(true); strokeColor != "" {
			widget.SetStrokeColor(strokeColor)
		}
		widget.SetStrokeSize(float64(sub.ReadFloat32()) + 1)
	}
	if sub.ReadBool() {
		_ = sub.Skip(12) // shadow data currently unused
	}
	if sub.ReadBool() {
		widget.SetTemplateVarsEnabled(true)
	}
	if sub.Seek(0, 6) {
		if text := sub.ReadS(); text != nil {
			widget.SetText(*text)
		}
	}
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

func (f *Factory) applyButtonInstanceSettings(widget *widgets.GButton, parent *core.GComponent, owner *assets.PackageItem, child *assets.ComponentChild, resolved *assets.PackageItem) {
	if widget == nil || owner == nil || owner.RawData == nil || child == nil || child.RawDataLength <= 0 {
		return
	}
	sub, err := owner.RawData.SubBuffer(child.RawDataOffset, child.RawDataLength)
	if err != nil {
		return
	}
	if !sub.Seek(0, 6) || sub.Remaining() <= 0 {
		return
	}
	objType := assets.ObjectType(sub.ReadByte())
	if objType != child.Type {
		if resolved != nil && objType == resolved.ObjectType {
			// ok: specialised component (e.g., Button)
		} else if child.Type == assets.ObjectTypeComponent {
			// allow components that encode their runtime object type here
		} else {
			return
		}
	}
	readS := func() *string {
		if sub.Remaining() < 2 {
			return nil
		}
		return sub.ReadS()
	}
	readBool := func() (bool, bool) {
		if sub.Remaining() < 1 {
			return false, false
		}
		return sub.ReadBool(), true
	}
	readInt16 := func() (int16, bool) {
		if sub.Remaining() < 2 {
			return 0, false
		}
		return sub.ReadInt16(), true
	}
	readInt32 := func() (int32, bool) {
		if sub.Remaining() < 4 {
			return 0, false
		}
		return sub.ReadInt32(), true
	}
	readFloat32 := func() (float32, bool) {
		if sub.Remaining() < 4 {
			return 0, false
		}
		return sub.ReadFloat32(), true
	}

	if title := readS(); title != nil && *title != "" {
		widget.SetTitle(*title)
	}
	if selectedTitle := readS(); selectedTitle != nil && *selectedTitle != "" {
		widget.SetSelectedTitle(*selectedTitle)
	}
	if icon := readS(); icon != nil && *icon != "" {
		widget.SetIcon(*icon)
	}
	if selectedIcon := readS(); selectedIcon != nil && *selectedIcon != "" {
		widget.SetSelectedIcon(*selectedIcon)
	}
	if hasColor, ok := readBool(); ok && hasColor {
		if sub.Remaining() >= 4 {
			if color := sub.ReadColorString(true); color != "" {
				widget.SetTitleColor(color)
			}
		}
	}
	if size, ok := readInt32(); ok && size != 0 {
		widget.SetTitleFontSize(int(size))
	}
	if idx, ok := readInt16(); ok && idx >= 0 && parent != nil {
		controllers := parent.Controllers()
		if int(idx) < len(controllers) {
			widget.SetRelatedController(controllers[idx])
		}
	}
	if page := readS(); page != nil {
		widget.SetRelatedPageID(*page)
	}
	if sound := readS(); sound != nil && *sound != "" {
		widget.SetSound(*sound)
	}
	if hasVolume, ok := readBool(); ok && hasVolume {
		if vol, ok := readFloat32(); ok {
			widget.SetSoundVolumeScale(float64(vol))
		}
	}
	if selected, ok := readBool(); ok {
		widget.SetSelected(selected)
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

func (f *Factory) applyLabelInstanceSettings(ctx context.Context, widget *widgets.GLabel, pkg *assets.Package, owner *assets.PackageItem, child *assets.ComponentChild, item *assets.PackageItem) {
	if widget == nil || owner == nil || owner.RawData == nil || child == nil || child.RawDataLength <= 0 {
		return
	}
	sub, err := owner.RawData.SubBuffer(child.RawDataOffset, child.RawDataLength)
	if err != nil {
		return
	}
	if !sub.Seek(0, 6) || sub.Remaining() <= 0 {
		return
	}
	objType := assets.ObjectType(sub.ReadByte())
	isLabelObj := objType == assets.ObjectTypeLabel
	if item != nil && objType != item.ObjectType {
		return
	}
	if item == nil && !isLabelObj {
		return
	}
	if title := sub.ReadS(); title != nil {
		widget.SetTitle(*title)
	}
	if icon := sub.ReadS(); icon != nil {
		widget.SetIcon(*icon)
		if iconItem := f.resolveIcon(ctx, pkg, *icon); iconItem != nil {
			widget.SetIconItem(iconItem)
		}
	}
	if sub.ReadBool() {
		widget.SetTitleColor(sub.ReadColorString(true))
	}
	if size := sub.ReadInt32(); size != 0 {
		widget.SetTitleFontSize(int(size))
	}
	if sub.ReadBool() {
		_ = sub.ReadS()
		_ = sub.ReadS()
		_ = sub.ReadInt32()
		_ = sub.ReadInt32()
		_ = sub.ReadBool()
	}
}

func (f *Factory) applyGraphSettings(widget *widgets.GGraph, owner *assets.PackageItem, child *assets.ComponentChild) {
	if widget == nil || owner == nil || owner.RawData == nil || child == nil || child.RawDataLength <= 0 {
		return
	}
	sub, err := owner.RawData.SubBuffer(child.RawDataOffset, child.RawDataLength)
	if err != nil {
		return
	}
	if !sub.Seek(0, 5) || sub.Remaining() <= 0 {
		return
	}
	typeByte := sub.ReadByte()
	widget.SetType(widgets.GraphType(typeByte))
	if widget.Type() == widgets.GraphTypeEmpty {
		return
	}
	if sub.Remaining() >= 4 {
		widget.SetLineSize(float64(sub.ReadInt32()))
	}
	if sub.Remaining() >= 4 {
		widget.SetLineColor(sub.ReadColorString(true))
	}
	if sub.Remaining() >= 4 {
		widget.SetFillColor(sub.ReadColorString(true))
	}
	if sub.Remaining() > 0 && sub.ReadBool() {
		radii := make([]float64, 0, 4)
		for i := 0; i < 4 && sub.Remaining() >= 4; i++ {
			radii = append(radii, float64(sub.ReadFloat32()))
		}
		widget.SetCornerRadius(radii)
	}
	switch widget.Type() {
	case widgets.GraphTypePolygon:
		if sub.Remaining() >= 2 {
			cnt := int(sub.ReadInt16())
			points := make([]float64, 0, cnt)
			for i := 0; i < cnt && sub.Remaining() >= 4; i++ {
				points = append(points, float64(sub.ReadFloat32()))
			}
			widget.SetPolygonPoints(points)
		}
	case widgets.GraphTypeRegularPolygon:
		sides := 0
		if sub.Remaining() >= 2 {
			sides = int(sub.ReadInt16())
		}
		angle := 0.0
		if sub.Remaining() >= 4 {
			angle = float64(sub.ReadFloat32())
		}
		var distances []float64
		if sub.Remaining() >= 2 {
			cnt := int(sub.ReadInt16())
			distances = make([]float64, 0, cnt)
			for i := 0; i < cnt && sub.Remaining() >= 4; i++ {
				distances = append(distances, float64(sub.ReadFloat32()))
			}
		}
		widget.SetRegularPolygon(sides, angle, distances)
	}
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
				setupComponentControllers(child, sub)
			}
			setupObjectAfterAdd(child, comp, item.RawData, start)
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

func setupObjectAfterAdd(obj *core.GObject, parent *core.GComponent, buf *utils.ByteBuffer, start int) {
	if obj == nil || buf == nil || start < 0 {
		return
	}
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if buf.Seek(start, 1) {
		if buf.Remaining() >= 2 {
			if tip := buf.ReadS(); tip != nil {
				obj.SetTooltips(*tip)
			}
		}
		if buf.Remaining() >= 2 {
			groupIndex := int(buf.ReadInt16())
			if groupIndex >= 0 && parent != nil {
				if groupObj := parent.ChildAt(groupIndex); groupObj != nil {
					obj.SetGroup(groupObj)
				}
			}
		}
	}
	component := extractComponent(obj)
	if component == nil && parent != nil && parent.GObject == obj {
		component = parent
	}
	if component != nil {
		if buf.Seek(start, 4) {
			if buf.Remaining() >= 2 {
				_ = buf.ReadInt16() // scroll pane page controller index (not wired yet)
			}
			if buf.Remaining() >= 2 {
				cnt := int(buf.ReadInt16())
				for i := 0; i < cnt; i++ {
					if buf.Remaining() < 4 {
						break
					}
					ctrlName := buf.ReadS()
					pageID := buf.ReadS()
					if ctrlName == nil || pageID == nil {
						continue
					}
					if ctrl := component.ControllerByName(*ctrlName); ctrl != nil {
						ctrl.SetSelectedPageID(*pageID)
					}
				}
			}
			if buf.Version >= 2 {
				if buf.Remaining() >= 2 {
					assignments := int(buf.ReadInt16())
					for i := 0; i < assignments; i++ {
						if buf.Remaining() < 4 {
							break
						}
						targetPath := buf.ReadS()
						if buf.Remaining() < 2 {
							break
						}
						propID := buf.ReadInt16()
						if buf.Remaining() < 2 {
							break
						}
						value := buf.ReadS()
						if targetPath == nil || value == nil {
							continue
						}
						if target := findChildByPath(component, *targetPath); target != nil {
							target.SetProp(gears.ObjectPropID(propID), *value)
						}
					}
				}
			}
		}
	}
}

func setupComponentControllers(obj *core.GObject, buf *utils.ByteBuffer) {
	component := extractComponent(obj)
	if component == nil || buf == nil {
		return
	}
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(0, 4) {
		return
	}
	// Page controller index is used when a scroll pane is present; we don't have scroll pane wiring yet.
	_ = buf.ReadInt16()

	count := int(buf.ReadInt16())
	for i := 0; i < count; i++ {
		ctrlName := buf.ReadS()
		pageID := buf.ReadS()
		if ctrlName == nil || pageID == nil {
			continue
		}
		if ctrl := component.ControllerByName(*ctrlName); ctrl != nil {
			ctrl.SetSelectedPageID(*pageID)
		}
	}

	if buf.Version >= 2 {
		assignments := int(buf.ReadInt16())
		for i := 0; i < assignments; i++ {
			targetPath := buf.ReadS()
			propID := buf.ReadInt16()
			value := buf.ReadS()
			if targetPath == nil || value == nil {
				continue
			}
			if target := findChildByPath(component, *targetPath); target != nil {
				target.SetProp(gears.ObjectPropID(propID), *value)
			}
		}
	}
}

func (f *Factory) applyComponentMaskAndHitTest(comp *core.GComponent, item *assets.PackageItem) {
	if comp == nil || item == nil || item.RawData == nil {
		return
	}
	buf := item.RawData
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(0, 4) {
		return
	}
	if err := buf.Skip(2); err != nil {
		return
	}
	if buf.Remaining() < 1+2+2+4+4 {
		return
	}
	comp.SetOpaque(buf.ReadBool())
	if buf.Remaining() < 2 {
		comp.SetMask(nil, false)
		comp.SetHitTest(core.HitTest{Mode: core.HitTestModeNone})
		return
	}
	maskIndex := int(buf.ReadInt16())
	if maskIndex != -1 {
		if buf.Remaining() < 1 {
			comp.SetMask(nil, false)
		} else {
			reversed := buf.ReadBool()
			if maskChild := comp.ChildAt(maskIndex); maskChild != nil {
				comp.SetMask(maskChild, reversed)
			} else {
				comp.SetMask(nil, reversed)
			}
		}
	} else {
		comp.SetMask(nil, false)
	}

	if buf.Remaining() < 2 {
		comp.SetHitTest(core.HitTest{Mode: core.HitTestModeNone})
		return
	}
	hitTestID := buf.ReadS()
	if buf.Remaining() < 8 {
		comp.SetHitTest(core.HitTest{Mode: core.HitTestModeNone})
		return
	}
	offsetX := int(buf.ReadInt32())
	offsetY := int(buf.ReadInt32())
	hit := core.HitTest{Mode: core.HitTestModeNone}
	if hitTestID != nil && *hitTestID != "" {
		hit.Mode = core.HitTestModePixel
		hit.ItemID = *hitTestID
		hit.OffsetX = offsetX
		hit.OffsetY = offsetY
	} else if offsetX != 0 && offsetY != -1 {
		hit.Mode = core.HitTestModeChild
		hit.OffsetX = offsetX
		hit.ChildIndex = offsetY
	}
	comp.SetHitTest(hit)

	var pixelData *assets.PixelHitTestData
	if hit.Mode == core.HitTestModePixel && hit.ItemID != "" && item.Owner != nil {
		if pi := item.Owner.ItemByID(hit.ItemID); pi != nil {
			pixelData = pi.PixelHitTest
		}
	}
	render.ConfigureComponentHitArea(comp, hit, pixelData)
}

func (f *Factory) applyComponentTransitions(comp *core.GComponent, item *assets.PackageItem) {
	if comp == nil || item == nil || item.RawData == nil {
		return
	}
	buf := item.RawData
	saved := buf.Pos()
	defer buf.SetPos(saved)
	if !buf.Seek(0, 5) {
		return
	}
	count := int(buf.ReadInt16())
	if count <= 0 {
		return
	}
	for i := 0; i < count; i++ {
		nextPos := int(buf.ReadInt16()) + buf.Pos()
		if nextPos > buf.Len() {
			nextPos = buf.Len()
		}
		if nextPos <= buf.Pos() {
			buf.SetPos(nextPos)
			continue
		}
		info := core.TransitionInfo{}
		remaining := func() int { return nextPos - buf.Pos() }
		if remaining() >= 2 {
			if name := buf.ReadS(); name != nil {
				info.Name = *name
			}
		}
		if remaining() >= 4 {
			info.Options = int(buf.ReadInt32())
		}
		if remaining() >= 1 {
			info.AutoPlay = buf.ReadBool()
		}
		if remaining() >= 4 {
			info.AutoPlayTimes = int(buf.ReadInt32())
		}
		if remaining() >= 4 {
			info.AutoPlayDelay = float64(buf.ReadFloat32())
		}
		itemCount := 0
		if remaining() >= 2 {
			itemCount = int(buf.ReadInt16())
		}
		info.Items = make([]core.TransitionItem, 0, itemCount)
		maxDuration := 0.0
		for j := 0; j < itemCount; j++ {
			if buf.Pos() >= nextPos || nextPos-buf.Pos() < 2 {
				break
			}
			dataLen := int(buf.ReadInt16())
			curPos := buf.Pos()
			if dataLen < 0 || curPos+dataLen > nextPos {
				buf.SetPos(nextPos)
				break
			}
			if parsed := parseTransitionItem(buf, comp, curPos, dataLen); parsed != nil {
				end := parsed.Time
				if parsed.Tween != nil {
					end += parsed.Tween.Duration
				}
				if end > maxDuration {
					maxDuration = end
				}
				info.Items = append(info.Items, *parsed)
			}
			buf.SetPos(curPos + dataLen)
		}
		info.ItemCount = len(info.Items)
		info.TotalDuration = maxDuration
		if info.ItemCount > 0 || info.Name != "" {
			comp.AddTransition(info)
		}
		buf.SetPos(nextPos)
	}
}

func parseTransitionItem(buf *utils.ByteBuffer, comp *core.GComponent, start int, length int) *core.TransitionItem {
	if buf == nil || length <= 0 {
		return nil
	}
	saved := buf.Pos()
	defer buf.SetPos(saved)
	limit := start + length
	if limit > buf.Len() || !buf.Seek(start, 0) {
		return nil
	}
	rem := func() int { return limit - buf.Pos() }
	if rem() <= 0 {
		return nil
	}
	action := transitionActionFromByte(int(buf.ReadByte()))
	item := core.TransitionItem{Type: action}
	if rem() >= 4 {
		item.Time = float64(buf.ReadFloat32())
	} else {
		return nil
	}
	if rem() >= 2 {
		targetIndex := int(buf.ReadInt16())
		if targetIndex >= 0 {
			item.TargetID = resolveTransitionTargetID(comp, targetIndex)
		}
	}
	if rem() >= 2 {
		if label := buf.ReadS(); label != nil {
			item.Label = *label
		}
	}
	hasTween := false
	if rem() > 0 {
		hasTween = buf.ReadBool()
	}
	if hasTween {
		tween := core.TransitionTween{
			Start: core.TransitionValue{B1: true, B2: true},
			End:   core.TransitionValue{B1: true, B2: true},
		}
		if buf.Seek(start, 1) {
			if limit-buf.Pos() >= 4 {
				tween.Duration = float64(buf.ReadFloat32())
			}
			if limit-buf.Pos() >= 1 {
				tween.EaseType = int(buf.ReadByte())
			}
			if limit-buf.Pos() >= 4 {
				tween.Repeat = int(buf.ReadInt32())
			}
			if limit-buf.Pos() >= 1 {
				tween.Yoyo = buf.ReadBool()
			}
			if limit-buf.Pos() >= 2 {
				if endLabel := buf.ReadS(); endLabel != nil {
					tween.EndLabel = *endLabel
				}
			}
		}
		if buf.Seek(start, 2) {
			decodeTransitionValue(buf, limit, action, &tween.Start)
		}
		if buf.Seek(start, 3) {
			decodeTransitionValue(buf, limit, action, &tween.End)
			if buf.Version >= 2 && limit-buf.Pos() >= 4 {
				pathLen := int(buf.ReadInt32())
				if pathLen > 0 {
					points := make([]core.TransitionPathPoint, 0, pathLen)
					for p := 0; p < pathLen; p++ {
						if limit-buf.Pos() < 1 {
							break
						}
						curveType := int(buf.ReadUint8())
						point := core.TransitionPathPoint{CurveType: curveType}
						if limit-buf.Pos() < 8 {
							points = append(points, point)
							break
						}
						point.X = float64(buf.ReadFloat32())
						point.Y = float64(buf.ReadFloat32())
						switch curveType {
						case 1:
							if limit-buf.Pos() >= 8 {
								point.CX1 = float64(buf.ReadFloat32())
								point.CY1 = float64(buf.ReadFloat32())
							}
						case 2:
							if limit-buf.Pos() >= 16 {
								point.CX1 = float64(buf.ReadFloat32())
								point.CY1 = float64(buf.ReadFloat32())
								point.CX2 = float64(buf.ReadFloat32())
								point.CY2 = float64(buf.ReadFloat32())
							} else if limit-buf.Pos() >= 8 {
								point.CX1 = float64(buf.ReadFloat32())
								point.CY1 = float64(buf.ReadFloat32())
							}
						}
						points = append(points, point)
					}
					tween.Path = points
				}
			}
		}
		item.Tween = &tween
	} else {
		if buf.Seek(start, 2) {
			decodeTransitionValue(buf, limit, action, &item.Value)
		}
	}
	return &item
}

func transitionActionFromByte(value int) core.TransitionAction {
	if value < 0 || value > int(core.TransitionActionUnknown) {
		return core.TransitionActionUnknown
	}
	return core.TransitionAction(value)
}

func resolveTransitionTargetID(comp *core.GComponent, index int) string {
	if comp == nil {
		return ""
	}
	child := comp.ChildAt(index)
	if child == nil {
		return ""
	}
	if id := child.ResourceID(); id != "" {
		return id
	}
	if name := child.Name(); name != "" {
		return name
	}
	return child.ID()
}

func decodeTransitionValue(buf *utils.ByteBuffer, limit int, action core.TransitionAction, out *core.TransitionValue) {
	if out == nil {
		return
	}
	readBool := func() bool {
		if limit-buf.Pos() < 1 {
			return false
		}
		return buf.ReadBool()
	}
	readFloat := func() float64 {
		if limit-buf.Pos() < 4 {
			return 0
		}
		return float64(buf.ReadFloat32())
	}
	readInt := func() int {
		if limit-buf.Pos() < 4 {
			return 0
		}
		return int(buf.ReadInt32())
	}
	readUint32 := func() uint32 {
		if limit-buf.Pos() < 4 {
			return 0
		}
		return buf.ReadUint32()
	}
	readString := func() string {
		if limit-buf.Pos() < 2 {
			return ""
		}
		if s := buf.ReadS(); s != nil {
			return *s
		}
		return ""
	}

	switch action {
	case core.TransitionActionXY, core.TransitionActionSize, core.TransitionActionPivot, core.TransitionActionSkew:
		out.B1 = readBool()
		out.B2 = readBool()
		out.F1 = readFloat()
		out.F2 = readFloat()
		if buf.Version >= 2 && action == core.TransitionActionXY && limit-buf.Pos() >= 1 {
			out.B3 = buf.ReadBool()
		}
	case core.TransitionActionAlpha, core.TransitionActionRotation:
		out.F1 = readFloat()
	case core.TransitionActionScale:
		out.F1 = readFloat()
		out.F2 = readFloat()
	case core.TransitionActionColor:
		out.Color = readUint32()
	case core.TransitionActionAnimation:
		out.Playing = readBool()
		out.Frame = readInt()
	case core.TransitionActionVisible:
		out.Visible = readBool()
	case core.TransitionActionSound:
		out.Sound = readString()
		out.Volume = readFloat()
	case core.TransitionActionTransition:
		out.TransName = readString()
		out.PlayTimes = readInt()
	case core.TransitionActionShake:
		out.Amplitude = readFloat()
		out.Duration = readFloat()
	case core.TransitionActionColorFilter:
		out.F1 = readFloat()
		out.F2 = readFloat()
		out.F3 = readFloat()
		out.F4 = readFloat()
	case core.TransitionActionText, core.TransitionActionIcon:
		out.Text = readString()
	}
}

func extractComponent(obj *core.GObject) *core.GComponent {
	if obj == nil {
		return nil
	}
	switch data := obj.Data().(type) {
	case *core.GComponent:
		return data
	case *widgets.GButton:
		return data.GComponent
	case *widgets.GLabel:
		return data.GComponent
	case *widgets.GList:
		return data.GComponent
	default:
		return nil
	}
}

type componentControllerResolver struct {
	controllers []gears.Controller
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

func findChildByPath(comp *core.GComponent, path string) *core.GObject {
	if comp == nil || path == "" {
		return nil
	}
	segments := strings.Split(path, ".")
	current := comp
	var obj *core.GObject
	for idx, segment := range segments {
		if segment == "" {
			continue
		}
		obj = current.ChildByName(segment)
		if obj == nil {
			return nil
		}
		if idx == len(segments)-1 {
			return obj
		}
		nested := extractComponent(obj)
		if nested == nil {
			return nil
		}
		current = nested
	}
	return obj
}
