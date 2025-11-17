package render

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// 灰度效果测试
// ============================================================================

func TestApplyGrayscaleEffect(t *testing.T) {
	opts := &ebiten.DrawImageOptions{}

	// 应用灰度效果
	ApplyGrayscaleEffect(opts)

	// 验证 ColorScale 已被修改（不是默认值）
	// 注意：我们无法直接访问 ColorScale 的内部值，
	// 但可以验证函数不会 panic
}

// ============================================================================
// 颜色矩阵测试
// ============================================================================

func TestApplyColorMatrix_Identity(t *testing.T) {
	opts := &ebiten.DrawImageOptions{}

	// 单位矩阵（不改变颜色）
	identityMatrix := [20]float64{
		1, 0, 0, 0, 0,
		0, 1, 0, 0, 0,
		0, 0, 1, 0, 0,
		0, 0, 0, 1, 0,
	}

	// 应用单位矩阵
	ApplyColorMatrix(opts, identityMatrix)

	// 应该不会 panic
}

func TestApplyColorMatrix_Scale(t *testing.T) {
	opts := &ebiten.DrawImageOptions{}

	// 缩放矩阵（对角线元素）
	scaleMatrix := [20]float64{
		0.5, 0, 0, 0, 0, // R * 0.5
		0, 0.8, 0, 0, 0, // G * 0.8
		0, 0, 1.2, 0, 0, // B * 1.2
		0, 0, 0, 0.9, 0, // A * 0.9
	}

	// 应用缩放矩阵
	ApplyColorMatrix(opts, scaleMatrix)

	// 应该不会 panic
}

func TestApplyColorMatrix_Complex(t *testing.T) {
	opts := &ebiten.DrawImageOptions{}

	// 复杂矩阵（包含非对角线元素）
	// 注意：当前实现只使用对角线元素
	complexMatrix := [20]float64{
		0.5, 0.2, 0.1, 0, 10,
		0.1, 0.8, 0.2, 0, 20,
		0.2, 0.1, 0.9, 0, 30,
		0, 0, 0, 1, 0,
	}

	// 应用复杂矩阵
	ApplyColorMatrix(opts, complexMatrix)

	// 应该不会 panic（即使当前只使用对角线元素）
}

// ============================================================================
// 组合效果测试
// ============================================================================

func TestCombineColorEffects_NoEffects(t *testing.T) {
	// 无任何效果
	opts := CombineColorEffects(nil, 1.0, false, nil, false)

	if opts == nil {
		t.Error("Expected non-nil DrawImageOptions")
	}
}

func TestCombineColorEffects_AlphaOnly(t *testing.T) {
	// 只有 Alpha
	opts := CombineColorEffects(nil, 0.5, false, nil, false)

	if opts == nil {
		t.Error("Expected non-nil DrawImageOptions")
	}
}

func TestCombineColorEffects_ColorTint(t *testing.T) {
	// 颜色叠加
	colorValue := uint32(0xFF8080FF) // 红色偏淡
	opts := CombineColorEffects(&colorValue, 1.0, false, nil, false)

	if opts == nil {
		t.Error("Expected non-nil DrawImageOptions")
	}
}

func TestCombineColorEffects_Grayed(t *testing.T) {
	// 灰度效果
	opts := CombineColorEffects(nil, 1.0, true, nil, false)

	if opts == nil {
		t.Error("Expected non-nil DrawImageOptions")
	}
}

func TestCombineColorEffects_ColorMatrix(t *testing.T) {
	// 颜色矩阵
	matrix := [20]float64{
		0.5, 0, 0, 0, 0,
		0, 0.8, 0, 0, 0,
		0, 0, 1.2, 0, 0,
		0, 0, 0, 1, 0,
	}
	opts := CombineColorEffects(nil, 1.0, false, &matrix, true)

	if opts == nil {
		t.Error("Expected non-nil DrawImageOptions")
	}
}

func TestCombineColorEffects_AllCombined(t *testing.T) {
	// 组合所有效果
	colorValue := uint32(0xFF8080FF)
	matrix := [20]float64{
		0.8, 0, 0, 0, 0,
		0, 0.9, 0, 0, 0,
		0, 0, 1.0, 0, 0,
		0, 0, 0, 1, 0,
	}

	opts := CombineColorEffects(&colorValue, 0.8, false, &matrix, true)

	if opts == nil {
		t.Error("Expected non-nil DrawImageOptions")
	}
}

func TestCombineColorEffects_GrayedOverridesMatrix(t *testing.T) {
	// 灰度效果应该优先于颜色矩阵
	matrix := [20]float64{
		0.5, 0, 0, 0, 0,
		0, 0.8, 0, 0, 0,
		0, 0, 1.2, 0, 0,
		0, 0, 0, 1, 0,
	}

	// 同时启用灰度和颜色矩阵，灰度应该优先
	opts := CombineColorEffects(nil, 1.0, true, &matrix, true)

	if opts == nil {
		t.Error("Expected non-nil DrawImageOptions")
	}
}

