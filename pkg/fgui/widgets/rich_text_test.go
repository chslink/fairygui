package widgets

import (
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

func TestGRichTextField_DefaultBehavior(t *testing.T) {
	richText := NewRichText()

	// 测试默认值
	if richText.HtmlEnabled() != true {
		t.Errorf("GRichTextField 应该默认启用 HTML 模式")
	}

	if richText.UBBEnabled() != true {
		t.Errorf("GRichTextField 应该默认启用 UBB 支持")
	}

	// 测试继承的 GTextField 功能
	if richText.AutoSize() != TextAutoSizeBoth {
		t.Errorf("GRichTextField 应该默认使用 TextAutoSizeBoth")
	}

	if richText.SingleLine() != false {
		t.Errorf("GRichTextField 应该默认不是单行模式")
	}

	t.Logf("GRichTextField 默认行为验证通过:")
	t.Logf("  HTML 启用: %v", richText.HtmlEnabled())
	t.Logf("  UBB 启用: %v", richText.UBBEnabled())
	t.Logf("  AutoSize: %v", richText.AutoSize())
	t.Logf("  SingleLine: %v", richText.SingleLine())
}

func TestGRichTextField_WrappingLogic(t *testing.T) {
	richText := NewRichText()

	// 默认设置与 Laya 一致：AutoSizeBoth => widthAutoSize=true -> 不换行
	if richText.WidthAutoSize() != true {
		t.Errorf("默认情况下 WidthAutoSize 应该为 true")
	}
	if richText.SingleLine() != false {
		t.Errorf("默认情况下 SingleLine 应该为 false")
	}
	allowWrap := !richText.WidthAutoSize() && !richText.SingleLine()
	if allowWrap {
		t.Errorf("富文本默认不应该换行，计算得到 allowWrap=%v", allowWrap)
	}

	// 设置固定宽度（AutoSizeHeight）后才允许换行
	richText.SetAutoSize(TextAutoSizeHeight)
	allowWrapFixed := !richText.WidthAutoSize() && !richText.SingleLine()
	if !allowWrapFixed {
		t.Errorf("AutoSizeHeight 时应该允许换行，allowWrap=%v", allowWrapFixed)
	}

	// 单行模式永远不换行
	richText.SetSingleLine(true)
	allowWrapSingleLine := !richText.WidthAutoSize() && !richText.SingleLine()
	if allowWrapSingleLine {
		t.Errorf("富文本在单行模式下不应该换行，allowWrap=%v", allowWrapSingleLine)
	}

	// 恢复默认配置方便日志
	richText.SetSingleLine(false)
	richText.SetAutoSize(TextAutoSizeBoth)
	t.Logf("GRichTextField 换行逻辑验证通过:")
	t.Logf("  默认: WidthAutoSize=%v, SingleLine=%v, allowWrap=%v", richText.WidthAutoSize(), richText.SingleLine(), !richText.WidthAutoSize() && !richText.SingleLine())
}

func TestGRichTextField_TypeIdentification(t *testing.T) {
	richText := NewRichText()
	plainText := NewText()

	// 测试类型识别 - 通过接口转换
	var richInterface interface{} = richText
	_, isRich := richInterface.(*GRichTextField)
	if !isRich {
		t.Errorf("richText 应该被识别为 *GRichTextField 类型")
	}

	var plainInterface interface{} = plainText
	_, isRichFromPlain := plainInterface.(*GRichTextField)
	if isRichFromPlain {
		t.Errorf("plainText 不应该被识别为 *GRichTextField 类型")
	}

	t.Logf("GRichTextField 类型识别验证通过:")
	t.Logf("  richText is GRichTextField: %v", isRich)
	t.Logf("  plainText is GRichTextField: %v", isRichFromPlain)
}

func TestGRichTextField_FactoryCreation(t *testing.T) {
	// 测试工厂创建富文本
	metaRich := &assets.ComponentChild{
		Type: assets.ObjectTypeRichText,
	}

	richWidget := CreateWidget(metaRich)
	if richWidget == nil {
		t.Fatal("工厂应该能够创建富文本组件")
	}

	richText, ok := richWidget.(*GRichTextField)
	if !ok {
		t.Errorf("工厂应该创建 GRichTextField，实际创建了 %T", richWidget)
	}

	if richText.HtmlEnabled() != true {
		t.Errorf("工厂创建的 GRichTextField 应该默认启用 HTML")
	}

	// 测试工厂创建普通文本
	metaPlain := &assets.ComponentChild{
		Type: assets.ObjectTypeText,
	}

	plainWidget := CreateWidget(metaPlain)
	if plainWidget == nil {
		t.Fatal("工厂应该能够创建普通文本组件")
	}

	_, plainOk := plainWidget.(*GTextField)
	if !plainOk {
		t.Errorf("工厂应该创建 GTextField，实际创建了 %T", plainWidget)
	}

	t.Logf("工厂创建验证通过:")
	t.Logf("  富文本类型: %T", richWidget)
	t.Logf("  普通文本类型: %T", plainWidget)
}

func TestGRichTextField_Inheritance(t *testing.T) {
	richText := NewRichText()

	// 测试继承的 GTextField 方法
	testText := "[color=#FF0000]测试[/color]文本"
	richText.SetText(testText)

	if richText.Text() != testText {
		t.Errorf("继承的 Text() 方法应该正常工作")
	}

	// 测试 AutoSize 功能
	richText.SetAutoSize(TextAutoSizeHeight)
	if richText.AutoSize() != TextAutoSizeHeight {
		t.Errorf("继承的 AutoSize 方法应该正常工作")
	}

	// 测试 WidthAutoSize 方法
	richText.SetAutoSize(TextAutoSizeBoth)
	if richText.WidthAutoSize() != true {
		t.Errorf("继承的 WidthAutoSize 方法应该正常工作")
	}

	t.Logf("GRichTextField 继承功能验证通过:")
	t.Logf("  Text: %s", richText.Text())
	t.Logf("  AutoSize: %v", richText.AutoSize())
	t.Logf("  WidthAutoSize: %v", richText.WidthAutoSize())
}

func TestGRichTextField_UbbHtmlConsistency(t *testing.T) {
	richText := NewRichText()

	// 测试 HTML 和 UBB 的一致性
	richText.SetHtmlEnabled(true)
	if richText.UBBEnabled() != true {
		t.Errorf("HTML 启用时，UBB 也应该启用")
	}

	richText.SetUBBEnabled(true)
	if richText.HtmlEnabled() != true {
		t.Errorf("UBB 启用时，HTML 也应该启用")
	}

	// 测试禁用
	richText.SetHtmlEnabled(false)
	if richText.UBBEnabled() != false {
		t.Errorf("HTML 禁用时，UBB 也应该禁用")
	}

	richText.SetUBBEnabled(false)
	if richText.HtmlEnabled() != false {
		t.Errorf("UBB 禁用时，HTML 也应该禁用")
	}

	t.Logf("GRichTextField HTML/UBB 一致性验证通过")
}
