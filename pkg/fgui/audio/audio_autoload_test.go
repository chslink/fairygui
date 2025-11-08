package audio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// loadTestPackage 加载测试用的FUI包
func loadTestPackage(t *testing.T, name string, loader assets.Loader) *assets.Package {
	// 检查包是否已加载
	if pkg := assets.GetPackageByName(name); pkg != nil {
		return pkg
	}

	// 加载FUI文件
	fuiPath := filepath.Join(name + ".fui")
	data, err := loader.LoadOne(nil, fuiPath, assets.ResourceBinary)
	if err != nil {
		t.Errorf("加载 %s 失败: %v", fuiPath, err)
		return nil
	}

	// 解析包
	pkg, err := assets.ParsePackage(data, name)
	if err != nil {
		t.Errorf("解析 %s 失败: %v", name, err)
		return nil
	}

	// 注册包
	assets.RegisterPackage(pkg)

	return pkg
}


// TestAudioAutoLoad 测试音效自动加载功能
func TestAudioAutoLoad(t *testing.T) {
	// 设置测试用的文件加载器（使用绝对路径）
	testDir := "../../../demo/assets"
	absDir, err := filepath.Abs(testDir)
	if err != nil {
		t.Fatalf("获取绝对路径失败: %v", err)
	}
	t.Logf("使用目录: %s", absDir)

	loader := assets.NewFileLoader(absDir)
	SetLoader(loader)

	// 先加载Basics包
	basicsPkg := loadTestPackage(t, "Basics", loader)
	if basicsPkg == nil {
		t.Fatal("无法加载Basics包")
	}

	// 初始化音频播放器
	player := GetInstance()
	player.Init(48000)

	// 测试1: 尝试播放不存在的音效（应该不会崩溃）
	t.Run("NonExistentSound", func(t *testing.T) {
		// 这个测试主要确保不会崩溃
		player.Play("ui://Basics/nonexistent", 1.0)
	})

	// 测试2: 检查GetItemByURL是否能找到音效资源
	t.Run("GetItemByURL", func(t *testing.T) {
		// 模拟TypeScript版本的逻辑
		// 获取Basics包
		basicsPkg := assets.GetPackageByName("Basics")
		if basicsPkg == nil {
			t.Skip("Basics包未加载，跳过测试")
		}

		// 获取音效资源（这里需要知道音效的确切ID或名称）
		// 从FUI文件中我们看到有Basics_gojg7u.wav和Basics_o4lt7w.wav文件
		// 但我们不知道确切的PackageItem名称或ID

		// 让我们查看Basics包中的所有Items
		t.Logf("Basics包包含 %d 个资源项", len(basicsPkg.Items))

		// 查找音效资源
		var soundItems []*assets.PackageItem
		for _, item := range basicsPkg.Items {
			if item.Type == assets.PackageItemTypeSound {
				soundItems = append(soundItems, item)
				t.Logf("找到音效资源: ID=%s, Name=%s, File=%s", item.ID, item.Name, item.File)
			}
		}

		if len(soundItems) == 0 {
			t.Error("未找到任何音效资源")
		}
	})

	// 测试3: 验证文件确实存在
	t.Run("SoundFileExists", func(t *testing.T) {
		// 检查demo/assets目录下的音频文件
		audioFiles := []string{
			"Basics_gojg7u.wav",
			"Basics_o4lt7w.wav",
		}

		for _, filename := range audioFiles {
			path := filepath.Join(testDir, filename)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("音频文件不存在: %s", path)
			} else {
				t.Logf("找到音频文件: %s", path)
			}
		}
	})
}

// TestAudioPlayFromPackage 测试从包中播放音效
func TestAudioPlayFromPackage(t *testing.T) {
	testDir := "../../../demo/assets"
	absDir, err := filepath.Abs(testDir)
	if err != nil {
		t.Fatalf("获取绝对路径失败: %v", err)
	}

	loader := assets.NewFileLoader(absDir)
	SetLoader(loader)

	// 先加载Basics包
	basicsPkg := loadTestPackage(t, "Basics", loader)
	if basicsPkg == nil {
		t.Fatal("Basics包未加载")
	}

	player := GetInstance()
	player.Init(48000)

	// 模拟TypeScript版本的按钮点击音效逻辑
	// 1. 获取包
	basicsPkg = assets.GetPackageByName("Basics")
	if basicsPkg == nil {
		t.Skip("Basics包未加载")
	}

	// 2. 查找音效资源
	var clickSoundItem *assets.PackageItem
	for _, item := range basicsPkg.Items {
		if item.Type == assets.PackageItemTypeSound && item.File != "" {
			// 尝试找到可能用于按钮点击的音效
			// 通常音效文件名包含 "click" 或 "gojg7u" 等
			if contains(item.File, "gojg7u") || contains(item.File, "o4lt7w") {
				clickSoundItem = item
				break
			}
		}
	}

	if clickSoundItem == nil {
		t.Error("未找到按钮点击音效")
		return
	}

	t.Logf("找到音效资源: File=%s", clickSoundItem.File)

	// 3. 尝试使用PackageItem的File字段加载和播放
	if clickSoundItem.File != "" {
		// 这里模拟TypeScript版本的逻辑
		// 在TypeScript中，playOneShotSound(pi.file)会直接使用文件路径
		t.Logf("TypeScript版本会调用: playOneShotSound('%s')", clickSoundItem.File)

		// 验证文件确实存在
		fullPath := filepath.Join(absDir, clickSoundItem.File)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("音效文件不存在: %s", fullPath)
		} else {
			t.Logf("音效文件存在: %s", fullPath)
		}

		// 实际测试播放
		t.Run("PlayWithFilePath", func(t *testing.T) {
			// 使用文件路径播放
			player.Play(clickSoundItem.File, 1.0)
			t.Logf("已调用 Play('%s', 1.0)", clickSoundItem.File)
		})
	}
}

// TestAudioPlayWithURL 测试使用URL播放音效
func TestAudioPlayWithURL(t *testing.T) {
	testDir := "../../../demo/assets"
	absDir, err := filepath.Abs(testDir)
	if err != nil {
		t.Fatalf("获取绝对路径失败: %v", err)
	}

	loader := assets.NewFileLoader(absDir)
	SetLoader(loader)

	// 先加载Basics包
	basicsPkg := loadTestPackage(t, "Basics", loader)
	if basicsPkg == nil {
		t.Fatal("Basics包未加载")
	}

	player := GetInstance()
	player.Init(48000)

	// 获取Basics包
	basicsPkg = assets.GetPackageByName("Basics")
	if basicsPkg == nil {
		t.Skip("Basics包未加载")
	}

	// 查找音效资源并获取其URL
	var soundItem *assets.PackageItem
	for _, item := range basicsPkg.Items {
		if item.Type == assets.PackageItemTypeSound {
			soundItem = item
			break
		}
	}

	if soundItem == nil {
		t.Skip("未找到音效资源")
	}

	// 生成URL（类似 "ui://packageId/itemId"）
	url := "ui://" + basicsPkg.ID + "/" + soundItem.ID
	t.Logf("音效URL: %s", url)

	// 使用URL尝试播放
	player.Play(url, 1.0)

	// 等待异步加载完成
	// 注意：这里我们无法直接等待，因为是异步的
	// 在实际使用中，用户需要手动点击按钮触发
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		 indexOf(s, substr) >= 0))
}

// indexOf 简单实现字符串查找
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
