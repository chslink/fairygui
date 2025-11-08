package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/gears"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestTextGearFontSize 测试 n17 文本的 gearFontSize
// 场景：Demo_Controller.xml 中的 n17 文本配置了 <gearFontSize controller="c1" pages="1" values="74" default="33"/>
// 预期：c1 在 page 0 时 fontSize=33，在 page 1 时 fontSize=74
func TestTextGearFontSize(t *testing.T) {
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

	// 构建 Demo_Controller 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 Demo_Controller 组件
	var demoItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_Controller" {
			demoItem = item
			break
		}
	}
	if demoItem == nil {
		t.Fatalf("未找到 Demo_Controller 组件")
	}

	demo, err := factory.BuildComponent(ctx, pkg, demoItem)
	if err != nil {
		t.Fatalf("构建 Demo_Controller 组件失败: %v", err)
	}

	// 获取 c1 控制器
	c1 := demo.ControllerByName("c1")
	if c1 == nil {
		t.Fatalf("未找到 c1 控制器")
	}

	// 获取 n17 文本对象
	n17 := demo.ChildByName("n17")
	if n17 == nil {
		t.Fatalf("未找到 n17 对象")
	}

	t.Logf("n17 类型: %T", n17.Data())

	// 检查 n17 是否有 GearFontSize
	gear := n17.GetGear(gears.IndexFontSize)
	if gear == nil {
		t.Fatalf("n17 没有 GearFontSize")
	}

	// 检查 gear 的 controller 是否是 c1
	if gear.Controller() != c1 {
		t.Errorf("n17 的 GearFontSize controller 不是 c1，而是: %v", gear.Controller())
	}

	// 检查 gear 的详细信息
	if fontSizeGear, ok := gear.(*gears.GearFontSize); ok {
		defaultVal := fontSizeGear.Owner().GetProp(gears.ObjectPropIDFontSize)
		t.Logf("GearFontSize 当前值: %v", defaultVal)
	}

	// 获取文本 widget
	var textWidget *widgets.GTextField
	switch data := n17.Data().(type) {
	case *widgets.GTextField:
		textWidget = data
	default:
		t.Fatalf("n17 不是 GTextField，而是: %T", data)
	}

	// 初始状态 (page 0): fontSize 应该是 33
	t.Logf("初始状态: c1.SelectedPageID=%s, fontSize=%d", c1.SelectedPageID(), textWidget.FontSize())
	if textWidget.FontSize() != 33 {
		t.Errorf("page 0 时 fontSize 应该是 33，实际: %d", textWidget.FontSize())
	}

	// 切换到 page 1
	c1.SetSelectedPageID("1")

	// page 1: fontSize 应该是 74
	t.Logf("切换后: c1.SelectedPageID=%s, fontSize=%d", c1.SelectedPageID(), textWidget.FontSize())
	if textWidget.FontSize() != 74 {
		t.Errorf("page 1 时 fontSize 应该是 74，实际: %d", textWidget.FontSize())
	}

	// 切换回 page 0
	c1.SetSelectedPageID("0")

	// page 0: fontSize 应该是 33
	t.Logf("切换回: c1.SelectedPageID=%s, fontSize=%d", c1.SelectedPageID(), textWidget.FontSize())
	if textWidget.FontSize() != 33 {
		t.Errorf("page 0 时 fontSize 应该是 33，实际: %d", textWidget.FontSize())
	}
}
