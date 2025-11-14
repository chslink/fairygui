package render

import (
	"image"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// RenderContext 渲染上下文
// 借鉴 Unity 版本的 UpdateContext 设计，统一管理渲染状态
type RenderContext struct {
	// 剪裁栈
	clipStack []ClipInfo

	// 当前渲染状态
	alpha        float64       // 全局透明度
	grayed       bool          // 是否灰度
	colorScale   ebiten.ColorScale // 颜色缩放
	blend        ebiten.Blend  // 混合模式

	// 渲染统计
	stats RenderStats
}

// ClipInfo 剪裁信息
// 借鉴 Unity 版本的 ClipInfo 设计
type ClipInfo struct {
	Rect      image.Rectangle // 剪裁矩形
	Soft      bool            // 是否软边
	Reversed  bool            // 是否反向遮罩
	Alpha     float64         // 剪裁时的透明度
	PrevState RenderState     // 前一个状态
}

// RenderState 渲染状态
type RenderState struct {
	Alpha     float64
	Grayed    bool
	Color     ebiten.ColorScale
	Blend     ebiten.Blend
}

// RenderStats 渲染统计信息
type RenderStats struct {
	DrawCallCount   int     // DrawCall 数量
	TrianglesCount  int     // 三角形数量
	BatchCount      int     // 批处理数量
	ClipDepth       int     // 当前剪裁深度
	TotalAlpha      float64 // 累计透明度
}

// NewRenderContext 创建渲染上下文
func NewRenderContext() *RenderContext {
	return &RenderContext{
		clipStack: make([]ClipInfo, 0, 8),
		alpha:     1.0,
		colorScale: ebiten.ColorScale{},
		blend:     ebiten.Blend{},
		stats:     RenderStats{},
	}
}

// Begin 开始渲染帧
// 借鉴 Unity 版本的 UpdateContext.Begin()
func (ctx *RenderContext) Begin() {
	// 重置状态
	ctx.alpha = 1.0
	ctx.grayed = false
	ctx.colorScale = ebiten.ColorScale{}
	ctx.blend = ebiten.Blend{}

	// 清空剪裁栈
	ctx.clipStack = ctx.clipStack[:0]

	// 重置统计
	ctx.stats.DrawCallCount = 0
	ctx.stats.TrianglesCount = 0
	ctx.stats.BatchCount = 0
	ctx.stats.ClipDepth = 0
	ctx.stats.TotalAlpha = 1.0
}

// End 结束渲染帧
// 借鉴 Unity 版本的 UpdateContext.End()
func (ctx *RenderContext) End() {
	// 清理状态
	ctx.clipStack = ctx.clipStack[:0]
	ctx.alpha = 1.0
	ctx.grayed = false
}

// EnterClipping 进入剪裁模式
// 借鉴 Unity 版本的 EnterClipping()
func (ctx *RenderContext) EnterClipping(rect image.Rectangle, soft bool) {
	// 保存当前状态
	prev := ClipInfo{
		Rect:      ctx.GetCurrentClipRect(),
		Soft:      ctx.HasSoftClipping(),
		Reversed:  false,
		Alpha:     ctx.alpha,
		PrevState: RenderState{
			Alpha:  ctx.alpha,
			Grayed: ctx.grayed,
			Color:  ctx.colorScale,
			Blend:  ctx.blend,
		},
	}
	ctx.clipStack = append(ctx.clipStack, prev)

	// 更新剪裁状态
	ctx.stats.ClipDepth++

	// 剪裁会改变后续渲染的区域，但不改变全局状态
	// 注意：Unity 中的 clipBox 计算需要 shader 支持，Ebiten 中使用临时图像方案
}

// LeaveClipping 离开剪裁模式
// 借鉴 Unity 版本的 LeaveClipping()
func (ctx *RenderContext) LeaveClipping() {
	if len(ctx.clipStack) > 0 {
		prev := ctx.clipStack[len(ctx.clipStack)-1]
		ctx.clipStack = ctx.clipStack[:len(ctx.clipStack)-1]

		// 恢复前一个状态
		ctx.alpha = prev.PrevState.Alpha
		ctx.grayed = prev.PrevState.Grayed
		ctx.colorScale = prev.PrevState.Color
		ctx.blend = prev.PrevState.Blend

		ctx.stats.ClipDepth--
	}
}

// EnterMask 进入遮罩模式
// 借鉴 Unity 版本的 stencil mask 实现
func (ctx *RenderContext) EnterMask(maskID uint, reversed bool) {
	// 保存当前状态
	prev := ClipInfo{
		Rect:      ctx.GetCurrentClipRect(),
		Soft:      false,
		Reversed:  reversed,
		Alpha:     ctx.alpha,
		PrevState: RenderState{
			Alpha:  ctx.alpha,
			Grayed: ctx.grayed,
			Color:  ctx.colorScale,
			Blend:  ctx.blend,
		},
	}
	ctx.clipStack = append(ctx.clipStack, prev)

	// 遮罩模式下，所有后续渲染会被 mask 形状限制
	ctx.stats.ClipDepth++
}

// LeaveMask 离开遮罩模式
func (ctx *RenderContext) LeaveMask() {
	ctx.LeaveClipping()
}

// SetAlpha 设置透明度
func (ctx *RenderContext) SetAlpha(alpha float64) {
	ctx.alpha = alpha
	ctx.stats.TotalAlpha = alpha
}

// GetAlpha 获取当前透明度
func (ctx *RenderContext) GetAlpha() float64 {
	return ctx.alpha
}

// SetGrayed 设置灰度模式
func (ctx *RenderContext) SetGrayed(grayed bool) {
	ctx.grayed = grayed
}

// IsGrayed 检查是否为灰度模式
func (ctx *RenderContext) IsGrayed() bool {
	return ctx.grayed
}

// SetColorScale 设置颜色缩放
func (ctx *RenderContext) SetColorScale(color ebiten.ColorScale) {
	ctx.colorScale = color
}

// GetColorScale 获取颜色缩放
func (ctx *RenderContext) GetColorScale() ebiten.ColorScale {
	return ctx.colorScale
}

// SetBlend 设置混合模式
func (ctx *RenderContext) SetBlend(blend ebiten.Blend) {
	ctx.blend = blend
}

// GetBlend 获取混合模式
func (ctx *RenderContext) GetBlend() ebiten.Blend {
	return ctx.blend
}

// GetCurrentClipRect 获取当前剪裁矩形
func (ctx *RenderContext) GetCurrentClipRect() image.Rectangle {
	if len(ctx.clipStack) > 0 {
		return ctx.clipStack[len(ctx.clipStack)-1].Rect
	}
	return image.Rectangle{}
}

// HasSoftClipping 是否有软边剪裁
func (ctx *RenderContext) HasSoftClipping() bool {
	if len(ctx.clipStack) > 0 {
		return ctx.clipStack[len(ctx.clipStack)-1].Soft
	}
	return false
}

// IsClipping 检查是否处于剪裁状态
func (ctx *RenderContext) IsClipping() bool {
	return len(ctx.clipStack) > 0
}

// IsMasking 检查是否处于遮罩状态
func (ctx *RenderContext) IsMasking() bool {
	for _, clip := range ctx.clipStack {
		if clip.Reversed || clip.Soft {
			return true
		}
	}
	return false
}

// GetClipDepth 获取当前剪裁深度
func (ctx *RenderContext) GetClipDepth() int {
	return ctx.stats.ClipDepth
}

// ApplyToOptions 将状态应用到 DrawImageOptions
// 借鉴 Unity 版本的 ApplyClippingProperties
func (ctx *RenderContext) ApplyToOptions(opts *ebiten.DrawImageOptions) {
	if opts == nil {
		return
	}

	// 应用颜色缩放
	opts.ColorScale = ctx.colorScale

	// 应用混合模式
	opts.Blend = ctx.blend

	// 应用透明度（作为 ColorScale 的一部分）
	if ctx.alpha < 1.0 {
		opts.ColorScale.ScaleAlpha(float32(ctx.alpha))
	}

	// 灰度效果通过 ColorScale 实现
	if ctx.grayed {
		// 灰度：R=G=B=(R+G+B)/3
		// 这里可以应用一个颜色矩阵
		// 简化实现：降低饱和度
		opts.ColorScale.Scale(0.8, 0.8, 0.8, 1.0)
	}
}

// GetStats 获取渲染统计
func (ctx *RenderContext) GetStats() RenderStats {
	return ctx.stats
}

// IncrementDrawCall 增加 DrawCall 计数
func (ctx *RenderContext) IncrementDrawCall() {
	ctx.stats.DrawCallCount++
}

// IncrementTriangles 增加三角形计数
func (ctx *RenderContext) IncrementTriangles(count int) {
	ctx.stats.TrianglesCount += count
}

// globalRenderContext 全局渲染上下文
var globalRenderContext = NewRenderContext()

// GetGlobalRenderContext 获取全局渲染上下文
func GetGlobalRenderContext() *RenderContext {
	return globalRenderContext
}

// RenderContextPool 渲染上下文对象池
// 用于多线程场景或多个渲染目标
var RenderContextPool = sync.Pool{
	New: func() interface{} {
		return NewRenderContext()
	},
}

// GetRenderContext 从对象池获取渲染上下文
func GetRenderContext() *RenderContext {
	return RenderContextPool.Get().(*RenderContext)
}

// PutRenderContext 将渲染上下文返回对象池
func PutRenderContext(ctx *RenderContext) {
	if ctx == nil {
		return
	}
	ctx.End()
	RenderContextPool.Put(ctx)
}
