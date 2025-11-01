package widgets

import (
	"math"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// GTextField is a minimal text widget.
type TextAlign string

const (
	TextAlignLeft   TextAlign = "left"
	TextAlignCenter TextAlign = "center"
	TextAlignRight  TextAlign = "right"
)

type TextVerticalAlign string

const (
	TextVerticalAlignTop    TextVerticalAlign = "top"
	TextVerticalAlignMiddle TextVerticalAlign = "middle"
	TextVerticalAlignBottom TextVerticalAlign = "bottom"
)

type TextAutoSize int

const (
	TextAutoSizeNone TextAutoSize = iota  // 0 - 对应 LayaAir 的 AutoSizeType.None
	TextAutoSizeBoth                      // 1 - 对应 LayaAir 的 AutoSizeType.Both
	TextAutoSizeHeight                    // 2 - 对应 LayaAir 的 AutoSizeType.Height
	TextAutoSizeShrink                    // 3 - 对应 LayaAir 的 AutoSizeType.Shrink
	TextAutoSizeEllipsis                  // 4 - 对应 LayaAir 的 AutoSizeType.Ellipsis
)

// GTextField is a minimal text widget.
type GTextField struct {
	*core.GObject
	text           string
	color          string
	outlineColor   string
	fontSize       int
	font           string
	align          TextAlign
	vertical       TextVerticalAlign
	autoSize       TextAutoSize
	singleLine     bool
	underline      bool
	italic         bool
	bold           bool
	letterSpace    int
	leading        int
	strokeSize     float64
	strokeColor    string
	ubbEnabled     bool
	templateVars   bool
	shadowColor    string
	shadowOffsetX  float64
	shadowOffsetY  float64
	shadowBlur     float64
	widthAutoSize  bool
	heightAutoSize bool
	layoutWidth    float64
	layoutHeight   float64
	textWidth      float64
	textHeight     float64
	linkRegions    []TextLinkRegion
	linkHandler    laya.Listener
	onLayoutUpdated func() // 参考 TS 版本的 _onPostLayout 回调
}

// TextLinkRegion describes a clickable link region within the text.
type TextLinkRegion struct {
	Target string
	Bounds laya.Rect
}

// NewText creates a new text field widget.
func NewText() *GTextField {
	field := &GTextField{GObject: core.NewGObject()}
	field.color = "#000000"
	field.fontSize = 12
	field.align = TextAlignLeft
	field.vertical = TextVerticalAlignTop
	field.autoSize = TextAutoSizeBoth
	field.widthAutoSize = true
	field.heightAutoSize = true
	field.GObject.SetData(field)
	// 修复：文本是显示性组件，不应该拦截鼠标事件，设置mouseThrough=true让事件穿透到父组件
	if sprite := field.GObject.DisplayObject(); sprite != nil {
		sprite.SetMouseThrough(true)
	}
	return field
}

// SetText updates the displayed text.
func (t *GTextField) SetText(value string) {
	if t.text == value {
		return
	}
	t.text = value
	if sprite := t.GObject.DisplayObject(); sprite != nil {
		sprite.Repaint()
	}
}

// Text returns the current text.
func (t *GTextField) Text() string {
	return t.text
}

// SetColor updates the text colour stored on this widget.
func (t *GTextField) SetColor(value string) {
	if t.color == value {
		return
	}
	t.color = value
	if sprite := t.GObject.DisplayObject(); sprite != nil {
		sprite.Repaint()
	}
}

// Color returns the current text colour.
func (t *GTextField) Color() string {
	return t.color
}

// SetFont stores the requested font face identifier.
func (t *GTextField) SetFont(value string) {
	if t.font == value {
		return
	}
	t.font = value
	if sprite := t.GObject.DisplayObject(); sprite != nil {
		sprite.Repaint()
	}
}

// Font returns the stored font identifier.
func (t *GTextField) Font() string {
	return t.font
}

// SetOutlineColor stores the outline colour value.
func (t *GTextField) SetOutlineColor(value string) {
	if t.outlineColor == value {
		return
	}
	t.outlineColor = value
	if sprite := t.GObject.DisplayObject(); sprite != nil {
		sprite.Repaint()
	}
}

// OutlineColor returns the current outline colour.
func (t *GTextField) OutlineColor() string {
	return t.outlineColor
}

// SetFontSize records the font size associated with the text.
func (t *GTextField) SetFontSize(size int) {
	if t.fontSize == size {
		return
	}
	t.fontSize = size
	// 参考 TypeScript 版本：设置 displayObject.fontSize 会触发重绘
	// Go 版本需要手动标记 Sprite 需要重绘
	if sprite := t.GObject.DisplayObject(); sprite != nil {
		sprite.Repaint()
	}
}

// FontSize returns the stored font size.
func (t *GTextField) FontSize() int {
	return t.fontSize
}

// SetAlign updates the horizontal alignment.
func (t *GTextField) SetAlign(value TextAlign) {
	if value == "" {
		value = TextAlignLeft
	}
	t.align = value
}

// Align returns the horizontal alignment.
func (t *GTextField) Align() TextAlign {
	return t.align
}

// SetVerticalAlign updates the vertical alignment.
func (t *GTextField) SetVerticalAlign(value TextVerticalAlign) {
	if value == "" {
		value = TextVerticalAlignTop
	}
	t.vertical = value
}

// VerticalAlign returns the vertical alignment.
func (t *GTextField) VerticalAlign() TextVerticalAlign {
	return t.vertical
}

// SetAutoSize configures the autosize behaviour.
func (t *GTextField) SetAutoSize(value TextAutoSize) {
	t.autoSize = value
	t.widthAutoSize = value == TextAutoSizeBoth
	t.heightAutoSize = value == TextAutoSizeBoth || value == TextAutoSizeHeight
	t.applyAutoSize()
}

// AutoSize returns the stored autosize mode.
func (t *GTextField) AutoSize() TextAutoSize {
	return t.autoSize
}

// SetSingleLine toggles single-line mode.
func (t *GTextField) SetSingleLine(value bool) {
	t.singleLine = value
}

// SingleLine reports whether single-line mode is active.
func (t *GTextField) SingleLine() bool {
	return t.singleLine
}

// WidthAutoSize reports whether width auto-size mode is active.
func (t *GTextField) WidthAutoSize() bool {
	return t.widthAutoSize
}

// SetUnderline toggles underline rendering.
func (t *GTextField) SetUnderline(value bool) {
	t.underline = value
}

// Underline reports whether underline should be drawn.
func (t *GTextField) Underline() bool {
	return t.underline
}

// SetItalic toggles italic styling.
func (t *GTextField) SetItalic(value bool) {
	t.italic = value
}

// Italic reports whether italic styling is requested.
func (t *GTextField) Italic() bool {
	return t.italic
}

// SetBold toggles bold styling.
func (t *GTextField) SetBold(value bool) {
	t.bold = value
}

// Bold reports whether bold styling is requested.
func (t *GTextField) Bold() bool {
	return t.bold
}

// SetLetterSpacing stores the additional spacing between characters (pixels).
func (t *GTextField) SetLetterSpacing(value int) {
	t.letterSpace = value
}

// LetterSpacing returns stored spacing between characters.
func (t *GTextField) LetterSpacing() int {
	return t.letterSpace
}

// SetLeading stores the additional line spacing in pixels.
func (t *GTextField) SetLeading(value int) {
	t.leading = value
}

// Leading returns the stored line spacing.
func (t *GTextField) Leading() int {
	return t.leading
}

// SetStrokeSize stores the outline thickness.
func (t *GTextField) SetStrokeSize(value float64) {
	t.strokeSize = value
}

// StrokeSize returns the outline thickness.
func (t *GTextField) StrokeSize() float64 {
	return t.strokeSize
}

// SetStrokeColor configures the outline colour.
func (t *GTextField) SetStrokeColor(value string) {
	t.strokeColor = value
}

// StrokeColor returns the outline colour.
func (t *GTextField) StrokeColor() string {
	return t.strokeColor
}

// SetUBBEnabled toggles UBB formatting support.
func (t *GTextField) SetUBBEnabled(value bool) {
	if t == nil {
		return
	}
	t.ubbEnabled = value
	if !value {
		t.detachLinkHandler()
	} else if len(t.linkRegions) > 0 {
		t.attachLinkHandler()
	}
}

// UBBEnabled reports whether UBB formatting is enabled.
func (t *GTextField) UBBEnabled() bool {
	return t.ubbEnabled
}

// SetShadow configures drop-shadow styling.
func (t *GTextField) SetShadow(color string, offsetX, offsetY, blur float64) {
	t.shadowColor = color
	t.shadowOffsetX = offsetX
	t.shadowOffsetY = offsetY
	t.shadowBlur = blur
}

// ShadowColor returns the configured drop-shadow colour.
func (t *GTextField) ShadowColor() string {
	return t.shadowColor
}

// ShadowOffset returns the drop-shadow offset in pixels.
func (t *GTextField) ShadowOffset() (float64, float64) {
	return t.shadowOffsetX, t.shadowOffsetY
}

// ShadowBlur returns the configured blur radius.
func (t *GTextField) ShadowBlur() float64 {
	return t.shadowBlur
}

// SetTemplateVarsEnabled records whether template variables are active.
func (t *GTextField) SetTemplateVarsEnabled(value bool) {
	t.templateVars = value
}

// TemplateVarsEnabled reports template variable usage.
func (t *GTextField) TemplateVarsEnabled() bool {
	return t.templateVars
}

// SetLinkRegions updates the clickable link regions referenced by this field.
func (t *GTextField) SetLinkRegions(regions []TextLinkRegion) {
	if t == nil {
		return
	}
	if len(regions) == 0 {
		t.linkRegions = nil
		t.detachLinkHandler()
		return
	}
	copyRegions := make([]TextLinkRegion, len(regions))
	copy(copyRegions, regions)
	t.linkRegions = copyRegions
	if t.ubbEnabled {
		t.attachLinkHandler()
	}
}

// LinkRegions returns a copy of the active link regions.
func (t *GTextField) LinkRegions() []TextLinkRegion {
	if t == nil || len(t.linkRegions) == 0 {
		return nil
	}
	out := make([]TextLinkRegion, len(t.linkRegions))
	copy(out, t.linkRegions)
	return out
}

// SetLayoutUpdateCallback 设置布局更新回调，类似 TS 版本的 _onPostLayout
func (t *GTextField) SetLayoutUpdateCallback(callback func()) {
	if t == nil {
		return
	}
	t.onLayoutUpdated = callback
}

// notifyLayoutUpdate 触发布局更新回调
func (t *GTextField) notifyLayoutUpdate() {
	if t == nil || t.onLayoutUpdated == nil {
		return
	}
	// 异步调用回调，避免递归调用
	defer func() {
		if recover() != nil {
			// 忽略回调中的 panic，避免影响主流程
		}
	}()
	t.onLayoutUpdated()
}

// RequestLayout 请求重新计算布局，参考 TS 版本的 typeset 方法
func (t *GTextField) RequestLayout() {
	if t == nil {
		return
	}
	// 标记需要重新计算，实际的布局计算在渲染时进行
	t.layoutWidth = 0
	t.layoutHeight = 0
	// 如果有父组件，通知父组件也需要重新布局
	if t.GObject != nil && t.GObject.Parent() != nil {
		// 这里可以添加父组件布局更新逻辑
	}
}

func (t *GTextField) attachLinkHandler() {
	if t == nil || t.linkHandler != nil || len(t.linkRegions) == 0 || !t.ubbEnabled {
		return
	}
	if t.GObject == nil {
		return
	}
	sprite := t.GObject.DisplayObject()
	if sprite == nil {
		return
	}
	handler := func(evt laya.Event) {
		pe, ok := evt.Data.(laya.PointerEvent)
		if !ok {
			return
		}
		local := sprite.GlobalToLocal(pe.Position)
		for _, region := range t.linkRegions {
			if region.Bounds.Contains(local) {
				sprite.EmitWithBubble(laya.EventLink, region.Target)
				return
			}
		}
	}
	sprite.Dispatcher().On(laya.EventClick, handler)
	t.linkHandler = handler
}

func (t *GTextField) detachLinkHandler() {
	if t == nil || t.linkHandler == nil {
		return
	}
	if t.GObject != nil {
		if sprite := t.GObject.DisplayObject(); sprite != nil {
			sprite.Dispatcher().Off(laya.EventClick, t.linkHandler)
		}
	}
	t.linkHandler = nil
}

// UpdateLayoutMetrics stores the measured layout dimensions and applies auto-size rules.
func (t *GTextField) UpdateLayoutMetrics(layoutWidth, layoutHeight, textWidth, textHeight float64) {
	if t == nil || t.GObject == nil {
		return
	}
	if math.IsNaN(layoutWidth) || math.IsInf(layoutWidth, 0) {
		layoutWidth = 0
	}
	if math.IsNaN(layoutHeight) || math.IsInf(layoutHeight, 0) {
		layoutHeight = 0
	}
	if layoutWidth < 0 {
		layoutWidth = 0
	}
	if layoutHeight < 0 {
		layoutHeight = 0
	}

	// 检查是否发生了显著变化
	sizeChanged := !almostEqual(t.layoutWidth, layoutWidth) ||
		!almostEqual(t.layoutHeight, layoutHeight) ||
		!almostEqual(t.textWidth, textWidth) ||
		!almostEqual(t.textHeight, textHeight)

	t.layoutWidth = layoutWidth
	t.layoutHeight = layoutHeight
	if textWidth < 0 {
		textWidth = 0
	}
	if textHeight < 0 {
		textHeight = 0
	}
	t.textWidth = textWidth
	t.textHeight = textHeight

	// 只有在尺寸发生变化时才应用自动大小和触发回调
	if sizeChanged {
		t.applyAutoSize()
	}
}

// SetupBeforeAdd populates text styling metadata from the component buffer.
// 对应 TypeScript 版本 GTextField.setup_beforeAdd
func (t *GTextField) SetupBeforeAdd(buf *utils.ByteBuffer, beginPos int) {
	if t == nil || buf == nil {
		return
	}

	// 首先调用父类处理基础属性（位置、尺寸、旋转等）
	t.GObject.SetupBeforeAdd(buf, beginPos)

	// 然后处理GTextField特定属性（font, fontSize, color等）
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(beginPos, 5) {
		return
	}
	if font := buf.ReadS(); font != nil && *font != "" {
		t.SetFont(*font)
	}
	if size := int(buf.ReadInt16()); size > 0 {
		t.SetFontSize(size)
	}
	if color := buf.ReadColorString(true); color != "" {
		t.SetColor(color)
	}
	mapAlign := func(code int8) TextAlign {
		switch code {
		case 1:
			return TextAlignCenter
		case 2:
			return TextAlignRight
		default:
			return TextAlignLeft
		}
	}
	mapVAlign := func(code int8) TextVerticalAlign {
		switch code {
		case 1:
			return TextVerticalAlignMiddle
		case 2:
			return TextVerticalAlignBottom
		default:
			return TextVerticalAlignTop
		}
	}
	t.SetAlign(mapAlign(buf.ReadByte()))
	t.SetVerticalAlign(mapVAlign(buf.ReadByte()))
	t.SetLeading(int(buf.ReadInt16()))
	t.SetLetterSpacing(int(buf.ReadInt16()))
	t.SetUBBEnabled(buf.ReadBool())
	t.SetAutoSize(TextAutoSize(buf.ReadByte()))
	t.SetUnderline(buf.ReadBool())
	t.SetItalic(buf.ReadBool())
	t.SetBold(buf.ReadBool())
	t.SetSingleLine(buf.ReadBool())
	if buf.ReadBool() {
		if strokeColor := buf.ReadColorString(true); strokeColor != "" {
			t.SetStrokeColor(strokeColor)
		}
		t.SetStrokeSize(float64(buf.ReadFloat32()) + 1)
	}
	if buf.ReadBool() {
		_ = buf.Skip(12) // shadow information currently unused
	}
	if buf.ReadBool() {
		t.SetTemplateVarsEnabled(true)
	}
	if buf.Seek(0, 6) {
		if text := buf.ReadS(); text != nil {
			t.SetText(*text)
		}
	}
}

// TextWidth returns the latest measured text width.
func (t *GTextField) TextWidth() float64 {
	return t.textWidth
}

// TextHeight returns the latest measured text height.
func (t *GTextField) TextHeight() float64 {
	return t.textHeight
}

func (t *GTextField) applyAutoSize() {
	if t == nil || t.GObject == nil {
		return
	}
	currentWidth := t.GObject.Width()
	currentHeight := t.GObject.Height()

	targetWidth := currentWidth
	targetHeight := currentHeight

	// 改进的自动大小逻辑，参考TS版本实现
	if t.widthAutoSize {
		// 优先使用布局度量值（由渲染层计算的结果）
		if t.layoutWidth > 0 {
			targetWidth = t.layoutWidth
		} else if t.textWidth > 0 {
			targetWidth = t.textWidth + t.getHorizontalPadding()
		}
		// 设置最小宽度保护
		if targetWidth < t.minTextWidth() {
			targetWidth = t.minTextWidth()
		}
	}

	if t.heightAutoSize {
		// 优先使用布局度量值（由渲染层计算的结果）
		if t.layoutHeight > 0 {
			targetHeight = t.layoutHeight
		} else if t.textHeight > 0 {
			targetHeight = t.textHeight + t.getVerticalPadding()
		}
		// 设置最小高度保护
		if targetHeight < t.minTextHeight() {
			targetHeight = t.minTextHeight()
		}
	}

	// 只有在尺寸发生显著变化时才更新，避免不必要的布局重计算
	if !almostEqual(targetWidth, currentWidth) || !almostEqual(targetHeight, currentHeight) {
		t.GObject.SetSize(targetWidth, targetHeight)
		t.notifyLayoutUpdate()
	}
}

func almostEqual(a, b float64) bool {
	const epsilon = 0.5
	diff := math.Abs(a - b)
	return diff <= epsilon
}

// getHorizontalPadding 返回水平方向的总内边距
func (t *GTextField) getHorizontalPadding() float64 {
	if t == nil {
		return 0
	}
	// LayaAir 的 labelPadding 是固定的 [2, 2, 2, 2]
	// 描边效果由渲染层处理，不影响文字容器的 padding
	padding := 4.0 // 左右各2像素

	// 只有阴影会影响 padding（如果有的话）
	if t.shadowOffsetX != 0 {
		padding += math.Abs(t.shadowOffsetX)
	}
	return padding
}

// getVerticalPadding 返回垂直方向的总内边距
func (t *GTextField) getVerticalPadding() float64 {
	if t == nil {
		return 0
	}
	// LayaAir 的 labelPadding 是固定的 [2, 2, 2, 2]
	// 描边效果由渲染层处理，不影响文字容器的 padding
	padding := 4.0 // 上下各2像素

	// 只有阴影会影响 padding（如果有的话）
	if t.shadowOffsetY != 0 {
		padding += math.Abs(t.shadowOffsetY)
	}
	return padding
}

// minTextWidth 返回最小文本宽度
func (t *GTextField) minTextWidth() float64 {
	if t.fontSize > 0 {
		return float64(t.fontSize) // 至少能容纳一个字符
	}
	return 12.0 // 默认最小宽度
}

// minTextHeight 返回最小文本高度
func (t *GTextField) minTextHeight() float64 {
	if t.fontSize > 0 {
		return float64(t.fontSize) * 1.2 // 行高通常是字体大小的1.2倍
	}
	return 14.0 // 默认最小高度
}
