package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// TestRootComponentInitialization 验证根组件的基础属性初始化正确
// 这个测试确保修复了之前根组件 alpha=0, rotation=垃圾值的问题
//
// 问题背景：
// - 之前 BuildComponent 错误地对根组件调用 SetupBeforeAdd
// - SetupBeforeAdd 从 RawData Section 0 读取 GObject 基础属性
// - 但根组件的 RawData Section 0 存储的是 ComponentData 元数据，不是基础属性
// - 导致读取到错误的 alpha=0.00, rotation=2322168020992000.0
//
// 修复方案：
// - 根组件不再调用 SetupBeforeAdd/SetupAfterAdd
// - 根组件的尺寸、pivot 等从 ComponentData 设置
// - 根组件的 alpha、rotation 保持默认值（1.0, 0）
func TestRootComponentInitialization(t *testing.T) {
	// 加载 MainMenu.fui 包
	fuiPath := filepath.Join("..", "..", "..", "demo", "assets", "MainMenu.fui")
	fuiData, err := os.ReadFile(fuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 .fui 文件: %v", err)
	}

	pkg, err := assets.ParsePackage(fuiData, "demo/assets/MainMenu")
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

	// 构建根组件（不需要 AtlasResolver，只测试基础属性）
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	root, err := factory.BuildComponent(ctx, pkg, mainItem)
	if err != nil {
		t.Fatalf("构建组件失败: %v", err)
	}

	// 验证根组件的基础属性
	t.Run("DefaultAlpha", func(t *testing.T) {
		alpha := root.GObject.Alpha()
		if alpha != 1.0 {
			t.Errorf("根组件的 alpha 应该是 1.0，实际: %.2f", alpha)
		} else {
			t.Logf("✓ 根组件 alpha 正确: %.2f", alpha)
		}
	})

	t.Run("DefaultRotation", func(t *testing.T) {
		rotation := root.GObject.Rotation()
		if rotation != 0.0 {
			t.Errorf("根组件的 rotation 应该是 0.0，实际: %.6f", rotation)
		} else {
			t.Logf("✓ 根组件 rotation 正确: %.2f", rotation)
		}
	})

	t.Run("DefaultVisibility", func(t *testing.T) {
		visible := root.GObject.Visible()
		if !visible {
			t.Errorf("根组件应该是可见的，实际: %v", visible)
		} else {
			t.Logf("✓ 根组件 visible 正确: %v", visible)
		}
	})

	t.Run("DefaultTouchable", func(t *testing.T) {
		touchable := root.GObject.Touchable()
		if !touchable {
			t.Errorf("根组件应该是可触摸的，实际: %v", touchable)
		} else {
			t.Logf("✓ 根组件 touchable 正确: %v", touchable)
		}
	})

	t.Run("SizeFromComponentData", func(t *testing.T) {
		width := root.GObject.Width()
		height := root.GObject.Height()
		expectedWidth := float64(mainItem.Component.InitWidth)
		expectedHeight := float64(mainItem.Component.InitHeight)

		if width != expectedWidth || height != expectedHeight {
			t.Errorf("根组件尺寸不正确: 期望(%g,%g), 实际(%.0f,%.0f)",
				expectedWidth, expectedHeight, width, height)
		} else {
			t.Logf("✓ 根组件尺寸正确: %.0fx%.0f", width, height)
		}
	})

	t.Run("SourceSizeFromComponentData", func(t *testing.T) {
		sourceWidth, sourceHeight := root.GObject.SourceSize()
		expectedWidth := float64(mainItem.Component.SourceWidth)
		expectedHeight := float64(mainItem.Component.SourceHeight)

		if sourceWidth != expectedWidth || sourceHeight != expectedHeight {
			t.Errorf("根组件源尺寸不正确: 期望(%g,%g), 实际(%.0f,%.0f)",
				expectedWidth, expectedHeight, sourceWidth, sourceHeight)
		} else {
			t.Logf("✓ 根组件源尺寸正确: %.0fx%.0f", sourceWidth, sourceHeight)
		}
	})

	t.Run("Position", func(t *testing.T) {
		x := root.GObject.X()
		y := root.GObject.Y()
		// 根组件的位置应该是 (0, 0)，因为它是场景根
		if x != 0 || y != 0 {
			t.Errorf("根组件位置应该是(0,0)，实际: (%.0f,%.0f)", x, y)
		} else {
			t.Logf("✓ 根组件位置正确: (%.0f,%.0f)", x, y)
		}
	})

	t.Run("ChildrenCount", func(t *testing.T) {
		children := root.Children()
		// MainMenu 场景应该有 16 个子元素（1个背景 + 15个按钮）
		expectedCount := len(mainItem.Component.Children)
		if len(children) != expectedCount {
			t.Errorf("子元素数量不正确: 期望%d, 实际%d", expectedCount, len(children))
		} else {
			t.Logf("✓ 子元素数量正确: %d", len(children))
		}

		// 验证子元素也有正确的默认值
		for i, child := range children {
			if child == nil {
				continue
			}
			// 每个子元素的 alpha 应该也是 1.0（除非在 XML 中明确设置）
			alpha := child.Alpha()
			if alpha != 1.0 {
				t.Logf("警告：子元素[%d] %s 的 alpha=%.2f", i, child.Name(), alpha)
			}
		}
	})

	t.Run("Controllers", func(t *testing.T) {
		controllers := root.Controllers()
		expectedCount := len(mainItem.Component.Controllers)
		if len(controllers) != expectedCount {
			t.Errorf("控制器数量不正确: 期望%d, 实际%d", expectedCount, len(controllers))
		} else {
			t.Logf("✓ 控制器数量正确: %d", len(controllers))
		}
	})
}

// TestBasicsComponentInitialization 验证 Basics 包中组件的初始化
func TestBasicsComponentInitialization(t *testing.T) {
	fuiPath := filepath.Join("..", "..", "..", "demo", "assets", "Basics.fui")
	fuiData, err := os.ReadFile(fuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 .fui 文件: %v", err)
	}

	pkg, err := assets.ParsePackage(fuiData, "demo/assets/Basics")
	if err != nil {
		t.Fatalf("解析 .fui 文件失败: %v", err)
	}

	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)
	ctx := context.Background()

	// 测试几个典型组件
	testCases := []struct {
		name string
	}{
		{"Button"},
		{"ProgressBar"},
		{"Slider_HZ"},
		{"ComboBox"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var item *assets.PackageItem
			for _, i := range pkg.Items {
				if i.Type == assets.PackageItemTypeComponent && i.Name == tc.name {
					item = i
					break
				}
			}

			if item == nil {
				t.Skipf("未找到组件: %s", tc.name)
			}

			root, err := factory.BuildComponent(ctx, pkg, item)
			if err != nil {
				t.Fatalf("构建组件失败: %v", err)
			}

			// 验证基础属性
			alpha := root.GObject.Alpha()
			if alpha != 1.0 {
				t.Errorf("%s 的 alpha 应该是 1.0，实际: %.2f", tc.name, alpha)
			}

			rotation := root.GObject.Rotation()
			if rotation != 0.0 {
				t.Errorf("%s 的 rotation 应该是 0.0，实际: %.6f", tc.name, rotation)
			}

			visible := root.GObject.Visible()
			if !visible {
				t.Errorf("%s 应该是可见的，实际: %v", tc.name, visible)
			}

			t.Logf("✓ %s 组件初始化正确: alpha=%.1f, rotation=%.1f, visible=%v",
				tc.name, alpha, rotation, visible)
		})
	}
}
