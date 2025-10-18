package assets

import "github.com/chslink/fairygui/pkg/fgui/utils"

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
}
