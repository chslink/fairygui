package fairygui_test

import (
	"testing"

	"github.com/chslink/fairygui"
)

// ============================================================================
// 资源系统集成测试
// ============================================================================

// TestResourceSystem_Integration_BasicFlow 测试资源系统的基本流程
func TestResourceSystem_Integration_BasicFlow(t *testing.T) {
	// 跳过测试如果没有资源文件
	// 这些测试需要 demo/assets 目录下的 .fui 文件
	t.Skip("需要 GUI 环境和真实资源文件")

	// 创建加载器
	loader := fairygui.NewFileLoader("./demo/assets")
	if loader == nil {
		t.Fatal("Expected non-nil loader")
	}

	// 加载包
	pkg, err := loader.LoadPackage("Basics")
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// 验证包信息
	if pkg == nil {
		t.Fatal("Expected non-nil package")
	}

	if pkg.Name() == "" {
		t.Error("Expected non-empty package name")
	}

	// 验证包项
	items := pkg.Items()
	if len(items) == 0 {
		t.Error("Expected at least one item in package")
	}
}

// TestResourceSystem_Integration_Factory 测试工厂创建对象
func TestResourceSystem_Integration_Factory(t *testing.T) {
	t.Skip("需要 GUI 环境和真实资源文件")

	// 创建加载器和工厂
	loader := fairygui.NewFileLoader("./demo/assets")
	factory := fairygui.NewComponentFactory(loader, nil)

	// 加载包
	pkg, err := loader.LoadPackage("Basics")
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// 注册包
	factory.RegisterPackage(pkg)

	// 获取第一个组件类型的项
	var componentItem fairygui.PackageItem
	for _, item := range pkg.Items() {
		if item.Type() == fairygui.ResourceTypeComponent {
			componentItem = item
			break
		}
	}

	if componentItem == nil {
		t.Skip("No component found in package")
	}

	// 创建对象
	obj, err := factory.CreateComponent(pkg, componentItem.Name())
	if err != nil {
		t.Fatalf("Failed to create component: %v", err)
	}

	if obj == nil {
		t.Fatal("Expected non-nil object")
	}
}

// TestResourceSystem_Integration_URLCreation 测试 URL 方式创建对象
func TestResourceSystem_Integration_URLCreation(t *testing.T) {
	t.Skip("需要 GUI 环境和真实资源文件")

	// 创建加载器和工厂
	loader := fairygui.NewFileLoader("./demo/assets")
	factory := fairygui.NewComponentFactory(loader, nil)

	// 使用 URL 创建对象（会自动加载包）
	obj, err := factory.CreateObjectFromURL("ui://Basics/Button")
	if err != nil {
		// 如果 Button 不存在，尝试其他项
		t.Logf("Button not found, trying to load package to find available items")

		pkg, loadErr := loader.LoadPackage("Basics")
		if loadErr != nil {
			t.Fatalf("Failed to load package: %v", loadErr)
		}

		items := pkg.Items()
		if len(items) == 0 {
			t.Skip("No items in package")
		}

		// 尝试第一个组件项
		for _, item := range items {
			if item.Type() == fairygui.ResourceTypeComponent {
				url := "ui://Basics/" + item.Name()
				obj, err = factory.CreateObjectFromURL(url)
				if err == nil {
					break
				}
			}
		}

		if err != nil {
			t.Fatalf("Failed to create any object: %v", err)
		}
	}

	if obj == nil {
		t.Fatal("Expected non-nil object")
	}
}

// TestResourceSystem_Integration_DependencyManagement 测试依赖管理
func TestResourceSystem_Integration_DependencyManagement(t *testing.T) {
	t.Skip("需要 GUI 环境和真实资源文件")

	// 创建加载器
	loader := fairygui.NewFileLoader("./demo/assets")

	// 加载一个有依赖的包
	pkg, err := loader.LoadPackage("Basics")
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// 获取依赖列表
	deps := pkg.Dependencies()
	t.Logf("Package %s has %d dependencies", pkg.Name(), len(deps))

	// 如果有依赖，验证它们都已加载
	for _, depName := range deps {
		depPkg := loader.GetPackage(depName)
		if depPkg == nil {
			t.Errorf("Dependency %s not loaded", depName)
		}
	}
}

// ============================================================================
// 资源系统性能测试
// ============================================================================

// TestResourceSystem_Integration_CacheEfficiency 测试缓存效率
func TestResourceSystem_Integration_CacheEfficiency(t *testing.T) {
	t.Skip("需要 GUI 环境和真实资源文件")

	loader := fairygui.NewFileLoader("./demo/assets")

	// 第一次加载
	pkg1, err := loader.LoadPackage("Basics")
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	// 第二次加载（应该从缓存）
	pkg2, err := loader.LoadPackage("Basics")
	if err != nil {
		t.Fatalf("Failed to load package second time: %v", err)
	}

	// 验证返回相同的包实例
	if pkg1 != pkg2 {
		t.Error("Expected same package instance from cache")
	}
}

// ============================================================================
// 边界情况测试
// ============================================================================

// TestResourceSystem_Edge_NonExistentPackage 测试加载不存在的包
func TestResourceSystem_Edge_NonExistentPackage(t *testing.T) {
	loader := fairygui.NewFileLoader("./demo/assets")

	_, err := loader.LoadPackage("NonExistentPackage")
	if err == nil {
		t.Error("Expected error when loading non-existent package")
	}
}

// TestResourceSystem_Edge_InvalidURL 测试无效的 URL
func TestResourceSystem_Edge_InvalidURL(t *testing.T) {
	loader := fairygui.NewFileLoader("./demo/assets")
	factory := fairygui.NewComponentFactory(loader, nil)

	tests := []string{
		"invalid://url",
		"ui://",
		"ui://package",
		"ui:///item",
		"",
	}

	for _, url := range tests {
		_, err := factory.CreateObjectFromURL(url)
		if err == nil {
			t.Errorf("Expected error for invalid URL: %s", url)
		}
	}
}

// TestResourceSystem_Edge_CreateFromNonExistentItem 测试创建不存在的项
func TestResourceSystem_Edge_CreateFromNonExistentItem(t *testing.T) {
	t.Skip("需要 GUI 环境和真实资源文件")

	loader := fairygui.NewFileLoader("./demo/assets")
	factory := fairygui.NewComponentFactory(loader, nil)

	pkg, err := loader.LoadPackage("Basics")
	if err != nil {
		t.Fatalf("Failed to load package: %v", err)
	}

	_, err = factory.CreateComponent(pkg, "NonExistentItem")
	if err == nil {
		t.Error("Expected error when creating non-existent item")
	}
}

// ============================================================================
// 接口一致性测试
// ============================================================================

// TestResourceSystem_InterfaceConsistency 测试接口一致性
func TestResourceSystem_InterfaceConsistency(t *testing.T) {
	// 验证 PackageWrapper 实现 Package 接口
	var _ fairygui.Package = (*fairygui.PackageWrapper)(nil)

	// 验证 PackageItemWrapper 实现 PackageItem 接口
	var _ fairygui.PackageItem = (*fairygui.PackageItemWrapper)(nil)

	// 验证 FileLoader 实现 AssetLoader 接口
	var _ fairygui.AssetLoader = (*fairygui.FileLoader)(nil)
}
