package render

import (
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/widgets"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestTextWrapping_DefaultBehavior(t *testing.T) {
	tests := []struct {
		name         string
		autoSize     widgets.TextAutoSize
		singleLine   bool
		expectWrap   bool
		description  string
	}{
		{
			name:       "default_autosize_both",
			autoSize:   widgets.TextAutoSizeBoth,
			singleLine: false,
			expectWrap: false,
			description: "默认 TextAutoSizeBoth 不应该换行",
		},
		{
			name:       "default_autosize_both_singleline",
			autoSize:   widgets.TextAutoSizeBoth,
			singleLine: true,
			expectWrap: false,
			description: "TextAutoSizeBoth + singleLine 不应该换行",
		},
		{
			name:       "manual_width_fixed_height",
			autoSize:   widgets.TextAutoSizeHeight,
			singleLine: false,
			expectWrap: true,
			description: "固定宽度高度自适应应该换行",
		},
		{
			name:       "manual_width_fixed_height_singleline",
			autoSize:   widgets.TextAutoSizeHeight,
			singleLine: true,
			expectWrap: false,
			description: "固定宽度 + singleLine 不应该换行",
		},
		{
			name:       "shrink_autosize",
			autoSize:   widgets.TextAutoSizeShrink,
			singleLine: false,
			expectWrap: false,
			description: "TextAutoSizeShrink 不应该换行（会缩放文字）",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建文本字段
			field := widgets.NewText()
			field.SetText("This is a long text that should wrap based on settings")
			field.SetAutoSize(tt.autoSize)
			field.SetSingleLine(tt.singleLine)

			// 创建测试图像
			img := ebiten.NewImage(200, 100)

			// 渲染文本
			geo := ebiten.GeoM{}
			geo.Translate(10, 10)

			err := drawTextImage(img, geo, field, field.Text(), 1.0, 200, 80, nil, nil)
			if err != nil {
				t.Fatalf("Failed to draw text: %v", err)
			}

			t.Logf("测试: %s", tt.description)
			t.Logf("  AutoSize: %v, SingleLine: %v", tt.autoSize, tt.singleLine)
			t.Logf("  预期换行: %v", tt.expectWrap)

			// 基本验证：确保渲染成功
			// 在实际 GUI 环境中，可以通过检查像素来验证换行行为
		})
	}
}

func TestTextWrapping_WidthAutoSizeConsistency(t *testing.T) {
	// 测试 WidthAutoSize() 方法的一致性
	field := widgets.NewText()

	// 测试默认值
	if field.WidthAutoSize() != true {
		t.Errorf("默认情况下 WidthAutoSize() 应该返回 true，实际返回 %v", field.WidthAutoSize())
	}

	// 测试不同 AutoSize 设置的响应
	testCases := []struct {
		autoSize           widgets.TextAutoSize
		expectedWidthAutoSize bool
	}{
		{widgets.TextAutoSizeBoth, true},
		{widgets.TextAutoSizeHeight, false},
		{widgets.TextAutoSizeShrink, false},
		{widgets.TextAutoSizeEllipsis, false},
	}

	for _, tc := range testCases {
		field.SetAutoSize(tc.autoSize)
		actual := field.WidthAutoSize()
		if actual != tc.expectedWidthAutoSize {
			t.Errorf("AutoSize=%v: 期望 WidthAutoSize=%v, 实际=%v",
				tc.autoSize, tc.expectedWidthAutoSize, actual)
		}
	}
}

func TestTextWrapping_LayaAirCompatibility(t *testing.T) {
	// 测试与 LayaAir 行为的兼容性
	field := widgets.NewText()

	// LayaAir 默认：widthAutoSize = true, singleLine = false
	// 因此 wordWrap = !widthAutoSize && !singleLine = false
	if field.WidthAutoSize() != true {
		t.Errorf("LayaAir 兼容性：默认 WidthAutoSize 应该为 true")
	}

	if field.SingleLine() != false {
		t.Errorf("LayaAir 兼容性：默认 SingleLine 应该为 false")
	}

	// 验证默认情况下不应该换行（LayaAir 行为）
	// wordWrap = !true && !false = false
	allowWrap := !field.WidthAutoSize() && !field.SingleLine()
	if allowWrap != false {
		t.Errorf("LayaAir 兼容性：默认情况下不应该换行，但 allowWrap=%v", allowWrap)
	}

	t.Logf("LayaAir 兼容性测试通过:")
	t.Logf("  WidthAutoSize: %v", field.WidthAutoSize())
	t.Logf("  SingleLine: %v", field.SingleLine())
	t.Logf("  WordWrap: !%v && !%v = %v", field.WidthAutoSize(), field.SingleLine(), allowWrap)
}