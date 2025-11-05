package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// TestGGroupPositionSync 测试 GGroup 位置同步
// 验证当 Group 移动时，子对象是否跟随移动
func TestGGroupPositionSync(t *testing.T) {
	// 加载 Basics.fui 包
	demoPath := filepath.Join("..", "..", "..", "demo", "assets")
	basicsData, err := os.ReadFile(filepath.Join(demoPath, "Basics.fui"))
	if err != nil {
		t.Fatalf("无法读取 Basics.fui: %v", err)
	}

	pkg, err := assets.ParsePackage(basicsData, "Basics")
	if err != nil {
		t.Fatalf("解析包失败: %v", err)
	}

	// 查找 Demo_Controller 组件
	demoCtrlItem := pkg.ItemByName("Demo_Controller")
	if demoCtrlItem == nil {
		t.Fatal("未找到 Demo_Controller 组件")
	}

	// 构建组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	rootComponent, err := factory.BuildComponent(ctx, pkg, demoCtrlItem)
	if err != nil {
		t.Fatalf("构建 Demo_Controller 失败: %v", err)
	}

	t.Run("ManualMove", func(t *testing.T) {
		// 获取对象
		n16 := rootComponent.ChildByName("n16")
		n13 := rootComponent.ChildByName("n13")
		n14 := rootComponent.ChildByName("n14")
		n15 := rootComponent.ChildByName("n15")

		if n16 == nil || n13 == nil || n14 == nil || n15 == nil {
			t.Fatal("未找到必需的对象")
		}

		// 记录初始位置
		initialN16X, initialN16Y := n16.X(), n16.Y()
		initialN13X, initialN13Y := n13.X(), n13.Y()
		initialN14X, initialN14Y := n14.X(), n14.Y()
		initialN15X, initialN15Y := n15.X(), n15.Y()

		t.Logf("初始位置:")
		t.Logf("  n16: (%.0f, %.0f)", initialN16X, initialN16Y)
		t.Logf("  n13: (%.0f, %.0f)", initialN13X, initialN13Y)
		t.Logf("  n14: (%.0f, %.0f)", initialN14X, initialN14Y)
		t.Logf("  n15: (%.0f, %.0f)", initialN15X, initialN15Y)

		// 手动移动 Group
		dx := 100.0
		dy := 50.0
		n16.SetPosition(initialN16X+dx, initialN16Y+dy)

		// 验证 Group 移动了
		newN16X, newN16Y := n16.X(), n16.Y()
		if newN16X != initialN16X+dx || newN16Y != initialN16Y+dy {
			t.Errorf("n16 移动失败: 期望 (%.0f, %.0f), 实际 (%.0f, %.0f)",
				initialN16X+dx, initialN16Y+dy, newN16X, newN16Y)
		}

		// 验证子对象跟随移动
		newN13X, newN13Y := n13.X(), n13.Y()
		expectedN13X := initialN13X + dx
		expectedN13Y := initialN13Y + dy
		if newN13X != expectedN13X || newN13Y != expectedN13Y {
			t.Errorf("n13 未跟随移动: 期望 (%.0f, %.0f), 实际 (%.0f, %.0f)",
				expectedN13X, expectedN13Y, newN13X, newN13Y)
		} else {
			t.Logf("n13 正确跟随移动: (%.0f, %.0f) → (%.0f, %.0f) ✓",
				initialN13X, initialN13Y, newN13X, newN13Y)
		}

		newN14X, newN14Y := n14.X(), n14.Y()
		expectedN14X := initialN14X + dx
		expectedN14Y := initialN14Y + dy
		if newN14X != expectedN14X || newN14Y != expectedN14Y {
			t.Errorf("n14 未跟随移动: 期望 (%.0f, %.0f), 实际 (%.0f, %.0f)",
				expectedN14X, expectedN14Y, newN14X, newN14Y)
		} else {
			t.Logf("n14 正确跟随移动 ✓")
		}

		newN15X, newN15Y := n15.X(), n15.Y()
		expectedN15X := initialN15X + dx
		expectedN15Y := initialN15Y + dy
		if newN15X != expectedN15X || newN15Y != expectedN15Y {
			t.Errorf("n15 未跟随移动: 期望 (%.0f, %.0f), 实际 (%.0f, %.0f)",
				expectedN15X, expectedN15Y, newN15X, newN15Y)
		} else {
			t.Logf("n15 正确跟随移动 ✓")
		}
	})

	t.Run("GearXYMove", func(t *testing.T) {
		// 重新构建组件以获取干净的初始状态
		rootComponent, err := factory.BuildComponent(ctx, pkg, demoCtrlItem)
		if err != nil {
			t.Fatalf("构建 Demo_Controller 失败: %v", err)
		}

		// 获取对象
		n16 := rootComponent.ChildByName("n16")
		n13 := rootComponent.ChildByName("n13")
		n14 := rootComponent.ChildByName("n14")
		n15 := rootComponent.ChildByName("n15")
		c2 := rootComponent.ControllerByName("c2")

		if n16 == nil || n13 == nil || n14 == nil || n15 == nil || c2 == nil {
			t.Fatal("未找到必需的对象或控制器")
		}

		// 初始状态应该是 page 0
		// 根据 XML: n16 在 page 0 的位置是 (1154, 450)
		initialN16X, initialN16Y := n16.X(), n16.Y()
		initialN13X, initialN13Y := n13.X(), n13.Y()
		initialN14X, initialN14Y := n14.X(), n14.Y()
		initialN15X, initialN15Y := n15.X(), n15.Y()

		t.Logf("page 0 初始位置:")
		t.Logf("  n16: (%.0f, %.0f)", initialN16X, initialN16Y)
		t.Logf("  n13: (%.0f, %.0f)", initialN13X, initialN13Y)
		t.Logf("  n14: (%.0f, %.0f)", initialN14X, initialN14Y)
		t.Logf("  n15: (%.0f, %.0f)", initialN15X, initialN15Y)

		// 切换到 page 1
		// 根据 XML: n16 在 page 1 的位置是 (661, 450) (default值)
		c2.SetSelectedIndex(1)

		newN16X, newN16Y := n16.X(), n16.Y()
		newN13X, newN13Y := n13.X(), n13.Y()
		newN14X, newN14Y := n14.X(), n14.Y()
		newN15X, newN15Y := n15.X(), n15.Y()

		t.Logf("page 1 新位置:")
		t.Logf("  n16: (%.0f, %.0f)", newN16X, newN16Y)
		t.Logf("  n13: (%.0f, %.0f)", newN13X, newN13Y)
		t.Logf("  n14: (%.0f, %.0f)", newN14X, newN14Y)
		t.Logf("  n15: (%.0f, %.0f)", newN15X, newN15Y)

		// 计算 Group 的实际移动量
		actualDx := newN16X - initialN16X
		actualDy := newN16Y - initialN16Y

		t.Logf("Group 移动量: (%.0f, %.0f)", actualDx, actualDy)

		// 验证子对象是否按相同量移动
		expectedN13X := initialN13X + actualDx
		expectedN13Y := initialN13Y + actualDy
		if newN13X != expectedN13X || newN13Y != expectedN13Y {
			t.Errorf("n13 未正确跟随 GearXY 移动: 期望 (%.0f, %.0f), 实际 (%.0f, %.0f)",
				expectedN13X, expectedN13Y, newN13X, newN13Y)
		} else {
			t.Logf("n13 正确跟随 GearXY 移动 ✓")
		}

		expectedN14X := initialN14X + actualDx
		expectedN14Y := initialN14Y + actualDy
		if newN14X != expectedN14X || newN14Y != expectedN14Y {
			t.Errorf("n14 未正确跟随 GearXY 移动: 期望 (%.0f, %.0f), 实际 (%.0f, %.0f)",
				expectedN14X, expectedN14Y, newN14X, newN14Y)
		} else {
			t.Logf("n14 正确跟随 GearXY 移动 ✓")
		}

		expectedN15X := initialN15X + actualDx
		expectedN15Y := initialN15Y + actualDy
		if newN15X != expectedN15X || newN15Y != expectedN15Y {
			t.Errorf("n15 未正确跟随 GearXY 移动: 期望 (%.0f, %.0f), 实际 (%.0f, %.0f)",
				expectedN15X, expectedN15Y, newN15X, newN15Y)
		} else {
			t.Logf("n15 正确跟随 GearXY 移动 ✓")
		}

		// 验证相对位置关系保持不变
		relN13X := newN13X - newN16X
		relN13Y := newN13Y - newN16Y
		expectedRelN13X := initialN13X - initialN16X
		expectedRelN13Y := initialN13Y - initialN16Y

		if relN13X != expectedRelN13X || relN13Y != expectedRelN13Y {
			t.Errorf("n13 相对于 n16 的位置改变了: 期望相对位置 (%.0f, %.0f), 实际 (%.0f, %.0f)",
				expectedRelN13X, expectedRelN13Y, relN13X, relN13Y)
		} else {
			t.Logf("n13 相对于 n16 的位置保持不变 ✓")
		}
	})
}
