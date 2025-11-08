package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestDemoLabelComponents 测试 Demo_Label 组件的构建
// 这个测试复现 Demo_Label 显示空白的问题
func TestDemoLabelComponents(t *testing.T) {
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

	// 查找 Demo_Label 组件
	var demoLabelItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_Label" {
			demoLabelItem = item
			break
		}
	}

	if demoLabelItem == nil {
		t.Fatalf("未找到 Demo_Label 组件")
	}

	// 构建 Demo_Label 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	demoLabel, err := factory.BuildComponent(ctx, pkg, demoLabelItem)
	if err != nil {
		t.Fatalf("构建 Demo_Label 组件失败: %v", err)
	}

	// 验证组件结构
	if demoLabel == nil {
		t.Fatalf("Demo_Label 组件为 nil")
	}

	// 检查子组件
	children := demoLabel.Children()
	t.Logf("Demo_Label 子组件数量: %d", len(children))

	// 根据 Demo_Label.xml，应该有以下子组件（按 name 属性）：
	// n1: 文本说明
	// frame (id=n3): Label 组件，标题 "Bag"
	// n4: Label 组件，标题 "Hello world"，图标 "ui://9leh0eyfrpmbg"
	// n5: Label 组件，标题 "Hello Unity"，自定义颜色和图标
	expectedChildren := map[string]struct {
		shouldBeLabel bool
		title         string
		icon          string
	}{
		"n1":    {shouldBeLabel: false}, // 纯文本
		"frame": {shouldBeLabel: true, title: "Bag"},
		"n4":    {shouldBeLabel: true, title: "Hello world", icon: "ui://9leh0eyfrpmbg"},
		"n5":    {shouldBeLabel: true, title: "Hello Unity", icon: "ui://9leh0eyfhixt1v"},
	}

	foundChildren := make(map[string]bool)

	for _, child := range children {
		if child == nil {
			continue
		}
		name := child.Name()
		foundChildren[name] = true

		expected, exists := expectedChildren[name]
		if !exists {
			t.Logf("发现未预期的子组件: %s", name)
			continue
		}

		t.Run(name, func(t *testing.T) {
			// 无论是否应该是 Label，都打印详细信息
			t.Logf("子组件 %s 详细信息:", name)
			t.Logf("  位置: (%.0f, %.0f)", child.X(), child.Y())
			t.Logf("  尺寸: %.0fx%.0f", child.Width(), child.Height())
			t.Logf("  可见: %v", child.Visible())
			t.Logf("  Alpha: %.2f", child.Alpha())

			if expected.shouldBeLabel {
				// 检查子组件数据类型
				data := child.Data()
				t.Logf("子组件 %s 的数据类型: %T", name, data)

				// 应该是 GLabel 或者包含 Label 扩展的组件
				label, isLabel := data.(*widgets.GLabel)
				if !isLabel {
					// 可能是 GComponent 类型，但应该有 Label 扩展
					comp, isComp := data.(*core.GComponent)
					if !isComp {
						t.Errorf("子组件 %s 既不是 GLabel 也不是 GComponent，类型: %T", name, data)
						return
					}

					// GComponent 的情况下，检查是否有子组件显示标题和图标
					t.Logf("子组件 %s 是 GComponent，子组件数量: %d", name, len(comp.Children()))
					for i, subChild := range comp.Children() {
						if subChild != nil {
							t.Logf("  子组件[%d]: name=%s, type=%T", i, subChild.Name(), subChild.Data())
						}
					}

					// 如果是普通 GComponent，检查是否有 "title" 和 "icon" 子对象
					if titleObj := comp.ChildByName("title"); titleObj != nil {
						t.Logf("  找到 title 对象: %s", titleObj.Name())
						if expected.title != "" {
							// 检查标题是否正确设置
							if tf, ok := titleObj.Data().(*widgets.GTextField); ok {
								if tf.Text() != expected.title {
									t.Errorf("标题不匹配：期望 %q，实际 %q", expected.title, tf.Text())
								} else {
									t.Logf("  标题正确: %q", tf.Text())
								}
							}
						}
					} else {
						t.Logf("  未找到 title 对象")
					}

					if iconObj := comp.ChildByName("icon"); iconObj != nil {
						t.Logf("  找到 icon 对象: %s", iconObj.Name())
					} else {
						t.Logf("  未找到 icon 对象")
					}

					return
				}

				// 是 GLabel 的情况
				t.Logf("子组件 %s 是 GLabel", name)

				// 验证标题
				if expected.title != "" {
					title := label.Title()
					if title != expected.title {
						t.Errorf("标题不匹配：期望 %q，实际 %q", expected.title, title)
					} else {
						t.Logf("✓ 标题正确: %q", title)
					}
				}

				// 验证图标
				if expected.icon != "" {
					icon := label.Icon()
					if icon != expected.icon {
						t.Errorf("图标不匹配：期望 %q，实际 %q", expected.icon, icon)
					} else {
						t.Logf("✓ 图标正确: %q", icon)
					}

					// 检查 iconItem 类型
					iconItem := label.IconItem()
					if iconItem != nil {
						t.Logf("  iconItem 类型: %v, ID: %s, Name: %s", iconItem.Type, iconItem.ID, iconItem.Name)
						if iconItem.Sprite != nil {
							t.Logf("  iconItem 有 Sprite 数据（Image 类型）")
						} else {
							t.Logf("  iconItem 没有 Sprite 数据")

							// 根据 iconItem 类型验证不同的处理方式
							if iconObj := label.IconObject(); iconObj != nil {
								if loader, ok := iconObj.Data().(*widgets.GLoader); ok && loader != nil {
									if iconItem.Type == assets.PackageItemTypeComponent {
										// Component 类型：应该构建 Component 实例
										if comp := loader.Component(); comp != nil {
											t.Logf("  ✓ GLoader 成功构建了 Component 实例")
											t.Logf("  Component 尺寸: %.0fx%.0f", comp.Width(), comp.Height())
										} else {
											t.Errorf("  ✗ GLoader 没有构建 Component 实例（这会导致图标不显示）")
										}
									} else if iconItem.Type == assets.PackageItemTypeMovieClip {
										// MovieClip 类型：应该创建内部 MovieClip 实例
										if loader.MovieClip() != nil {
											t.Logf("  ✓ GLoader 成功创建了 MovieClip 实例")
											t.Logf("  MovieClip playing: %v", loader.MovieClip().Playing())
										} else {
											t.Errorf("  ✗ GLoader 没有创建 MovieClip 实例（这会导致图标不显示）")
										}
									}
								}
							}
						}
					} else {
						t.Logf("  iconItem 为 nil")
					}
				}

				// 验证模板组件
				template := label.TemplateComponent()
				if template == nil {
					t.Errorf("Label 的模板组件为 nil")
				} else {
					t.Logf("✓ Label 有模板组件，子组件数量: %d", len(template.Children()))
					for i, subChild := range template.Children() {
						if subChild != nil {
							t.Logf("  模板子组件[%d]: name=%s, type=%T", i, subChild.Name(), subChild.Data())
						}
					}
				}

				// 验证 titleObject 和 iconObject
				if expected.title != "" {
					titleObj := label.TitleObject()
					if titleObj == nil {
						t.Errorf("Label 的 titleObject 为 nil")
					} else {
						t.Logf("✓ Label 有 titleObject: %s", titleObj.Name())
					}
				}

				if expected.icon != "" {
					iconObj := label.IconObject()
					if iconObj == nil {
						t.Errorf("Label 的 iconObject 为 nil")
					} else {
						t.Logf("✓ Label 有 iconObject: %s", iconObj.Name())
					}
				}
			} else {
				// 非 Label 组件，打印其类型
				data := child.Data()
				t.Logf("子组件 %s 的数据类型: %T", name, data)

				// 如果是 GTextField，打印其文本内容
				if tf, ok := data.(*widgets.GTextField); ok {
					t.Logf("  文本内容: %q", tf.Text())
					t.Logf("  字体大小: %d", tf.FontSize())
					t.Logf("  颜色: %s", tf.Color())
				}
			}
		})
	}

	// 验证所有期望的子组件都被找到
	for name := range expectedChildren {
		if !foundChildren[name] {
			t.Errorf("未找到期望的子组件: %s", name)
		}
	}
}
