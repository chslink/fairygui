package fairygui

import (
	"testing"
)

// TestNewLabel 测试创建新的标签
func TestNewLabel(t *testing.T) {
	label := NewLabel()
	if label == nil {
		t.Fatal("NewLabel() returned nil")
	}

	if label.ComponentImpl == nil {
		t.Error("Label.ComponentImpl is nil")
	}

	// 验证默认值
	if label.TitleColor() != "#ffffff" {
		t.Errorf("默认标题颜色不正确: got %s, want #ffffff", label.TitleColor())
	}

	if label.TitleFontSize() != 12 {
		t.Errorf("默认字体大小不正确: got %d, want 12", label.TitleFontSize())
	}

	if label.Title() != "" {
		t.Errorf("默认标题文本应该为空: got %s", label.Title())
	}
}

// TestLabel_SetTitle 测试设置标题文本
func TestLabel_SetTitle(t *testing.T) {
	label := NewLabel()

	label.SetTitle("Hello World")
	if label.Title() != "Hello World" {
		t.Errorf("标题文本设置失败: got %s, want Hello World", label.Title())
	}

	// 测试空文本
	label.SetTitle("")
	if label.Title() != "" {
		t.Error("设置空文本失败")
	}
}

// TestLabel_TitleFormatting 测试标题格式化
func TestLabel_TitleFormatting(t *testing.T) {
	label := NewLabel()

	// 测试设置颜色
	label.SetTitleColor("#ff0000")
	if label.TitleColor() != "#ff0000" {
		t.Errorf("标题颜色设置失败: got %s", label.TitleColor())
	}

	// 测试设置描边颜色
	label.SetTitleOutlineColor("#000000")
	if label.TitleOutlineColor() != "#000000" {
		t.Errorf("描边颜色设置失败: got %s", label.TitleOutlineColor())
	}

	// 测试设置字体大小
	label.SetTitleFontSize(24)
	if label.TitleFontSize() != 24 {
		t.Errorf("字体大小设置失败: got %d", label.TitleFontSize())
	}
}

// TestLabel_Icon 测试图标设置
func TestLabel_Icon(t *testing.T) {
	label := NewLabel()

	label.SetIcon("icon.png")
	if label.Icon() != "icon.png" {
		t.Errorf("图标设置失败: got %s", label.Icon())
	}

	// 测试空图标
	label.SetIcon("")
	if label.Icon() != "" {
		t.Error("设置空图标失败")
	}
}

// TestLabel_IconItem 测试图标项设置
func TestLabel_IconItem(t *testing.T) {
	label := NewLabel()

	// 测试设置 nil - 应该接受nil而不panic
	// 注意：IconItem()返回interface，即使内部是nil，也不直接等于nil
	label.SetIconItem(nil)

	// 创建虚拟资源项
	item := &PackageItemWrapper{}
	label.SetIconItem(item)

	if label.IconItem() != item {
		t.Error("图标项设置失败")
	}
}

// TestLabel_PackageItem 测试资源项操作
func TestLabel_PackageItem(t *testing.T) {
	label := NewLabel()

	// 测试设置 nil
	label.SetPackageItem(nil)
	// 应该接受 nil 而不 panic

	// 创建虚拟资源项
	item := &PackageItemWrapper{}
	label.SetPackageItem(item)

	if label.PackageItem() != item {
		t.Error("资源项设置失败")
	}
}

// TestLabel_Resource 测试资源设置
func TestLabel_Resource(t *testing.T) {
	label := NewLabel()

	label.SetResource("ui://package/component")
	if label.Resource() != "ui://package/component" {
		t.Errorf("资源设置失败: got %s", label.Resource())
	}

	// 测试空资源
	label.SetResource("")
	if label.Resource() != "" {
		t.Error("设置空资源失败")
	}
}

// TestLabel_TitleObject 测试标题对象管理
func TestLabel_TitleObject(t *testing.T) {
	label := NewLabel()

	// 测试设置 TextField 作为标题对象
	tf := NewTextField()
	label.SetTitleObject(tf)

	if label.TitleObject() != tf {
		t.Error("标题对象设置失败")
	}

	// 设置标题并验证应用
	label.SetTitle("Test Title")
	if tf.Text() != "Test Title" {
		t.Errorf("标题文本未应用到 TextField: got %s", tf.Text())
	}
}

