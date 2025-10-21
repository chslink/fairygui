package widgets

import (
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
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
}

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
	}
}

// PackageItem returns the package item backing this image.
func (i *GImage) PackageItem() *assets.PackageItem {
	return i.packageItem
}

// SetColor records the applied tint colour in hex format.
func (i *GImage) SetColor(value string) {
	i.color = value
}

// Color returns the stored tint colour.
func (i *GImage) Color() string {
	return i.color
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
