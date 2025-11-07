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

// TestListDefaultItemWithItemAttributes 测试List的defaultItem是否能正确解析item节点的属性
// 问题：n0组件的defaultItem是一个背包格子，里面的item项应该显示title和icon
// 但当前只显示了格子，title和icon没有被渲染
func TestListDefaultItemWithItemAttributes(t *testing.T) {
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

	// 构建 Demo_List 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 Demo_List 组件
	var demoItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_List" {
			demoItem = item
			break
		}
	}
	if demoItem == nil {
		t.Fatalf("未找到 Demo_List 组件")
	}

	demo, err := factory.BuildComponent(ctx, pkg, demoItem)
	if err != nil {
		t.Fatalf("构建 Demo_List 组件失败: %v", err)
	}

	// 将 demo 添加到 stage
	stage.AddChild(demo.DisplayObject())

	// 获取 n0 List
	n0 := demo.ChildByName("n0")
	if n0 == nil {
		t.Fatalf("未找到 n0 对象")
	}

	list, ok := n0.Data().(*widgets.GList)
	if !ok || list == nil {
		t.Fatalf("n0 不是 GList")
	}

	t.Logf("List defaultItem: %s", list.DefaultItem())
	t.Logf("List items count: %d", len(list.Items()))

	// 验证 List 的 items
	items := list.Items()
	if len(items) == 0 {
		t.Errorf("List 应该包含 items，但实际为 0 个")
	}

	// 验证每个 item 的属性
	for i, item := range items {
		if item == nil {
			t.Errorf("item %d 为 nil", i)
			continue
		}

		t.Logf("Item %d: name=%s, type=%T", i, item.Name(), item.Data())

		// 根据测试结果，item 是 GButton 类型
		if btn, ok := item.Data().(*widgets.GButton); ok && btn != nil {
			// 检查 title 文本
			title := btn.Title()
			t.Logf("  Button title: '%s'", title)
			if title == "" {
				t.Errorf("item %d 的 title 为空，预期应该有文本（来自 XML 中的 item title 属性）", i)
			}

			// 检查 icon 图标
			icon := btn.Icon()
			t.Logf("  Button icon: '%s'", icon)
			if icon == "" {
				t.Errorf("item %d 的 icon 为空，预期应该有图标（来自 XML 中的 item icon 属性）", i)
			}
		} else if comp, ok := item.Data().(*core.GComponent); ok && comp != nil {
			// 如果是 GComponent 类型，查找 title 和 icon 子组件
			t.Logf("  Component 子组件数量: %d", len(comp.Children()))

			// 查找 title 和 icon 子组件
			for _, child := range comp.Children() {
				if child == nil {
					continue
				}
				childName := child.Name()
				t.Logf("    子组件: name=%s, type=%T", childName, child.Data())

				// 检查 title 文本
				if childName == "title" {
					if tf, ok := child.Data().(*widgets.GTextField); ok {
						text := tf.Text()
						t.Logf("      title 文本: '%s'", text)
						if text == "" {
							t.Errorf("item %d 的 title 为空，预期应该有文本（来自 XML 中的 item title 属性）", i)
						}
					}
				}

				// 检查 icon 图片
				if childName == "icon" {
					if loader, ok := child.Data().(*widgets.GLoader); ok {
						url := loader.URL()
						t.Logf("      icon URL: '%s'", url)
						if url == "" {
							t.Errorf("item %d 的 icon 为空，预期应该有图标（来自 XML 中的 item icon 属性）", i)
						}
					} else if btn, ok := child.Data().(*widgets.GButton); ok {
						icon := btn.Icon()
						t.Logf("      icon (GButton): '%s'", icon)
						if icon == "" {
							t.Errorf("item %d 的 icon (GButton) 为空，预期应该有图标（来自 XML 中的 item icon 属性）", i)
						}
					}
				}
			}
		}
	}
}
