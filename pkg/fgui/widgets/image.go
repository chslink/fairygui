package widgets

import (
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
func (i *GImage) OwnerSizeChanged(_, _ float64) {
	i.updateGraphics()
}

func (i *GImage) updateGraphics() {
	if i == nil || i.GObject == nil {
		return
	}
	if sprite := i.GObject.DisplayObject(); sprite != nil {
		sprite.SetMouseEnabled(i.GObject.Touchable())
		sprite.Repaint()
	}
}

// SetupBeforeAdd parses image-specific metadata from the component buffer.
func (i *GImage) SetupBeforeAdd(_ *SetupContext, buf *utils.ByteBuffer) {
	if i == nil || buf == nil {
		return
	}
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(0, 5) || buf.Remaining() <= 0 {
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
		if method != 0 && buf.Remaining() >= 6 {
			origin := int(buf.ReadByte())
			clockwise := buf.ReadBool()
			amount := float64(buf.ReadFloat32())
			i.SetFill(method, origin, clockwise, amount)
		}
	}
}
