package render

import (
	"github.com/chslink/fairygui"
	"github.com/chslink/fairygui/internal/display"
	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// EbitenRenderer - Ebiten 渲染器
// ============================================================================

// EbitenRenderer 是基于 Ebiten 的渲染器实现。
type EbitenRenderer struct {
	// 渲染统计
	drawCalls int
	vertices  int
}

// NewEbitenRenderer 创建一个新的 Ebiten 渲染器。
func NewEbitenRenderer() *EbitenRenderer {
	return &EbitenRenderer{}
}

// Draw 渲染显示对象树到屏幕。
func (r *EbitenRenderer) Draw(screen *ebiten.Image, root fairygui.DisplayObject) {
	if root == nil {
		return
	}

	// 重置统计
	r.drawCalls = 0
	r.vertices = 0

	// 递归渲染显示对象树
	r.drawObject(screen, root)
}

// drawObject 递归渲染单个显示对象及其子对象。
func (r *EbitenRenderer) drawObject(screen *ebiten.Image, obj fairygui.DisplayObject) {
	if obj == nil || !obj.Visible() {
		return
	}

	// 获取 alpha
	alpha := obj.Alpha()
	if alpha <= 0 {
		return
	}

	// 绘制对象本身
	r.drawSingle(screen, obj, alpha)

	// 递归绘制子对象
	children := obj.Children()
	for _, child := range children {
		r.drawObject(screen, child)
	}
}

// drawSingle 绘制单个对象（不包括子对象）。
func (r *EbitenRenderer) drawSingle(screen *ebiten.Image, obj fairygui.DisplayObject, alpha float64) {
	// 获取 sprite（通过类型断言）
	sprite := r.getSprite(obj)
	if sprite == nil {
		return
	}

	// 获取纹理
	texture := sprite.Texture()
	if texture == nil {
		return
	}

	// 获取绘制选项
	opts := sprite.DrawOptions()
	if opts == nil {
		opts = &ebiten.DrawImageOptions{}
	}

	// 应用 alpha
	if alpha < 1.0 {
		opts.ColorScale.ScaleAlpha(float32(alpha))
	}

	// 绘制纹理
	screen.DrawImage(texture, opts)
	r.drawCalls++
}

// getSprite 从 DisplayObject 获取内部 sprite。
func (r *EbitenRenderer) getSprite(obj fairygui.DisplayObject) *display.Sprite {
	// 这里需要一个辅助方法来获取内部 sprite
	// 由于我们控制了实现，可以使用类型断言
	type spriteGetter interface {
		Sprite() *display.Sprite
	}

	if sg, ok := obj.(spriteGetter); ok {
		return sg.Sprite()
	}

	return nil
}

// DrawText 渲染文本。
func (r *EbitenRenderer) DrawText(screen *ebiten.Image, text string, x, y float64, style fairygui.TextStyle) {
	// 文本渲染将在 Phase 3.3 实现
	// 目前只是占位实现
}

// DrawTexture 渲染纹理。
func (r *EbitenRenderer) DrawTexture(screen *ebiten.Image, texture *ebiten.Image, options fairygui.DrawOptions) {
	if texture == nil {
		return
	}

	// 创建 Ebiten 绘制选项
	opts := &ebiten.DrawImageOptions{}

	// 应用变换
	if options.ScaleX != 0 || options.ScaleY != 0 {
		scaleX := options.ScaleX
		scaleY := options.ScaleY
		if scaleX == 0 {
			scaleX = 1.0
		}
		if scaleY == 0 {
			scaleY = 1.0
		}
		opts.GeoM.Scale(scaleX, scaleY)
	}

	if options.Rotation != 0 {
		opts.GeoM.Rotate(options.Rotation)
	}

	if options.X != 0 || options.Y != 0 {
		opts.GeoM.Translate(options.X, options.Y)
	}

	// 应用 alpha
	if options.Alpha > 0 && options.Alpha < 1.0 {
		opts.ColorScale.ScaleAlpha(float32(options.Alpha))
	}

	// 应用颜色叠加
	if options.Color != nil {
		r.applyColorTint(opts, *options.Color)
	}

	// 应用混合模式
	r.applyBlendMode(opts, options.BlendMode)

	// 处理九宫格
	if options.NineSlice != nil {
		r.drawNineSlice(screen, texture, options, opts)
		return
	}

	// 处理平铺
	if options.Tiling {
		r.drawTiling(screen, texture, options, opts)
		return
	}

	// 普通绘制
	screen.DrawImage(texture, opts)
	r.drawCalls++
}

// DrawShape 渲染形状。
func (r *EbitenRenderer) DrawShape(screen *ebiten.Image, shape fairygui.Shape, options fairygui.DrawOptions) {
	// 形状渲染将在 Phase 3.4 实现
	// 目前只是占位实现
}

// applyColorTint 应用颜色叠加。
func (r *EbitenRenderer) applyColorTint(opts *ebiten.DrawImageOptions, color uint32) {
	// RGBA 格式: 0xRRGGBBAA
	rr := float32((color >> 24) & 0xFF) / 255.0
	gg := float32((color >> 16) & 0xFF) / 255.0
	bb := float32((color >> 8) & 0xFF) / 255.0
	aa := float32(color & 0xFF) / 255.0

	// 使用 Scale 方法应用颜色缩放
	opts.ColorScale.Scale(rr, gg, bb, aa)
}

// applyBlendMode 应用混合模式。
func (r *EbitenRenderer) applyBlendMode(opts *ebiten.DrawImageOptions, mode fairygui.BlendMode) {
	switch mode {
	case fairygui.BlendModeNormal:
		opts.Blend = ebiten.BlendSourceOver
	case fairygui.BlendModeAdd:
		opts.Blend = ebiten.BlendLighter
	case fairygui.BlendModeMultiply:
		// Ebiten 不直接支持 Multiply，使用默认
		opts.Blend = ebiten.BlendSourceOver
	case fairygui.BlendModeScreen:
		// Ebiten 不直接支持 Screen，使用默认
		opts.Blend = ebiten.BlendSourceOver
	default:
		opts.Blend = ebiten.BlendSourceOver
	}
}

// drawNineSlice 绘制九宫格纹理。
func (r *EbitenRenderer) drawNineSlice(screen *ebiten.Image, texture *ebiten.Image, options fairygui.DrawOptions, baseOpts *ebiten.DrawImageOptions) {
	r.DrawNineSliceTexture(screen, texture, options, baseOpts)
}

// drawTiling 绘制平铺纹理。
func (r *EbitenRenderer) drawTiling(screen *ebiten.Image, texture *ebiten.Image, options fairygui.DrawOptions, baseOpts *ebiten.DrawImageOptions) {
	r.DrawTilingTexture(screen, texture, options, baseOpts)
}

// DrawCalls 返回上一帧的绘制调用次数。
func (r *EbitenRenderer) DrawCalls() int {
	return r.drawCalls
}

// Vertices 返回上一帧的顶点数。
func (r *EbitenRenderer) Vertices() int {
	return r.vertices
}
