package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestProgressBarComponentParsing 测试ProgressBar组件的解析和模板应用
// 场景：Demo_ProgressBar.xml中引用了多个ProgressBar组件
// 预期：所有ProgressBar组件都应该被正确解析为GProgressBar，模板组件应该被正确应用
func TestProgressBarComponentParsing(t *testing.T) {
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

	// 创建工厂
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 Demo_ProgressBar 组件
	var demoItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_ProgressBar" {
			demoItem = item
			break
		}
	}
	if demoItem == nil {
		t.Fatalf("未找到 Demo_ProgressBar 组件")
	}

	demo, err := factory.BuildComponent(ctx, pkg, demoItem)
	if err != nil {
		t.Fatalf("构建 Demo_ProgressBar 组件失败: %v", err)
	}

	t.Logf("✅ Demo_ProgressBar 组件构建成功")

	// 检查各个ProgressBar子组件
	progressBars := []struct {
		name      string
		expectMin int
		expectMax int
		expectVal int
	}{
		{"n1", 0, 100, 50},  // t9fj50
		{"n2", 0, 100, 78},  // t9fj53
		{"n3", 0, 100, 80},  // t9fj57
		{"n4", 0, 100, 75},  // t9fj5b
		{"n6", 0, 100, 80},  // t9fj5g
		{"n7", 0, 100, 67},  // fjqr7j (Component11.xml, extention="ProgressBar")
		{"n9", 0, 100, 30},  // gzpr81
	}

	allPassed := true
	for _, pb := range progressBars {
		t.Logf("\n--- 检查 ProgressBar: %s ---", pb.name)

		obj := demo.ChildByName(pb.name)
		if obj == nil {
			t.Errorf("❌ 未找到 %s 组件", pb.name)
			allPassed = false
			continue
		}
		t.Logf("✅ 找到 %s 对象", pb.name)

		// 检查对象数据
		data := obj.Data()
		t.Logf("   - Data() 类型: %T", data)

		// 检查是否是GComponent
		comp := core.ComponentFrom(obj)
		if comp == nil {
			t.Errorf("❌ %s 不是有效的组件", pb.name)
			allPassed = false
			continue
		}
		t.Logf("✅ %s 是有效的组件", pb.name)

		// 查找GProgressBar
		var bar *widgets.GProgressBar
		// 方法1：直接检查obj.Data()
		if b, ok := data.(*widgets.GProgressBar); ok {
			bar = b
			t.Logf("✅ %s 找到 GProgressBar (通过 obj.Data())", pb.name)
		} else {
			// 方法2：检查组件的Data
			compData := comp.Data()
			if b, ok := compData.(*widgets.GProgressBar); ok {
				bar = b
				t.Logf("✅ %s 找到 GProgressBar (通过 comp.Data())", pb.name)
			} else {
				t.Logf("⚠️  %s Data类型: %T", pb.name, compData)
			}
		}

		if bar == nil {
			t.Errorf("❌ %s 不是 GProgressBar 类型", pb.name)
			allPassed = false
			continue
		}

		// 验证ProgressBar属性
		if bar.Min() != float64(pb.expectMin) {
			t.Errorf("❌ %s Min值错误: 期望 %d, 实际 %v", pb.name, pb.expectMin, bar.Min())
			allPassed = false
		} else {
			t.Logf("✅ %s Min值正确: %v", pb.name, bar.Min())
		}

		if bar.Max() != float64(pb.expectMax) {
			t.Errorf("❌ %s Max值错误: 期望 %d, 实际 %v", pb.name, pb.expectMax, bar.Max())
			allPassed = false
		} else {
			t.Logf("✅ %s Max值正确: %v", pb.name, bar.Max())
		}

		if int(bar.Value()) != pb.expectVal {
			t.Errorf("❌ %s Value值错误: 期望 %d, 实际 %v", pb.name, pb.expectVal, bar.Value())
			allPassed = false
		} else {
			t.Logf("✅ %s Value值正确: %v", pb.name, bar.Value())
		}

		// 验证ProgressBar尺寸（关键修复：确保从模板继承尺寸）
		barW := bar.GComponent.Width()
		barH := bar.GComponent.Height()
		if barW > 0 && barH > 0 {
			t.Logf("✅ %s 尺寸正确: %.0fx%.0f (已从模板继承)", pb.name, barW, barH)
		} else {
			t.Errorf("❌ %s 尺寸错误: %.0fx%.0f (应该从模板继承，非0尺寸)", pb.name, barW, barH)
			allPassed = false
		}

		// 检查模板组件
		if tmpl := bar.TemplateComponent(); tmpl != nil {
			tmplW := tmpl.Width()
			tmplH := tmpl.Height()
			if tmplW > 0 || tmplH > 0 {
				t.Logf("✅ %s 模板组件存在且有尺寸: %.0fx%.0f", pb.name, tmplW, tmplH)
			} else {
				t.Logf("⚠️  %s 模板组件存在但尺寸为0x0", pb.name)
			}
			tmplData := tmpl.Data()
			if _, ok := tmplData.(*widgets.GProgressBar); ok {
				t.Logf("✅ %s 模板数据是 GProgressBar", pb.name)
			} else {
				t.Logf("⚠️  %s 模板数据类型: %T", pb.name, tmplData)
			}
		} else {
			t.Logf("⚠️  %s 模板组件为空", pb.name)
		}

		t.Logf("--- %s 检查完成 ---\n", pb.name)
	}

	if allPassed {
		t.Logf("\n✅ 所有ProgressBar组件解析正确！")
	} else {
		t.Errorf("\n❌ 部分ProgressBar组件解析失败")
	}
}

