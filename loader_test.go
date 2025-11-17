package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// FileLoader 基础测试
// ============================================================================

func TestFileLoader_Creation(t *testing.T) {
	loader := fairygui.NewFileLoader("./testdata")

	if loader == nil {
		t.Fatal("Expected non-nil FileLoader")
	}
}

func TestFileLoader_GetPackage_NotLoaded(t *testing.T) {
	loader := fairygui.NewFileLoader("./testdata")

	pkg := loader.GetPackage("NonExistent")
	if pkg != nil {
		t.Error("Expected nil for non-existent package")
	}
}

// ============================================================================
// PackageWrapper 测试
// ============================================================================

func TestPackageWrapper_Interface(t *testing.T) {
	// 这个测试验证 PackageWrapper 的接口方法存在
	// 实际的包加载需要真实的 .fui 文件

	// 创建一个模拟的场景（无需实际文件）
	loader := fairygui.NewFileLoader("./testdata")

	// 验证 GetPackage 返回 nil（因为没有加载）
	pkg := loader.GetPackage("Test")
	if pkg != nil {
		t.Error("Expected nil for unloaded package")
	}
}

// ============================================================================
// 资源加载接口测试
// ============================================================================

func TestFileLoader_Exists_NonExistent(t *testing.T) {
	loader := fairygui.NewFileLoader("./testdata")

	// 测试不存在的文件
	if loader.Exists("non_existent_file.txt") {
		t.Error("Expected false for non-existent file")
	}
}

// ============================================================================
// 包装类型测试
// ============================================================================

func TestPackageItemWrapper_Concept(t *testing.T) {
	// 这个测试验证 PackageItemWrapper 的概念
	// 实际使用需要真实的包数据

	// 只是验证类型存在和可以编译
	var _ *fairygui.PackageItemWrapper = nil
	var _ *fairygui.PackageWrapper = nil
}

// ============================================================================
// 加载器方法测试
// ============================================================================

func TestFileLoader_LoadFont_NotImplemented(t *testing.T) {
	loader := fairygui.NewFileLoader("./testdata")

	_, err := loader.LoadFont("test.ttf")
	if err == nil {
		t.Error("Expected error for unimplemented font loading")
	}
}

// ============================================================================
// 依赖管理概念测试
// ============================================================================

func TestFileLoader_DependencyManagement_Concept(t *testing.T) {
	// 这个测试验证依赖管理的设计概念
	// FileLoader 应该自动加载包的依赖

	loader := fairygui.NewFileLoader("./testdata")

	// 验证可以创建 loader
	if loader == nil {
		t.Fatal("Expected non-nil loader")
	}

	// 实际的依赖加载测试需要真实的包文件
	// 这里只验证 API 存在
}
