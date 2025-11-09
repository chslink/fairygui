package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestComboBoxComponentAccess 验证 ComboBox 组件的构建和获取
// 场景：Demo_ComboBox.xml 中的 n1 组件引用了 rt103t 模板
// 预期：能够正确构建和访问 ComboBox 组件
func TestComboBoxComponentAccess(t *testing.T) {
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

	// 构建 Factory
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 Demo_ComboBox 组件
	var demoItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_ComboBox" {
			demoItem = item
			break
		}
	}
	if demoItem == nil {
		t.Fatalf("未找到 Demo_ComboBox 组件")
	}

	demo, err := factory.BuildComponent(ctx, pkg, demoItem)
	if err != nil {
		t.Fatalf("构建 Demo_ComboBox 组件失败: %v", err)
	}

	// 将 demo 添加到 stage
	stage.AddChild(demo.DisplayObject())

	// 测试获取 n1 组件
	t.Run("n1 ComboBox", func(t *testing.T) {
		n1 := demo.ChildByName("n1")
		if n1 == nil {
			t.Fatalf("未找到 n1 对象")
		}

		// 检查 n1 的类型
		combobox, ok := n1.Data().(*widgets.GComboBox)
		if !ok {
			t.Fatalf("n1 不是 GComboBox，实际类型: %T", n1.Data())
		}

		t.Logf("n1 组件信息:")
		t.Logf("  Name: %s", n1.Name())
		t.Logf("  Type: %T", n1.Data())
		t.Logf("  Touchable: %v", n1.Touchable())
		t.Logf("  ComboBox items: %d", len(combobox.Items()))
		t.Logf("  ComboBox selectedIndex: %d", combobox.SelectedIndex())
		t.Logf("  ComboBox visibleItemCount: %d", combobox.VisibleItemCount())

		// 关键：检查dropdown是否被正确设置
		if combobox.Dropdown() == nil {
			t.Errorf("n1 ComboBox 没有设置 dropdown 组件")
		} else {
			t.Logf("✓ n1 ComboBox dropdown 已设置: %s", combobox.Dropdown().GObject.Name())
			t.Logf("  dropdown 子元素数量: %d", combobox.Dropdown().NumChildren())
			for i := 0; i < combobox.Dropdown().NumChildren(); i++ {
				child := combobox.Dropdown().ChildAt(i)
				t.Logf("    子元素[%d]: %s (类型: %T)", i, child.Name(), child.Data())
			}
		}

		// 关键：检查list是否被正确设置
		if combobox.List() == nil {
			t.Errorf("n1 ComboBox 没有设置 list 组件")
		} else {
			t.Logf("✓ n1 ComboBox list 已设置，项目数: %d", combobox.List().NumItems())
			t.Logf("  list 默认item: %s", combobox.List().DefaultItem())
		}

		// 验证 ComboBox 基本属性
		items := combobox.Items()
		if len(items) == 0 {
			t.Errorf("ComboBox 应该有 8 个项目，实际: %d", len(items))
		} else {
			t.Logf("✓ ComboBox 项目数量正确: %d", len(items))
		}

		// 验证可见项目数
		if combobox.VisibleItemCount() != 10 {
			t.Errorf("ComboBox 可见项目数应该是 10，实际: %d", combobox.VisibleItemCount())
		} else {
			t.Logf("✓ ComboBox 可见项目数正确: %d", combobox.VisibleItemCount())
		}

		// 测试获取第一个项目
		if len(items) > 0 {
			firstItem := items[0]
			if firstItem == "" {
				t.Errorf("第一个项目不应该为空")
			} else {
				t.Logf("✓ 第一个项目: '%s'", firstItem)
			}
		}
	})

	// 测试获取 n4 组件
	t.Run("n4 ComboBox", func(t *testing.T) {
		n4 := demo.ChildByName("n4")
		if n4 == nil {
			t.Fatalf("未找到 n4 对象")
		}

		combobox, ok := n4.Data().(*widgets.GComboBox)
		if !ok {
			t.Fatalf("n4 不是 GComboBox，实际类型: %T", n4.Data())
		}

		t.Logf("n4 组件信息:")
		t.Logf("  ComboBox items: %d", len(combobox.Items()))
		t.Logf("  ComboBox visibleItemCount: %d", combobox.VisibleItemCount())

		if combobox.VisibleItemCount() != 5 {
			t.Errorf("n4 ComboBox 可见项目数应该是 5，实际: %d", combobox.VisibleItemCount())
		} else {
			t.Logf("✓ n4 ComboBox 可见项目数正确")
		}
	})

	// 测试获取 n5 组件
	t.Run("n5 ComboBox", func(t *testing.T) {
		n5 := demo.ChildByName("n5")
		if n5 == nil {
			t.Fatalf("未找到 n5 对象")
		}

		combobox, ok := n5.Data().(*widgets.GComboBox)
		if !ok {
			t.Fatalf("n5 不是 GComboBox，实际类型: %T", n5.Data())
		}

		t.Logf("n5 组件信息:")
		t.Logf("  ComboBox items: %d", len(combobox.Items()))
		t.Logf("  ComboBox visibleItemCount: %d", combobox.VisibleItemCount())

		if combobox.VisibleItemCount() != 10 {
			t.Errorf("n5 ComboBox 可见项目数应该是 10，实际: %d", combobox.VisibleItemCount())
		} else {
			t.Logf("✓ n5 ComboBox 可见项目数正确")
		}
	})

	// 测试获取 n6 组件
	t.Run("n6 ComboBox", func(t *testing.T) {
		n6 := demo.ChildByName("n6")
		if n6 == nil {
			t.Fatalf("未找到 n6 对象")
		}

		combobox, ok := n6.Data().(*widgets.GComboBox)
		if !ok {
			t.Fatalf("n6 不是 GComboBox，实际类型: %T", n6.Data())
		}

		t.Logf("n6 组件信息:")
		t.Logf("  ComboBox items: %d", len(combobox.Items()))
		t.Logf("  ComboBox visibleItemCount: %d", combobox.VisibleItemCount())

		if combobox.VisibleItemCount() != 5 {
			t.Errorf("n6 ComboBox 可见项目数应该是 5，实际: %d", combobox.VisibleItemCount())
		} else {
			t.Logf("✓ n6 ComboBox 可见项目数正确")
		}
	})

	// 测试触发下拉显示
	t.Run("TestShowDropdown", func(t *testing.T) {
		// 查找 n1 ComboBox
		n1 := demo.ChildByName("n1")
		if n1 == nil {
			t.Fatalf("未找到 n1 对象")
		}

		combobox, ok := n1.Data().(*widgets.GComboBox)
		if !ok {
			t.Fatalf("n1 不是 GComboBox")
		}

		t.Log("=== 触发 ComboBox 下拉显示 ===")
		// 模拟鼠标点击事件，触发下拉
		combobox.ShowDropdown()

		// 等待一下让操作完成
		time.Sleep(100 * time.Millisecond)

		// 检查列表项数量
		if combobox.List() != nil {
			t.Logf("下拉后列表项数量: %d", len(combobox.Items()))
			if len(combobox.Items()) == 0 {
				t.Errorf("下拉后列表项数量为 0，说明下拉显示失败")
			}
		}
	})

	// 测试检查包中的资源项
	t.Run("TestCheckPackageItems", func(t *testing.T) {
		t.Logf("=== 检查包 %s (地址=%p) 中的所有资源项 ===", pkg.Name, pkg)
		t.Logf("包中有 %d 个资源项", len(pkg.Items))

		// 特别检查是否存在 ComboBoxItem 模板（rt1040）
		comboBoxItem := pkg.ItemByID("rt1040")
		if comboBoxItem == nil {
			t.Logf("❌ 在包 %s 中未找到 ID 为 rt1040 的 ComboBoxItem 模板", pkg.Name)

			// 尝试通过名称查找
			comboBoxItem2 := pkg.ItemByName("ComboBoxItem")
			if comboBoxItem2 != nil {
				t.Logf("⚠️  但是通过名称找到了: ID=%s, Name=%s", comboBoxItem2.ID, comboBoxItem2.Name)
			} else {
				t.Logf("⚠️  通过名称也未找到 ComboBoxItem")
			}

			t.Errorf("缺少 ComboBoxItem 模板，这可能是导致下拉列表无法显示项目的原因")
		} else {
			t.Logf("✅ 找到 ComboBoxItem 模板: %s (类型=%v)", comboBoxItem.Name, comboBoxItem.Type)
		}

		// 测试 FactoryObjectCreator 能否找到这个资源
		t.Logf("=== 测试 FactoryObjectCreator 查找资源 ===")
		creator := &FactoryObjectCreator{
			factory: factory,
			pkg:     pkg,
			ctx:     ctx,
		}
		item := creator.CreateObject("ui://rt1040")
		if item == nil {
			t.Errorf("❌ FactoryObjectCreator 无法创建 ui://rt1040")
		} else {
			t.Logf("✅ FactoryObjectCreator 成功创建 ui://rt1040: %s", item.Name())
		}
	})
}
