package fairygui

import (
	"fmt"
	"testing"
)

// TestNewLoader 测试创建新的加载器
func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	if loader == nil {
		t.Fatal("NewLoader() returned nil")
	}

	if loader.Object == nil {
		t.Error("Loader.Object is nil")
	}

	// 验证默认值
	if loader.Align() != TextAlignCenter {
		t.Errorf("默认水平对齐不正确: got %v, want %v", loader.Align(), TextAlignCenter)
	}

	if loader.VerticalAlign() != VerticalAlignMiddle {
		t.Errorf("默认垂直对齐不正确: got %v, want %v", loader.VerticalAlign(), VerticalAlignMiddle)
	}

	if loader.AutoSize() {
		t.Error("新创建的 Loader 不应该默认启用自动大小")
	}

	if loader.URL() != "" {
		t.Error("新创建的 Loader 默认 URL 应该为空")
	}
}

// TestLoader_URL 测试 URL 设置
func TestLoader_URL(t *testing.T) {
	loader := NewLoader()

	// 设置 URL
	loader.SetURL("ui://package/item")
	if loader.URL() != "ui://package/item" {
		t.Errorf("URL 设置失败: got %s", loader.URL())
	}

	// 更新 URL
	loader.SetURL("ui://package/item2")
	if loader.URL() != "ui://package/item2" {
		t.Errorf("URL 更新失败: got %s", loader.URL())
	}

	// 清空 URL
	loader.SetURL("")
	if loader.URL() != "" {
		t.Error("清空 URL 失败")
	}
}

// TestLoader_PackageItem 测试资源项设置
func TestLoader_PackageItem(t *testing.T) {
	loader := NewLoader()

	// 测试设置 nil
	loader.SetPackageItem(nil)
	if loader.PackageItem() != nil {
		t.Error("设置为 nil 应该成功")
	}

	// 创建虚拟资源项
	item := &PackageItemWrapper{}
	loader.SetPackageItem(item)

	if loader.PackageItem() != item {
		t.Error("PackageItem 设置失败")
	}
}

// TestLoader_AutoSize 测试自动大小
func TestLoader_AutoSize(t *testing.T) {
	loader := NewLoader()

	// 默认禁用
	if loader.AutoSize() {
		t.Error("自动大小应该默认禁用")
	}

	// 启用自动大小
	loader.SetAutoSize(true)
	if !loader.AutoSize() {
		t.Error("启用自动大小失败")
	}

	// 禁用自动大小
	loader.SetAutoSize(false)
	if loader.AutoSize() {
		t.Error("禁用自动大小失败")
	}
}

// TestLoader_Fill 测试填充模式
func TestLoader_Fill(t *testing.T) {
	loader := NewLoader()

	// 默认填充模式
	if loader.Fill() != 0 {
		t.Errorf("默认填充模式应该是 0 (none), got %d", loader.Fill())
	}

	// 设置不同的填充模式
	fillModes := []int{0, 1, 2, 3, 4}
	for _, mode := range fillModes {
		loader.SetFill(mode)
		if loader.Fill() != mode {
			t.Errorf("填充模式设置失败: got %d, want %d", loader.Fill(), mode)
		}
	}
}

// TestLoader_Align 测试水平对齐
func TestLoader_Align(t *testing.T) {
	loader := NewLoader()

	// 默认对齐
	if loader.Align() != TextAlignCenter {
		t.Errorf("默认水平对齐应该是 Center, got %v", loader.Align())
	}

	// 设置左对齐
	loader.SetAlign(TextAlignLeft)
	if loader.Align() != TextAlignLeft {
		t.Errorf("设置左对齐失败: got %v", loader.Align())
	}

	// 设置右对齐
	loader.SetAlign(TextAlignRight)
	if loader.Align() != TextAlignRight {
		t.Errorf("设置右对齐失败: got %v", loader.Align())
	}

	// 恢复居中
	loader.SetAlign(TextAlignCenter)
	if loader.Align() != TextAlignCenter {
		t.Errorf("恢复居中对齐失败: got %v", loader.Align())
	}
}

