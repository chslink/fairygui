package render

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestGLoaderGeneratesTextureCommand 验证 GLoader 会生成纹理命令
func TestGLoaderGeneratesTextureCommand(t *testing.T) {
	// 创建一个 GLoader
	loader := widgets.NewLoader()

	// 设置基本属性
	loader.SetSize(100, 100)
	loader.SetColor("#ff0000")

	// 创建一个模拟的 PackageItem
	item := &assets.PackageItem{
		ID:   "test-texture",
		Type: assets.PackageItemTypeImage,
		Sprite: &assets.AtlasSprite{
			Rect: assets.Rect{Width: 50, Height: 50},
		},
	}
	loader.SetPackageItem(item)

	// 获取 sprite 和 graphics
	sprite := loader.DisplayObject()
	gfx := sprite.Graphics()

	// 验证命令已生成
	if gfx == nil {
		t.Fatal("Graphics should not be nil")
	}

	if gfx.IsEmpty() {
		t.Fatal("Graphics should have commands")
	}

	commands := gfx.Commands()
	if len(commands) == 0 {
		t.Fatal("Expected at least one command")
	}

	// 验证第一个命令是纹理命令
	cmd := commands[0]
	if cmd.Type != laya.GraphicsCommandTexture {
		t.Errorf("Expected texture command, got type %d", cmd.Type)
	}

	// 验证纹理命令的内容
	texCmd := cmd.Texture
	if texCmd == nil {
		t.Fatal("Texture command should not be nil")
	}

	if texCmd.Texture != item {
		t.Error("Texture command should reference the package item")
	}

	if texCmd.Mode != laya.TextureModeSimple {
		t.Errorf("Expected Simple mode, got %d", texCmd.Mode)
	}

	if texCmd.Color != "#ff0000" {
		t.Errorf("Expected color #ff0000, got %s", texCmd.Color)
	}

	t.Log("✅ GLoader successfully generates texture command")
}

// TestGLoaderScale9Mode 验证 Scale9 模式的命令生成
func TestGLoaderScale9Mode(t *testing.T) {
	loader := widgets.NewLoader()
	loader.SetSize(100, 100)

	item := &assets.PackageItem{
		ID:   "test",
		Type: assets.PackageItemTypeImage,
		Sprite: &assets.AtlasSprite{
			Rect: assets.Rect{Width: 50, Height: 50},
		},
		Scale9Grid: &assets.Rect{X: 10, Y: 10, Width: 30, Height: 30},
	}
	loader.SetPackageItem(item)

	sprite := loader.DisplayObject()
	gfx := sprite.Graphics()

	if gfx.IsEmpty() {
		t.Fatal("Graphics should have commands")
	}

	cmd := gfx.Commands()[0]
	if cmd.Texture.Mode != laya.TextureModeScale9 {
		t.Errorf("Expected Scale9 mode, got %d", cmd.Texture.Mode)
	}

	if cmd.Texture.Scale9Grid == nil {
		t.Error("Scale9Grid should not be nil")
	}

	t.Log("✅ GLoader Scale9 mode works correctly")
}

// TestGLoaderContentOffset 验证内容偏移和缩放
func TestGLoaderContentOffset(t *testing.T) {
	loader := widgets.NewLoader()
	loader.SetSize(200, 200)
	loader.SetAlign(widgets.LoaderAlignCenter)
	loader.SetVerticalAlign(widgets.LoaderAlignMiddle)

	item := &assets.PackageItem{
		ID:   "test",
		Type: assets.PackageItemTypeImage,
		Sprite: &assets.AtlasSprite{
			Rect: assets.Rect{Width: 100, Height: 100},
		},
	}
	loader.SetPackageItem(item)

	sprite := loader.DisplayObject()
	gfx := sprite.Graphics()

	if gfx.IsEmpty() {
		t.Fatal("Graphics should have commands")
	}

	cmd := gfx.Commands()[0]
	texCmd := cmd.Texture

	// 居中对齐应该有偏移
	if texCmd.OffsetX != 50 || texCmd.OffsetY != 50 {
		t.Errorf("Expected offset (50, 50), got (%.0f, %.0f)", texCmd.OffsetX, texCmd.OffsetY)
	}

	t.Log("✅ GLoader content offset works correctly")
}

// TestGLoaderMovieClipSkipsCommand 验证 MovieClip 不生成命令
func TestGLoaderMovieClipSkipsCommand(t *testing.T) {
	loader := widgets.NewLoader()

	item := &assets.PackageItem{
		ID:   "test-mc",
		Type: assets.PackageItemTypeMovieClip,
		Frames: []*assets.MovieClipFrame{
			{
				Width:  50,
				Height: 50,
			},
		},
	}
	loader.SetPackageItem(item)

	sprite := loader.DisplayObject()
	gfx := sprite.Graphics()

	// MovieClip 应该不生成命令（使用旧渲染路径）
	if !gfx.IsEmpty() {
		t.Error("MovieClip should not generate texture commands")
	}

	t.Log("✅ GLoader correctly skips command for MovieClip")
}

// TestGLoaderFillMethodSkipsCommand 验证 FillMethod 不生成命令
func TestGLoaderFillMethodSkipsCommand(t *testing.T) {
	loader := widgets.NewLoader()
	loader.SetSize(100, 100)

	// 先设置 FillMethod
	loader.SetFillMethod(int(widgets.LoaderFillMethodRadial90))
	loader.SetFillAmount(0.5)

	// 再设置 PackageItem
	item := &assets.PackageItem{
		ID:   "test",
		Type: assets.PackageItemTypeImage,
		Sprite: &assets.AtlasSprite{
			Rect: assets.Rect{Width: 50, Height: 50},
		},
	}
	loader.SetPackageItem(item)

	sprite := loader.DisplayObject()
	gfx := sprite.Graphics()

	// FillMethod 应该不生成命令（使用旧渲染路径）
	if !gfx.IsEmpty() {
		t.Error("FillMethod should not generate texture commands")
	}

	t.Log("✅ GLoader correctly skips command for FillMethod")
}
