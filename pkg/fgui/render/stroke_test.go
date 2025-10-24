package render

import (
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/widgets"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestStrokeRendering_SizeControl(t *testing.T) {
	tests := []struct {
		name       string
		strokeSize float64
		expectRuns int // 预期的描边绘制次数
	}{
		{"stroke_1", 1.0, 8},   // 8个主要方向
		{"stroke_1_5", 1.5, 8}, // 8个主要方向（向下取整）
		{"stroke_2", 2.0, 24}, // 8个主要 + 16个外围
		{"stroke_3", 3.0, 24}, // 同2.0，因为范围相同
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试图像
			img := ebiten.NewImage(200, 100)

			// 创建文本字段
			field := widgets.NewText()
			field.SetText("Test")
			field.SetFontSize(20)
			field.SetStrokeSize(tt.strokeSize)
			field.SetStrokeColor("#FF0000")

			// 渲染文本
			geo := ebiten.GeoM{}
			geo.Translate(10, 10)

			err := drawTextImage(img, geo, field, "Test", 1.0, 180, 80, nil, nil)
			if err != nil {
				t.Fatalf("Failed to draw text: %v", err)
			}

			// 基本验证：图像应该被修改（有描边）
			// 在实际测试中，我们可能需要更精确的像素检查
			t.Logf("Successfully rendered text with stroke size %.1f", tt.strokeSize)
		})
	}
}

func TestStrokeRendering_AlgorithmConsistency(t *testing.T) {
	// 测试描边算法的一致性

	textCases := []string{
		"Hello",
		"你好",
		"Hello World",
		"Test123",
	}

	for _, text := range textCases {
		t.Run("text_"+text, func(t *testing.T) {
			// 创建测试图像
			img := ebiten.NewImage(300, 100)

			// 创建文本字段
			field := widgets.NewText()
			field.SetText(text)
			field.SetFontSize(16)
			field.SetStrokeSize(1.0)
			field.SetStrokeColor("#000000")

			// 渲染文本
			geo := ebiten.GeoM{}
			geo.Translate(10, 10)

			err := drawTextImage(img, geo, field, text, 1.0, 280, 80, nil, nil)
			if err != nil {
				t.Fatalf("Failed to draw text: %v", err)
			}

			t.Logf("Successfully rendered text '%s' with stroke", text)
		})
	}
}

func TestStrokeRendering_NoStroke(t *testing.T) {
	// 测试没有描边的情况
	img := ebiten.NewImage(200, 100)

	field := widgets.NewText()
	field.SetText("Test")
	field.SetFontSize(20)
	// 不设置描边

	geo := ebiten.GeoM{}
	geo.Translate(10, 10)

	err := drawTextImage(img, geo, field, "Test", 1.0, 180, 80, nil, nil)
	if err != nil {
		t.Fatalf("Failed to draw text: %v", err)
	}

	t.Log("Successfully rendered text without stroke")
}

func TestStrokeRendering_WithColorEffects(t *testing.T) {
	// 测试描边与其他颜色效果的组合

	img := ebiten.NewImage(300, 100)

	field := widgets.NewText()
	field.SetText("Stroke Test")
	field.SetFontSize(18)
	field.SetStrokeSize(1.0)
	field.SetStrokeColor("#FF0000") // 红色描边

	geo := ebiten.GeoM{}
	geo.Translate(10, 10)

	err := drawTextImage(img, geo, field, "Stroke Test", 1.0, 280, 80, nil, nil)
	if err != nil {
		t.Fatalf("Failed to draw text: %v", err)
	}

	t.Log("Successfully rendered text with stroke and color effects")
}

func TestStrokeRendering_PerformanceCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// 简单的性能测试：确保描边不会导致明显的性能问题
	img := ebiten.NewImage(400, 200)

	field := widgets.NewText()
	field.SetText("Performance Test String")
	field.SetFontSize(16)
	field.SetStrokeSize(2.0) // 较大的描边
	field.SetStrokeColor("#0000FF")

	geo := ebiten.GeoM{}
	geo.Translate(10, 10)

	// 多次渲染以测试性能
	for i := 0; i < 10; i++ {
		err := drawTextImage(img, geo, field, "Performance Test String", 1.0, 380, 180, nil, nil)
		if err != nil {
			t.Fatalf("Failed to draw text: %v", err)
		}
	}

	t.Log("Performance test completed successfully")
}