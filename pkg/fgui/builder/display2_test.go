package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// TestGearDisplay2 测试 GearDisplay2 的组合可见性逻辑
// n18 同时受 c1 和 c2 两个控制器影响
func TestGearDisplay2(t *testing.T) {
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

	// 查找控制器
	c1 := rootComponent.ControllerByName("c1")
	c2 := rootComponent.ControllerByName("c2")
	if c1 == nil || c2 == nil {
		t.Fatal("找不到 c1 或 c2 控制器")
	}

	// 查找 n18
	n18 := rootComponent.ChildByName("n18")
	if n18 == nil {
		t.Fatal("找不到 n18")
	}

	// n18 的配置：
	// <gearDisplay controller="c1" pages="1"/>
	// <gearDisplay2 controller="c2" pages="1" condition="0"/>
	// 意思是：只有当 c1=page1 AND c2=page1 时才显示

	// 初始状态：c1=0, c2=0
	t.Logf("初始状态: c1=%s, c2=%s", c1.SelectedPageID(), c2.SelectedPageID())
	t.Logf("  n18.Visible=%v, DisplayObject.Visible=%v",
		n18.Visible(), n18.DisplayObject().Visible())

	// 应该不可见（c1 不在 page 1）
	if n18.DisplayObject().Visible() {
		t.Errorf("初始状态 (c1=0, c2=0) n18 应该不可见，实际可见")
	}

	// 切换 c1 到 page 1，c2 仍是 page 0
	c1.SetSelectedPageID("1")
	t.Logf("切换 c1: c1=%s, c2=%s", c1.SelectedPageID(), c2.SelectedPageID())
	t.Logf("  n18.Visible=%v, DisplayObject.Visible=%v",
		n18.Visible(), n18.DisplayObject().Visible())

	// 应该不可见（c2 不在 page 1，condition=0 是 AND）
	if n18.DisplayObject().Visible() {
		t.Errorf("c1=1, c2=0 时 n18 应该不可见（需要 AND 条件），实际可见")
	}

	// 切换 c2 到 page 1，c1 是 page 1
	c2.SetSelectedPageID("1")
	t.Logf("切换 c2: c1=%s, c2=%s", c1.SelectedPageID(), c2.SelectedPageID())
	t.Logf("  n18.Visible=%v, DisplayObject.Visible=%v",
		n18.Visible(), n18.DisplayObject().Visible())

	// 应该可见（c1=page1 AND c2=page1）
	if !n18.DisplayObject().Visible() {
		t.Errorf("c1=1, c2=1 时 n18 应该可见，实际不可见")
	}

	// 切换 c1 回 page 0，c2 仍是 page 1
	c1.SetSelectedPageID("0")
	t.Logf("切换回 c1: c1=%s, c2=%s", c1.SelectedPageID(), c2.SelectedPageID())
	t.Logf("  n18.Visible=%v, DisplayObject.Visible=%v",
		n18.Visible(), n18.DisplayObject().Visible())

	// 应该不可见（c1 不在 page 1）
	if n18.DisplayObject().Visible() {
		t.Errorf("c1=0, c2=1 时 n18 应该不可见，实际可见")
	}

	// 切换 c2 回 page 0，c1 是 page 0
	c2.SetSelectedPageID("0")
	t.Logf("全部切换回: c1=%s, c2=%s", c1.SelectedPageID(), c2.SelectedPageID())
	t.Logf("  n18.Visible=%v, DisplayObject.Visible=%v",
		n18.Visible(), n18.DisplayObject().Visible())

	// 应该不可见
	if n18.DisplayObject().Visible() {
		t.Errorf("c1=0, c2=0 时 n18 应该不可见，实际可见")
	}
}
