package fairygui

import (
	"testing"

	textutil "github.com/chslink/fairygui/internal/text"
)

// TestNewRichTextField 测试创建新的富文本控件
func TestNewRichTextField(t *testing.T) {
	rtf := NewRichTextField()
	if rtf == nil {
		t.Fatal("NewRichTextField() returned nil")
	}

	if rtf.TextField == nil {
		t.Error("RichTextField.TextField is nil")
	}

	// 验证默认值
	if !rtf.UBBEnabled() {
		t.Error("新创建的 RichTextField 应该默认启用 UBB")
	}

	if rtf.BaseStyle().Color != "#000000" {
		t.Errorf("默认颜色不正确: got %s, want #000000", rtf.BaseStyle().Color)
	}

	if rtf.BaseStyle().FontSize != 12 {
		t.Errorf("默认字体大小不正确: got %d, want 12", rtf.BaseStyle().FontSize)
	}

	// 验证自动换行
	if !rtf.WordWrap() {
		t.Error("富文本应该默认启用自动换行")
	}
}

// TestRichTextField_UBBEnabled 测试 UBB 启用/禁用
func TestRichTextField_UBBEnabled(t *testing.T) {
	rtf := NewRichTextField()

	// 默认启用
	if !rtf.UBBEnabled() {
		t.Error("UBB 应该默认启用")
	}

	// 禁用 UBB
	rtf.SetUBBEnabled(false)
	if rtf.UBBEnabled() {
		t.Error("禁用 UBB 失败")
	}

	// 启用 UBB
	rtf.SetUBBEnabled(true)
	if !rtf.UBBEnabled() {
		t.Error("启用 UBB 失败")
	}
}

// TestRichTextField_SetText 测试设置文本
func TestRichTextField_SetText(t *testing.T) {
	rtf := NewRichTextField()

	// 普通文本
	rtf.SetText("Hello World")
	if rtf.Text() != "Hello World" {
		t.Errorf("文本设置失败: got %s", rtf.Text())
	}

	// UBB 文本
	rtf.SetText("[b]Bold[/b]")
	if rtf.Text() != "[b]Bold[/b]" {
		t.Errorf("UBB 文本设置失败: got %s", rtf.Text())
	}

	// 空文本
	rtf.SetText("")
	if rtf.Text() != "" {
		t.Error("设置空文本失败")
	}
}

// TestRichTextField_SetBaseStyle 测试设置基础样式
func TestRichTextField_SetBaseStyle(t *testing.T) {
	rtf := NewRichTextField()

	newStyle := textutil.Style{
		Color:     "#ff0000",
		FontSize:  24,
		Bold:      true,
		Italic:    true,
		Underline: true,
		Font:      "SimHei",
	}

	rtf.SetBaseStyle(newStyle)

	if rtf.BaseStyle().Color != "#ff0000" {
		t.Errorf("颜色设置失败: got %s", rtf.BaseStyle().Color)
	}

	if rtf.BaseStyle().FontSize != 24 {
		t.Errorf("字体大小设置失败: got %d", rtf.BaseStyle().FontSize)
	}

	if !rtf.BaseStyle().Bold {
		t.Error("粗体设置失败")
	}

	if !rtf.BaseStyle().Italic {
		t.Error("斜体设置失败")
	}

	if !rtf.BaseStyle().Underline {
		t.Error("下划线设置失败")
	}

	if rtf.BaseStyle().Font != "SimHei" {
		t.Errorf("字体设置失败: got %s", rtf.BaseStyle().Font)
	}
}

// TestRichTextField_Segments 测试富文本段解析
func TestRichTextField_Segments(t *testing.T) {
	rtf := NewRichTextField()

	// 普通文本应该只有一个段
	rtf.SetText("Hello")
	segments := rtf.GetSegments()
	if segments == nil {
		t.Fatal("普通文本应该有段")
	}

	if len(segments) != 1 {
		t.Errorf("普通文本段数量错误: got %d, want 1", len(segments))
	}

	// 复杂 UBB 文本
	rtf.SetText("[b][color=#ff0000]Bold Red[/color][/b]")
	segments = rtf.GetSegments()
	if segments == nil {
		t.Fatal("UBB 文本应该有段")
	}

	if len(segments) != 1 {
		t.Errorf("UBB 文本段数量错误: got %d, want 1", len(segments))
	}

	// 验证样式
	seg := segments[0]
	if seg.Text != "Bold Red" {
		t.Errorf("段文本错误: got %s, want 'Bold Red'", seg.Text)
	}

	if seg.Style.Bold != true {
		t.Error("段应该为粗体")
	}

	if seg.Style.Color != "#ff0000" {
		t.Errorf("段颜色错误: got %s, want #ff0000", seg.Style.Color)
	}
}

