package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestComponentWithRelation 测试使用 Relation 的组件渲染问题
// 复现 n17/n20 组件（使用 Component6，有 Relation）子组件不显示的问题
func TestComponentWithRelation(t *testing.T) {
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

	// 构建 Demo_Component 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	demoComp, err := factory.BuildComponent(ctx, pkg, pkg.ItemByName("Demo_Component"))
	if err != nil {
		t.Fatalf("构建 Demo_Component 组件失败: %v", err)
	}

	if demoComp == nil {
		t.Fatalf("Demo_Component 组件为 nil")
	}

	t.Logf("Demo_Component 子组件数量: %d", len(demoComp.Children()))

	// 测试 n12 (无 Relation 的组件)
	t.Run("n12(无Relation)", func(t *testing.T) {
		n12Obj := demoComp.ChildByName("n12")
		if n12Obj == nil {
			t.Fatalf("未找到 n12 组件")
		}

		// 获取 GComponent
		n12 := n12Obj.Data()
		if n12 == nil {
			t.Errorf("n12 数据为 nil")
			return
		}

		comp, ok := n12.(*core.GComponent)
		if !ok {
			t.Errorf("n12 不是 GComponent，类型: %T", n12)
			return
		}

		t.Logf("n12 尺寸: %.0fx%.0f", n12Obj.Width(), n12Obj.Height())
		t.Logf("n12 子组件数量: %d", len(comp.Children()))

		// 验证子组件是否存在和可见
		for i, child := range comp.Children() {
			if child == nil {
				continue
			}
			t.Logf("  子组件[%d]: name=%s, 可见=%v, 尺寸=%.0fx%.0f",
				i, child.Name(), child.Visible(), child.Width(), child.Height())

			// 验证子组件数据
			if data := child.Data(); data != nil {
				t.Logf("    数据类型: %T", data)
			}
		}

		// n12 应该正常显示子组件
		if len(comp.Children()) == 0 {
			t.Errorf("n12 没有子组件")
		}
	})

	// 测试 n13 (无 Relation，但设置了尺寸)
	t.Run("n13(无Relation+尺寸)", func(t *testing.T) {
		n13Obj := demoComp.ChildByName("n13")
		if n13Obj == nil {
			t.Fatalf("未找到 n13 组件")
		}

		// 获取 GComponent
		n13Data := n13Obj.Data()
		if n13Data == nil {
			t.Errorf("n13 数据为 nil")
			return
		}

		comp, ok := n13Data.(*core.GComponent)
		if !ok {
			t.Errorf("n13 不是 GComponent，类型: %T", n13Data)
			return
		}

		t.Logf("n13 尺寸: %.0fx%.0f", n13Obj.Width(), n13Obj.Height())
		t.Logf("n13 子组件数量: %d", len(comp.Children()))

		// 验证子组件是否存在和可见
		for i, child := range comp.Children() {
			if child == nil {
				continue
			}
			t.Logf("  子组件[%d]: name=%s, 可见=%v, 尺寸=%.0fx%.0f",
				i, child.Name(), child.Visible(), child.Width(), child.Height())

			// 验证子组件数据
			if data := child.Data(); data != nil {
				t.Logf("    数据类型: %T", data)
			}
		}

		// n13 应该正常显示子组件
		if len(comp.Children()) == 0 {
			t.Errorf("n13 没有子组件")
		}
	})

	// 测试 n17 (有 Relation 的组件，未设置尺寸)
	t.Run("n17(有Relation)", func(t *testing.T) {
		n17Obj := demoComp.ChildByName("n17")
		if n17Obj == nil {
			t.Fatalf("未找到 n17 组件")
		}

		// 获取 GComponent
		n17Data := n17Obj.Data()
		if n17Data == nil {
			t.Errorf("n17 数据为 nil")
			return
		}

		comp, ok := n17Data.(*core.GComponent)
		if !ok {
			t.Errorf("n17 不是 GComponent，类型: %T", n17Data)
			return
		}

		t.Logf("n17 尺寸: %.0fx%.0f", n17Obj.Width(), n17Obj.Height())
		t.Logf("n17 子组件数量: %d", len(comp.Children()))

		// 验证子组件是否存在和可见
		for i, child := range comp.Children() {
			if child == nil {
				t.Logf("  子组件[%d] 为 nil", i)
				continue
			}
			t.Logf("  子组件[%d]: name=%s, 可见=%v, 尺寸=%.0fx%.0f, 位置=%.0f,%.0f",
				i, child.Name(), child.Visible(), child.Width(), child.Height(), child.X(), child.Y())

			// 验证子组件数据
			if data := child.Data(); data != nil {
				t.Logf("    数据类型: %T", data)

				// 如果是文本，检查文本内容
				if tf, ok := data.(*widgets.GTextField); ok {
					t.Logf("    文本内容: %q", tf.Text())
				}
			}
		}

		// n17 应该也有子组件
		if len(comp.Children()) == 0 {
			t.Errorf("n17 没有子组件（这是一个问题！）")
		}
	})

	// 测试 n20 (有 Relation 的组件，设置了尺寸)
	t.Run("n20(有Relation+尺寸)", func(t *testing.T) {
		n20Obj := demoComp.ChildByName("n20")
		if n20Obj == nil {
			t.Fatalf("未找到 n20 组件")
		}

		// 获取 GComponent
		n20Data := n20Obj.Data()
		if n20Data == nil {
			t.Errorf("n20 数据为 nil")
			return
		}

		comp, ok := n20Data.(*core.GComponent)
		if !ok {
			t.Errorf("n20 不是 GComponent，类型: %T", n20Data)
			return
		}

		t.Logf("n20 尺寸: %.0fx%.0f", n20Obj.Width(), n20Obj.Height())
		t.Logf("n20 子组件数量: %d", len(comp.Children()))

		// 验证子组件是否存在和可见
		for i, child := range comp.Children() {
			if child == nil {
				t.Logf("  子组件[%d] 为 nil", i)
				continue
			}
			t.Logf("  子组件[%d]: name=%s, 可见=%v, 尺寸=%.0fx%.0f, 位置=%.0f,%.0f",
				i, child.Name(), child.Visible(), child.Width(), child.Height(), child.X(), child.Y())

			// 验证子组件数据
			if data := child.Data(); data != nil {
				t.Logf("    数据类型: %T", data)

				// 如果是文本，检查文本内容
				if tf, ok := data.(*widgets.GTextField); ok {
					t.Logf("    文本内容: %q", tf.Text())
				}
			}
		}

		// n20 应该也有子组件
		if len(comp.Children()) == 0 {
			t.Errorf("n20 没有子组件（这是一个问题！）")
		}
	})
}

