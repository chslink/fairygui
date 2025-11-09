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

// TestButtonRadioGroup 单选按钮组功能测试
// 场景：Demo_Button.xml 中的 n18, n19, n20 配置了 RadioGroup
// n18 配置了 checked="true"，应该是初始选中的选项
// 点击 n19 或 n20 时，n18 应该取消选中
func TestButtonRadioGroup(t *testing.T) {
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

	// 测试初始状态
	t.Run("InitialState", func(t *testing.T) {
		// 获取三个RadioButton
		n18 := demoButton.ChildByName("n18")
		n19 := demoButton.ChildByName("n19")
		n20 := demoButton.ChildByName("n20")

		if n18 == nil || n19 == nil || n20 == nil {
			t.Fatalf("未找到 RadioButton 组件")
		}

		btn18, ok := n18.Data().(*widgets.GButton)
		btn19, ok := n19.Data().(*widgets.GButton)
		btn20, ok := n20.Data().(*widgets.GButton)

		if !ok || btn18 == nil || btn19 == nil || btn20 == nil {
			t.Fatalf("组件不是 GButton")
		}

		t.Logf("n18 初始状态:")
		t.Logf("  selected: %v (预期: true)", btn18.Selected())
		t.Logf("  mode: %d (ButtonModeRadio=%d)", btn18.Mode(), widgets.ButtonModeRadio)
		t.Logf("  relatedController: %v", btn18.RelatedController() != nil)
		t.Logf("  relatedPageID: %s", btn18.RelatedPageID())

		// 验证 n18 被选中
		if !btn18.Selected() {
			t.Errorf("n18 配置了 checked=\"true\"，应该被选中，实际: %v", btn18.Selected())
		}

		// 验证按钮模式为 Radio
		if btn18.Mode() != widgets.ButtonModeRadio {
			t.Errorf("n18 应该是 Radio 模式，实际: %d", btn18.Mode())
		}

		// 验证 n19, n20 未被选中
		if btn19.Selected() {
			t.Errorf("n19 没有 checked 属性，不应该被选中，实际: %v", btn19.Selected())
		}
		if btn20.Selected() {
			t.Errorf("n20 没有 checked 属性，不应该被选中，实际: %v", btn20.Selected())
		}

		// 验证有相关的控制器
		if btn18.RelatedController() == nil {
			t.Errorf("n18 应该配置了 controller")
		}
		if btn18.RelatedPageID() != "0" {
			t.Errorf("n18 的 page 应该是 \"0\"，实际: %s", btn18.RelatedPageID())
		}
	})

	// 测试点击切换
	t.Run("ClickToggle", func(t *testing.T) {
		// 获取按钮
		n18 := demoButton.ChildByName("n18")
		n19 := demoButton.ChildByName("n19")

		btn18, ok := n18.Data().(*widgets.GButton)
		btn19, ok := n19.Data().(*widgets.GButton)

		if !ok || btn18 == nil || btn19 == nil {
			t.Fatalf("组件不是 GButton")
		}

		// 初始状态：n18 选中，n19 未选中
		if !btn18.Selected() {
			t.Fatalf("测试前提：n18 应该初始选中")
		}
		if btn19.Selected() {
			t.Fatalf("测试前提：n19 应该初始未选中")
		}

		// 点击 n19
		clickX := 408.0 + 43.0 // n19.x + center offset
		clickY := 190.0 + 9.5  // n19.y + center offset

		// 按下
		env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: true})
		// 弹起
		env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})

		t.Logf("点击 n19 后:")
		t.Logf("  n18.selected: %v (预期: false)", btn18.Selected())
		t.Logf("  n19.selected: %v (预期: true)", btn19.Selected())

		// 验证切换
		if btn18.Selected() {
			t.Errorf("点击 n19 后，n18 应该取消选中")
		}
		if !btn19.Selected() {
			t.Errorf("点击 n19 后，n19 应该被选中")
		}

		// 再次点击 n18
		clickX = 408.0 + 43.0 // n18.x + center offset
		clickY = 158.0 + 9.5  // n18.y + center offset

		// 按下
		env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: true})
		// 弹起
		env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})

		t.Logf("再次点击 n18 后:")
		t.Logf("  n18.selected: %v (预期: true)", btn18.Selected())
		t.Logf("  n19.selected: %v (预期: false)", btn19.Selected())

		// 验证切回
		if !btn18.Selected() {
			t.Errorf("再次点击 n18 后，n18 应该被选中")
		}
		if btn19.Selected() {
			t.Errorf("再次点击 n18 后，n19 应该取消选中")
		}
	})
}