// TestRichTextField_NestedTags 测试嵌套标签
func TestRichTextField_NestedTags(t *testing.T) {
	rtf := NewRichTextField()

	// 嵌套样式
	rtf.SetText("[size=24][font=SimSun]Large Text[/font][/size]")

	segments := rtf.GetSegments()
	if len(segments) != 1 {
		t.Fatalf("段数量错误: got %d, want 1", len(segments))
	}

	seg := segments[0]
	if seg.Text != "Large Text" {
		t.Errorf("段文本错误: got %s", seg.Text)
	}

	if seg.Style.FontSize != 24 {
		t.Errorf("字体大小错误: got %d, want 24", seg.Style.FontSize)
	}

	if seg.Style.Font != "SimSun" {
		t.Errorf("字体错误: got %s, want SimSun", seg.Style.Font)
	}
}

// TestRichTextField_Link 测试链接
func TestRichTextField_Link(t *testing.T) {
	rtf := NewRichTextField()

	// 链接文本
	rtf.SetText("[url=event:click]Click Me[/url]")

	segments := rtf.GetSegments()
	if len(segments) != 1 {
		t.Fatalf("段数量错误: got %d, want 1", len(segments))
	}

	seg := segments[0]
	if seg.Text != "Click Me" {
		t.Errorf("段文本错误: got %s", seg.Text)
	}

	if seg.Link != "event:click" {
		t.Errorf("链接错误: got %s, want 'event:click'", seg.Link)
	}

	if !seg.Style.Underline {
		t.Error("链接应该自动下划线")
	}
}

// TestRichTextField_Image 测试图片
func TestRichTextField_Image(t *testing.T) {
	rtf := NewRichTextField()

	// 内嵌图片
	rtf.SetText("Text [img]ui://package/icon[/img] Image")

	segments := rtf.GetSegments()
	// 应该有 3 段: "Text ", 图片占位符, " Image"
	if len(segments) != 3 {
		t.Fatalf("段数量错误: got %d, want 3", len(segments))
	}

	// 第一段
	if segments[0].Text != "Text " {
		t.Errorf("第一段文本错误: got %s", segments[0].Text)
	}

	// 第二段（图片）
	if segments[1].ImageURL != "ui://package/icon" {
		t.Errorf("图片URL错误: got %s", segments[1].ImageURL)
	}

	// 第三段
	if segments[2].Text != " Image" {
		t.Errorf("第三段文本错误: got %s", segments[2].Text)
	}
}

// TestRichTextField_MultipleSegments 测试多个段
func TestRichTextField_MultipleSegments(t *testing.T) {
	rtf := NewRichTextField()

	// 多个段
	rtf.SetText("[b]Bold[/b][i]Italic[/i][color=#ff0000]Red[/color]")

	segments := rtf.GetSegments()
	if len(segments) != 3 {
		t.Fatalf("段数量错误: got %d, want 3", len(segments))
	}

	// 验证每个段
	if !segments[0].Style.Bold {
		t.Error("第一段应该是粗体")
	}

	if !segments[1].Style.Italic {
		t.Error("第二段应该是斜体")
	}

	if segments[2].Style.Color != "#ff0000" {
		t.Error("第三段应该是红色")
	}

	// 测试带空格的版本，可能会产生更多的段
	rtf.SetText("Normal [b]Bold[/b] [i]Italic[/i] [color=#ff0000]Red[/color]")
	segments = rtf.GetSegments()
	// 至少应该有 4 段（带样式），空行也可能被分成段
	if len(segments) < 4 {
		t.Errorf("段数量过少: got %d, want at least 4", len(segments))
	}
}

// TestRichTextField_UNBDisabled 测试禁用 UBB
func TestRichTextField_UBBDisabled(t *testing.T) {
	rtf := NewRichTextField()

	// 禁用 UBB
	rtf.SetUBBEnabled(false)
	rtf.SetText("[b]Not Bold[/b]")

	segments := rtf.GetSegments()
	if segments != nil {
		t.Error("禁用 UBB 后应该返回 nil segments")
	}

	// 文本应该保持原样
	if rtf.Text() != "[b]Not Bold[/b]" {
		t.Errorf("禁用 UBB 后文本应该保持不变: got %s", rtf.Text())
	}
}

// TestRichTextField_HasRichContent 测试富文本内容检测
func TestRichTextField_HasRichContent(t *testing.T) {
	rtf := NewRichTextField()

	// 普通文本
	rtf.SetText("Plain Text")
	if rtf.HasRichContent() {
		t.Error("普通文本不应该有富文本内容")
	}

	// 有样式的文本
	rtf.SetText("[b]Bold[/b]")
	if !rtf.HasRichContent() {
		t.Error("UBB 文本应该有富文本内容")
	}

	// 禁用 UBB
	rtf.SetUBBEnabled(false)
	if rtf.HasRichContent() {
		t.Error("禁用 UBB 后不应该有富文本内容")
	}
}

