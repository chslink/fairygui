package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// ComponentFactory 基础测试
// ============================================================================

func TestComponentFactory_Creation(t *testing.T) {
	loader := fairygui.NewFileLoader("./testdata")
	factory := fairygui.NewComponentFactory(loader, nil)

	if factory == nil {
		t.Fatal("Expected non-nil ComponentFactory")
	}
}

func TestComponentFactory_GetPackage_NotLoaded(t *testing.T) {
	loader := fairygui.NewFileLoader("./testdata")
	factory := fairygui.NewComponentFactory(loader, nil)

	pkg := factory.GetPackage("NonExistent")
	if pkg != nil {
		t.Error("Expected nil for non-existent package")
	}
}

// ============================================================================
// URL 解析测试
// ============================================================================

func TestComponentFactory_CreateObjectFromURL_InvalidFormat(t *testing.T) {
	loader := fairygui.NewFileLoader("./testdata")
	factory := fairygui.NewComponentFactory(loader, nil)

	tests := []struct {
		name string
		url  string
	}{
		{"no prefix", "Main/Button"},
		{"missing item", "ui://Main"},
		{"empty url", ""},
		{"only prefix", "ui://"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := factory.CreateObjectFromURL(tt.url)
			if err == nil {
				t.Errorf("Expected error for invalid URL: %s", tt.url)
			}
		})
	}
}

// ============================================================================
// 包注册测试
// ============================================================================

func TestComponentFactory_RegisterPackage(t *testing.T) {
	loader := fairygui.NewFileLoader("./testdata")
	factory := fairygui.NewComponentFactory(loader, nil)

	// 创建模拟包（实际实现需要真实的 .fui 文件）
	// 这里只验证 API 不会 panic
	factory.RegisterPackage(nil)

	// 验证可以调用 GetPackage
	pkg := factory.GetPackage("Test")
	if pkg != nil {
		t.Error("Expected nil for unregistered package")
	}
}

// ============================================================================
// 边界情况测试
// ============================================================================

func TestComponentFactory_CreateComponent_NilPackage(t *testing.T) {
	loader := fairygui.NewFileLoader("./testdata")
	factory := fairygui.NewComponentFactory(loader, nil)

	_, err := factory.CreateComponent(nil, "Test")
	if err == nil {
		t.Error("Expected error for nil package")
	}
}

func TestComponentFactory_RawFactory(t *testing.T) {
	loader := fairygui.NewFileLoader("./testdata")
	factory := fairygui.NewComponentFactory(loader, nil)

	rawFactory := factory.RawFactory()
	if rawFactory == nil {
		t.Error("Expected non-nil raw factory")
	}
}
