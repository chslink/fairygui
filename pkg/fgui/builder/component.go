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

	// 性能优化：包状态缓存
	loadedPackages  map[string]bool // 已加载Atlas的包
	registeredFonts map[string]bool // 已注册字体的包
}

// FactoryObjectCreator 将Factory包装为ObjectCreator，用于GList虚拟列表
type FactoryObjectCreator struct {
	factory *Factory
	pkg     *assets.Package
	ctx     context.Context
}

// CreateObject 实现ObjectCreator接口
func (c *FactoryObjectCreator) CreateObject(url string) *core.GObject {
	if c.factory == nil {
		return nil
	}

	var item *assets.PackageItem

	// 首先尝试全局URL查找（支持ui://packageId/itemId格式）
	if strings.HasPrefix(url, "ui://") {
		item = assets.GetItemByURL(url)
	}

	// 如果没找到，尝试在当前包中查找
	if item == nil && c.pkg != nil {
		// 尝试通过名称查找
		item = c.pkg.ItemByName(url)
		if item == nil {
			// 尝试通过ID查找
			item = c.pkg.ItemByID(url)
		}
	}

	if item == nil {
		return nil
	}

	// 确定使用哪个包
	pkg := c.pkg
	if item.Owner != nil {
		pkg = item.Owner
	}

	// 使用Factory构建组件
	comp, err := c.factory.BuildComponent(c.ctx, pkg, item)
	if err != nil {
		return nil
	}

	return comp.GObject
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
		loadedPackages:  make(map[string]bool),
		registeredFonts: make(map[string]bool),
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

	// 同时注册到全局注册表,以支持 GetItemByURL
	assets.RegisterPackage(pkg)
}

// ensurePackageReady 确保包已准备好（Atlas已加载、已注册、字体已注册）
// 使用缓存避免重复操作，提升构建性能
func (f *Factory) ensurePackageReady(ctx context.Context, pkg *assets.Package) error {
	if pkg == nil {
		return nil
	}

	// 生成包键（优先使用ID，其次使用Name）
	pkgKey := pkg.ID
	if pkgKey == "" {
		pkgKey = pkg.Name
	}
	if pkgKey == "" {
		return nil // 匿名包，跳过缓存
	}

	// 只加载一次Atlas（避免重复加载纹理）
	if f.atlasManager != nil && !f.loadedPackages[pkgKey] {
		if err := f.atlasManager.LoadPackage(ctx, pkg); err != nil {
			return err
		}
		f.loadedPackages[pkgKey] = true
	}

	// 注册包（内部会检查重复）
	f.RegisterPackage(pkg)

	// 只注册一次字体（避免重复注册）
	if !f.registeredFonts[pkgKey] {
		render.RegisterBitmapFonts(pkg)
		f.registeredFonts[pkgKey] = true
	}

	return nil
}

// extractGComponent 从 widget 中提取 GComponent
func extractGComponent(widget interface{}) *core.GComponent {
	if widget == nil {
		return nil
	}
	switch w := widget.(type) {
	case *core.GComponent:
		return w
	case *widgets.GScrollBar:
		return w.GComponent
	case *widgets.GButton:
		return w.GComponent
	case *widgets.GLabel:
		return w.GComponent
	case *widgets.GList:
		return w.GComponent
	case *widgets.GProgressBar:
		return w.GComponent
	case *widgets.GSlider:
		return w.GComponent
	case *widgets.GComboBox:
		return w.GComponent
	case *widgets.GTree:
		return w.GComponent
	default:
		return nil
	}
}