// TestRichTextField_GetPlainText 测试获取纯文本
func TestRichTextField_GetPlainText(t *testing.T) {
	rtf := NewRichTextField()

	// 带 UBB 的文本
	rtf.SetText("Hello [b]Bold[/b] [color=#ff0000]Red[/color]")

	plainText := rtf.GetPlainText()
	expected := "Hello Bold Red"
	if plainText != expected {
		t.Errorf("纯文本提取错误: got '%s', want '%s'", plainText, expected)
	}
}

// TestRichTextField_StyleInheritance 测试样式继承
func TestRichTextField_StyleInheritance(t *testing.T) {
	rtf := NewRichTextField()

	// 设置基础样式
	rtf.SetBaseStyle(textutil.Style{
		Color:    "#00ff00",
		FontSize: 20,
		Font:     "SimSun",
	})

	rtf.SetText("[b]Bold Text[/b]")

	segments := rtf.GetSegments()
	if len(segments) != 1 {
		t.Fatalf("段数量错误: got %d", len(segments))
	}

	seg := segments[0]
	// 应该继承基础样式的颜色和大小
	if seg.Style.Color != "#00ff00" {
		t.Errorf("颜色未继承: got %s", seg.Style.Color)
	}

	if seg.Style.FontSize != 20 {
		t.Errorf("字体大小未继承: got %d", seg.Style.FontSize)
	}

	if seg.Style.Font != "SimSun" {
		t.Errorf("字体未继承: got %s", seg.Style.Font)
	}

	// 但应该是粗体
	if !seg.Style.Bold {
		t.Error("粗体未应用")
	}
}

// TestRichTextField_SetWithColorFont 测试设置颜色、字体和大小
func TestRichTextField_SetWithColorFont(t *testing.T) {
	rtf := NewRichTextField()

	// 设置颜色（同时更新基础样式）
	rtf.SetColor("#ff0000")
	if rtf.BaseStyle().Color != "#ff0000" {
		t.Error("设置颜色未更新基础样式")
	}

	// 设置字体大小
	rtf.SetFontSize(24)
	if rtf.BaseStyle().FontSize != 24 {
		t.Error("设置字体大小未更新基础样式")
	}

	// 设置字体
	rtf.SetFont("SimHei")
	if rtf.BaseStyle().Font != "SimHei" {
		t.Error("设置字体未更新基础样式")
	}
}

// TestRichTextField_EmptyText 测试空文本
func TestRichTextField_EmptyText(t *testing.T) {
	rtf := NewRichTextField()

	// 空文本
	rtf.SetText("")

	segments := rtf.GetSegments()
	if segments != nil {
		t.Error("空文本应该返回 nil segments")
	}
}

// TestAssertRichTextField 测试类型断言
func TestAssertRichTextField(t *testing.T) {
	rtf := NewRichTextField()

	result, ok := AssertRichTextField(rtf)
	if !ok {
		t.Error("AssertRichTextField 应该成功")
	}
	if result != rtf {
		t.Error("AssertRichTextField 返回的对象不正确")
	}

	if !IsRichTextField(rtf) {
		t.Error("IsRichTextField 应该返回 true")
	}

	obj := NewObject()
	_, ok = AssertRichTextField(obj)
	if ok {
		t.Error("AssertRichTextField 对非 RichTextField 对象应该失败")
	}

	if IsRichTextField(obj) {
		t.Error("IsRichTextField 对非 RichTextField 对象应该返回 false")
	}
}

// TestRichTextField_Chaining 测试链式调用
func TestRichTextField_Chaining(t *testing.T) {
	rtf := NewRichTextField()

	// 连续设置多个属性
	rtf.SetText("Chained Text")
	rtf.SetColor("#ff00ff")
	rtf.SetFontSize(18)

	// 验证所有设置都生效
	if rtf.Text() != "Chained Text" {
		t.Error("文本设置失败")
	}

	if rtf.BaseStyle().Color != "#ff00ff" {
		t.Error("颜色设置失败")
	}

	if rtf.BaseStyle().FontSize != 18 {
		t.Error("字体大小设置失败")
	}
}

// TestRichTextField_LineBreak 测试换行
func TestRichTextField_LineBreak(t *testing.T) {
	rtf := NewRichTextField()

	rtf.SetText("Line1[br]Line2[br]Line3")

	segments := rtf.GetSegments()
	if len(segments) != 5 {
		t.Fatalf("段数量错误: got %d, want 5 (3 行文本 + 2 个换行符)", len(segments))
	}

	// 验证换行符段
	if segments[1].Text != "\n" {
		t.Errorf("换行符错误: got %q, want %q", segments[1].Text, "\n")
	}

	if segments[3].Text != "\n" {
		t.Errorf("换行符错误: got %q, want %q", segments[3].Text, "\n")
	}

	// 验证文本
	if segments[0].Text != "Line1" {
		t.Errorf("第一行错误: got %s", segments[0].Text)
	}

	if segments[2].Text != "Line2" {
		t.Errorf("第二行错误: got %s", segments[2].Text)
	}

	if segments[4].Text != "Line3" {
		t.Errorf("第三行错误: got %s", segments[4].Text)
	}
}
