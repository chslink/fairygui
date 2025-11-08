package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestButtonSizeInheritance 验证按钮从模板继承尺寸
// 这个测试确保修复了按钮尺寸为 0x0 的问题
//
// 问题背景：
// - Basics/Main.xml 中的按钮大多没有指定 size 属性
// - 按钮应该从模板组件 (Button10.xml, size="163,69") 继承尺寸
// - 但是 applyButtonTemplate 没有将模板尺寸复制到按钮的 GObject 上
// - 导致按钮显示为 size:0x0
//
// 修复方案：
// - 在 applyButtonTemplate 中，应用模板后检查按钮尺寸
// - 如果按钮尺寸为 0，从模板继承尺寸
func TestButtonSizeInheritance(t *testing.T) {
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

	// 构建 Main 组件（不需要 AtlasResolver，只测试尺寸）
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	root, err := factory.BuildComponent(ctx, pkg, mainItem)
	if err != nil {
		t.Fatalf("构建组件失败: %v", err)
	}

	// 测试用例：验证按钮尺寸
	testCases := []struct {
		name         string
		expectedSize struct{ w, h float64 }
	}{
		{"btn_Button", struct{ w, h float64 }{163, 69}},
		{"btn_Image", struct{ w, h float64 }{163, 69}},
		{"btn_Graph", struct{ w, h float64 }{163, 69}},
		{"btn_MovieClip", struct{ w, h float64 }{163, 69}},
		{"btn_Depth", struct{ w, h float64 }{163, 69}},
		{"btn_Loader", struct{ w, h float64 }{163, 69}},
		{"btn_List", struct{ w, h float64 }{163, 69}},
		{"btn_ProgressBar", struct{ w, h float64 }{163, 69}},
		{"btn_Slider", struct{ w, h float64 }{163, 69}},
		{"btn_ComboBox", struct{ w, h float64 }{163, 69}},
		{"btn_Clip&Scroll", struct{ w, h float64 }{163, 69}}, // 这个在 XML 中明确指定了 size="163,69"
		{"btn_Back", struct{ w, h float64 }{163, 69}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			child := root.ChildByName(tc.name)
			if child == nil {
				t.Skipf("未找到按钮: %s", tc.name)
			}

			// 检查是否是 GButton
			button, ok := child.Data().(*widgets.GButton)
			if !ok {
				t.Errorf("按钮 %s 不是 GButton 类型", tc.name)
				return
			}

			width := button.GComponent.GObject.Width()
			height := button.GComponent.GObject.Height()

			if width != tc.expectedSize.w || height != tc.expectedSize.h {
				t.Errorf("按钮 %s 尺寸不正确: 期望(%.0f,%.0f), 实际(%.0f,%.0f)",
					tc.name, tc.expectedSize.w, tc.expectedSize.h, width, height)
			} else {
				t.Logf("✓ 按钮 %s 尺寸正确: %.0fx%.0f", tc.name, width, height)
			}

			// 验证模板组件存在且尺寸正确
			template := button.TemplateComponent()
			if template == nil {
				t.Errorf("按钮 %s 没有模板组件", tc.name)
				return
			}

			templateW := template.GObject.Width()
			templateH := template.GObject.Height()
			if templateW != tc.expectedSize.w || templateH != tc.expectedSize.h {
				t.Errorf("按钮 %s 的模板尺寸不正确: 期望(%.0f,%.0f), 实际(%.0f,%.0f)",
					tc.name, tc.expectedSize.w, tc.expectedSize.h, templateW, templateH)
			} else {
				t.Logf("✓ 按钮 %s 的模板尺寸正确: %.0fx%.0f", tc.name, templateW, templateH)
			}
		})
	}
}

// TestButtonTemplateSize 验证按钮模板本身的尺寸定义
func TestButtonTemplateSize(t *testing.T) {
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

	// 查找 Button10 模板组件（id="hixt1j"）
	var buttonTemplateItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.ID == "hixt1j" {
			buttonTemplateItem = item
			break
		}
	}

	if buttonTemplateItem == nil {
		t.Fatalf("未找到 Button10 模板组件 (id=hixt1j)")
	}

	// 验证模板组件的元数据
	if buttonTemplateItem.Component == nil {
		t.Fatalf("Button10 模板缺少 Component 数据")
	}

	t.Run("TemplateComponentData", func(t *testing.T) {
		expectedW := int(163)
		expectedH := int(69)

		if int(buttonTemplateItem.Component.InitWidth) != expectedW {
			t.Errorf("模板 InitWidth 不正确: 期望%d, 实际%d", expectedW, buttonTemplateItem.Component.InitWidth)
		} else {
			t.Logf("✓ 模板 InitWidth 正确: %d", buttonTemplateItem.Component.InitWidth)
		}

		if int(buttonTemplateItem.Component.InitHeight) != expectedH {
			t.Errorf("模板 InitHeight 不正确: 期望%d, 实际%d", expectedH, buttonTemplateItem.Component.InitHeight)
		} else {
			t.Logf("✓ 模板 InitHeight 正确: %d", buttonTemplateItem.Component.InitHeight)
		}
	})

	// 构建模板组件验证运行时尺寸
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	template, err := factory.BuildComponent(ctx, pkg, buttonTemplateItem)
	if err != nil {
		t.Fatalf("构建模板组件失败: %v", err)
	}

	t.Run("TemplateRuntimeSize", func(t *testing.T) {
		width := template.GObject.Width()
		height := template.GObject.Height()

		expectedW := 163.0
		expectedH := 69.0

		if width != expectedW || height != expectedH {
			t.Errorf("模板运行时尺寸不正确: 期望(%.0f,%.0f), 实际(%.0f,%.0f)",
				expectedW, expectedH, width, height)
		} else {
			t.Logf("✓ 模板运行时尺寸正确: %.0fx%.0f", width, height)
		}
	})
}
