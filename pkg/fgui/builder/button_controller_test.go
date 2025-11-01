package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestButtonControllerToggle 测试 Check 模式按钮点击切换控制器
// 场景：Demo_Controller.xml 中的 n11 按钮配置了 <Button controller="c1" page="1"/>
// 预期：第一次点击切换到 page 1，第二次点击切换回 page 0
func TestButtonControllerToggle(t *testing.T) {
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
		t.Fatalf("未找到 Demo_Controller 组件")
	}

	demo, err := factory.BuildComponent(ctx, pkg, demoItem)
	if err != nil {
		t.Fatalf("构建 Demo_Controller 组件失败: %v", err)
	}

	// 将 demo 添加到 stage
	stage.AddChild(demo.DisplayObject())

	// 获取 c1 控制器
	c1 := demo.ControllerByName("c1")
	if c1 == nil {
		t.Fatalf("未找到 c1 控制器")
	}
	t.Logf("控制器 c1 初始状态: selectedIndex=%d, selectedPage=%s",
		c1.SelectedIndex(), c1.SelectedPageID())

	// 获取 n11 按钮
	n11 := demo.ChildByName("n11")
	if n11 == nil {
		t.Fatalf("未找到 n11 对象")
	}

	btn, ok := n11.Data().(*widgets.GButton)
	if !ok || btn == nil {
		t.Fatalf("n11 不是 GButton")
	}

	t.Logf("按钮 n11 初始状态: selected=%v, mode=%d, relatedController=%v, relatedPageID=%s",
		btn.Selected(), btn.Mode(), btn.RelatedController() != nil, btn.RelatedPageID())

	// 第一次点击：应该从 page 0 切换到 page 1
	clickX := 89.0 + 50.0 // n11.x + center offset
	clickY := 195.0 + 25.0

	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: true})
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})

	t.Logf("第一次点击后: c1.selectedIndex=%d, c1.selectedPage=%s, btn.selected=%v",
		c1.SelectedIndex(), c1.SelectedPageID(), btn.Selected())

	if c1.SelectedPageID() != "1" {
		t.Errorf("第一次点击后，c1 应该切换到 page 1，实际: %s", c1.SelectedPageID())
	}
	if !btn.Selected() {
		t.Errorf("第一次点击后，按钮应该被选中")
	}

	// 第二次点击：应该从 page 1 切换回 page 0
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: true})
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})

	t.Logf("第二次点击后: c1.selectedIndex=%d, c1.selectedPage=%s, btn.selected=%v",
		c1.SelectedIndex(), c1.SelectedPageID(), btn.Selected())

	if c1.SelectedPageID() != "0" {
		t.Errorf("第二次点击后，c1 应该切换回 page 0，实际: %s", c1.SelectedPageID())
	}
	if btn.Selected() {
		t.Errorf("第二次点击后，按钮应该取消选中")
	}
}
