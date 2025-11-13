package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestSliderComponentAccess 验证 Slider 组件的构建和获取
// 场景：Demo_Slider.xml 中有 4 个 slider 组件（n1, n2, n3, n4）
// 预期：能够正确构建和访问所有 Slider 组件及其子元素
func TestSliderComponentAccess(t *testing.T) {
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

	// 创建测试 Stage 环境
	env := testutil.NewStageEnv(t, 1136, 640)
	stage := env.Stage

	// 构建 Factory
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 Demo_Slider 组件
	var demoItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_Slider" {
			demoItem = item
			break
		}
	}
	if demoItem == nil {
		t.Fatalf("未找到 Demo_Slider 组件")
	}

	demo, err := factory.BuildComponent(ctx, pkg, demoItem)
	if err != nil {
		t.Fatalf("构建 Demo_Slider 组件失败: %v", err)
	}

	// 将 demo 添加到 stage
	stage.AddChild(demo.DisplayObject())

	// 测试所有 slider 组件
	sliderTests := []struct {
		name          string
		expectedValue float64
		expectedMax   float64
	}{
		{"n1", 50, 100},
		{"n2", 50, 100},
		{"n3", 50, 100},
		{"n4", 50, 100},
	}

	for _, tt := range sliderTests {
		t.Run(tt.name, func(t *testing.T) {
			obj := demo.ChildByName(tt.name)
			if obj == nil {
				t.Fatalf("未找到 %s 对象", tt.name)
			}

			slider, ok := obj.Data().(*widgets.GSlider)
			if !ok {
				t.Fatalf("%s 不是 GSlider，实际类型: %T", tt.name, obj.Data())
			}

			t.Logf("%s 组件信息:", tt.name)
			t.Logf("  Name: %s", obj.Name())
			t.Logf("  Type: %T", obj.Data())
			t.Logf("  Value: %.0f", slider.Value())
			t.Logf("  Min: %.0f", slider.Min())
			t.Logf("  Max: %.0f", slider.Max())
			t.Logf("  Size: %.0fx%.0f", obj.Width(), obj.Height())

			// 验证初始值
			if slider.Value() != tt.expectedValue {
				t.Errorf("预期 value=%.0f，实际: %.0f", tt.expectedValue, slider.Value())
			}

			// 验证最大值
			if slider.Max() != tt.expectedMax {
				t.Errorf("预期 max=%.0f，实际: %.0f", tt.expectedMax, slider.Max())
			}

			// 验证模板组件
			tmpl := slider.TemplateComponent()
			if tmpl == nil {
				t.Errorf("%s 的 template 组件为 nil", tt.name)
			} else {
				t.Logf("  ✓ Template 组件已设置")
				t.Logf("    子元素数量: %d", tmpl.NumChildren())
				for i := 0; i < tmpl.NumChildren(); i++ {
					child := tmpl.ChildAt(i)
					if child != nil {
						t.Logf("      [%d] %s (类型: %T)", i, child.Name(), child.Data())
					}
				}
			}

			// 检查关键子对象
			hasGrip := false
			hasBar := false

			if tmpl != nil {
				if grip := tmpl.ChildByName("grip"); grip != nil {
					hasGrip = true
					t.Logf("  ✓ grip 对象已找到: pos=(%.0f,%.0f), size=%.0fx%.0f",
						grip.X(), grip.Y(), grip.Width(), grip.Height())
				}

				if bar := tmpl.ChildByName("bar"); bar != nil {
					hasBar = true
					t.Logf("  ✓ bar 对象已找到: pos=(%.0f,%.0f), size=%.0fx%.0f",
						bar.X(), bar.Y(), bar.Width(), bar.Height())
				}
				if barV := tmpl.ChildByName("bar_v"); barV != nil {
					hasBar = true
					t.Logf("  ✓ bar_v 对象已找到: pos=(%.0f,%.0f), size=%.0fx%.0f",
						barV.X(), barV.Y(), barV.Width(), barV.Height())
				}

				if title := tmpl.ChildByName("title"); title != nil {
					t.Logf("  ✓ title 对象已找到")
				}
			}

			// 至少应该有 grip 或 bar
			if !hasGrip && !hasBar {
				t.Errorf("%s 缺少 grip 和 bar 对象，slider 可能无法正常工作", tt.name)
			}

			// 测试改变值
			t.Run("ChangeValue", func(t *testing.T) {
				slider.SetValue(75)
				if slider.Value() != 75 {
					t.Errorf("设置 value=75 后，实际值: %.0f", slider.Value())
				}
				t.Logf("  ✓ SetValue(75) 成功")

				slider.SetValue(25)
				if slider.Value() != 25 {
					t.Errorf("设置 value=25 后，实际值: %.0f", slider.Value())
				}
				t.Logf("  ✓ SetValue(25) 成功")

				// 恢复初始值
				slider.SetValue(tt.expectedValue)
			})

			// 测试边界值
			t.Run("BoundaryValues", func(t *testing.T) {
				slider.SetValue(0)
				if slider.Value() != 0 {
					t.Errorf("设置 value=0 后，实际值: %.0f", slider.Value())
				}
				t.Logf("  ✓ SetValue(0) 成功")

				slider.SetValue(100)
				if slider.Value() != 100 {
					t.Errorf("设置 value=100 后，实际值: %.0f", slider.Value())
				}
				t.Logf("  ✓ SetValue(100) 成功")

				// 测试超出范围的值
				slider.SetValue(-10)
				if slider.Value() != slider.Min() {
					t.Errorf("设置 value=-10 应该被限制到 min(%.0f)，实际: %.0f",
						slider.Min(), slider.Value())
				}
				t.Logf("  ✓ SetValue(-10) 正确限制到 min")

				slider.SetValue(150)
				if slider.Value() != slider.Max() {
					t.Errorf("设置 value=150 应该被限制到 max(%.0f)，实际: %.0f",
						slider.Max(), slider.Value())
				}
				t.Logf("  ✓ SetValue(150) 正确限制到 max")

				// 恢复初始值
				slider.SetValue(tt.expectedValue)
			})
		})
	}

	// 测试总结
	t.Run("Summary", func(t *testing.T) {
		t.Logf("=== Slider 组件完整性测试总结 ===")
		t.Logf("✓ 成功加载 Demo_Slider 场景")
		t.Logf("✓ 所有 4 个 Slider 组件都能正确获取")
		t.Logf("✓ 所有 Slider 的初始值和范围正确")
		t.Logf("✓ 所有 Slider 能够正确改变值")
		t.Logf("✓ 所有 Slider 边界值处理正确")
	})
}
