package render

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestGImageGeneratesTextureCommand 验证 GImage 会生成纹理命令
func TestGImageGeneratesTextureCommand(t *testing.T) {
	// 创建一个 GImage
	img := widgets.NewImage()

	// 设置基本属性
	img.SetSize(100, 100)
	img.SetColor("#ff0000")
	img.SetFlip(widgets.FlipTypeHorizontal)

	// 创建一个模拟的 PackageItem
	item := &assets.PackageItem{
		ID:   "test-texture",
		Type: assets.PackageItemTypeImage,
		Sprite: &assets.AtlasSprite{
			Rect: assets.Rect{Width: 50, Height: 50},
		},
		Scale9Grid: &assets.Rect{
			X:      10,
			Y:      10,
			Width:  30,
			Height: 30,
		},
	}
	img.SetPackageItem(item)

	// 获取 sprite 和 graphics
	sprite := img.DisplayObject()
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

	if texCmd.Dest.W != 100 || texCmd.Dest.H != 100 {
		t.Errorf("Expected dest size 100x100, got %.0fx%.0f", texCmd.Dest.W, texCmd.Dest.H)
	}

	if texCmd.Mode != laya.TextureModeScale9 {
		t.Errorf("Expected Scale9 mode, got %d", texCmd.Mode)
	}

	if texCmd.ScaleX != -1.0 {
		t.Errorf("Expected ScaleX=-1 for horizontal flip, got %f", texCmd.ScaleX)
	}

	if texCmd.ScaleY != 1.0 {
		t.Errorf("Expected ScaleY=1, got %f", texCmd.ScaleY)
	}

	if texCmd.Color != "#ff0000" {
		t.Errorf("Expected color #ff0000, got %s", texCmd.Color)
	}

	t.Log("✅ GImage successfully generates texture command")
}

// TestGImageModeTileSimple 验证不同模式的命令生成
func TestGImageModeDetection(t *testing.T) {
	tests := []struct {
		name          string
		hasScale9Grid bool
		scaleByTile   bool
		expectedMode  laya.TextureCommandMode
	}{
		{"Simple mode", false, false, laya.TextureModeSimple},
		{"Scale9 mode", true, false, laya.TextureModeScale9},
		{"Tile mode", false, true, laya.TextureModeTile},
		{"Scale9 overrides tile", true, true, laya.TextureModeScale9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := widgets.NewImage()
			img.SetSize(100, 100)

			item := &assets.PackageItem{
				ID:   "test",
				Type: assets.PackageItemTypeImage,
				Sprite: &assets.AtlasSprite{
					Rect: assets.Rect{Width: 50, Height: 50},
				},
				ScaleByTile: tt.scaleByTile,
			}

			if tt.hasScale9Grid {
				item.Scale9Grid = &assets.Rect{X: 10, Y: 10, Width: 30, Height: 30}
			}

			img.SetPackageItem(item)

			sprite := img.DisplayObject()
			gfx := sprite.Graphics()

			if gfx.IsEmpty() {
				t.Fatal("Graphics should have commands")
			}

			cmd := gfx.Commands()[0]
			if cmd.Texture.Mode != tt.expectedMode {
				t.Errorf("Expected mode %d, got %d", tt.expectedMode, cmd.Texture.Mode)
			}
		})
	}
}

// TestGImageUpdateOnPropertyChange 验证属性变化会重新生成命令
func TestGImageUpdateOnPropertyChange(t *testing.T) {
	img := widgets.NewImage()
	img.SetSize(100, 100)

	item := &assets.PackageItem{
		ID:   "test",
		Type: assets.PackageItemTypeImage,
		Sprite: &assets.AtlasSprite{
			Rect: assets.Rect{Width: 50, Height: 50},
		},
	}
	img.SetPackageItem(item)

	sprite := img.DisplayObject()
	gfx := sprite.Graphics()
	oldVersion := gfx.Version()

	// 修改属性应该触发更新
	img.SetColor("#00ff00")
	newVersion := gfx.Version()

	if newVersion == oldVersion {
		t.Error("Graphics version should change after property update")
	}

	// 验证新命令的颜色
	cmd := gfx.Commands()[0]
	if cmd.Texture.Color != "#00ff00" {
		t.Errorf("Expected color #00ff00, got %s", cmd.Texture.Color)
	}

	t.Log("✅ GImage updates command on property change")
}
