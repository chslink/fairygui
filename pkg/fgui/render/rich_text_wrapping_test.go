package render

import (
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/widgets"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestRichTextWrapping_DefaultBehavior(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		ubbEnabled  bool
		autoSize    widgets.TextAutoSize
		singleLine  bool
		expectWrap  bool
		description string
	}{
		{
			name:        "rich_text_default_autosize",
			text:        "[color=#FF0000]这是一个[/color][b]富文本[/b]内容，用于测试默认的换行行为。",
			ubbEnabled:  true,
			autoSize:    widgets.TextAutoSizeBoth,
			singleLine:  false,
			expectWrap:  false,
			description: "与 Laya 一致：AutoSizeBoth + 非单行不换行",
		},
		{
			name:        "rich_text_autosize_singleline",
			text:        "[color=#FF0000]富文本[/color][b]单行模式[/b]不应该换行",
			ubbEnabled:  true,
			autoSize:    widgets.TextAutoSizeBoth,
			singleLine:  true,
			expectWrap:  false,
			description: "富文本 + 单行模式不应该换行",
		},
		{
			name:        "rich_text_fixed_width",
			text:        "[color=#FF0000]富文本[/color]在固定宽度下应该换行，这是[b]正常的[/b]行为。",
			ubbEnabled:  true,
			autoSize:    widgets.TextAutoSizeHeight,
			singleLine:  false,
			expectWrap:  true,
			description: "富文本固定宽度才换行",
		},
		{
			name:        "plain_text_default_autosize",
			text:        "这是一个普通文本，用于对比富文本的换行行为。",
			ubbEnabled:  false,
			autoSize:    widgets.TextAutoSizeBoth,
			singleLine:  false,
			expectWrap:  false,
			description: "普通文本默认不应该换行（对比测试）",
		},
		{
			name:        "plain_text_fixed_width",
			text:        "普通文本在固定宽度下应该换行。",
			ubbEnabled:  false,
			autoSize:    widgets.TextAutoSizeHeight,
			singleLine:  false,
			expectWrap:  true,
			description: "普通文本固定宽度应该换行",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建文本字段
			field := widgets.NewText()
			field.SetText(tt.text)
			field.SetUBBEnabled(tt.ubbEnabled)
			field.SetAutoSize(tt.autoSize)
			field.SetSingleLine(tt.singleLine)

			// 创建测试图像
			img := ebiten.NewImage(300, 100)

			// 渲染文本
			geo := ebiten.GeoM{}
			geo.Translate(10, 10)

			err := drawTextImage(img, geo, field, field.Text(), 1.0, 280, 80, nil, nil)
			if err != nil {
				t.Fatalf("Failed to draw text: %v", err)
			}

			allowWrap := !field.WidthAutoSize() && !field.SingleLine()
			if allowWrap != tt.expectWrap {
				t.Errorf("WordWrap 计算不符：widthAuto=%v singleLine=%v => %v, 期望 %v",
					field.WidthAutoSize(), field.SingleLine(), allowWrap, tt.expectWrap)
			}

			t.Logf("测试: %s", tt.description)
			t.Logf("  UBB: %v, AutoSize: %v, SingleLine: %v", tt.ubbEnabled, tt.autoSize, tt.singleLine)
			t.Logf("  宽度自适应: %v, allowWrap: %v", field.WidthAutoSize(), allowWrap)
			t.Logf("  文本内容: %s", tt.text)

			// 基本验证：确保渲染成功
			// 在实际 GUI 环境中，可以通过检查像素来验证换行行为
		})
	}
}

func TestRichTextWrapping_UbbFeatures(t *testing.T) {
	// 测试富文本的各种 UBB 功能在换行时的表现
	ubbTests := []struct {
		name       string
		ubbText    string
		expectWrap bool
	}{
		{
			name:       "color_and_bold",
			ubbText:    "[color=#FF0000]红色文字[/color]和[b]粗体文字[/color]的混合内容",
			expectWrap: true,
		},
		{
			name:       "underline_and_italic",
			ubbText:    "[u]下划线[/u]和[i]斜体[/i]文字的换行测试",
			expectWrap: true,
		},
		{
			name:       "mixed_formatting",
			ubbText:    "[color=#FF0000][b][u]红粗下划线[/u][/b][/color]复杂格式文本换行测试",
			expectWrap: true,
		},
		{
			name:       "link_simulation",
			ubbText:    "[a=link://test]这是链接文字[/a]后面还有普通文字继续换行测试",
			expectWrap: true,
		},
	}

	for _, tt := range ubbTests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建富文本字段
			field := widgets.NewText()
			field.SetText(tt.ubbText)
			field.SetUBBEnabled(true) // 启用 UBB
			field.SetAutoSize(widgets.TextAutoSizeHeight)
			field.SetSingleLine(false)

			// 创建测试图像
			img := ebiten.NewImage(300, 100)

			// 渲染文本
			geo := ebiten.GeoM{}
			geo.Translate(10, 10)

			err := drawTextImage(img, geo, field, field.Text(), 1.0, 280, 80, nil, nil)
			if err != nil {
				t.Fatalf("Failed to draw rich text: %v", err)
			}

			allowWrap := !field.WidthAutoSize() && !field.SingleLine()
			if allowWrap != tt.expectWrap {
				t.Errorf("UBB 测试 %s: allowWrap=%v, 期望=%v", tt.name, allowWrap, tt.expectWrap)
			}

			t.Logf("富文本功能测试: %s", tt.name)
			t.Logf("  UBB 内容: %s", tt.ubbText)
			t.Logf("  宽度自适应: %v, allowWrap: %v", field.WidthAutoSize(), allowWrap)
		})
	}
}

func TestRichTextWrapping_CompatibilityWithLayaAir(t *testing.T) {
	// 测试与 LayaAir 富文本行为的兼容性

	// LayaAir 中 GRichTextField 设置：
	// - html = true（等价于启用 UBB 解析）
	// - 换行逻辑完全继承 GTextField（wordWrap = !widthAuto && !singleLine）

	// 测试默认富文本行为
	richField := widgets.NewText()
	richField.SetText("[color=#FF0000]这是富文本[/color]内容，验证默认换行逻辑。")
	richField.SetUBBEnabled(true) // 相当于 html = true

	// 验证默认设置
	if richField.UBBEnabled() != true {
		t.Errorf("富文本应该启用 UBB 支持")
	}

	if richField.WidthAutoSize() != true {
		t.Errorf("富文本默认 WidthAutoSize 应该为 true")
	}
	if richField.SingleLine() != false {
		t.Errorf("富文本默认 SingleLine 应该为 false")
	}
	allowWrap := !richField.WidthAutoSize() && !richField.SingleLine()
	if allowWrap != false {
		t.Errorf("富文本默认不应换行，wordWrap=%v", allowWrap)
	}

	t.Logf("LayaAir 兼容性测试通过:")
	t.Logf("  UBB 启用: %v", richField.UBBEnabled())
	t.Logf("  WidthAutoSize: %v", richField.WidthAutoSize())
	t.Logf("  SingleLine: %v", richField.SingleLine())
	t.Logf("  WordWrap: !%v && !%v = %v", richField.WidthAutoSize(), richField.SingleLine(), allowWrap)
}
