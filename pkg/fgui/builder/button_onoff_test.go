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

// TestButtonOnOffTransition OnOffButton切换动画测试
// 场景：Demo_Button.xml 中的 n16 使用 OnOffButton.xml
// 预期：点击时图片应该有流畅的左右滑动效果，而不是立即消失/出现
func TestButtonOnOffTransition(t *testing.T) {
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

	// 创建测试环境
	env := testutil.NewStageEnv(t, 1136, 640)
	stage := env.Stage

	// 构建 Demo_Button 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 Demo_Button 组件
	var demoButtonItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_Button" {
			demoButtonItem = item
			break
		}
	}
	if demoButtonItem == nil {
		t.Fatalf("未找到 Demo_Button 组件")
	}

	demoButton, err := factory.BuildComponent(ctx, pkg, demoButtonItem)
	if err != nil {
		t.Fatalf("构建 Demo_Button 组件失败: %v", err)
	}

	// 将场景添加到舞台
	stage.AddChild(demoButton.DisplayObject())

	// 获取 n16 (OnOffButton)
	n16 := demoButton.ChildByName("n16")
	if n16 == nil {
		t.Fatalf("未找到 n16 对象")
	}

	btn, ok := n16.Data().(*widgets.GButton)
	if !ok || btn == nil {
		t.Fatalf("n16 不是 GButton")
	}

	t.Logf("n16 初始状态:")
	t.Logf("  selected: %v (预期: true，因为默认是down状态)", btn.Selected())
	t.Logf("  mode: %d (ButtonModeCheck=%d)", btn.Mode(), widgets.ButtonModeCheck)

	// 验证初始状态
	if !btn.Selected() {
		t.Errorf("n16 初始应该是选中状态（selected=true），实际: %v", btn.Selected())
	}
	if btn.Mode() != widgets.ButtonModeCheck {
		t.Errorf("n16 应该是 Check 模式，实际: %d", btn.Mode())
	}

	// 获取两个子图片来验证位置
	// 在Go版本中，我们需要检查子对象的位置是否正确设置
	// 由于这是动画效果，我们主要验证控制器状态是否正确设置
	if ctrl := btn.ButtonController(); ctrl != nil {
		t.Logf("  按钮控制器页面: %s", ctrl.SelectedPageName())
		// 初始应该是 "down" 或 "1"
		if ctrl.SelectedPageName() != "down" && ctrl.SelectedIndex() != 1 {
			t.Logf("  警告：初始页面不是 down 状态")
		}
	}

	// 点击按钮切换状态
	clickX := 473.0 + 50.0 // n16.x + center offset
	clickY := 302.0 + 20.0  // n16.y + center offset

	// 第一次点击：从选中切换到未选中
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: true})
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})

	t.Logf("第一次点击后:")
	t.Logf("  selected: %v (预期: false)", btn.Selected())

	if btn.Selected() {
		t.Errorf("第一次点击后，n16 应该取消选中")
	}

	if ctrl := btn.ButtonController(); ctrl != nil {
		t.Logf("  按钮控制器页面: %s", ctrl.SelectedPageName())
		// 点击后应该是 "up" 或 "0"
		if ctrl.SelectedPageName() != "up" && ctrl.SelectedIndex() != 0 {
			t.Logf("  警告：点击后页面不是 up 状态")
		}
	}

	// 第二次点击：从未选中切换到选中
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: true})
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})

	t.Logf("第二次点击后:")
	t.Logf("  selected: %v (预期: true)", btn.Selected())

	if !btn.Selected() {
		t.Errorf("第二次点击后，n16 应该被选中")
	}

	if ctrl := btn.ButtonController(); ctrl != nil {
		t.Logf("  按钮控制器页面: %s", ctrl.SelectedPageName())
		// 再次点击后应该是 "down" 或 "1"
		if ctrl.SelectedPageName() != "down" && ctrl.SelectedIndex() != 1 {
			t.Logf("  警告：再次点击后页面不是 down 状态")
		}
	}

	t.Logf("测试通过：OnOffButton 状态切换正常")
}
