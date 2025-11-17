package render

import (
	"image/color"

	"github.com/chslink/fairygui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// ============================================================================
// 文本渲染
// ============================================================================

// DrawTextBasic 绘制基础文本（使用 Ebiten 内置文本渲染）。
func (r *EbitenRenderer) DrawTextBasic(
	screen *ebiten.Image,
	content string,
	x, y float64,
	style fairygui.TextStyle,
) {
	if content == "" {
		return
	}

	// 转换颜色 (RGBA 格式: 0xRRGGBBAA)
	textColor := r.uint32ToColor(style.Color)

	// 创建绘制选项
	opts := &text.DrawOptions{}
	opts.GeoM.Translate(x, y)
	opts.ColorScale.ScaleWithColor(textColor)

	// 获取字体
	face := r.getFontFace(style)
	if face == nil {
		// 如果没有字体，使用默认字体
		// 注意：Ebiten v2.7+ 默认字体需要特殊处理
		return
	}

	// 绘制文本
	text.Draw(screen, content, face, opts)
	r.drawCalls++
}

// DrawTextWithStroke 绘制带描边的文本。
func (r *EbitenRenderer) DrawTextWithStroke(
	screen *ebiten.Image,
	content string,
	x, y float64,
	style fairygui.TextStyle,
) {
	if content == "" || style.Stroke == nil {
		r.DrawTextBasic(screen, content, x, y, style)
		return
	}

	// 绘制描边（在 8 个方向绘制）
	stroke := style.Stroke
	thickness := stroke.Thickness

	// 临时修改样式为描边颜色
	strokeStyle := style
	strokeStyle.Color = stroke.Color
	strokeStyle.Stroke = nil // 避免递归

	offsets := []struct{ dx, dy float64 }{
		{-thickness, -thickness}, // 左上
		{0, -thickness},          // 上
		{thickness, -thickness},  // 右上
		{-thickness, 0},          // 左
		{thickness, 0},           // 右
		{-thickness, thickness},  // 左下
		{0, thickness},           // 下
		{thickness, thickness},   // 右下
	}

	for _, offset := range offsets {
		r.DrawTextBasic(screen, content, x+offset.dx, y+offset.dy, strokeStyle)
	}

	// 绘制原文本
	r.DrawTextBasic(screen, content, x, y, style)
}

// DrawTextWithShadow 绘制带阴影的文本。
func (r *EbitenRenderer) DrawTextWithShadow(
	screen *ebiten.Image,
	content string,
	x, y float64,
	style fairygui.TextStyle,
) {
	if content == "" || style.Shadow == nil {
		r.DrawTextBasic(screen, content, x, y, style)
		return
	}

	// 绘制阴影
	shadow := style.Shadow
	shadowStyle := style
	shadowStyle.Color = shadow.Color
	shadowStyle.Shadow = nil // 避免递归

	// 阴影偏移
	shadowX := x + shadow.OffsetX
	shadowY := y + shadow.OffsetY

	// TODO: 实现模糊效果 (shadow.Blur)
	// Ebiten 不直接支持模糊，可能需要使用多次绘制或 shader

	r.DrawTextBasic(screen, content, shadowX, shadowY, shadowStyle)

	// 绘制原文本
	r.DrawTextBasic(screen, content, x, y, style)
}

// MeasureText 测量文本尺寸。
func (r *EbitenRenderer) MeasureText(content string, style fairygui.TextStyle) (width, height float64) {
	if content == "" {
		return 0, 0
	}

	face := r.getFontFace(style)
	if face == nil {
		// 使用默认尺寸估算
		return float64(len(content)) * style.Size * 0.6, style.Size
	}

	// 使用 Ebiten 文本测量
	w, h := text.Measure(content, face, 0)
	return w, h
}

// uint32ToColor 将 uint32 颜色转换为 color.Color。
func (r *EbitenRenderer) uint32ToColor(c uint32) color.Color {
	// RGBA 格式: 0xRRGGBBAA
	rr := uint8((c >> 24) & 0xFF)
	gg := uint8((c >> 16) & 0xFF)
	bb := uint8((c >> 8) & 0xFF)
	aa := uint8(c & 0xFF)

	return color.RGBA{R: rr, G: gg, B: bb, A: aa}
}

// getFontFace 根据样式获取字体 Face。
func (r *EbitenRenderer) getFontFace(style fairygui.TextStyle) text.Face {
	// TODO: 实现字体管理系统
	// 目前返回 nil，使用调用方的默认处理

	// 字体管理应该包括：
	// 1. 系统字体加载
	// 2. 自定义字体加载
	// 3. 字体缓存
	// 4. 字号缩放
	// 5. 粗体/斜体支持

	return nil
}

// ============================================================================
// 文本对齐辅助函数
// ============================================================================

// AlignTextX 根据对齐方式计算文本 X 坐标。
func AlignTextX(x, width, textWidth float64, align fairygui.TextAlign) float64 {
	switch align {
	case fairygui.TextAlignLeft:
		return x
	case fairygui.TextAlignCenter:
		return x + (width-textWidth)/2
	case fairygui.TextAlignRight:
		return x + width - textWidth
	default:
		return x
	}
}

// AlignTextY 根据垂直对齐方式计算文本 Y 坐标。
func AlignTextY(y, height, textHeight float64, verticalAlign string) float64 {
	switch verticalAlign {
	case "top":
		return y
	case "middle":
		return y + (height-textHeight)/2
	case "bottom":
		return y + height - textHeight
	default:
		return y
	}
}
