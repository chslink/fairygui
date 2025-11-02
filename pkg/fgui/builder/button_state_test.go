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

// TestMainMenuButtonControllerState 测试 MainMenu 场景下按钮的 controller 状态
// 验证按钮的初始状态应该是 "up" 而不是其他状态
func TestMainMenuButtonControllerState(t *testing.T) {
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
	t.Logf("MainMenu 包中的组件：")
	for _, item := range mainMenuPkg.Items {
		if item.Type == assets.PackageItemTypeComponent {
			t.Logf("  - %s (ID: %s)", item.Name, item.ID)
			if item.Name == "MainMenu" || item.Name == "Main" {
				mainMenuItem = item
			}
		}
	}
	if mainMenuItem == nil {
		t.Fatal("找不到 MainMenu 或 Main 组件")
	}
	t.Logf("使用组件: %s", mainMenuItem.Name)

	// 创建组件实例
	rootComponent, err := factory.BuildComponent(ctx, mainMenuPkg, mainMenuItem)
	if err != nil {
		t.Fatalf("构建 MainMenu 失败: %v", err)
	}

	// 打印所有子元素
	t.Logf("Main 组件的子元素：")
	for _, child := range rootComponent.Children() {
		if child != nil {
			t.Logf("  - %s (type: %T)", child.Name(), child.Data())
		}
	}

	// 测试所有按钮 (Main 组件中的按钮名称是 n1-n16)
	buttonTests := []struct {
		name           string
		expectedState  string
		expectedIndex  int
		expectedVisible bool
	}{
		{name: "n1", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n2", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n4", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n5", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n6", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n7", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n8", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n9", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n10", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n11", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n12", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n13", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n14", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n15", expectedState: "up", expectedIndex: 0, expectedVisible: true},
		{name: "n16", expectedState: "up", expectedIndex: 0, expectedVisible: true},
	}

	for _, tt := range buttonTests {
		t.Run(tt.name, func(t *testing.T) {
			// 查找按钮
			btnObj := rootComponent.ChildByName(tt.name)
			if btnObj == nil {
				t.Fatalf("找不到按钮 %s", tt.name)
			}

			// 验证按钮可见性
			if btnObj.Visible() != tt.expectedVisible {
				t.Errorf("%s.Visible() = %v, 期望 %v",
					tt.name, btnObj.Visible(), tt.expectedVisible)
			}

			// 验证 DisplayObject 可见性
			if btnObj.DisplayObject().Visible() != tt.expectedVisible {
				t.Errorf("%s.DisplayObject().Visible() = %v, 期望 %v",
					tt.name, btnObj.DisplayObject().Visible(), tt.expectedVisible)
			}

			// 获取按钮 widget
			btnWidget, ok := btnObj.Data().(*widgets.GButton)
			if !ok {
				t.Fatalf("%s 不是 GButton，而是 %T", tt.name, btnObj.Data())
			}

			// 验证按钮有 controller
			ctrl := btnWidget.ButtonController()
			if ctrl == nil {
				t.Fatalf("%s 没有 button controller", tt.name)
			}

			// 验证 controller 的初始状态
			t.Logf("%s: controller.Name=%s, selectedIndex=%d, selectedPage=%s, pageCount=%d",
				tt.name,
				ctrl.Name,
				ctrl.SelectedIndex(),
				ctrl.SelectedPageID(),
				ctrl.PageCount())

			// 打印 PageIDs 和 PageNames 以调试
			t.Logf("  PageIDs: %v", ctrl.PageIDs)
			t.Logf("  PageNames: %v", ctrl.PageNames)

			if ctrl.SelectedIndex() != tt.expectedIndex {
				t.Errorf("%s controller 的 selectedIndex = %d, 期望 %d",
					tt.name, ctrl.SelectedIndex(), tt.expectedIndex)
			}

			if ctrl.SelectedPageID() != tt.expectedState {
				t.Errorf("%s controller 的 selectedPage = %s, 期望 %s",
					tt.name, ctrl.SelectedPageID(), tt.expectedState)
			}

			// 验证模板的 DisplayObject 可见性
			template := btnWidget.TemplateComponent()
			if template != nil {
				t.Logf("  template.Visible=%v, DisplayObject.Visible=%v",
					template.GObject.Visible(), template.GObject.DisplayObject().Visible())

				// 模板应该可见
				if !template.GObject.DisplayObject().Visible() {
					t.Errorf("%s 的模板 DisplayObject 应该可见，但实际是隐藏的", tt.name)
				}
			}
		})
	}
}

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