// TestProgressBarStarComponent 测试Grid中使用的star组件（fjqr7j）
// 这个测试专门验证Component11.xml (extention="ProgressBar")的解析
func TestProgressBarStarComponent(t *testing.T) {
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

	// 创建工厂
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 fjqr7j 组件 (Component11.xml)
	var fjqr7jItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.ID == "fjqr7j" {
			fjqr7jItem = item
			break
		}
	}
	if fjqr7jItem == nil {
		t.Fatalf("未找到 fjqr7j 组件")
	}

	t.Logf("组件信息:")
	t.Logf("  - ID: %s", fjqr7jItem.ID)
	t.Logf("  - Name: %s", fjqr7jItem.Name)
	t.Logf("  - ObjectType: %v", fjqr7jItem.ObjectType)
	t.Logf("  - Type: %v", fjqr7jItem.Type)

	// 构建组件
	comp, err := factory.BuildComponent(ctx, pkg, fjqr7jItem)
	if err != nil {
		t.Fatalf("构建 fjqr7j 组件失败: %v", err)
	}

	t.Logf("\n✅ fjqr7j 组件构建成功")

	// 检查组件结构
	t.Logf("\n组件结构:")
	t.Logf("  - Component: %v", comp)
	t.Logf("  - GObject: %v", comp.GObject)
	t.Logf("  - Data: %T", comp.Data())

	// 查找GProgressBar
	var bar *widgets.GProgressBar
	if b, ok := comp.Data().(*widgets.GProgressBar); ok {
		bar = b
		t.Logf("✅ 组件 Data 是 GProgressBar")
	} else {
		t.Errorf("❌ 组件 Data 不是 GProgressBar，实际类型: %T", comp.Data())
	}

	if bar == nil {
		t.Fatal("无法获取 GProgressBar")
	}

	// 验证ProgressBar属性
	t.Logf("\nProgressBar 属性:")
	t.Logf("  - Min: %v", bar.Min())
	t.Logf("  - Max: %v", bar.Max())
	t.Logf("  - Value: %v", bar.Value())

	// 检查模板组件
	tmpl := bar.TemplateComponent()
	if tmpl != nil {
		t.Logf("✅ 模板组件存在")
		tmplObj := tmpl.GObject
		if tmplObj != nil {
			t.Logf("  - 模板对象: %v", tmplObj)
			t.Logf("  - 模板GComponent: %v", tmpl)
			t.Logf("  - 模板子组件数: %d", len(tmpl.Children()))
			for i, child := range tmpl.Children() {
				if child != nil {
					t.Logf("    [%d] %s (类型: %T)", i, child.Name(), child.Data())
				}
			}
		}
	} else {
		t.Logf("⚠️  模板组件为空")

		// 添加调试：检查模板设置过程
		t.Logf("\n--- 手动调试模板创建过程 ---")
		targetPkg := fjqr7jItem.Owner
		if targetPkg != nil {
			t.Logf("目标包: %s", targetPkg.Name)
			// 尝试重新构建模板
			tmpl2, err := factory.BuildComponent(ctx, targetPkg, fjqr7jItem)
			if err != nil {
				t.Logf("❌ 重建模板失败: %v", err)
			} else {
				t.Logf("✅ 重建模板成功: %v", tmpl2)
				tmpl2Data := tmpl2.Data()
				t.Logf("  模板Data: %T", tmpl2Data)
				t.Logf("  模板子组件数: %d", len(tmpl2.Children()))
				for i, child := range tmpl2.Children() {
					if child != nil {
						t.Logf("    [%d] %s (类型: %T)", i, child.Name(), child.Data())
					}
				}
			}
		}
	}

	// 检查子组件
	t.Logf("\n主组件子组件:")
	for i, child := range comp.Children() {
		if child != nil {
			t.Logf("  [%d] %s (类型: %T)", i, child.Name(), child.Data())
		}
	}
}