// BuildComponent instantiates a component hierarchy for the given package item.
func (f *Factory) BuildComponent(ctx context.Context, pkg *assets.Package, item *assets.PackageItem) (*core.GComponent, error) {
	if item == nil || item.Type != assets.PackageItemTypeComponent {
		return nil, fmt.Errorf("builder: package item must be a component")
	}
	if item.Component == nil {
		return nil, fmt.Errorf("builder: component data missing for %s", item.Name)
	}

	// 优化：使用缓存的包准备方法，避免重复操作
	if err := f.ensurePackageReady(ctx, pkg); err != nil {
		return nil, err
	}

	// 根据 ObjectType 创建对应的 widget
	// 对于特殊类型（如 ScrollBar, Button 等），需要创建对应的 widget 实例
	var root *core.GComponent
	widget := widgets.CreateWidgetFromPackage(item)
	if widget != nil {
		// 从 widget 中提取 GComponent
		root = extractGComponent(widget)
		if root != nil {
			// 设置 Data 为 widget 实例
			root.GObject.SetData(widget)
		} else {
			// 降级到普通组件
			root = core.NewGComponent()
		}
	} else {
		root = core.NewGComponent()
	}
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
		// 注意：PageIDs 和 PageNames 是独立的两个数组
		// PageIDs 用于 gear 等系统内部匹配（通常是数字字符串 "0","1","2"...）
		// PageNames 用于显示名称（如 "up","down","over"...）
		// 不能把 pageNames 替换 pageIDs！
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

	// 如果根组件是 GButton，设置 button controller 和其他属性
	// 注意：不能调用 applyButtonTemplate，因为它会尝试构建模板组件导致无限递归
	// 这里手动读取按钮属性（mode, sound等）
	if btnWidget, ok := root.GObject.Data().(*widgets.GButton); ok && btnWidget != nil {
		// 设置 button controller（如果还没有设置）
		if btnWidget.ButtonController() == nil {
			for _, ctrl := range btnWidget.Controllers() {
				if strings.EqualFold(ctrl.Name, "button") {
					btnWidget.SetButtonController(ctrl)
					break
				}
			}
		}

		// 读取按钮扩展属性（对应 applyButtonTemplate 中 section 6 的内容）
		if buf := item.RawData; buf != nil {
			saved := buf.Pos()
			defer buf.SetPos(saved)
			if buf.Seek(0, 6) {
				mode := widgets.ButtonMode(buf.ReadByte())
				btnWidget.SetMode(mode)
				if sound := buf.ReadS(); sound != nil && *sound != "" {
					btnWidget.SetSound(*sound)
				}
				btnWidget.SetSoundVolumeScale(float64(buf.ReadFloat32()))
				btnWidget.SetDownEffect(int(buf.ReadByte()))
				btnWidget.SetDownEffectValue(float64(buf.ReadFloat32()))
				if btnWidget.DownEffect() == 2 {
					btnWidget.SetPivotWithAnchor(0.5, 0.5, btnWidget.PivotAsAnchor())
				}
			}
		}
	}

	// 如果根组件是 GScrollBar，解析子组件并设置属性
	// 注意：GScrollBar 的子组件（grip, bar, arrow1, arrow2）已经在上面的 for 循环中构建到 root 中了
	// 我们需要让 scrollBarWidget 识别这些子组件，但不能调用 SetTemplateComponent(root)
	// 因为那会导致 root 添加自己为子对象（循环引用）
	if scrollBarWidget, ok := root.GObject.Data().(*widgets.GScrollBar); ok && scrollBarWidget != nil {
		// resolveTemplate 现在支持从 GComponent 本身查找子组件（当 template 为 nil 时）
		// 所以我们不需要设置 template，直接调用 ResolveChildren 即可
		if scrollBarWidget.TemplateComponent() == nil {
			scrollBarWidget.ResolveChildren()
		}

		// 读取 ScrollBar 扩展属性（对应 applyScrollBarTemplate 中 section 6 的内容）
		if buf := item.RawData; buf != nil {
			saved := buf.Pos()
			defer buf.SetPos(saved)
			if buf.Seek(0, 6) && buf.Remaining() > 0 {
				scrollBarWidget.SetFixedGrip(buf.ReadBool())
			}
		}
	}

	// 注意：根组件不能调用 SetupBeforeAdd/SetupAfterAdd，因为：
	// 1. 根组件的尺寸、pivot 等已经从 ComponentData 设置
	// 2. 根组件的 alpha、rotation 应该保持默认值（1.0, 0）
	// 3. SetupBeforeAdd 会错误地从 RawData Section 0 读取 ComponentData 元数据，而不是 GObject 基础属性
	//
	// 但我们需要手动设置 mask、hitTest 和 transitions
	// 参考 GComponent.SetupBeforeAdd 的实现（gcomponent.go:616-659）
	if buf := item.RawData; buf != nil {
		saved := buf.Pos()
		defer buf.SetPos(saved)

		// 读取 mask 和 hitTest（Section 4）
		if buf.Seek(0, 4) {
			if err := buf.Skip(2); err == nil && buf.Remaining() >= 1+2 {
				root.SetOpaque(buf.ReadBool())
				maskIndex := int(buf.ReadInt16())
				reversed := false
				var maskObj *core.GObject
				if maskIndex >= 0 {
					if buf.Remaining() > 0 {
						reversed = buf.ReadBool()
					}
					// 从根组件的子对象中查找 mask
					maskObj = root.ChildAt(maskIndex)
				}
				root.SetMask(maskObj, reversed)

				// 读取 hitTest
				if buf.Remaining() >= 4+4 {
					hitID := buf.ReadS()
					offsetX := int(buf.ReadInt32())
					offsetY := int(buf.ReadInt32())
					hitMode := core.HitTest{Mode: core.HitTestModeNone}
					if hitID != nil && *hitID != "" {
						hitMode = core.HitTest{Mode: core.HitTestModePixel, ItemID: *hitID, OffsetX: offsetX, OffsetY: offsetY}
					} else if offsetX != 0 && offsetY != -1 {
						hitMode = core.HitTest{Mode: core.HitTestModeChild, OffsetX: offsetX, ChildIndex: offsetY}
					}
					root.SetHitTest(hitMode)
				}
			}
		}
	}

	// 设置 transitions（Section 5）
	root.SetupTransitions(item.RawData, 0)

	// 设置 margin 和 overflow（已经由 parseComponentData 解析）
	// 参考 TypeScript 版本：GComponent.ts setup (1039-1054行)
	if item.Component != nil {
		// 设置 margin
		root.SetMargin(core.Margin{
			Top:    item.Component.Margin.Top,
			Bottom: item.Component.Margin.Bottom,
			Left:   item.Component.Margin.Left,
			Right:  item.Component.Margin.Right,
		})

		// 设置 overflow
		overflow := item.Component.Overflow
		if overflow == core.OverflowScroll {
			// 切换到 section 7 读取 scroll 配置
			if buf := item.RawData; buf != nil {
				saved := buf.Pos()
				if buf.Seek(0, 7) {
					root.SetupScroll(buf)
				}
				buf.SetPos(saved)
			}
		} else {
			root.SetupOverflow(overflow)
		}
	}

	f.finalizeComponentSize(root)

	// 创建并绑定滚动条（如果有）
	if pane := root.ScrollPane(); pane != nil {
		f.setupScrollBars(ctx, pkg, root, pane)
	}

	// 标记边界需要计算并立即计算（对应 TypeScript GComponent.ts:1204）
	// TypeScript 使用 callLater 延迟调用，Go 版本同步执行
	root.SetBoundsChangedFlag()
	root.EnsureBoundsCorrect()

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

	// 通用的 SetupBeforeAdd/AfterAdd 调用函数
	callSetupBeforeAdd := func(widget interface{}) {
		if sub != nil {
			if before, ok := widget.(widgets.BeforeAdder); ok {
				before.SetupBeforeAdd(sub, 0)
			}
		}
	}
	callSetupAfterAdd := func(widget interface{}) {
		if sub != nil {
			if after, ok := widget.(widgets.AfterAdder); ok {
				after.SetupAfterAdd(ensureCtx(), sub)
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
		callSetupBeforeAdd(widget)

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
		callSetupBeforeAdd(widget)

	case *widgets.GRichTextField:
		obj = widget.GObject()
		obj.SetData(widget)
		callSetupBeforeAdd(widget)

	case *widgets.GTextField:
		obj = widget.GObject
		obj.SetData(widget)
		callSetupBeforeAdd(widget)

	case *widgets.GTextInput:
		obj = widget.GObject
		obj.SetData(widget)
		callSetupBeforeAdd(widget)

	case *widgets.GButton:
		obj = widget.GComponent.GObject
		obj.SetData(widget)
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
			f.applyButtonTemplate(ctx, widget, pkg, owner, resolvedItem)
		}
		callSetupBeforeAdd(widget)
		callSetupAfterAdd(widget)

	case *widgets.GLoader:
		obj = widget.GObject
		obj.SetData(widget)
		callSetupBeforeAdd(widget)
		if resolvedItem == nil {
			resolvedItem = f.resolveIcon(ctx, pkg, widget.URL())
		}
		if setupCtx != nil {
			setupCtx.ResolvedItem = resolvedItem
		}
		f.assignLoaderPackage(ctx, widget, pkg, resolvedItem)

	case *widgets.GLabel:
		obj = widget.GComponent.GObject
		obj.SetData(widget)
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
		}
		f.applyLabelTemplate(ctx, widget, pkg, owner, resolvedItem)
		callSetupAfterAdd(widget)

	case *widgets.GList:
		obj = widget.GComponent.GObject
		obj.SetData(widget)
		widget.SetPackageItem(resolvedItem)
		// 为虚拟列表设置对象创建器
		widget.SetObjectCreator(&FactoryObjectCreator{
			factory: f,
			pkg:     pkg,
			ctx:     ctx,
		})
		callSetupBeforeAdd(widget)
		callSetupAfterAdd(widget)
		// SetupBeforeAdd会创建ScrollPane，现在创建滚动条
		if pane := widget.GComponent.ScrollPane(); pane != nil {
			f.setupScrollBars(ctx, pkg, widget.GComponent, pane)
		}

	case *widgets.GProgressBar:
		obj = widget.GComponent.GObject
		obj.SetData(widget)
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
		}
		f.applyProgressBarTemplate(ctx, widget, pkg, owner, resolvedItem)
		callSetupAfterAdd(widget)

	case *widgets.GSlider:
		obj = widget.GComponent.GObject
		obj.SetData(widget)
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
		}
		f.applySliderTemplate(ctx, widget, pkg, owner, resolvedItem)
		callSetupAfterAdd(widget)

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
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
		}
		callSetupBeforeAdd(widget)
		if sub != nil {
			f.populateTree(ctx, widget, pkg, owner, child, sub)
		}
		callSetupAfterAdd(widget)

	case *widgets.GComboBox:
		obj = widget.GComponent.GObject
		obj.SetData(widget)
		if resolvedItem != nil {
			widget.SetPackageItem(resolvedItem)
		}
		f.applyComboBoxTemplate(ctx, widget, pkg, owner, resolvedItem)
		callSetupAfterAdd(widget)

	case *widgets.GGroup:
		obj = widget.GObject
		callSetupBeforeAdd(widget)
		callSetupAfterAdd(widget)

	case *widgets.GGraph:
		obj = widget.GObject
		obj.SetData(widget)
		callSetupBeforeAdd(widget)

	default:
		obj = core.NewGObject()
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
				// 关键修复：对于嵌套组件，直接使用 nested 的 GObject
				// 而不是创建新的 GObject 并把 nested 存储在 Data 中
				// 这确保 DisplayObject 层级正确连接
				obj = nested.GObject
				if nestedItem != nil {
					resolvedItem = nestedItem
				}
				// 注意：nested 已经有正确的尺寸，但如果 child 指定了尺寸覆盖，
				// 后面的代码会处理（第573-590行）
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

	// 设置基础属性（对应 TypeScript 版本 constructFromResource2）
	// 参考 GComponent.ts:1024: this.setSize(this.sourceWidth, this.sourceHeight);
	obj.SetSize(initW, initH)
	if child.Name != "" {
		obj.SetName(child.Name)
	}
	obj.SetPosition(float64(child.X), float64(child.Y))

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
				// 优化：使用ensurePackageReady避免重复操作
				if err := f.ensurePackageReady(ctx, resolved); err != nil {
					fmt.Printf("builder: load dependent package failed: %v\n", err)
					continue
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
				// 优化：使用ensurePackageReady避免重复操作
				if err := f.ensurePackageReady(ctx, resolved); err != nil {
					fmt.Printf("builder: load icon package failed: %v\n", err)
				} else {
					pkg = resolved
				}
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
	// 修复：查找并设置 button controller
	// 参考 TypeScript 版本：GButton.constructExtension() 中使用 this.getController("button")
	// 注意：controllers 已经在 BuildComponent 中添加到组件了，这里只需要查找并设置
	for _, ctrl := range widget.Controllers() {
		if strings.EqualFold(ctrl.Name, "button") {
			widget.SetButtonController(ctrl)
			break
		}
	}

	// 然后处理 template 相关的设置
	if template := widget.TemplateComponent(); template != nil {
		titleObj := template.ChildByName("title")
		if titleObj != nil {
			widget.SetTitleObject(titleObj)
		}
		iconObj := template.ChildByName("icon")
		if iconObj != nil {
			widget.SetIconObject(iconObj)
		}
		// 只有在按钮自己没有 button controller 时，才从 template 查找
		if widget.ButtonController() == nil {
			if ctrl := template.ControllerByName("button"); ctrl != nil {
				widget.SetButtonController(ctrl)
			} else if ctrl := template.ControllerByName("Button"); ctrl != nil {
				widget.SetButtonController(ctrl)
			}
		}

		// 修复：如果按钮尺寸为0，从模板继承尺寸
		// 对应 TypeScript 版本中模板组件的尺寸会自动影响按钮尺寸
		// 参考 Button10.xml: <component size="163,69" extention="Button">
		if widget.GComponent.GObject.Width() == 0 || widget.GComponent.GObject.Height() == 0 {
			widget.GComponent.GObject.SetSize(template.GObject.Width(), template.GObject.Height())
		}
	}

	// 如果还是没有 button controller，使用第一个 controller
	if widget.ButtonController() == nil {
		for _, ctrl := range widget.Controllers() {
			widget.SetButtonController(ctrl)
			break
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

	// 所有 gear 设置完成后，计算 GearDisplay 和 GearDisplay2 的组合可见性
	// 参考 TypeScript 版本 GComponent.ts constructFromResource2 (1039行)
	obj.CheckGearDisplay()
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

// setupScrollBars 创建并绑定滚动条到ScrollPane
// 对应 TypeScript 版本 ScrollPane.ts:149-178
func (f *Factory) setupScrollBars(ctx context.Context, pkg *assets.Package, owner *core.GComponent, pane *core.ScrollPane) {
	if pane == nil || owner == nil {
		return
	}

	vtURL := pane.VtScrollBarURL()
	hzURL := pane.HzScrollBarURL()

	// 如果滚动条URL为空，使用全局配置的默认值（对应 TypeScript 版本 ScrollPane.ts:150,160）
	// TypeScript: var res: string = vtScrollBarRes ? vtScrollBarRes : UIConfig.verticalScrollBar;
	if vtURL == "" {
		vtURL = core.GetUIConfig().VerticalScrollBar
	}
	if hzURL == "" {
		hzURL = core.GetUIConfig().HorizontalScrollBar
	}

	// 创建垂直滚动条
	if vtURL != "" {
		if vtItem := f.resolveIcon(ctx, pkg, vtURL); vtItem != nil {
			targetPkg := vtItem.Owner
			if targetPkg == nil {
				targetPkg = pkg
			}
			if vtComp, err := f.BuildComponent(ctx, targetPkg, vtItem); err == nil && vtComp != nil {
				vtComp.GObject.SetName("__VT_SCROLLBAR__")
				pane.SetVtScrollBar(vtComp.GObject)

				// 关键修复：滚动条必须添加到 owner.displayObject（根 DisplayObject），而不是通过 AddChild
				// AddChild 会把子对象添加到 childContainer()，而 childContainer() 可能是被裁剪的容器
				// 参考 TypeScript ScrollPane.ts:156 - this._owner.displayObject.addChild(this._vtScrollBar.displayObject);
				if owner.GObject.DisplayObject() != nil && vtComp.GObject.DisplayObject() != nil {
					scrollBarDisplay := vtComp.GObject.DisplayObject()
					scrollBarDisplay.SetOwner(vtComp.GObject)
					owner.GObject.DisplayObject().AddChild(scrollBarDisplay)
				}

				// 调用GScrollBar的SetScrollPane绑定
				if scrollBar, ok := vtComp.GObject.Data().(*widgets.GScrollBar); ok {
					scrollBar.SetScrollPane(pane, true) // true = vertical
				}
			}
		}
	}

	// 创建水平滚动条
	if hzURL != "" {
		if hzItem := f.resolveIcon(ctx, pkg, hzURL); hzItem != nil {
			targetPkg := hzItem.Owner
			if targetPkg == nil {
				targetPkg = pkg
			}
			if hzComp, err := f.BuildComponent(ctx, targetPkg, hzItem); err == nil && hzComp != nil {
				hzComp.GObject.SetName("__HZ_SCROLLBAR__")
				pane.SetHzScrollBar(hzComp.GObject)

				// 关键修复：滚动条必须添加到 owner.displayObject（根 DisplayObject）
				// 参考 TypeScript ScrollPane.ts:166 - this._owner.displayObject.addChild(this._hzScrollBar.displayObject);
				if owner.GObject.DisplayObject() != nil && hzComp.GObject.DisplayObject() != nil {
					scrollBarDisplay := hzComp.GObject.DisplayObject()
					scrollBarDisplay.SetOwner(hzComp.GObject)
					owner.GObject.DisplayObject().AddChild(scrollBarDisplay)
				}

				// 调用GScrollBar的SetScrollPane绑定
				if scrollBar, ok := hzComp.GObject.Data().(*widgets.GScrollBar); ok {
					scrollBar.SetScrollPane(pane, false) // false = horizontal
				}
			}
		}
	}

	// 滚动条创建后，重新计算 ViewSize（包含 margin 和滚动条尺寸）
	// 对应 TypeScript 版本 ScrollPane.ts:197 (setup方法最后调用 this.setSize)
	pane.OnOwnerSizeChanged()
}
