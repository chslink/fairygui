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

// TestWindowFrameSize 测试WindowFrame组件的尺寸设置是否生效
func TestWindowFrameSize(t *testing.T) {
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

	// 查找WindowFrame组件（rt103l）
	var windowFrameItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Name == "WindowFrame" || item.ID == "rt103l" {
			windowFrameItem = item
			break
		}
	}

	if windowFrameItem == nil {
		t.Fatalf("未找到WindowFrame组件")
	}

	t.Logf("WindowFrame原始尺寸: %dx%d", windowFrameItem.Component.InitWidth, windowFrameItem.Component.InitHeight)

	// 构建WindowFrame组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	windowFrame, err := factory.BuildComponent(ctx, pkg, windowFrameItem)
	if err != nil {
		t.Fatalf("构建WindowFrame组件失败: %v", err)
	}

	// 验证原始尺寸
	if windowFrame.Width() != 157 || windowFrame.Height() != 202 {
		t.Errorf("WindowFrame原始尺寸不正确：期望 157x202，实际 %.0fx%.0f", windowFrame.Width(), windowFrame.Height())
	} else {
		t.Logf("✓ WindowFrame原始尺寸正确: %.0fx%.0f", windowFrame.Width(), windowFrame.Height())
	}

	// 设置新的尺寸
	newW, newH := 223.0, 226.0
	windowFrame.GObject.SetSize(newW, newH)

	// 验证尺寸是否生效
	if windowFrame.Width() != newW || windowFrame.Height() != newH {
		t.Errorf("WindowFrame尺寸设置失败：期望 %.0fx%.0f，实际 %.0fx%.0f", newW, newH, windowFrame.Width(), windowFrame.Height())
	} else {
		t.Logf("✓ WindowFrame尺寸设置成功: %.0fx%.0f", windowFrame.Width(), windowFrame.Height())
	}

	// 检查内部元素
	t.Logf("WindowFrame子组件数量: %d", len(windowFrame.Children()))
	for i, child := range windowFrame.Children() {
		if child == nil {
			continue
		}
		t.Logf("  子组件[%d]: name=%s, size=%.0fx%.0f", i, child.Name(), child.Width(), child.Height())

		// 检查contentArea（n4）的尺寸
		if child.Name() == "contentArea" {
			expectedContentW := newW - 7 // 实际计算结果为216
			expectedContentH := newH - 30 // 实际计算结果为196

			t.Logf("    contentArea期望尺寸: %.0fx%.0f", expectedContentW, expectedContentH)

			// 验证contentArea是否根据Relation调整了尺寸
			if child.Width() != expectedContentW {
				t.Errorf("contentArea宽度不正确：期望 %.0f，实际 %.0f", expectedContentW, child.Width())
			} else {
				t.Logf("    ✓ contentArea宽度正确: %.0f", child.Width())
			}
		}
	}
}

// TestWindowFrameInLabelContext 测试WindowFrame作为Label引用时的尺寸
func TestWindowFrameInLabelContext(t *testing.T) {
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

	// 查找Demo_Label组件
	var demoLabelItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_Label" {
			demoLabelItem = item
			break
		}
	}

	if demoLabelItem == nil {
		t.Fatalf("未找到Demo_Label组件")
	}

	// 构建Demo_Label组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()
	demoLabel, err := factory.BuildComponent(ctx, pkg, demoLabelItem)
	if err != nil {
		t.Fatalf("构建Demo_Label组件失败: %v", err)
	}

	// 查找frame子组件
	var frameObj *core.GObject
	for _, child := range demoLabel.Children() {
		if child.Name() == "frame" {
			frameObj = child
			break
		}
	}

	if frameObj == nil {
		t.Fatalf("未找到frame子组件")
	}

	t.Logf("frame子组件尺寸: %.0fx%.0f", frameObj.Width(), frameObj.Height())

	// 期望尺寸：223x226
	expectedW, expectedH := 223.0, 226.0
	if frameObj.Width() != expectedW || frameObj.Height() != expectedH {
		t.Errorf("frame子组件尺寸不正确：期望 %.0fx%.0f，实际 %.0fx%.0f", expectedW, expectedH, frameObj.Width(), frameObj.Height())
	} else {
		t.Logf("✓ frame子组件尺寸正确: %.0fx%.0f", frameObj.Width(), frameObj.Height())
	}

	// 检查frame的类型
	if label, ok := frameObj.Data().(*widgets.GLabel); ok {
		t.Logf("frame是GLabel")
		if tpl := label.TemplateComponent(); tpl != nil {
			t.Logf("  TemplateComponent尺寸: %.0fx%.0f", tpl.Width(), tpl.Height())

			// TemplateComponent应该也有正确的尺寸
			if tpl.Width() != expectedW || tpl.Height() != expectedH {
				t.Errorf("TemplateComponent尺寸不正确：期望 %.0fx%.0f，实际 %.0fx%.0f", expectedW, expectedH, tpl.Width(), tpl.Height())
			} else {
				t.Logf("  ✓ TemplateComponent尺寸正确: %.0fx%.0f", tpl.Width(), tpl.Height())
			}

			// 检查contentArea
			if contentArea := tpl.ChildByName("contentArea"); contentArea != nil {
				t.Logf("  contentArea尺寸: %.0fx%.0f", contentArea.Width(), contentArea.Height())
				// contentArea应该有正确的尺寸（223 - 7 = 216 宽度，226 - 30 = 196 高度）
				// 实际尺寸可能因为Relation计算和边距设置而有微小差异
				expectedContentW := 216.0
				expectedContentH := 196.0
				if contentArea.Width() != expectedContentW {
					t.Errorf("contentArea宽度不正确：期望 %.0f，实际 %.0f", expectedContentW, contentArea.Width())
				} else {
					t.Logf("    ✓ contentArea宽度正确: %.0f", contentArea.Width())
				}
				if contentArea.Height() != expectedContentH {
					t.Errorf("contentArea高度不正确：期望 %.0f，实际 %.0f", expectedContentH, contentArea.Height())
				} else {
					t.Logf("    ✓ contentArea高度正确: %.0f", contentArea.Height())
				}
			}
		}
	} else {
		t.Errorf("frame不是GLabel，类型: %T", frameObj.Data())
	}
}
