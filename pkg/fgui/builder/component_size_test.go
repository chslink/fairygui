package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// TestComponentSizeInheritance 验证组件尺寸的正确设置
// 测试场景：Main.xml 中的 container 组件指定了 size="1136,570"
// 应该覆盖 Component4.xml 中定义的默认尺寸 640x890
func TestComponentSizeInheritance(t *testing.T) {
	// 加载 Basics.fui 包
	fuiPath := filepath.Join("..", "..", "..", "demo", "assets", "Basics.fui")
	fuiData, err := os.ReadFile(fuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 .fui 文件: %v", err)
	}

	pkg, err := assets.ParsePackage(fuiData, "demo/assets/Basics")
	if err != nil {
		t.Fatalf("解析 .fui 文件失败: %v", err)
	}

	// 查找 Main 组件
	var mainItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Main" {
			mainItem = item
			break
		}
	}

	if mainItem == nil {
		t.Fatalf("未找到 Main 组件")
	}

	// 构建 Main 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	main, err := factory.BuildComponent(ctx, pkg, mainItem)
	if err != nil {
		t.Fatalf("构建 Main 组件失败: %v", err)
	}

	// 测试 container 的尺寸
	t.Run("container_Size", func(t *testing.T) {
		container := main.ChildByName("container")
		if container == nil {
			t.Fatalf("未找到 container 对象")
		}

		// XML 中定义：<component id="n25" name="container" src="f2ax74" xy="-1143,70" size="1136,570">
		// 期望尺寸：1136x570 (来自 Main.xml 的覆盖值)
		expectedWidth, expectedHeight := 1136.0, 570.0
		actualWidth := container.Width()
		actualHeight := container.Height()

		if actualWidth != expectedWidth || actualHeight != expectedHeight {
			t.Errorf("container 尺寸不正确:\n期望: %.0fx%.0f\n实际: %.0fx%.0f",
				expectedWidth, expectedHeight, actualWidth, actualHeight)
		} else {
			t.Logf("✓ container 尺寸正确: %.0fx%.0f", actualWidth, actualHeight)
		}

		// 验证 SourceSize 和 InitSize 也正确设置
		sourceW, sourceH := container.SourceSize()
		if sourceW != expectedWidth || sourceH != expectedHeight {
			t.Errorf("container SourceSize 不正确: 期望(%.0f,%.0f), 实际(%.0f,%.0f)",
				expectedWidth, expectedHeight, sourceW, sourceH)
		}

		initW, initH := container.InitSize()
		if initW != expectedWidth || initH != expectedHeight {
			t.Errorf("container InitSize 不正确: 期望(%.0f,%.0f), 实际(%.0f,%.0f)",
				expectedWidth, expectedHeight, initW, initH)
		}
	})

	// 测试控制器切换后尺寸保持不变（因为没有 GearSize）
	t.Run("controller_Switch_Preserves_Size", func(t *testing.T) {
		container := main.ChildByName("container")
		if container == nil {
			t.Fatalf("未找到 container 对象")
		}

		// 获取 c1 控制器
		c1 := main.ControllerByName("c1")
		if c1 == nil {
			t.Fatalf("未找到 c1 控制器")
		}

		expectedWidth, expectedHeight := 1136.0, 570.0

		// 切换到 page 0
		c1.SetSelectedIndex(0)
		if container.Width() != expectedWidth || container.Height() != expectedHeight {
			t.Errorf("切换到 page 0 后尺寸错误: 期望(%.0f,%.0f), 实际(%.0f,%.0f)",
				expectedWidth, expectedHeight, container.Width(), container.Height())
		}

		// 切换到 page 1
		c1.SetSelectedIndex(1)
		if container.Width() != expectedWidth || container.Height() != expectedHeight {
			t.Errorf("切换到 page 1 后尺寸错误: 期望(%.0f,%.0f), 实际(%.0f,%.0f)",
				expectedWidth, expectedHeight, container.Width(), container.Height())
		}

		t.Logf("✓ 控制器切换后尺寸保持正确: %.0fx%.0f", container.Width(), container.Height())
	})
}

// TestNestedComponentDefaultSize 验证嵌套组件的默认尺寸
// 当 ComponentChild 没有指定 size 时，应该使用模板组件的默认尺寸
func TestNestedComponentDefaultSize(t *testing.T) {
	fuiPath := filepath.Join("..", "..", "..", "demo", "assets", "Basics.fui")
	fuiData, err := os.ReadFile(fuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 .fui 文件: %v", err)
	}

	pkg, err := assets.ParsePackage(fuiData, "demo/assets/Basics")
	if err != nil {
		t.Fatalf("解析 .fui 文件失败: %v", err)
	}

	// 查找 Component4 组件（f2ax74）
	var component4Item *assets.PackageItem
	for _, item := range pkg.Items {
		if item.ID == "f2ax74" {
			component4Item = item
			break
		}
	}

	if component4Item == nil {
		t.Fatalf("未找到 Component4 (f2ax74)")
	}

	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	comp4, err := factory.BuildComponent(ctx, pkg, component4Item)
	if err != nil {
		t.Fatalf("构建 Component4 失败: %v", err)
	}

	// Component4.xml 定义：<component size="640,890">
	expectedWidth, expectedHeight := 640.0, 890.0
	actualWidth := comp4.Width()
	actualHeight := comp4.Height()

	if actualWidth != expectedWidth || actualHeight != expectedHeight {
		t.Errorf("Component4 默认尺寸不正确:\n期望: %.0fx%.0f\n实际: %.0fx%.0f",
			expectedWidth, expectedHeight, actualWidth, actualHeight)
	} else {
		t.Logf("✓ Component4 默认尺寸正确: %.0fx%.0f", actualWidth, actualHeight)
	}
}
