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

// TestButtonSoundParsing 测试按钮音效的解析
// 场景：Demo_Button.xml 中的 n7 按钮配置了自定义音效 <Button sound="ui://9leh0eyfgojg7u"/>
// 预期：解析后的按钮实例应该有自定义音效，而不是全局音效
func TestButtonSoundParsing(t *testing.T) {
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

	// 构建 Demo_Button 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 Demo_Button 组件
	var demoItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_Button" {
			demoItem = item
			break
		}
	}
	if demoItem == nil {
		t.Fatalf("未找到 Demo_Button 组件")
	}

	demo, err := factory.BuildComponent(ctx, pkg, demoItem)
	if err != nil {
		t.Fatalf("构建 Demo_Button 组件失败: %v", err)
	}

	// 将 demo 添加到 stage
	stage.AddChild(demo.DisplayObject())

	// 获取 n7 按钮 (配置了自定义音效的按钮)
	n7 := demo.ChildByName("n7")
	if n7 == nil {
		t.Fatalf("未找到 n7 对象")
	}

	btn, ok := n7.Data().(*widgets.GButton)
	if !ok || btn == nil {
		t.Fatalf("n7 不是 GButton")
	}

	t.Logf("按钮 n7 音效: sound=%s, soundVolumeScale=%.2f",
		btn.Sound(), btn.SoundVolumeScale())

	// 验证：n7 应该有自定义音效 "ui://9leh0eyfgojg7u"
	expectedSound := "ui://9leh0eyfgojg7u"
	if btn.Sound() != expectedSound {
		t.Errorf("n7 按钮音效解析错误：期望 %s，实际 %s", expectedSound, btn.Sound())
	}

	// 解析 sound URL 查看对应的实际文件
	if btn.Sound() != "" {
		pi := assets.GetItemByURL(btn.Sound())
		if pi != nil {
			t.Logf("n7 音效 URL 解析结果: ID=%s, Name=%s, File=%s", pi.ID, pi.Name, pi.File)
			expectedFile := "demo/assets/Basics_gojg7u.wav"
			if pi.File != expectedFile {
				t.Errorf("n7 音效文件错误：期望 %s，实际 %s", expectedFile, pi.File)
			}
		} else {
			t.Errorf("无法解析 n7 音效 URL: %s", btn.Sound())
		}
	}

	// 检查 Button4 组件模板的默认音效
	var button4Item *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Button4" {
			button4Item = item
			break
		}
	}
	if button4Item != nil {
		button4, err := factory.BuildComponent(ctx, pkg, button4Item)
		if err == nil {
			if btn4, ok := button4.Data().(*widgets.GButton); ok {
				t.Logf("Button4 模板默认 sound: '%s'", btn4.Sound())
			}
		}
	}

	// 同时测试 n16 按钮 (OnOffButton，也配置了相同的音效)
	n16 := demo.ChildByName("n16")
	if n16 == nil {
		t.Fatalf("未找到 n16 对象")
	}

	btn16, ok := n16.Data().(*widgets.GButton)
	if !ok || btn16 == nil {
		t.Fatalf("n16 不是 GButton")
	}

	t.Logf("按钮 n16 音效: sound=%s, soundVolumeScale=%.2f",
		btn16.Sound(), btn16.SoundVolumeScale())

	// 验证：n16 也应该有自定义音效
	if btn16.Sound() != expectedSound {
		t.Errorf("n16 按钮音效解析错误：期望 %s，实际 %s", expectedSound, btn16.Sound())
	}

	// 测试没有自定义音效的按钮 (n3)
	n3 := demo.ChildByName("n3")
	if n3 == nil {
		t.Fatalf("未找到 n3 对象")
	}

	btn3, ok := n3.Data().(*widgets.GButton)
	if !ok || btn3 == nil {
		t.Fatalf("n3 不是 GButton")
	}

	t.Logf("按钮 n3 音效: sound=%s, soundVolumeScale=%.2f",
		btn3.Sound(), btn3.SoundVolumeScale())

	// n3 没有自定义音效，应该使用全局音效或为空（取决于具体的组件模板配置）
	// 这里只是输出日志，不做断言
}