// TestLabel_IconObject 测试图标对象管理
func TestLabel_IconObject(t *testing.T) {
	label := NewLabel()

	// 测试设置 Loader 作为图标对象
	loader := NewLoader()
	label.SetIconObject(loader)

	if label.IconObject() != loader {
		t.Error("图标对象设置失败")
	}

	// 设置图标并验证应用
	label.SetIcon("icon.png")
	// 注意：这里只是测试不 panic，实际效果需要 Loader 实现
}

// TestLabel_TemplateComponent 测试模板组件
func TestLabel_TemplateComponent(t *testing.T) {
	label := NewLabel()

	// 测试设置 nil
	label.SetTemplateComponent(nil)
	if label.TemplateComponent() != nil {
		t.Error("模板组件应该为 nil")
	}

	// 测试设置组件模板
	template := NewComponent()
	label.SetTemplateComponent(template)

	if label.TemplateComponent() != template {
		t.Error("模板组件设置失败")
	}

	// 验证模板被添加为子对象（如果有实现）
	// 注意：实际实现可能使用内部机制而非标准子对象列表
}

// TestLabel_Chaining 测试链式调用效果
func TestLabel_Chaining(t *testing.T) {
	label := NewLabel()

	// 连续设置多个属性
	label.SetTitle("Chained")
	label.SetTitleColor("#ff00ff")
	label.SetIcon("chain.png")
	label.SetTitleFontSize(18)

	// 验证所有属性都被正确设置
	if label.Title() != "Chained" {
		t.Error("链式调用后标题错误")
	}

	if label.TitleColor() != "#ff00ff" {
		t.Error("链式调用后颜色错误")
	}

	if label.Icon() != "chain.png" {
		t.Error("链式调用后图标错误")
	}

	if label.TitleFontSize() != 18 {
		t.Error("链式调用后字体大小错误")
	}
}

// TestLabel_DifferentTitleObjectTypes 测试不同类型的标题对象
func TestLabel_DifferentTitleObjectTypes(t *testing.T) {
	label := NewLabel()
	label.SetTitle("Test")

	// 测试 TextField
	tf := NewTextField()
	label.SetTitleObject(tf)
	if tf.Text() != "Test" {
		t.Error("TextField 标题未更新")
	}

	// 测试 ComponentImpl
	comp := NewComponent()
	label.SetTitleObject(comp)
	if comp.Data() != "Test" {
		t.Error("Component 数据未更新")
	}

	// 测试 Label
	lbl := NewLabel()
	label.SetTitleObject(lbl)
	if lbl.Title() != "Test" {
		t.Error("Label 标题未更新")
	}

	// 测试 Button
	btn := NewButton()
	label.SetTitleObject(btn)
	if btn.Title() != "Test" {
		t.Error("Button 标题未更新")
	}
}

// TestAssertLabel 测试类型断言
func TestAssertLabel(t *testing.T) {
	label := NewLabel()

	result, ok := AssertLabel(label)
	if !ok {
		t.Error("AssertLabel 应该成功")
	}
	if result != label {
		t.Error("AssertLabel 返回的对象不正确")
	}

	if !IsLabel(label) {
		t.Error("IsLabel 应该返回 true")
	}

	obj := NewObject()
	_, ok = AssertLabel(obj)
	if ok {
		t.Error("AssertLabel 对非 Label 对象应该失败")
	}

	if IsLabel(obj) {
		t.Error("IsLabel 对非 Label 对象应该返回 false")
	}
}

// TestLabel_EmptyStates 测试空状态
func TestLabel_EmptyStates(t *testing.T) {
	label := NewLabel()

	// 在没有设置对象的情况下设置属性
	label.SetTitle("No Object")
	label.SetIcon("No Icon Object")
	label.SetTitleColor("#ffff00")

	// 应该不 panic
	if label.Title() != "No Object" {
		t.Error("标题设置失败")
	}

	if label.Icon() != "No Icon Object" {
		t.Error("图标设置失败")
	}

	if label.TitleColor() != "#ffff00" {
		t.Error("颜色设置失败")
	}
}
