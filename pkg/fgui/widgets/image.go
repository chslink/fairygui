package widgets

import (
	"log"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// GImage is an image widget supporting scale9 grid and tiling.
type GImage struct {
	*core.GObject
	packageItem    *assets.PackageItem
	color          string
	fillMethod     int
	fillAmount     float64
	fillOrigin     int
	fillClockwise  bool
	scaleByTile    bool
	tileGridIndice int
	flip           FlipType
}

// FlipType mirrors FairyGUI's image flip enumeration.
type FlipType int

const (
	FlipTypeNone FlipType = iota
	FlipTypeHorizontal
	FlipTypeVertical
	FlipTypeBoth
)

// NewImage constructs a GImage.
func NewImage() *GImage {
	img := &GImage{
		GObject: core.NewGObject(),
		color:   "#ffffff",
	}
	img.GObject.SetData(img)
	// 修复：图片是显示性组件，不应该拦截鼠标事件，设置mouseThrough=true让事件穿透到父组件
	if sprite := img.GObject.DisplayObject(); sprite != nil {
		sprite.SetMouseThrough(true)
	}
	return img
}

// SetPackageItem assigns the package item that provides the texture for this image.
func (i *GImage) SetPackageItem(item *assets.PackageItem) {
	i.packageItem = item
	if item != nil {
		i.scaleByTile = item.ScaleByTile
		i.tileGridIndice = item.TileGridIndice
	} else {
		i.scaleByTile = false
		i.tileGridIndice = 0
	}
	i.updateGraphics()
}

// PackageItem returns the package item backing this image.
func (i *GImage) PackageItem() *assets.PackageItem {
	return i.packageItem
}

// SetColor records the applied tint colour in hex format.
func (i *GImage) SetColor(value string) {
	if i.color == value {
		return
	}
	i.color = value
	i.updateGraphics()
}

// Color returns the stored tint colour.
func (i *GImage) Color() string {
	return i.color
}

// Flip returns the current flip mode.
func (i *GImage) Flip() FlipType {
	return i.flip
}

// SetFlip updates the flip mode and refreshes the rendering command.
func (i *GImage) SetFlip(value FlipType) {
	if i.flip == value {
		return
	}
	i.flip = value
	i.updateGraphics()
}

// SetFill configures the fill method/origin/amount metadata (currently advisory).
func (i *GImage) SetFill(method int, origin int, clockwise bool, amount float64) {
	i.fillMethod = method
	i.fillOrigin = origin
	i.fillClockwise = clockwise
	i.fillAmount = amount
	// 触发重绘以应用fillAmount
	i.updateGraphics()
}

// Fill returns the stored fill parameters.
func (i *GImage) Fill() (method int, origin int, clockwise bool, amount float64) {
	return i.fillMethod, i.fillOrigin, i.fillClockwise, i.fillAmount
}

// ScaleSettings exposes the tiling parameters from the package item.
func (i *GImage) ScaleSettings() (scaleByTile bool, tileGridIndice int) {
	return i.scaleByTile, i.tileGridIndice
}

// OwnerSizeChanged 在宿主尺寸变化时刷新绘制。

func (i *GImage) OwnerSizeChanged(oldW, oldH float64) {

	i.updateGraphics()

}

func (i *GImage) updateGraphics() {
	if i == nil || i.GObject == nil {
		return
	}
	sprite := i.GObject.DisplayObject()
	if sprite == nil {
		return
	}

	sprite.SetMouseEnabled(i.GObject.Touchable())

	// 获取或创建 Graphics
	gfx := sprite.Graphics()
	gfx.Clear()

	// 如果没有纹理，不生成命令
	if i.packageItem == nil {
		sprite.Repaint()
		return
	}

	// 确定渲染模式
	mode := i.determineMode()

	// 构建纹理命令
	cmd := laya.TextureCommand{
		Texture: i.packageItem,
		Mode:    mode,
		Dest: laya.Rect{
			W: i.GObject.Width(),
			H: i.GObject.Height(),
		},
		Color:          i.color,
		ScaleX:         i.flipScaleX(),
		ScaleY:         i.flipScaleY(),
		OffsetX:        i.flipOffsetX(),
		OffsetY:        i.flipOffsetY(),
		ScaleByTile:    i.scaleByTile,
		TileGridIndice: i.tileGridIndice,
		// 填充相关参数
		FillMethod:     i.fillMethod,
		FillAmount:     i.fillAmount,
		FillOrigin:     i.fillOrigin,
		FillClockwise:  i.fillClockwise,
	}

	// 设置 Scale9Grid（如果存在）
	if i.packageItem.Scale9Grid != nil {
		cmd.Scale9Grid = &laya.Rect{
			X: float64(i.packageItem.Scale9Grid.X),
			Y: float64(i.packageItem.Scale9Grid.Y),
			W: float64(i.packageItem.Scale9Grid.Width),
			H: float64(i.packageItem.Scale9Grid.Height),
		}
	}

	// 记录命令
	gfx.DrawTexture(cmd)
	sprite.Repaint()
}

// determineMode 根据配置确定渲染模式
func (i *GImage) determineMode() laya.TextureCommandMode {
	if i.packageItem.Scale9Grid != nil {
		return laya.TextureModeScale9
	}
	if i.scaleByTile {
		return laya.TextureModeTile
	}
	return laya.TextureModeSimple
}

// flipScaleX 返回水平翻转的缩放系数
func (i *GImage) flipScaleX() float64 {
	if i.flip == FlipTypeHorizontal || i.flip == FlipTypeBoth {
		return -1.0
	}
	return 1.0
}

// flipScaleY 返回垂直翻转的缩放系数
func (i *GImage) flipScaleY() float64 {
	if i.flip == FlipTypeVertical || i.flip == FlipTypeBoth {
		return -1.0
	}
	return 1.0
}

// flipOffsetX 返回水平翻转的偏移
func (i *GImage) flipOffsetX() float64 {
	if i.flip == FlipTypeHorizontal || i.flip == FlipTypeBoth {
		return i.GObject.Width()
	}
	return 0
}

// flipOffsetY 返回垂直翻转的偏移
func (i *GImage) flipOffsetY() float64 {
	if i.flip == FlipTypeVertical || i.flip == FlipTypeBoth {
		return i.GObject.Height()
	}
	return 0
}

// SetupBeforeAdd parses image-specific metadata from the component buffer.
// 对应 TypeScript 版本 GImage.setup_beforeAdd
func (i *GImage) SetupBeforeAdd(buf *utils.ByteBuffer, beginPos int) {
	if i == nil || buf == nil {
		return
	}

	// 首先调用父类处理基础属性（位置、尺寸、旋转等）
	i.GObject.SetupBeforeAdd(buf, beginPos)

	// 然后处理GImage特定属性（color, flip, fillMethod等）
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(beginPos, 5) || buf.Remaining() <= 0 {
		return
	}
	if buf.ReadBool() && buf.Remaining() >= 4 {
		i.SetColor(buf.ReadColorString(true))
	}
	if buf.Remaining() > 0 {
		i.SetFlip(FlipType(buf.ReadByte()))
	}
	if buf.Remaining() > 0 {
		method := int(buf.ReadByte())
		log.Printf("[SetupBeforeAdd] GImage name=%s, method=%d, remaining=%d", i.GObject.Name(), method, buf.Remaining())
		if method != 0 && buf.Remaining() >= 6 {
			origin := int(buf.ReadByte())
			clockwise := buf.ReadBool()
			amount := float64(buf.ReadFloat32())
			log.Printf("[SetupBeforeAdd] GImage name=%s, SetFill called: method=%d, origin=%d, clockwise=%v, amount=%.2f",
				i.GObject.Name(), method, origin, clockwise, amount)
			i.SetFill(method, origin, clockwise, amount)
		} else {
			log.Printf("[SetupBeforeAdd] GImage name=%s, skip SetFill: method=%d, need6bytes=%v", i.GObject.Name(), method, buf.Remaining() >= 6)
		}
	}
}