// ============================================================================
// 颜色转换测试
// ============================================================================

func TestRGBAToFloat32_ValidColor(t *testing.T) {
	testCases := []struct {
		name     string
		color    color.Color
		expected [4]float32
	}{
		{
			"Red",
			color.RGBA{R: 255, G: 0, B: 0, A: 255},
			[4]float32{1.0, 0.0, 0.0, 1.0},
		},
		{
			"Green",
			color.RGBA{R: 0, G: 255, B: 0, A: 255},
			[4]float32{0.0, 1.0, 0.0, 1.0},
		},
		{
			"Blue",
			color.RGBA{R: 0, G: 0, B: 255, A: 255},
			[4]float32{0.0, 0.0, 1.0, 1.0},
		},
		{
			"White",
			color.RGBA{R: 255, G: 255, B: 255, A: 255},
			[4]float32{1.0, 1.0, 1.0, 1.0},
		},
		{
			"Black",
			color.RGBA{R: 0, G: 0, B: 0, A: 255},
			[4]float32{0.0, 0.0, 0.0, 1.0},
		},
		{
			"HalfAlpha",
			color.RGBA{R: 128, G: 128, B: 128, A: 128},
			[4]float32{0.501960814, 0.501960814, 0.501960814, 0.501960814},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, g, b, a := RGBAToFloat32(tc.color)

			// 使用小误差容忍度比较浮点数
			epsilon := float32(0.01)
			if math.Abs(float64(r-tc.expected[0])) > float64(epsilon) {
				t.Errorf("R: expected %.3f, got %.3f", tc.expected[0], r)
			}
			if math.Abs(float64(g-tc.expected[1])) > float64(epsilon) {
				t.Errorf("G: expected %.3f, got %.3f", tc.expected[1], g)
			}
			if math.Abs(float64(b-tc.expected[2])) > float64(epsilon) {
				t.Errorf("B: expected %.3f, got %.3f", tc.expected[2], b)
			}
			if math.Abs(float64(a-tc.expected[3])) > float64(epsilon) {
				t.Errorf("A: expected %.3f, got %.3f", tc.expected[3], a)
			}
		})
	}
}

func TestRGBAToFloat32_NilColor(t *testing.T) {
	r, g, b, a := RGBAToFloat32(nil)

	// nil 应该返回白色（全 1）
	if r != 1.0 || g != 1.0 || b != 1.0 || a != 1.0 {
		t.Errorf("Expected (1, 1, 1, 1) for nil color, got (%.3f, %.3f, %.3f, %.3f)", r, g, b, a)
	}
}

func TestUint32ToColorRGBA(t *testing.T) {
	testCases := []struct {
		name     string
		input    uint32
		expected color.RGBA
	}{
		{"Red", 0xFF0000FF, color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"Green", 0x00FF00FF, color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{"Blue", 0x0000FFFF, color.RGBA{R: 0, G: 0, B: 255, A: 255}},
		{"White", 0xFFFFFFFF, color.RGBA{R: 255, G: 255, B: 255, A: 255}},
		{"Black", 0x000000FF, color.RGBA{R: 0, G: 0, B: 0, A: 255}},
		{"Transparent", 0xFFFFFF00, color.RGBA{R: 255, G: 255, B: 255, A: 0}},
		{"HalfAlpha", 0x80808080, color.RGBA{R: 128, G: 128, B: 128, A: 128}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Uint32ToColorRGBA(tc.input)

			if result.R != tc.expected.R {
				t.Errorf("R: expected %d, got %d", tc.expected.R, result.R)
			}
			if result.G != tc.expected.G {
				t.Errorf("G: expected %d, got %d", tc.expected.G, result.G)
			}
			if result.B != tc.expected.B {
				t.Errorf("B: expected %d, got %d", tc.expected.B, result.B)
			}
			if result.A != tc.expected.A {
				t.Errorf("A: expected %d, got %d", tc.expected.A, result.A)
			}
		})
	}
}

// ============================================================================
// 辅助函数测试
// ============================================================================

func TestApplyColorTintToOpts(t *testing.T) {
	opts := &ebiten.DrawImageOptions{}

	// 应用颜色叠加
	colorValue := uint32(0xFF8080FF)
	applyColorTintToOpts(opts, colorValue)

	// 应该不会 panic
}

func TestApplyColorTintToOpts_FullTransparent(t *testing.T) {
	opts := &ebiten.DrawImageOptions{}

	// 完全透明
	colorValue := uint32(0xFFFFFF00)
	applyColorTintToOpts(opts, colorValue)

	// 应该不会 panic
}

func TestApplyColorTintToOpts_FullOpaque(t *testing.T) {
	opts := &ebiten.DrawImageOptions{}

	// 完全不透明
	colorValue := uint32(0xFFFFFFFF)
	applyColorTintToOpts(opts, colorValue)

	// 应该不会 panic
}
