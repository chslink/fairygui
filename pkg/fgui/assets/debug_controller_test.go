package assets

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDebugControllerData 调试控制器数据的二进制格式
func TestDebugControllerData(t *testing.T) {
	// 加载 Basics.fui 包
	fuiPath := filepath.Join("..", "..", "..", "demo", "assets", "Basics.fui")
	fuiData, err := os.ReadFile(fuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 .fui 文件: %v", err)
	}

	pkg, err := ParsePackage(fuiData, "demo/assets/Basics")
	if err != nil {
		t.Fatalf("解析 .fui 文件失败: %v", err)
	}

	// 查找 OnOffButton 组件
	var onoffItem *PackageItem
	for _, item := range pkg.Items {
		if item.Type == PackageItemTypeComponent && (item.Name == "OnOffButton" || item.Name == "components/OnOffButton") {
			onoffItem = item
			break
		}
	}
	if onoffItem == nil {
		t.Fatalf("未找到 OnOffButton 组件")
	}

	t.Logf("OnOffButton 组件:")
	t.Logf("  Component: %v", onoffItem.Component != nil)
	if onoffItem.Component != nil {
		t.Logf("  Controllers 数量: %d", len(onoffItem.Component.Controllers))
		for i, ctrl := range onoffItem.Component.Controllers {
			t.Logf("  Controller[%d]:", i)
			t.Logf("    Name: %s", ctrl.Name)
			t.Logf("    PageNames: %v", ctrl.PageNames)
			t.Logf("    PageIDs: %v", ctrl.PageIDs)
			t.Logf("    AutoRadio: %v", ctrl.AutoRadio)
			t.Logf("    Selected: %d", ctrl.Selected)

			// 如果是 "button" 控制器，打印更多信息
			if ctrl.Name == "button" {
				t.Logf("    -> 这是 button 控制器！")
			}
		}
	}
}