// TestComponent6Directly 直接测试 Component6 组件
func TestComponent6Directly(t *testing.T) {
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

	// 查找 Component6 (qa477h)
	component6Item := pkg.ItemByID("qa477h")
	if component6Item == nil {
		t.Fatalf("未找到 Component6 组件")
	}

	// 构建 Component6
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	component6, err := factory.BuildComponent(ctx, pkg, component6Item)
	if err != nil {
		t.Fatalf("构建 Component6 失败: %v", err)
	}

	if component6 == nil {
		t.Fatalf("Component6 为 nil")
	}

	t.Logf("Component6 尺寸: %.0fx%.0f", component6.Width(), component6.Height())
	t.Logf("Component6 子组件数量: %d", len(component6.Children()))

	// 检查所有子组件
	for i, child := range component6.Children() {
		if child == nil {
			t.Logf("子组件[%d] 为 nil", i)
			continue
		}
		t.Logf("子组件[%d]:", i)
		t.Logf("  name=%s, 可见=%v", child.Name(), child.Visible())
		t.Logf("  尺寸=%.0fx%.0f, 位置=%.0f,%.0f", child.Width(), child.Height(), child.X(), child.Y())
		t.Logf("  Alpha=%.2f", child.Alpha())

		// 验证子组件数据
		if data := child.Data(); data != nil {
			t.Logf("  数据类型: %T", data)

			// 如果是文本，检查文本内容
			if tf, ok := data.(*widgets.GTextField); ok {
				t.Logf("  文本内容: %q", tf.Text())
			}

			// 如果是图片，检查 PackageItem
			if img, ok := data.(*widgets.GImage); ok {
				if item := img.PackageItem(); item != nil {
					t.Logf("  PackageItem: %s (%s)", item.ID, item.Name)
				}
			}
		}

		// 检查 Relation
		if relations := child.Relations(); relations != nil {
			items := relations.Items()
			t.Logf("  Relation 数量: %d", len(items))
			if len(items) > 0 {
				t.Logf("    有 Relation 关系（Component6 的特征）")
			}
		}
	}

	// 验证预期子组件
	expectedChildren := []string{"n1", "n2", "n3"}
	for _, name := range expectedChildren {
		child := component6.ChildByName(name)
		if child == nil {
			t.Errorf("未找到子组件: %s", name)
		} else {
			t.Logf("✓ 找到子组件: %s", name)
		}
	}
}
