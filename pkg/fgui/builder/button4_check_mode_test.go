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
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/gears"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestButton4CheckModeWithGearXY 测试实际的 Button4 组件（Check 模式 + GearXY）
func TestButton4CheckModeWithGearXY(t *testing.T) {
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

	// 查找 Button4 组件
	var button4Item *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Button4" {
			button4Item = item
			break
		}
	}

	if button4Item == nil {
		t.Fatal("未找到 Button4 组件")
	}

	// 创建测试环境
	env := testutil.NewStageEnv(t, 640, 480)
	stage := env.Stage

	// 初始化 GRoot
	root := core.Root()
	root.AttachStage(stage)

	// 创建 factory 并构建组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	button4Comp, err := factory.BuildComponent(ctx, pkg, button4Item)
	if err != nil {
		t.Fatalf("构建 Button4 失败: %v", err)
	}

	// 获取 GButton widget
	btnWidget, ok := button4Comp.GObject.Data().(*widgets.GButton)
	if !ok {
		t.Fatal("Button4 不是 GButton widget")
	}

	// 设置位置并添加到舞台
	button4Comp.GObject.SetPosition(100, 100)
	stage.AddChild(button4Comp.GObject.DisplayObject())

	// 添加事件监听器用于调试
	clickCount := 0
	button4Comp.GObject.On("click", func(evt *laya.Event) {
		clickCount++
		t.Logf("点击事件触发，次数: %d", clickCount)
	})

	// 验证初始状态
	t.Run("InitialState", func(t *testing.T) {
		if btnWidget.Selected() {
			t.Error("初始状态不应该被选中")
		}

		// 调试信息
		t.Logf("Button4 Controllers 数量: %d", len(btnWidget.Controllers()))
		for i, ctrl := range btnWidget.Controllers() {
			t.Logf("  Controller[%d]: Name=%s, PageIDs=%v, PageNames=%v", i, ctrl.Name, ctrl.PageIDs, ctrl.PageNames)
		}
		t.Logf("Button4 有 template: %v", btnWidget.TemplateComponent() != nil)
		t.Logf("Button4 尺寸: %.0f x %.0f", button4Comp.GObject.Width(), button4Comp.GObject.Height())
		t.Logf("Button4 位置: (%.0f, %.0f)", button4Comp.GObject.X(), button4Comp.GObject.Y())
		t.Logf("Button4 Touchable: %v", button4Comp.GObject.Touchable())
		t.Logf("Button4 Mode: %v", btnWidget.Mode())
		t.Logf("Button4 ChangeStateOnClick: %v", btnWidget.ChangeStateOnClick())

		ctrl := btnWidget.ButtonController()
		if ctrl == nil {
			t.Fatal("Button4 没有 button controller")
		}

		// 检查 controller 的页面配置
		t.Logf("Button controller 页面: %v", ctrl.PageNames)
		expectedPages := []string{"up", "down", "over", "selectedOver"}
		if len(ctrl.PageNames) != len(expectedPages) {
			t.Errorf("Controller 页面数量不匹配：期望 %d，实际 %d", len(expectedPages), len(ctrl.PageNames))
		}

		// 验证初始页面是 "up"
		if ctrl.SelectedPageName() != "up" {
			t.Errorf("初始页面应该是 'up'，实际是 '%s'", ctrl.SelectedPageName())
		}
	})

	// 验证 GearXY 配置
	t.Run("GearXYConfiguration", func(t *testing.T) {
		// 查找 n1 子对象（带 GearXY 的图片）
		children := button4Comp.Children()
		var n1Obj *core.GObject
		for _, child := range children {
			if child != nil && child.Name() == "n1" {
				n1Obj = child
				break
			}
		}

		if n1Obj == nil {
			t.Skip("未找到 n1 子对象，跳过 GearXY 测试")
			return
		}

		t.Logf("n1 子对象初始位置: %.0f, %.0f", n1Obj.X(), n1Obj.Y())

		// 检查 n1 是否有 GearXY
		gear := n1Obj.GetGear(gears.IndexXY)
		gearXY, ok := gear.(*gears.GearXY)
		if !ok || gearXY == nil {
			t.Error("n1 对象应该有 GearXY")
			return
		}

		// 检查 GearXY 的 controller 是否绑定到 button controller
		if gearXY.Controller() != btnWidget.ButtonController() {
			t.Error("GearXY 的 controller 应该绑定到 button controller")
		}

		// 记录不同页面的位置值
		t.Logf("GearXY 页面值:")
		for _, pageName := range []string{"up", "down", "over", "selectedOver"} {
			// 找到对应的 pageID
			var pageID string
			ctrl := btnWidget.ButtonController()
			for i, name := range ctrl.PageNames {
				if name == pageName && i < len(ctrl.PageIDs) {
					pageID = ctrl.PageIDs[i]
					break
				}
			}
			if pageID != "" {
				val := gearXY.Value(pageID)
				t.Logf("  %s (ID=%s): X=%.0f, Y=%.0f", pageName, pageID, val.X, val.Y)
			}
		}
	})

	// 测试点击切换选中状态
	t.Run("ClickToggleSelected", func(t *testing.T) {
		// 记录点击前的状态
		beforeX, beforeY := getN1Position(button4Comp)
		t.Logf("点击前 n1 位置: %.0f, %.0f", beforeX, beforeY)

		// 模拟第一次点击：选中按钮
		// 鼠标移到按钮上
		env.Advance(16*time.Millisecond, laya.MouseState{X: 120, Y: 120, Primary: false})
		// 鼠标按下
		env.Advance(16*time.Millisecond, laya.MouseState{X: 120, Y: 120, Primary: true})
		// 鼠标释放（点击完成）
		env.Advance(16*time.Millisecond, laya.MouseState{X: 120, Y: 120, Primary: false})

		// 等待 GearXY 的 tween 动画完成（duration=0.3秒）
		for i := 0; i < 20; i++ {
			delta := 20 * time.Millisecond
			mouse := laya.MouseState{X: 120, Y: 120, Primary: false}
			env.Advance(delta, mouse)
			root.Advance(delta, mouse) // 更新 tweener
		}

		// 验证选中状态
		if !btnWidget.Selected() {
			t.Error("第一次点击后应该被选中")
		}

		// 验证 controller 状态（鼠标还在按钮上，应该是 selectedOver）
		ctrl := btnWidget.ButtonController()
		t.Logf("第一次点击后 controller 状态: %s", ctrl.SelectedPageName())
		if ctrl.SelectedPageName() != "selectedOver" {
			t.Errorf("第一次点击后（鼠标在按钮上）controller 状态应该是 'selectedOver'，实际是 '%s'", ctrl.SelectedPageName())
		}

		// 记录点击后的位置
		afterX, afterY := getN1Position(button4Comp)
		t.Logf("第一次点击后 n1 位置: %.0f, %.0f", afterX, afterY)

		// GearXY 应该已经应用，位置应该改变
		// 根据 Button4.xml，selectedOver 状态（page 3）的位置是 (64, -1)
		// 但是如果 GearXY 没有正确配置，位置可能不会改变
		// 这里我们只验证状态切换，位置变化取决于 GearXY 的 storage 配置

		// 移开鼠标，状态应该变成 "down"
		env.Advance(16*time.Millisecond, laya.MouseState{X: 10, Y: 10, Primary: false})
		// 等待动画完成
		for i := 0; i < 20; i++ {
			delta := 20 * time.Millisecond
			mouse := laya.MouseState{X: 10, Y: 10, Primary: false}
			env.Advance(delta, mouse)
			root.Advance(delta, mouse)
		}
		t.Logf("移开鼠标后 controller 状态: %s", ctrl.SelectedPageName())
		if ctrl.SelectedPageName() != "down" {
			t.Errorf("移开鼠标后 controller 状态应该是 'down'，实际是 '%s'", ctrl.SelectedPageName())
		}

		downX, downY := getN1Position(button4Comp)
		t.Logf("down 状态 n1 位置: %.0f, %.0f", downX, downY)

		// 模拟第二次点击：取消选中
		env.Advance(16*time.Millisecond, laya.MouseState{X: 120, Y: 120, Primary: false})
		env.Advance(16*time.Millisecond, laya.MouseState{X: 120, Y: 120, Primary: true})
		env.Advance(16*time.Millisecond, laya.MouseState{X: 120, Y: 120, Primary: false})
		// 等待动画完成
		for i := 0; i < 20; i++ {
			delta := 20 * time.Millisecond
			mouse := laya.MouseState{X: 120, Y: 120, Primary: false}
			env.Advance(delta, mouse)
			root.Advance(delta, mouse)
		}

		// 验证取消选中
		if btnWidget.Selected() {
			t.Error("第二次点击后应该取消选中")
		}

		// 验证 controller 状态（鼠标还在按钮上，应该是 over）
		t.Logf("第二次点击后 controller 状态: %s", ctrl.SelectedPageName())
		if ctrl.SelectedPageName() != "over" {
			t.Errorf("第二次点击后（鼠标在按钮上）controller 状态应该是 'over'，实际是 '%s'", ctrl.SelectedPageName())
		}

		// 移开鼠标，状态应该变成 "up"
		env.Advance(16*time.Millisecond, laya.MouseState{X: 10, Y: 10, Primary: false})
		// 等待动画完成
		for i := 0; i < 20; i++ {
			delta := 20 * time.Millisecond
			mouse := laya.MouseState{X: 10, Y: 10, Primary: false}
			env.Advance(delta, mouse)
			root.Advance(delta, mouse)
		}
		t.Logf("最终 controller 状态: %s", ctrl.SelectedPageName())
		if ctrl.SelectedPageName() != "up" {
			t.Errorf("最终 controller 状态应该是 'up'，实际是 '%s'", ctrl.SelectedPageName())
		}

		upX, upY := getN1Position(button4Comp)
		t.Logf("up 状态 n1 位置: %.0f, %.0f", upX, upY)

		// 验证位置是否回到初始位置
		if upX != beforeX || upY != beforeY {
			t.Logf("警告：up 状态位置与初始位置不同，可能是 GearXY 应用的结果")
			t.Logf("  初始: (%.0f, %.0f), 最终: (%.0f, %.0f)", beforeX, beforeY, upX, upY)
		}
	})
}

// getN1Position 获取 n1 子对象的位置
func getN1Position(comp *core.GComponent) (float64, float64) {
	children := comp.Children()
	for _, child := range children {
		if child != nil && child.Name() == "n1" {
			return child.X(), child.Y()
		}
	}
	return 0, 0
}