// TestLoader_VerticalAlign 测试垂直对齐
func TestLoader_VerticalAlign(t *testing.T) {
	loader := NewLoader()

	// 默认对齐
	if loader.VerticalAlign() != VerticalAlignMiddle {
		t.Errorf("默认垂直对齐应该是 Middle, got %v", loader.VerticalAlign())
	}

	// 设置顶部对齐
	loader.SetVerticalAlign(VerticalAlignTop)
	if loader.VerticalAlign() != VerticalAlignTop {
		t.Errorf("设置顶部对齐失败: got %v", loader.VerticalAlign())
	}

	// 设置底部对齐
	loader.SetVerticalAlign(VerticalAlignBottom)
	if loader.VerticalAlign() != VerticalAlignBottom {
		t.Errorf("设置底部对齐失败: got %v", loader.VerticalAlign())
	}

	// 恢复居中
	loader.SetVerticalAlign(VerticalAlignMiddle)
	if loader.VerticalAlign() != VerticalAlignMiddle {
		t.Errorf("恢复居中对齐失败: got %v", loader.VerticalAlign())
	}
}

// TestLoader_Chaining 测试链式调用效果
func TestLoader_Chaining(t *testing.T) {
	loader := NewLoader()

	// 连续设置多个属性
	loader.SetURL("ui://package/item")
	loader.SetAutoSize(true)
	loader.SetFill(1)
	loader.SetAlign(TextAlignLeft)
	loader.SetVerticalAlign(VerticalAlignTop)

	// 验证所有属性都被正确设置
	if loader.URL() != "ui://package/item" {
		t.Error("链式调用后 URL 错误")
	}

	if !loader.AutoSize() {
		t.Error("链式调用后自动大小错误")
	}

	if loader.Fill() != 1 {
		t.Errorf("链式调用后填充模式错误: got %d", loader.Fill())
	}

	if loader.Align() != TextAlignLeft {
		t.Error("链式调用后水平对齐错误")
	}

	if loader.VerticalAlign() != VerticalAlignTop {
		t.Error("链式调用后垂直对齐错误")
	}
}

// TestAssertLoader 测试类型断言
func TestAssertLoader(t *testing.T) {
	loader := NewLoader()

	result, ok := AssertLoader(loader)
	if !ok {
		t.Error("AssertLoader 应该成功")
	}
	if result != loader {
		t.Error("AssertLoader 返回的对象不正确")
	}

	if !IsLoader(loader) {
		t.Error("IsLoader 应该返回 true")
	}

	obj := NewObject()
	_, ok = AssertLoader(obj)
	if ok {
		t.Error("AssertLoader 对非 Loader 对象应该失败")
	}

	if IsLoader(obj) {
		t.Error("IsLoader 对非 Loader 对象应该返回 false")
	}
}

// TestLoader_EmptyStates 测试空状态
func TestLoader_EmptyStates(t *testing.T) {
	loader := NewLoader()

	// 测试空 URL、空 PackageItem
	loader.SetURL("")
	loader.SetPackageItem(nil)

	if loader.URL() != "" {
		t.Error("空 URL 设置失败")
	}

	if loader.PackageItem() != nil {
		t.Error("空 PackageItem 设置失败")
	}
}

// TestLoader_MultipleChanges 测试多次变更
func TestLoader_MultipleChanges(t *testing.T) {
	loader := NewLoader()

	// 多次变更 URL
	for i := 0; i < 10; i++ {
		url := fmt.Sprintf("ui://package/item%d", i)
		loader.SetURL(url)
		if loader.URL() != url {
			t.Errorf("第 %d 次 URL 变更失败: got %s", i, loader.URL())
		}
	}

	// 多次变更填充模式
	for i := 0; i < 5; i++ {
		loader.SetFill(i)
		if loader.Fill() != i {
			t.Errorf("第 %d 次填充模式变更失败: got %d", i, loader.Fill())
		}
	}
}

// TestLoader_URLValidation 测试 URL 验证
func TestLoader_URLValidation(t *testing.T) {
	loader := NewLoader()

	// 测试空 URL
	loader.SetURL("")
	if loader.URL() != "" {
		t.Error("空 URL 应该被接受")
	}

	// 测试 FairyGUI URL
	loader.SetURL("ui://package/item")
	if loader.URL() != "ui://package/item" {
		t.Error("FairyGUI URL 应该被接受")
	}

	// 测试普通路径
	loader.SetURL("path/to/resource.png")
	if loader.URL() != "path/to/resource.png" {
		t.Error("普通路径应该被接受")
	}

	// 测试带参数的 URL
	loader.SetURL("ui://package/item?param=value")
	if loader.URL() != "ui://package/item?param=value" {
		t.Error("带参数 URL 应该被接受")
	}
}
