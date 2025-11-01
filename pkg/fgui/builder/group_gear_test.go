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

// TestGroupGearDisplay 测试 Group 的 gearDisplay 功能
// 验证当 Group 的可见性改变时，属于该 Group 的子元素也会相应改变
func TestGroupGearDisplay(t *testing.T) {
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
	_ = env.Stage

	// 构建 Demo_Controller 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 Demo_Controller 组件
	var demoItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_Controller" {
			demoItem = item
			break
		}
	}
	if demoItem == nil {
		t.Fatal("找不到 Demo_Controller 组件")
	}

	// 创建组件实例
	rootComponent, err := factory.BuildComponent(ctx, pkg, demoItem)
	if err != nil {
		t.Fatalf("构建 Demo_Controller 失败: %v", err)
	}

	// 查找 c2 控制器
	c2 := rootComponent.ControllerByName("c2")
	if c2 == nil {
		t.Fatal("找不到 c2 控制器")
	}

	// 查找 n16 Group
	n16 := rootComponent.ChildByName("n16")
	if n16 == nil {
		t.Fatal("找不到 n16 Group")
	}
	t.Logf("n16 类型: %T", n16.Data())

	// 验证 n16 是 GGroup
	groupWidget, ok := n16.Data().(*widgets.GGroup)
	if !ok {
		t.Fatalf("n16 不是 GGroup，而是 %T", n16.Data())
	}

	// 查找属于 n16 Group 的子元素
	n13 := rootComponent.ChildByName("n13")
	n14 := rootComponent.ChildByName("n14")
	n15 := rootComponent.ChildByName("n15")

	if n13 == nil || n14 == nil || n15 == nil {
		t.Fatal("找不到 n13/n14/n15 子元素")
	}

	// 验证子元素的 group 关联
	if n13.Group() != n16 {
		t.Errorf("n13 的 group 应该是 n16，实际是 %v", n13.Group())
	}
	if n14.Group() != n16 {
		t.Errorf("n14 的 group 应该是 n16，实际是 %v", n14.Group())
	}
	if n15.Group() != n16 {
		t.Errorf("n15 的 group 应该是 n16，实际是 %v", n15.Group())
	}

	// 检查初始状态（c2 page 0）
	t.Logf("初始状态: c2.SelectedPageID=%s", c2.SelectedPageID())
	t.Logf("  n16.Visible=%v, DisplayObject.Visible=%v",
		groupWidget.GObject.Visible(), groupWidget.GObject.DisplayObject().Visible())
	t.Logf("  n13.Visible=%v, DisplayObject.Visible=%v",
		n13.Visible(), n13.DisplayObject().Visible())
	t.Logf("  n14.Visible=%v, DisplayObject.Visible=%v",
		n14.Visible(), n14.DisplayObject().Visible())
	t.Logf("  n15.Visible=%v, DisplayObject.Visible=%v",
		n15.Visible(), n15.DisplayObject().Visible())

	// 根据 XML，c2 page 0 时，n16 应该在默认位置 (661,450)
	// c2 page 1 时，n16 有 gearDisplay，应该显示；gearXY 改为 (1154,450)
	if c2.SelectedPageID() == "0" {
		// page 0: gearDisplay 没有定义，所以应该隐藏？
		// 实际上 gearDisplay controller="c2" pages="1" 表示只在 page 1 显示
		// 所以 page 0 时应该隐藏
		expectedVisible := false
		if groupWidget.GObject.Visible() != expectedVisible {
			t.Errorf("page 0 时 n16.Visible 应该是 %v，实际是 %v",
				expectedVisible, groupWidget.GObject.Visible())
		}
		// 子元素的 DisplayObject 也应该隐藏（跟随 Group）
		if n13.DisplayObject().Visible() != expectedVisible {
			t.Errorf("page 0 时 n13.DisplayObject().Visible 应该是 %v（跟随 Group），实际是 %v",
				expectedVisible, n13.DisplayObject().Visible())
		}
		if n14.DisplayObject().Visible() != expectedVisible {
			t.Errorf("page 0 时 n14.DisplayObject().Visible 应该是 %v（跟随 Group），实际是 %v",
				expectedVisible, n14.DisplayObject().Visible())
		}
		if n15.DisplayObject().Visible() != expectedVisible {
			t.Errorf("page 0 时 n15.DisplayObject().Visible 应该是 %v（跟随 Group），实际是 %v",
				expectedVisible, n15.DisplayObject().Visible())
		}
	}

	// 切换到 page 1
	c2.SetSelectedPageID("1")
	t.Logf("切换后: c2.SelectedPageID=%s", c2.SelectedPageID())
	t.Logf("  n16.Visible=%v, DisplayObject.Visible=%v",
		groupWidget.GObject.Visible(), groupWidget.GObject.DisplayObject().Visible())
	t.Logf("  n13.Visible=%v, DisplayObject.Visible=%v",
		n13.Visible(), n13.DisplayObject().Visible())
	t.Logf("  n14.Visible=%v, DisplayObject.Visible=%v",
		n14.Visible(), n14.DisplayObject().Visible())
	t.Logf("  n15.Visible=%v, DisplayObject.Visible=%v",
		n15.Visible(), n15.DisplayObject().Visible())

	// page 1: n16 应该显示
	if !groupWidget.GObject.Visible() {
		t.Errorf("page 1 时 n16.Visible 应该是 true，实际是 false")
	}
	// 子元素的 DisplayObject 也应该显示
	if !n13.DisplayObject().Visible() {
		t.Errorf("page 1 时 n13.DisplayObject().Visible 应该是 true（跟随 Group），实际是 false")
	}
	if !n14.DisplayObject().Visible() {
		t.Errorf("page 1 时 n14.DisplayObject().Visible 应该是 true（跟随 Group），实际是 false")
	}
	if !n15.DisplayObject().Visible() {
		t.Errorf("page 1 时 n15.DisplayObject().Visible 应该是 true（跟随 Group），实际是 false")
	}

	// 切换回 page 0
	c2.SetSelectedPageID("0")
	t.Logf("切换回: c2.SelectedPageID=%s", c2.SelectedPageID())
	t.Logf("  n16.Visible=%v, DisplayObject.Visible=%v",
		groupWidget.GObject.Visible(), groupWidget.GObject.DisplayObject().Visible())
	t.Logf("  n13.Visible=%v, DisplayObject.Visible=%v",
		n13.Visible(), n13.DisplayObject().Visible())
	t.Logf("  n14.Visible=%v, DisplayObject.Visible=%v",
		n14.Visible(), n14.DisplayObject().Visible())
	t.Logf("  n15.Visible=%v, DisplayObject.Visible=%v",
		n15.Visible(), n15.DisplayObject().Visible())

	// page 0: 应该隐藏
	if groupWidget.GObject.Visible() {
		t.Errorf("page 0 时 n16.Visible 应该是 false，实际是 true")
	}
	// 子元素的 DisplayObject 也应该隐藏
	if n13.DisplayObject().Visible() {
		t.Errorf("page 0 时 n13.DisplayObject().Visible 应该是 false（跟随 Group），实际是 true")
	}
	if n14.DisplayObject().Visible() {
		t.Errorf("page 0 时 n14.DisplayObject().Visible 应该是 false（跟随 Group），实际是 true")
	}
	if n15.DisplayObject().Visible() {
		t.Errorf("page 0 时 n15.DisplayObject().Visible 应该是 false（跟随 Group），实际是 true")
	}
}
