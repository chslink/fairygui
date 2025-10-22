package assets

import (
	"math"
	"sync"

	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// PackageItemType mirrors the TypeScript enum used by FairyGUI.
type PackageItemType uint8

const (
	PackageItemTypeImage PackageItemType = iota
	PackageItemTypeMovieClip
	PackageItemTypeSound
	PackageItemTypeComponent
	PackageItemTypeAtlas
	PackageItemTypeFont
	PackageItemTypeSwf
	PackageItemTypeMisc
	PackageItemTypeUnknown
	PackageItemTypeSpine
	PackageItemTypeDragonBones
)

// OverflowType mirrors FairyGUI 的 overflow 枚举。
type OverflowType uint8

const (
	OverflowTypeVisible OverflowType = iota
	OverflowTypeHidden
	OverflowTypeScroll
)

// ObjectType mirrors FairyGUI's runtime object types.
type ObjectType uint8

const (
	ObjectTypeImage ObjectType = iota
	ObjectTypeMovieClip
	ObjectTypeSwf
	ObjectTypeGraph
	ObjectTypeLoader
	ObjectTypeGroup
	ObjectTypeText
	ObjectTypeRichText
	ObjectTypeInputText
	ObjectTypeComponent
	ObjectTypeList
	ObjectTypeLabel
	ObjectTypeButton
	ObjectTypeComboBox
	ObjectTypeProgressBar
	ObjectTypeSlider
	ObjectTypeScrollBar
	ObjectTypeTree
	ObjectTypeLoader3D
)

// Rect represents an axis-aligned rectangle.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// Point represents a 2D point.
type Point struct {
	X float32
	Y float32
}

// PackageItem holds metadata for a single UI resource within a package.
type PackageItem struct {
	Owner *Package

	Type       PackageItemType
	ObjectType ObjectType

	ID     string
	Name   string
	File   string
	Width  int
	Height int

	Scale9Grid     *Rect
	TileGridIndice int
	ScaleByTile    bool
	Smoothing      bool

	RawData *utils.ByteBuffer

	Branches       []string
	HighResolution []string
	SkeletonAnchor *Point

	Atlas        *PackageItem
	Sprite       *AtlasSprite
	PixelHitTest *PixelHitTestData
	Component    *ComponentData

	bitmapFont *BitmapFont
	fontOnce   sync.Once
	fontErr    error
}

// AtlasSprite describes a sprite on an atlas texture.
type AtlasSprite struct {
	Atlas        *PackageItem
	Rect         Rect
	Offset       Point
	OriginalSize Point
	Rotated      bool
}

// PixelHitTestData mirrors FairyGUI's hit test data.
type PixelHitTestData struct {
	Width  int
	Height int
	Scale  float32
	Data   []byte
}

// Load fills the hit test data from the provided buffer.
func (d *PixelHitTestData) Load(buf *utils.ByteBuffer) {
	if buf == nil {
		return
	}
	_ = buf.ReadInt32() // reserved
	d.Width = int(buf.ReadInt32())
	scale := int(buf.ReadByte())
	if scale < 0 {
		scale += 256
	}
	if scale == 0 {
		d.Scale = 1
	} else {
		d.Scale = 1 / float32(scale)
	}
	length := int(buf.ReadInt32())
	d.Data = buf.ReadBytes(length)
	if d.Width > 0 {
		bits := length * 8
		d.Height = (bits + d.Width - 1) / d.Width
	}
}

// Contains reports whether the local coordinate hits an opaque pixel.
func (d *PixelHitTestData) Contains(x, y float64) bool {
	if d == nil || len(d.Data) == 0 || d.Width == 0 {
		return true
	}
	px := int(math.Floor(x * float64(d.Scale)))
	py := int(math.Floor(y * float64(d.Scale)))
	if px < 0 || py < 0 || px >= d.Width {
		return false
	}
	if d.Height > 0 && py >= d.Height {
		return false
	}
	index := py*d.Width + px
	byteIndex := index >> 3
	bit := uint(index & 7)
	if byteIndex < 0 || byteIndex >= len(d.Data) {
		return false
	}
	return (d.Data[byteIndex]>>bit)&0x1 == 1
}

// ComponentData represents parsed metadata for component package items.
type ComponentData struct {
	SourceWidth  int
	SourceHeight int
	InitWidth    int
	InitHeight   int
	MinWidth     int
	MaxWidth     int
	MinHeight    int
	MaxHeight    int
	PivotX       float32
	PivotY       float32
	PivotAnchor  bool
	Margin       Margin
	Overflow     OverflowType

	Children    []ComponentChild
	Controllers []ControllerData
}

// Margin describes component margins.
type Margin struct {
	Top    int
	Bottom int
	Left   int
	Right  int
}

// ComponentChild describes a child declared in component raw data.
type ComponentChild struct {
	ID            string
	Name          string
	Type          ObjectType
	Src           string
	PackageID     string
	X             int
	Y             int
	Width         int
	Height        int
	MinWidth      int
	MaxWidth      int
	MinHeight     int
	MaxHeight     int
	ScaleX        float32
	ScaleY        float32
	SkewX         float32
	SkewY         float32
	PivotX        float32
	PivotY        float32
	PivotAnchor   bool
	Alpha         float32
	Rotation      float32
	Visible       bool
	Touchable     bool
	Grayed        bool
	Data          string
	Text          string
	Icon          string
	URL           string
	RawDataOffset int
	RawDataLength int
}

// ControllerData describes a component controller.
type ControllerData struct {
	Name      string
	PageNames []string
	PageIDs   []string
	AutoRadio bool
}
