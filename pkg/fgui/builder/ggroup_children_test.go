package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestGGroupChildren 测试 GGroup 的子对象关联和渲染
func TestGGroupChildren(t *testing.T) {
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

	t.Run("VerifyGroupStructure", func(t *testing.T) {
		// 验证 n16 是 GGroup
		n16 := rootComponent.ChildByName("n16")
		if n16 == nil {
			t.Fatal("未找到 n16")
		}

		groupWidget, ok := n16.Data().(*widgets.GGroup)
		if !ok {
			t.Fatalf("n16 不是 GGroup，而是 %T", n16.Data())
		}
		t.Logf("n16 是 GGroup ✓")

		// 验证 n13、n14、n15 存在
		n13 := rootComponent.ChildByName("n13")
		n14 := rootComponent.ChildByName("n14")
		n15 := rootComponent.ChildByName("n15")

		if n13 == nil {
			t.Fatal("未找到 n13")
		}
		if n14 == nil {
			t.Fatal("未找到 n14")
		}
		if n15 == nil {
			t.Fatal("未找到 n15")
		}
		t.Logf("n13, n14, n15 都存在 ✓")

		// 验证 n13、n14、n15 的 Group() 返回 n16
		if n13.Group() != n16 {
			t.Errorf("n13.Group() = %v, 期望 n16", n13.Group())
		} else {
			t.Logf("n13.Group() == n16 ✓")
		}

		if n14.Group() != n16 {
			t.Errorf("n14.Group() = %v, 期望 n16", n14.Group())
		} else {
			t.Logf("n14.Group() == n16 ✓")
		}

		if n15.Group() != n16 {
			t.Errorf("n15.Group() = %v, 期望 n16", n15.Group())
		} else {
			t.Logf("n15.Group() == n16 ✓")
		}

		// 输出子对象信息
		t.Logf("n13 位置: (%.0f, %.0f), 尺寸: (%.0f, %.0f)", n13.X(), n13.Y(), n13.Width(), n13.Height())
		t.Logf("n14 位置: (%.0f, %.0f), 尺寸: (%.0f, %.0f)", n14.X(), n14.Y(), n14.Width(), n14.Height())
		t.Logf("n15 位置: (%.0f, %.0f), 尺寸: (%.0f, %.0f)", n15.X(), n15.Y(), n15.Width(), n15.Height())

		// 验证 DisplayObject 存在
		if n13.DisplayObject() == nil {
			t.Error("n13.DisplayObject() 为 nil")
		} else {
			t.Logf("n13 DisplayObject 存在 ✓")
		}
		if n14.DisplayObject() == nil {
			t.Error("n14.DisplayObject() 为 nil")
		} else {
			t.Logf("n14 DisplayObject 存在 ✓")
		}
		if n15.DisplayObject() == nil {
			t.Error("n15.DisplayObject() 为 nil")
		} else {
			t.Logf("n15 DisplayObject 存在 ✓")
		}

		_ = groupWidget // 使用变量避免警告
	})

	t.Run("VerifyGearDisplayEffect", func(t *testing.T) {
		// 查找 c2 controller
		c2 := rootComponent.ControllerByName("c2")
		if c2 == nil {
			t.Fatal("未找到 c2 controller")
		}

		n16 := rootComponent.ChildByName("n16")
		n13 := rootComponent.ChildByName("n13")
		n14 := rootComponent.ChildByName("n14")
		n15 := rootComponent.ChildByName("n15")

		// c2 初始页面是 0，检查初始可见性
		t.Logf("c2 初始页面: %d (%s)", c2.SelectedIndex(), c2.SelectedPageName())
		t.Logf("n16 初始 Visible: %v", n16.Visible())
		t.Logf("n13 初始 Visible: %v", n13.Visible())
		t.Logf("n14 初始 Visible: %v", n14.Visible())
		t.Logf("n15 初始 Visible: %v", n15.Visible())

		// 根据 XML，n16 有 <gearDisplay controller="c2" pages="1"/>
		// 这意味着只在 page 1 显示，page 0 应该隐藏
		if c2.SelectedIndex() == 0 {
			// page 0: n16 应该隐藏
			if n16.Visible() {
				t.Errorf("page 0: n16 应该隐藏，但 Visible = true")
			} else {
				t.Logf("page 0: n16 正确隐藏 ✓")
			}

			// 子对象应该跟随隐藏
			if n13.Visible() {
				t.Errorf("page 0: n13 应该跟随 n16 隐藏，但 Visible = true")
			} else {
				t.Logf("page 0: n13 正确隐藏 ✓")
			}
			if n14.Visible() {
				t.Errorf("page 0: n14 应该跟随 n16 隐藏，但 Visible = true")
			}
			if n15.Visible() {
				t.Errorf("page 0: n15 应该跟随 n16 隐藏，但 Visible = true")
			}
		}

		// 切换到 page 1
		c2.SetSelectedIndex(1)
		t.Logf("\n切换到 page 1")
		t.Logf("n16 Visible: %v", n16.Visible())
		t.Logf("n13 Visible: %v", n13.Visible())
		t.Logf("n14 Visible: %v", n14.Visible())
		t.Logf("n15 Visible: %v", n15.Visible())

		// page 1: n16 应该显示
		if !n16.Visible() {
			t.Errorf("page 1: n16 应该显示，但 Visible = false")
		} else {
			t.Logf("page 1: n16 正确显示 ✓")
		}

		// 子对象应该跟随显示
		if !n13.Visible() {
			t.Errorf("page 1: n13 应该跟随 n16 显示，但 Visible = false")
		} else {
			t.Logf("page 1: n13 正确显示 ✓")
		}
		if !n14.Visible() {
			t.Errorf("page 1: n14 应该跟随 n16 显示，但 Visible = false")
		}
		if !n15.Visible() {
			t.Errorf("page 1: n15 应该跟随 n16 显示，但 Visible = false")
		}

		// 切换回 page 0
		c2.SetSelectedIndex(0)
		t.Logf("\n切换回 page 0")
		t.Logf("n16 Visible: %v", n16.Visible())
		t.Logf("n13 Visible: %v", n13.Visible())

		// 再次验证隐藏
		if n16.Visible() {
			t.Errorf("page 0: n16 应该再次隐藏，但 Visible = true")
		}
		if n13.Visible() {
			t.Errorf("page 0: n13 应该再次隐藏，但 Visible = true")
		}
	})

	t.Run("VerifyChildrenInParent", func(t *testing.T) {
		// 验证 n13、n14、n15 是否在父组件的 children 列表中
		n13 := rootComponent.ChildByName("n13")
		n14 := rootComponent.ChildByName("n14")
		n15 := rootComponent.ChildByName("n15")
		n16 := rootComponent.ChildByName("n16")

		found13, found14, found15, found16 := false, false, false, false
		for _, child := range rootComponent.Children() {
			if child == n13 {
				found13 = true
			}
			if child == n14 {
				found14 = true
			}
			if child == n15 {
				found15 = true
			}
			if child == n16 {
				found16 = true
			}
		}

		if !found13 {
			t.Error("n13 不在 rootComponent.Children() 中")
		} else {
			t.Logf("n13 在父组件的 children 列表中 ✓")
		}
		if !found14 {
			t.Error("n14 不在 rootComponent.Children() 中")
		}
		if !found15 {
			t.Error("n15 不在 rootComponent.Children() 中")
		}
		if !found16 {
			t.Error("n16 不在 rootComponent.Children() 中")
		} else {
			t.Logf("n16 在父组件的 children 列表中 ✓")
		}

		// 验证 Parent() 关系
		if n13.Parent() != rootComponent {
			t.Errorf("n13.Parent() != rootComponent")
		} else {
			t.Logf("n13.Parent() == rootComponent ✓")
		}
		if n16.Parent() != rootComponent {
			t.Errorf("n16.Parent() != rootComponent")
		} else {
			t.Logf("n16.Parent() == rootComponent ✓")
		}
	})
}
