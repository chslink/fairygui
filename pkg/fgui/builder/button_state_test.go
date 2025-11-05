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

// TestButtonControllerStateAfterGearSetup 测试 gear 设置后按钮 controller 状态
// 验证 CheckGearDisplay 不应该影响没有 gearDisplay 的按钮
func TestButtonControllerStateAfterGearSetup(t *testing.T) {
	// 加载 Basics.fui 包（包含 Button 模板）
	basicsFuiPath := filepath.Join("..", "..", "..", "demo", "assets", "Basics.fui")
	basicsFuiData, err := os.ReadFile(basicsFuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 Basics.fui 文件: %v", err)
	}

	basicsPkg, err := assets.ParsePackage(basicsFuiData, "demo/assets/Basics")
	if err != nil {
		t.Fatalf("解析 Basics.fui 文件失败: %v", err)
	}

	// 加载 MainMenu.fui 包
	mainMenuFuiPath := filepath.Join("..", "..", "..", "demo", "assets", "MainMenu.fui")
	mainMenuFuiData, err := os.ReadFile(mainMenuFuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 MainMenu.fui 文件: %v", err)
	}

	mainMenuPkg, err := assets.ParsePackage(mainMenuFuiData, "demo/assets/MainMenu")
	if err != nil {
		t.Fatalf("解析 MainMenu.fui 文件失败: %v", err)
	}

	// 创建测试 Stage 环境
	env := testutil.NewStageEnv(t, 1136, 640)
	_ = env.Stage

	// 构建 MainMenu 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(basicsPkg)
	factory.RegisterPackage(mainMenuPkg)

	ctx := context.Background()

	// 查找 MainMenu 组件
	var mainMenuItem *assets.PackageItem
	for _, item := range mainMenuPkg.Items {
		if item.Type == assets.PackageItemTypeComponent {
			if item.Name == "MainMenu" || item.Name == "Main" {
				mainMenuItem = item
			}
		}
	}
	if mainMenuItem == nil {
		t.Fatal("找不到 MainMenu 或 Main 组件")
	}

	// 创建组件实例
	rootComponent, err := factory.BuildComponent(ctx, mainMenuPkg, mainMenuItem)
	if err != nil {
		t.Fatalf("构建 MainMenu 失败: %v", err)
	}

	// 选择一个按钮测试
	btnObj := rootComponent.ChildByName("n1")
	if btnObj == nil {
		t.Fatal("找不到 n1 按钮")
	}

	btnWidget, ok := btnObj.Data().(*widgets.GButton)
	if !ok {
		t.Fatalf("btn_Button 不是 GButton")
	}

	// 记录初始状态
	ctrl := btnWidget.ButtonController()
	if ctrl == nil {
		t.Fatal("btn_Button 没有 button controller")
	}

	t.Logf("初始状态: selectedIndex=%d, selectedPage=%s",
		ctrl.SelectedIndex(), ctrl.SelectedPageID())

	// 验证初始状态是 "up"
	if ctrl.SelectedPageID() != "up" {
		t.Errorf("初始状态应该是 'up'，实际是 '%s'", ctrl.SelectedPageID())
	}

	// 手动调用 CheckGearDisplay（模拟 gear 更新）
	btnObj.CheckGearDisplay()

	// 验证调用后状态仍然是 "up"
	t.Logf("CheckGearDisplay 后: selectedIndex=%d, selectedPage=%s",
		ctrl.SelectedIndex(), ctrl.SelectedPageID())

	if ctrl.SelectedPageID() != "up" {
		t.Errorf("CheckGearDisplay 后状态应该仍然是 'up'，实际是 '%s'", ctrl.SelectedPageID())
	}

	// 验证按钮可见性
	if !btnObj.DisplayObject().Visible() {
		t.Error("CheckGearDisplay 后按钮应该可见")
	}
}
