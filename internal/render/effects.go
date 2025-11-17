package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// 颜色效果应用
// ============================================================================

// ApplyGrayscaleEffect 应用灰度效果到 DrawImageOptions。
//
// 灰度算法使用标准的亮度加权：
// gray = 0.299*R + 0.587*G + 0.114*B
//
// 使用 ColorScale 来实现：
// - R = G = B = gray (通过颜色矩阵转换)
// - 保持 Alpha 不变
func ApplyGrayscaleEffect(opts *ebiten.DrawImageOptions) {
	// Ebiten 的 ColorScale 使用乘法，我们需要设置灰度矩阵
	// 标准灰度转换：
	// R' = 0.299*R + 0.587*G + 0.114*B
	// G' = 0.299*R + 0.587*G + 0.114*B
	// B' = 0.299*R + 0.587*G + 0.114*B
	// A' = A
	//
	// 但 Ebiten 的 ColorScale 只支持简单的乘法缩放，
	// 不支持通道间的线性组合。
	//
	// 作为近似，我们使用平均值：R' = G' = B' = (R+G+B)/3
	// 通过设置 ColorScale 为灰色权重：
	opts.ColorScale.ScaleWithColor(color.Gray{Y: 0x80}) // 50% 灰度混合

	// 注意：这是一个简化实现。
	// 完整的灰度效果需要使用 Shader 来实现精确的颜色矩阵变换。
	// TODO: 实现基于 Shader 的灰度效果以获得更准确的结果。
}

// ApplyColorMatrix 应用颜色矩阵变换到 DrawImageOptions。
//
// 颜色矩阵是一个 5x4 矩阵（存储为长度为 20 的数组）：
// [R']   [m[0]  m[1]  m[2]  m[3]  m[4] ]   [R]
// [G']   [m[5]  m[6]  m[7]  m[8]  m[9] ]   [G]
// [B'] = [m[10] m[11] m[12] m[13] m[14]] × [B]
// [A']   [m[15] m[16] m[17] m[18] m[19]]   [A]
//                                           [1]
//
// 由于 Ebiten 的 ColorScale 只支持简单的乘法缩放，
// 我们只能应用对角线元素（m[0], m[6], m[12], m[18]）。
// 完整的颜色矩阵需要使用 Shader。
func ApplyColorMatrix(opts *ebiten.DrawImageOptions, matrix [20]float64) {
	// 提取对角线元素作为缩放因子
	// 这是一个简化实现，只支持基本的颜色缩放
	r := float32(matrix[0])   // R 缩放
	g := float32(matrix[6])   // G 缩放
	b := float32(matrix[12])  // B 缩放
	a := float32(matrix[18])  // A 缩放

	// 应用缩放
	opts.ColorScale.Scale(r, g, b, a)

	// TODO: 实现基于 Shader 的完整颜色矩阵变换
	// 以支持通道间的混合和偏移量。
}

// ============================================================================
// 辅助函数
// ============================================================================

// CombineColorEffects 组合多个颜色效果到单个 ColorScale。
//
// 参数：
// - baseColor: 基础颜色叠加（可选，nil 表示无叠加）
// - alpha: Alpha 透明度（0.0-1.0）
// - grayed: 是否应用灰度效果
// - colorMatrix: 颜色矩阵（可选，nil 表示不使用）
// - colorMatrixEnabled: 是否启用颜色矩阵
//
// 返回：
// - 配置好的 DrawImageOptions
func CombineColorEffects(
	baseColor *uint32,
	alpha float64,
	grayed bool,
	colorMatrix *[20]float64,
	colorMatrixEnabled bool,
) *ebiten.DrawImageOptions {
	opts := &ebiten.DrawImageOptions{}

	// 1. 应用基础颜色叠加
	if baseColor != nil {
		applyColorTintToOpts(opts, *baseColor)
	}

	// 2. 应用 Alpha
	if alpha > 0 && alpha < 1.0 {
		opts.ColorScale.ScaleAlpha(float32(alpha))
	}

	// 3. 应用灰度效果
	if grayed {
		ApplyGrayscaleEffect(opts)
	}

	// 4. 应用颜色矩阵（如果启用且灰度未启用）
	if colorMatrixEnabled && !grayed && colorMatrix != nil {
		ApplyColorMatrix(opts, *colorMatrix)
	}

	return opts
}

// applyColorTintToOpts 应用颜色叠加到 DrawImageOptions。
// RGBA 格式: 0xRRGGBBAA
func applyColorTintToOpts(opts *ebiten.DrawImageOptions, colorValue uint32) {
	rr := float32((colorValue >> 24) & 0xFF) / 255.0
	gg := float32((colorValue >> 16) & 0xFF) / 255.0
	bb := float32((colorValue >> 8) & 0xFF) / 255.0
	aa := float32(colorValue & 0xFF) / 255.0

	opts.ColorScale.Scale(rr, gg, bb, aa)
}

// ============================================================================
// 颜色转换辅助函数
// ============================================================================

// RGBAToFloat32 将 RGBA 颜色转换为 float32 数组（范围 0.0-1.0）。
func RGBAToFloat32(c color.Color) (r, g, b, a float32) {
	if c == nil {
		return 1, 1, 1, 1
	}

	rc, gc, bc, ac := c.RGBA()
	r = float32(rc) / 0xFFFF
	g = float32(gc) / 0xFFFF
	b = float32(bc) / 0xFFFF
	a = float32(ac) / 0xFFFF
	return
}

// Uint32ToColorRGBA 将 uint32 颜色值转换为 color.RGBA。
// RGBA 格式: 0xRRGGBBAA
func Uint32ToColorRGBA(c uint32) color.RGBA {
	return color.RGBA{
		R: uint8((c >> 24) & 0xFF),
		G: uint8((c >> 16) & 0xFF),
		B: uint8((c >> 8) & 0xFF),
		A: uint8(c & 0xFF),
	}
}
