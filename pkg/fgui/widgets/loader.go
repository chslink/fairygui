package widgets

import (
	"math"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// GLoader represents a resource loader widget.
type GLoader struct {
	*core.GObject
	packageItem *assets.PackageItem
	component   *core.GComponent
	movieClip   *GMovieClip // Internal MovieClip for animated content
	url         string
	autoSize    bool
	useResize   bool
	fill        LoaderFillType
	align       LoaderAlign
	vertical    LoaderAlign
	shrinkOnly  bool
	updating    bool

	contentOffsetX float64
	contentOffsetY float64
	contentScaleX  float64
	contentScaleY  float64
	contentWidth   float64
	contentHeight  float64
	playing        bool
	frame          int
	color          string
	fillMethod     LoaderFillMethod
	fillOrigin     int
	fillClockwise  bool
	fillAmount     float64
	scale9Grid     *assets.Rect
	scaleByTile    bool
	tileGridIndice int
}

// NewLoader creates a loader widget.
func NewLoader() *GLoader {
	loader := &GLoader{
		GObject:       core.NewGObject(),
		fill:          LoaderFillNone,
		align:         LoaderAlignLeft,
		vertical:      LoaderAlignTop,
		contentScaleX: 1,
		contentScaleY: 1,
		playing:       true,
		color:         "#ffffff",
		fillMethod:    LoaderFillMethodNone,
	}
	// 修复：加载器是显示性组件，不应该拦截鼠标事件，设置mouseThrough=true让事件穿透到父组件
	if sprite := loader.GObject.DisplayObject(); sprite != nil {
		sprite.SetMouseThrough(true)
	}
	return loader
}

// LoaderFillType describes how the loader fits content to its bounds.
type LoaderFillType int

const (
	LoaderFillNone LoaderFillType = iota
	LoaderFillScale
	LoaderFillScaleMatchHeight
	LoaderFillScaleMatchWidth
	LoaderFillScaleFree
	LoaderFillScaleNoBorder
)

// LoaderAlign enumerates horizontal and vertical alignment modes.
type LoaderAlign string

const (
	LoaderAlignLeft   LoaderAlign = "left"
	LoaderAlignCenter LoaderAlign = "center"
	LoaderAlignRight  LoaderAlign = "right"

	LoaderAlignTop    LoaderAlign = "top"
	LoaderAlignMiddle LoaderAlign = "middle"
	LoaderAlignBottom LoaderAlign = "bottom"
)

// LoaderFillMethod describes the internal masking strategy.
type LoaderFillMethod int

const (
	LoaderFillMethodNone LoaderFillMethod = iota
	LoaderFillMethodHorizontal
	LoaderFillMethodVertical
	LoaderFillMethodRadial90
	LoaderFillMethodRadial180
	LoaderFillMethodRadial360
)

// SetPackageItem assigns the loader's package item source.
func (l *GLoader) SetPackageItem(item *assets.PackageItem) {
	// 清理旧的 MovieClip（如果存在）
	if l.movieClip != nil {
		// 停止播放以清理 ticker
		l.movieClip.SetPlaying(false)
		l.movieClip = nil
	}

	l.packageItem = item
	if item != nil {
		l.url = item.ID

		// 为 MovieClip 类型创建内部 MovieClip 实例
		if item.Type == assets.PackageItemTypeMovieClip {
			l.movieClip = NewMovieClip()
			l.movieClip.SetPackageItem(item)
			l.movieClip.SetPlaying(l.playing)
			l.movieClip.SetFrame(l.frame)
		}
	}
	l.updateAutoSize()
	l.updateLayout() // 确保 updateGraphics 被调用
}

// PackageItem returns the resolved package item.
func (l *GLoader) PackageItem() *assets.PackageItem {
	return l.packageItem
}

// SetComponent sets the component rendered by this loader.
func (l *GLoader) SetComponent(comp *core.GComponent) {
	if l.component == comp {
		return
	}
	if l.component != nil && l.component.DisplayObject() != nil && l.DisplayObject() != nil {
		l.DisplayObject().RemoveChild(l.component.DisplayObject())
	}
	l.component = comp
	if comp != nil && comp.DisplayObject() != nil && l.DisplayObject() != nil {
		comp.DisplayObject().SetPosition(0, 0)
		l.DisplayObject().AddChild(comp.DisplayObject())
	}
	l.updateAutoSize()
	l.updateLayout()
}

// Component returns the component rendered by this loader, if any.
func (l *GLoader) Component() *core.GComponent {
	return l.component
}

// MovieClip returns the internal MovieClip instance, if any.
func (l *GLoader) MovieClip() *GMovieClip {
	return l.movieClip
}

// SetURL stores the loader url (ui:// or external) and loads the content.
// External URLs are not yet handled.
func (l *GLoader) SetURL(url string) {
	if l.url == url {
		return
	}
	l.url = url
	l.loadContent()
}

// URL returns the current loader URL.
func (l *GLoader) URL() string {
	return l.url
}

// SetPlaying toggles playback for content with frames.
func (l *GLoader) SetPlaying(playing bool) {
	l.playing = playing
	// 同步到内部 MovieClip
	if l.movieClip != nil {
		l.movieClip.SetPlaying(playing)
	}
}

// Playing reports whether playback is active.
func (l *GLoader) Playing() bool {
	return l.playing
}

// SetFrame configures the frame for frame-based content.
func (l *GLoader) SetFrame(frame int) {
	l.frame = frame
	// 同步到内部 MovieClip
	if l.movieClip != nil {
		l.movieClip.SetFrame(frame)
	}
}

// Frame returns the current frame index.
func (l *GLoader) Frame() int {
	return l.frame
}

// SetColor stores the tint colour string.
func (l *GLoader) SetColor(value string) {
	if value == "" {
		value = "#ffffff"
	}
	if l.color == value {
		return
	}
	l.color = value
	l.updateGraphics()
}

// Color returns the tint colour string.
func (l *GLoader) Color() string {
	return l.color
}

// SetAutoSize toggles whether the loader should adopt its source dimensions when unset.
func (l *GLoader) SetAutoSize(enabled bool) {
	if l.autoSize == enabled {
		return
	}
	l.autoSize = enabled
	l.updateAutoSize()
	l.updateLayout()
}

// AutoSize reports whether the loader will resize to its source.
func (l *GLoader) AutoSize() bool {
	return l.autoSize
}

// SetUseResize toggles whether nested components resize instead of scale.
func (l *GLoader) SetUseResize(enabled bool) {
	if l.useResize == enabled {
		return
	}
	l.useResize = enabled
	l.updateLayout()
}

// UseResize reports whether resize semantics are enabled for nested components.
func (l *GLoader) UseResize() bool {
	return l.useResize
}

// SetFill configures how source content fits the loader bounds.
func (l *GLoader) SetFill(fill LoaderFillType) {
	if l.fill == fill {
		return
	}
	l.fill = fill
	l.updateLayout()
}

// Fill returns the current fill rule.
func (l *GLoader) Fill() LoaderFillType {
	return l.fill
}

// SetAlign configures horizontal alignment for the rendered content.
func (l *GLoader) SetAlign(align LoaderAlign) {
	if l.align == align {
		return
	}
	l.align = align
	l.updateLayout()
}

// Align returns the current horizontal alignment.
func (l *GLoader) Align() LoaderAlign {
	return l.align
}

// SetVerticalAlign configures vertical alignment for the rendered content.
func (l *GLoader) SetVerticalAlign(align LoaderAlign) {
	if l.vertical == align {
		return
	}
	l.vertical = align
	l.updateLayout()
}

// VerticalAlign returns the vertical alignment mode.
func (l *GLoader) VerticalAlign() LoaderAlign {
	return l.vertical
}

// SetShrinkOnly prevents scaling up when fill is active.
func (l *GLoader) SetShrinkOnly(enabled bool) {
	if l.shrinkOnly == enabled {
		return
	}
	l.shrinkOnly = enabled
	l.updateLayout()
}

// ShrinkOnly reports whether the loader avoids enlarging content.
func (l *GLoader) ShrinkOnly() bool {
	return l.shrinkOnly
}

// ContentOffset returns the computed rendering offset.
func (l *GLoader) ContentOffset() (float64, float64) {
	return l.contentOffsetX, l.contentOffsetY
}

// ContentScale returns the computed content scale factors.
func (l *GLoader) ContentScale() (float64, float64) {
	return l.contentScaleX, l.contentScaleY
}

// RefreshLayout recomputes layout based on current state.
func (l *GLoader) RefreshLayout() {
	l.updateLayout()
}

// SetFillMethod configures the image fill method (0 = none).
func (l *GLoader) SetFillMethod(method int) {
	l.fillMethod = LoaderFillMethod(method)
}

// FillMethod returns the image fill method.
func (l *GLoader) FillMethod() int {
	return int(l.fillMethod)
}

// SetFillOrigin stores the origin for radial fill.
func (l *GLoader) SetFillOrigin(origin int) {
	l.fillOrigin = origin
}

// FillOrigin returns the fill origin.
func (l *GLoader) FillOrigin() int {
	return l.fillOrigin
}

// SetFillClockwise indicates whether the radial fill runs clockwise.
func (l *GLoader) SetFillClockwise(clockwise bool) {
	l.fillClockwise = clockwise
}

// FillClockwise reports the radial fill direction.
func (l *GLoader) FillClockwise() bool {
	return l.fillClockwise
}

// SetFillAmount stores the fill amount (0..1).
func (l *GLoader) SetFillAmount(amount float64) {
	if amount < 0 {
		amount = 0
	} else if amount > 1 {
		amount = 1
	}
	l.fillAmount = amount
}

// FillAmount returns the radial fill amount.
func (l *GLoader) FillAmount() float64 {
	return l.fillAmount
}

// ContentSize returns the current content width and height after layout.
func (l *GLoader) ContentSize() (float64, float64) {
	return l.contentWidth, l.contentHeight
}

// SetupBeforeAdd reads loader configuration from the component buffer.
func (l *GLoader) SetupBeforeAdd(buf *utils.ByteBuffer, beginPos int) {
	if l == nil || buf == nil {
		return
	}

	// 首先调用父类GObject处理基础属性
	l.GObject.SetupBeforeAdd(buf, beginPos)

	// 然后处理GLoader特定属性（block 5）
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(beginPos, 5) {
		return
	}
	if url := buf.ReadS(); url != nil && *url != "" {
		l.SetURL(*url)
	}
	mapAlign := func(code int8, horizontal bool) LoaderAlign {
		switch code {
		case 1:
			if horizontal {
				return LoaderAlignCenter
			}
			return LoaderAlignMiddle
		case 2:
			if horizontal {
				return LoaderAlignRight
			}
			return LoaderAlignBottom
		default:
			if horizontal {
				return LoaderAlignLeft
			}
			return LoaderAlignTop
		}
	}
	l.SetAlign(mapAlign(buf.ReadByte(), true))
	l.SetVerticalAlign(mapAlign(buf.ReadByte(), false))
	l.SetFill(LoaderFillType(buf.ReadByte()))
	l.SetShrinkOnly(buf.ReadBool())
	l.SetAutoSize(buf.ReadBool())
	_ = buf.ReadBool() // showErrorSign flag not yet wired
	l.SetPlaying(buf.ReadBool())
	l.SetFrame(int(buf.ReadInt32()))
	if buf.ReadBool() {
		l.SetColor(buf.ReadColorString(true))
	}
	fillMethod := LoaderFillMethod(buf.ReadByte())
	l.SetFillMethod(int(fillMethod))
	if fillMethod != LoaderFillMethodNone {
		l.SetFillOrigin(int(buf.ReadByte()))
		l.SetFillClockwise(buf.ReadBool())
		l.SetFillAmount(float64(buf.ReadFloat32()))
	}
	if buf.Version >= 7 {
		l.SetUseResize(buf.ReadBool())
	}
	l.RefreshLayout()
}

// SetScale9Grid applies nine-slice data to the loader.
func (l *GLoader) SetScale9Grid(grid *assets.Rect) {
	if grid == nil {
		l.scale9Grid = nil
		return
	}
	rect := *grid
	l.scale9Grid = &rect
}

// Scale9Grid returns the active nine-slice rectangle, if any.
func (l *GLoader) Scale9Grid() *assets.Rect {
	if l.scale9Grid == nil {
		return nil
	}
	rect := *l.scale9Grid
	return &rect
}

// SetScaleByTile toggles grid tiling mode.
func (l *GLoader) SetScaleByTile(enabled bool) {
	l.scaleByTile = enabled
}

// ScaleByTile reports whether tile scaling is enabled.
func (l *GLoader) ScaleByTile() bool {
	return l.scaleByTile
}

// SetTileGridIndice stores tile grid indices for scale9 rendering.
func (l *GLoader) SetTileGridIndice(value int) {
	l.tileGridIndice = value
}

// TileGridIndice returns the tile grid index value.
func (l *GLoader) TileGridIndice() int {
	return l.tileGridIndice
}

// SourceSize returns the dimensions implied by the current content.
func (l *GLoader) SourceSize() (float64, float64) {
	if l.component != nil {
		return l.component.Width(), l.component.Height()
	}
	if l.packageItem != nil {
		if sprite := l.packageItem.Sprite; sprite != nil {
			w := float64(sprite.OriginalSize.X)
			h := float64(sprite.OriginalSize.Y)
			if w <= 0 {
				w = float64(sprite.Rect.Width)
			}
			if h <= 0 {
				h = float64(sprite.Rect.Height)
			}
			return w, h
		}
		if l.packageItem.Width > 0 || l.packageItem.Height > 0 {
			return float64(l.packageItem.Width), float64(l.packageItem.Height)
		}
	}
	return 0, 0
}

func (l *GLoader) updateAutoSize() {
	if !l.autoSize {
		return
	}
	sourceW, sourceH := l.SourceSize()
	width := sourceW
	height := sourceH

	if width <= 0 {
		if l.Width() > 0 {
			width = l.Width()
		} else {
			width = 50
		}
	}
	if height <= 0 {
		if l.Height() > 0 {
			height = l.Height()
		} else {
			height = 30
		}
	}

	width = math.Max(width, 0)
	height = math.Max(height, 0)

	l.updating = true
	l.SetSize(width, height)
	l.updating = false
}

func (l *GLoader) updateLayout() {
	if l.updating {
		return
	}
	sourceW, sourceH := l.SourceSize()
	width := l.Width()
	height := l.Height()

	if l.autoSize {
		targetW := sourceW
		targetH := sourceH
		if targetW == 0 {
			targetW = 50
		}
		if targetH == 0 {
			targetH = 30
		}
		l.updating = true
		l.SetSize(targetW, targetH)
		l.updating = false
		width = l.Width()
		height = l.Height()
	}

	if width <= 0 {
		width = 0
	}
	if height <= 0 {
		height = 0
	}

	sx := 1.0
	sy := 1.0
	cw := sourceW
	ch := sourceH

	if cw == 0 && width > 0 {
		cw = width
	}
	if ch == 0 && height > 0 {
		ch = height
	}

	if l.fill != LoaderFillNone && sourceW > 0 && sourceH > 0 && width > 0 && height > 0 {
		sx = width / sourceW
		sy = height / sourceH

		switch l.fill {
		case LoaderFillScaleMatchHeight:
			sx = sy
		case LoaderFillScaleMatchWidth:
			sy = sx
		case LoaderFillScale:
			if sx > sy {
				sx = sy
			} else {
				sy = sx
			}
		case LoaderFillScaleNoBorder:
			if sx > sy {
				sy = sx
			} else {
				sx = sy
			}
		case LoaderFillScaleFree:
			// intentionally left blank
		}

		if l.shrinkOnly {
			if sx > 1 {
				sx = 1
			}
			if sy > 1 {
				sy = 1
			}
		}

		cw = sourceW * sx
		ch = sourceH * sy
	} else {
		if l.useResize && l.component != nil {
			cw = width
			ch = height
			if sourceW > 0 {
				sx = width / sourceW
			}
			if sourceH > 0 {
				sy = height / sourceH
			}
		}
	}

	if cw <= 0 {
		cw = width
	}
	if ch <= 0 {
		ch = height
	}

	var nx, ny float64
	switch l.align {
	case LoaderAlignCenter:
		nx = (width - cw) / 2
	case LoaderAlignRight:
		nx = width - cw
	default:
		nx = 0
	}
	switch l.vertical {
	case LoaderAlignMiddle:
		ny = (height - ch) / 2
	case LoaderAlignBottom:
		ny = height - ch
	default:
		ny = 0
	}

	l.contentOffsetX = nx
	l.contentOffsetY = ny
	l.contentScaleX = sx
	l.contentScaleY = sy
	l.contentWidth = cw
	l.contentHeight = ch

	if l.component != nil {
		l.component.SetPosition(nx, ny)
		if l.useResize {
			l.component.SetScale(1, 1)
			l.component.SetSize(cw, ch)
		} else {
			l.component.SetScale(sx, sy)
		}
	}

	// 更新 Graphics 命令（仅用于简单图像和 Scale9）
	l.updateGraphics()
}

// updateGraphics 生成 Graphics 命令用于渲染
// 注意：仅处理简单图像和 Scale9 模式，MovieClip 和 FillMethod 仍使用旧渲染路径
func (l *GLoader) updateGraphics() {
	if l == nil || l.GObject == nil {
		return
	}

	sprite := l.GObject.DisplayObject()
	if sprite == nil {
		return
	}

	sprite.SetMouseEnabled(l.GObject.Touchable())

	// 获取或创建 Graphics
	gfx := sprite.Graphics()
	gfx.Clear()

	// Component、MovieClip 和 FillMethod 不使用命令模式
	if l.component != nil || l.movieClip != nil || l.fillMethod != LoaderFillMethodNone {
		sprite.Repaint()
		return
	}

	// 没有纹理，不生成命令
	if l.packageItem == nil {
		sprite.Repaint()
		return
	}

	// MovieClip 类型使用旧渲染路径
	if l.packageItem.Type == assets.PackageItemTypeMovieClip {
		sprite.Repaint()
		return
	}

	// 确定渲染模式
	mode := l.determineMode()

	// 获取内容尺寸和偏移
	dstW, dstH := l.ContentSize()
	offsetX, offsetY := l.ContentOffset()

	// 构建纹理命令
	cmd := laya.TextureCommand{
		Texture: l.packageItem,
		Mode:    mode,
		Dest: laya.Rect{
			W: dstW,
			H: dstH,
		},
		Color:          l.color,
		ScaleX:         1.0, // GLoader 的翻转通过 DisplayObject 处理
		ScaleY:         1.0,
		OffsetX:        offsetX,
		OffsetY:        offsetY,
		ScaleByTile:    l.scaleByTile,
		TileGridIndice: l.tileGridIndice,
	}

	// 设置 Scale9Grid（如果存在）
	if grid := l.Scale9Grid(); grid != nil {
		cmd.Scale9Grid = &laya.Rect{
			X: float64(grid.X),
			Y: float64(grid.Y),
			W: float64(grid.Width),
			H: float64(grid.Height),
		}
	} else if l.packageItem.Scale9Grid != nil {
		// 使用 PackageItem 的 Scale9Grid
		cmd.Scale9Grid = &laya.Rect{
			X: float64(l.packageItem.Scale9Grid.X),
			Y: float64(l.packageItem.Scale9Grid.Y),
			W: float64(l.packageItem.Scale9Grid.Width),
			H: float64(l.packageItem.Scale9Grid.Height),
		}
	}

	// 记录命令
	gfx.DrawTexture(cmd)
	sprite.Repaint()
}

// determineMode 根据配置确定渲染模式
func (l *GLoader) determineMode() laya.TextureCommandMode {
	if l.Scale9Grid() != nil || l.packageItem.Scale9Grid != nil {
		return laya.TextureModeScale9
	}
	if l.scaleByTile {
		return laya.TextureModeTile
	}
	return laya.TextureModeSimple
}

// loadContent 加载 URL 指定的内容
// 参考 TypeScript GLoader.ts:218-228
func (l *GLoader) loadContent() {
	if l == nil {
		return
	}

	// 清理旧内容
	l.clearContent()

	// 检查 URL
	if l.url == "" {
		return
	}

	// 根据 URL 类型加载
	if len(l.url) >= 5 && l.url[:5] == "ui://" {
		l.loadFromPackage(l.url)
	} else {
		// 外部 URL 暂不支持
		// l.loadExternal()
	}
}

// clearContent 清理加载器的当前内容
func (l *GLoader) clearContent() {
	if l == nil {
		return
	}

	// 清理 MovieClip
	if l.movieClip != nil {
		l.movieClip.SetPlaying(false)
		l.movieClip = nil
	}

	// 清理 Component
	if l.component != nil {
		if l.component.DisplayObject() != nil && l.DisplayObject() != nil {
			l.DisplayObject().RemoveChild(l.component.DisplayObject())
		}
		l.component = nil
	}

	// 清理 PackageItem
	l.packageItem = nil
}

// loadFromPackage 从 UIPackage 加载资源
// 参考 TypeScript GLoader.ts:230-276
func (l *GLoader) loadFromPackage(itemURL string) {
	if l == nil {
		return
	}

	// 通过 URL 获取 PackageItem
	item := assets.GetItemByURL(itemURL)
	if item == nil {
		return
	}

	// 设置 PackageItem
	l.SetPackageItem(item)

	// autoSize 会在 SetPackageItem → updateAutoSize 中处理
}
