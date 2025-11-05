package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// TestGGroupRenderDebug 调试 GGroup 渲染问题
func TestGGroupRenderDebug(t *testing.T) {
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

	// 输出所有子对象信息
	t.Logf("=== Demo_Controller 子对象信息 ===")
	t.Logf("总共 %d 个子对象", len(rootComponent.Children()))

	for i, child := range rootComponent.Children() {
		if child == nil {
			continue
		}

		name := child.Name()
		visible := child.Visible()
		x, y := child.X(), child.Y()
		w, h := child.Width(), child.Height()
		group := child.Group()
		groupName := ""
		if group != nil {
			groupName = group.Name()
		}

		displayObj := child.DisplayObject()
		hasDisplay := displayObj != nil
		hasGraphics := false
		if hasDisplay && displayObj.Graphics() != nil {
			hasGraphics = !displayObj.Graphics().IsEmpty()
		}

		t.Logf("[%d] %s: visible=%v, pos=(%.0f,%.0f), size=(%.0f,%.0f), group=%s, display=%v, graphics=%v, type=%T",
			i, name, visible, x, y, w, h, groupName, hasDisplay, hasGraphics, child.Data())
	}

	// 检查 c2 controller 状态
	c2 := rootComponent.ControllerByName("c2")
	if c2 != nil {
		t.Logf("\n=== c2 Controller 信息 ===")
		t.Logf("当前页面: %d (%s)", c2.SelectedIndex(), c2.SelectedPageName())
		t.Logf("页面列表: IDs=%v, Names=%v", c2.PageIDs, c2.PageNames)

		// 测试切换到 page 1
		c2.SetSelectedIndex(1)
		t.Logf("\n切换到 page 1 后:")

		n16 := rootComponent.ChildByName("n16")
		n13 := rootComponent.ChildByName("n13")
		n14 := rootComponent.ChildByName("n14")
		n15 := rootComponent.ChildByName("n15")

		if n16 != nil {
			t.Logf("n16 (GGroup): visible=%v", n16.Visible())
		}
		if n13 != nil {
			t.Logf("n13: visible=%v, pos=(%.0f,%.0f)", n13.Visible(), n13.X(), n13.Y())
			if displayObj := n13.DisplayObject(); displayObj != nil {
				gfx := displayObj.Graphics()
				if gfx != nil {
					t.Logf("  - Graphics: empty=%v, commands=%d", gfx.IsEmpty(), len(gfx.Commands()))
					for i, cmd := range gfx.Commands() {
						t.Logf("    [%d] type=%d", i, cmd.Type)
					}
				}
			}
		}
		if n14 != nil {
			t.Logf("n14: visible=%v, pos=(%.0f,%.0f)", n14.Visible(), n14.X(), n14.Y())
		}
		if n15 != nil {
			t.Logf("n15: visible=%v, pos=(%.0f,%.0f)", n15.Visible(), n15.X(), n15.Y())
		}
	}
}
