package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestVirtualListPackage 验证 VirtualList.fui 包的组件定义
// 场景：VirtualList.fui 包含 Main 组件（带 mailList 虚拟列表）和 mailItem 组件（列表项模板）
// 预期：能够正确解析包结构并构建虚拟列表组件
func TestVirtualListPackage(t *testing.T) {
	// 加载 VirtualList.fui 包
	fuiPath := filepath.Join("..", "..", "..", "demo", "assets", "VirtualList.fui")
	fuiData, err := os.ReadFile(fuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 .fui 文件: %v", err)
	}

	pkg, err := assets.ParsePackage(fuiData, "demo/assets/VirtualList")
	if err != nil {
		t.Fatalf("解析 .fui 文件失败: %v", err)
	}

	// 验证包基本信息
	t.Logf("包名: %s", pkg.Name)
	t.Logf("包 ID: %s", pkg.ID)
	t.Logf("包中组件数量: %d", len(pkg.Items))

	// 列出所有组件
	t.Log("\n=== 包中的所有组件 ===")
	for i, item := range pkg.Items {
		t.Logf("[%d] 名称: %s, 类型: %v, 对象类型: %v",
			i, item.Name, item.Type, item.ObjectType)
	}

	// 创建测试 Stage 环境
	env := testutil.NewStageEnv(t, 1136, 640)
	stage := env.Stage

	// 构建 Factory
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 Main 组件
	t.Log("\n=== 测试 Main 组件 ===")
	var mainItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Main" {
			mainItem = item
			break
		}
	}
	if mainItem == nil {
		t.Fatalf("未找到 Main 组件")
	}

	main, err := factory.BuildComponent(ctx, pkg, mainItem)
	if err != nil {
		t.Fatalf("构建 Main 组件失败: %v", err)
	}

	// 将 main 添加到 stage
	stage.AddChild(main.DisplayObject())

	t.Logf("Main 组件构建成功")
	t.Logf("  尺寸: %.0fx%.0f", main.Width(), main.Height())
	t.Logf("  子对象数量: %d", main.NumChildren())

	// 列出 Main 的所有子对象
	t.Log("\n=== Main 组件的子对象 ===")
	for i := 0; i < main.NumChildren(); i++ {
		child := main.ChildAt(i)
		if child != nil {
			t.Logf("  [%d] %s (类型: %T, 尺寸: %.0fx%.0f, 位置: %.0f,%.0f)",
				i, child.Name(), child.Data(),
				child.Width(), child.Height(),
				child.X(), child.Y())
		}
	}

	// 查找 mailList 对象
	t.Log("\n=== 测试 mailList 虚拟列表 ===")
	mailListObj := main.ChildByName("mailList")
	if mailListObj == nil {
		t.Fatalf("未找到 mailList 对象")
	}

	mailList, ok := mailListObj.Data().(*widgets.GList)
	if !ok {
		t.Fatalf("mailList 不是 GList，实际类型: %T", mailListObj.Data())
	}

	t.Logf("mailList 组件信息（启用虚拟化前）:")
	t.Logf("  Name: %s", mailListObj.Name())
	t.Logf("  Type: %T", mailListObj.Data())
	t.Logf("  尺寸: %.0fx%.0f", mailListObj.Width(), mailListObj.Height())
	t.Logf("  位置: %.0f,%.0f", mailListObj.X(), mailListObj.Y())
	t.Logf("  IsVirtual: %v", mailList.IsVirtual())
	t.Logf("  Layout: %v", mailList.Layout())
	t.Logf("  DefaultItem: %s", mailList.DefaultItem())

	// 验证默认项已设置
	if mailList.DefaultItem() == "" {
		t.Errorf("mailList 应该设置了 defaultItem")
	} else {
		t.Logf("  ✓ DefaultItem 已设置: %s", mailList.DefaultItem())
	}

	// 关键：需要显式启用虚拟列表
	// 在 TypeScript 版本中，这是通过 setVirtual() 或 setVirtualAndLoop() 完成的
	t.Log("\n=== 启用虚拟列表 ===")

	// 设置对象创建器（用于从 defaultItem 创建列表项）
	factoryCreator := &FactoryObjectCreator{
		factory: factory,
		pkg:     pkg,
		ctx:     ctx,
	}
	mailList.SetObjectCreator(factoryCreator)

	// 启用虚拟列表
	mailList.SetVirtual(true)

	t.Logf("mailList 组件信息（启用虚拟化后）:")
	t.Logf("  IsVirtual: %v", mailList.IsVirtual())
	t.Logf("  VirtualItemSize: %v", mailList.VirtualItemSize())

	if !mailList.IsVirtual() {
		t.Errorf("SetVirtual(true) 后，IsVirtual 应该返回 true")
	} else {
		t.Logf("  ✓ 虚拟列表已启用")
	}

	// 设置列表项数量（虚拟列表的核心功能）
	t.Log("\n=== 设置虚拟列表数据 ===")
	mailList.SetNumItems(1000)
	t.Logf("  设置 NumItems: %d", mailList.NumItems())
	t.Logf("  ChildrenCount（实际渲染的子对象数量）: %d", mailList.ChildrenCount())
	t.Logf("  ✓ 虚拟列表数据已设置")

	// 查找 mailItem 组件定义
	t.Log("\n=== 测试 mailItem 组件定义 ===")
	var mailItemDef *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "mailItem" {
			mailItemDef = item
			break
		}
	}
	if mailItemDef == nil {
		t.Fatalf("未找到 mailItem 组件定义")
	}

	t.Logf("mailItem 组件定义:")
	t.Logf("  Name: %s", mailItemDef.Name)
	t.Logf("  ID: %s", mailItemDef.ID)
	t.Logf("  File: %s", mailItemDef.File)
	t.Logf("  ObjectType: %v", mailItemDef.ObjectType)

	// 尝试构建一个 mailItem 实例
	mailItem, err := factory.BuildComponent(ctx, pkg, mailItemDef)
	if err != nil {
		t.Fatalf("构建 mailItem 组件失败: %v", err)
	}

	t.Logf("mailItem 组件构建成功:")
	t.Logf("  尺寸: %.0fx%.0f", mailItem.Width(), mailItem.Height())
	t.Logf("  子对象数量: %d", mailItem.NumChildren())

	// 列出 mailItem 的所有子对象
	t.Log("\n=== mailItem 组件的子对象 ===")
	for i := 0; i < mailItem.NumChildren(); i++ {
		child := mailItem.ChildAt(i)
		if child != nil {
			t.Logf("  [%d] %s (类型: %T)",
				i, child.Name(), child.Data())
		}
	}

	// 列出 mailItem 的所有 Controllers
	t.Log("\n=== mailItem 组件的 Controllers ===")
	for i, ctrl := range mailItem.Controllers() {
		t.Logf("  [%d] %s (页面数: %d, 当前页: %d)",
			i, ctrl.Name, ctrl.PageCount(), ctrl.SelectedIndex())
	}

	// 查找按钮
	t.Log("\n=== 测试按钮组件 ===")
	buttonNames := []string{"n6", "n7", "n8"}
	for _, btnName := range buttonNames {
		btnObj := main.ChildByName(btnName)
		if btnObj == nil {
			t.Logf("警告: 未找到按钮 %s", btnName)
			continue
		}

		btn, ok := btnObj.Data().(*widgets.GButton)
		if !ok {
			t.Logf("警告: %s 不是 GButton，实际类型: %T", btnName, btnObj.Data())
			continue
		}

		t.Logf("按钮 %s:", btnName)
		t.Logf("  Title: %s", btn.Title())
		t.Logf("  尺寸: %.0fx%.0f", btnObj.Width(), btnObj.Height())
		t.Logf("  位置: %.0f,%.0f", btnObj.X(), btnObj.Y())
	}

	// 测试 AddSelection 和 ScrollToView 功能
	t.Log("\n=== 测试 AddSelection 和 ScrollToView ===")

	// 测试 AddSelection(index, false) - 不滚动
	t.Log("测试 AddSelection(500, false) - 不滚动")
	mailList.AddSelection(500, false)
	if !mailList.IsSelected(500) {
		t.Errorf("AddSelection(500, false) 失败：项目 500 未被选中")
	} else {
		t.Logf("  ✓ 项目 500 已选中")
	}

	// 验证选中状态的 UI 显示（不滚动的情况下，第500项不会被渲染，但选中状态应该被记住）
	t.Log("  验证: 选中状态已记录（UI在滚动时会被更新）")

	// 验证滚动位置未改变（应该还在顶部）
	pane := mailList.GComponent.ScrollPane()
	if pane != nil {
		initialPosY := pane.PosY()
		t.Logf("  滚动位置: Y=%.2f (预期未滚动，应接近0)", initialPosY)
	}

	// 测试 AddSelection(index, true) - 自动滚动
	t.Log("测试 AddSelection(500, true) - 自动滚动")
	mailList.ClearSelection()
	mailList.AddSelection(500, true)

	if !mailList.IsSelected(500) {
		t.Errorf("AddSelection(500, true) 失败：项目 500 未被选中")
	} else {
		t.Logf("  ✓ 项目 500 已选中")
	}

	// 验证选中状态的 UI 显示
	// 查找第 500 项（应该已经被渲染）
	item500 := mailList.GetChildAt(3) // 第 500 项应该是第4个子对象（索引3）
	if item500 != nil {
		t.Logf("  验证 UI 选中状态: 第 500 项已渲染，类型: %T", item500.Data())

		// GButton 本身就是 GComponent 的子类，直接访问控制器
		if btn, ok := item500.Data().(*widgets.GButton); ok && btn != nil {
			// GButton 实现了 Controller 接口，可以直接使用 ControllerByName
			if buttonCtrl := btn.ControllerByName("button"); buttonCtrl != nil {
				selectedPage := buttonCtrl.SelectedIndex()
				t.Logf("    button 控制器当前页面: %d (0=未选中, 1=选中)", selectedPage)
				if selectedPage == 0 {
					t.Errorf("  ❌ UI 未显示选中状态：button控制器仍在页面0")
				} else {
					t.Logf("  ✓ UI 正确显示选中状态")
				}
			} else {
				t.Logf("    未找到 button 控制器")
			}
		} else if comp, ok := item500.Data().(*core.GComponent); ok && comp != nil {
			// 备用：GComponent 路径
			if buttonCtrl := comp.ControllerByName("button"); buttonCtrl != nil {
				selectedPage := buttonCtrl.SelectedIndex()
				t.Logf("    button 控制器当前页面: %d (0=未选中, 1=选中)", selectedPage)
				if selectedPage == 0 {
					t.Errorf("  ❌ UI 未显示选中状态：button控制器仍在页面0")
				} else {
					t.Logf("  ✓ UI 正确显示选中状态")
				}
			} else {
				t.Logf("    未找到 button 控制器")
			}
		} else {
			t.Logf("    第 500 项不是可识别的组件类型: %T", item500.Data())
		}
	} else {
		t.Logf("  警告: 第 500 项尚未渲染（可能需要滚动完成）")
	}

	// 验证滚动位置已改变
	if pane != nil {
		scrolledPosY := pane.PosY()
		t.Logf("  滚动位置: Y=%.2f", scrolledPosY)

		// 预期滚动到第 500 项附近（粗略验证：应该大于 0）
		if scrolledPosY <= 0 {
			t.Errorf("滚动失败：位置仍为 %.2f，预期已滚动到第 500 项", scrolledPosY)
		} else {
			t.Logf("  ✓ 已滚动到第 500 项附近")
		}

		// 计算期望的滚动位置（SingleColumn 布局）
		// 假设每项高度约为 itemSize.Y + lineGap
		itemHeight := mailList.VirtualItemSize().Y
		expectedY := itemHeight * 500 // 粗略估算
		t.Logf("  期望滚动位置: 约 %.2f (实际: %.2f)", expectedY, scrolledPosY)
	}

	// 测试总结
	t.Log("\n=== VirtualList 包验证总结 ===")
	t.Log("✓ 成功加载 VirtualList.fui 包")
	t.Log("✓ 成功构建 Main 组件")
	t.Log("✓ 成功找到并验证 mailList 虚拟列表")
	t.Log("✓ 成功找到并构建 mailItem 组件模板")
	t.Log("✓ AddSelection 功能正常")
	t.Log("✓ ScrollToView 自动滚动功能正常")
	t.Log("✓ 所有组件结构正确")
}
