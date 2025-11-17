package fairygui

import (
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ============================================================================
// Image - 简化的图片控件
// ============================================================================

// Image 是简化的图片控件，包装了 pkg/fgui/widgets.GImage。
type Image struct {
	img *widgets.GImage
}

// FlipType 定义图片翻转类型。
type FlipType int

const (
	// FlipTypeNone 不翻转
	FlipTypeNone FlipType = iota
	// FlipTypeHorizontal 水平翻转
	FlipTypeHorizontal
	// FlipTypeVertical 垂直翻转
	FlipTypeVertical
	// FlipTypeBoth 水平和垂直翻转
	FlipTypeBoth
)

// NewImage 创建一个新的图片控件。
//
// 示例：
//
//	img := fairygui.NewImage()
//	img.SetColor("#FF0000")  // 设置红色色调
func NewImage() *Image {
	return &Image{
		img: widgets.NewImage(),
	}
}

// Color 返回图片的色调颜色（十六进制格式）。
func (i *Image) Color() string {
	return i.img.Color()
}

// SetColor 设置图片的色调颜色（十六进制格式）。
//
// 示例：
//
//	img.SetColor("#FF0000")  // 红色
//	img.SetColor("#00FF00")  // 绿色
func (i *Image) SetColor(color string) {
	i.img.SetColor(color)
}

// Flip 返回图片的翻转模式。
func (i *Image) Flip() FlipType {
	return FlipType(i.img.Flip())
}

// SetFlip 设置图片的翻转模式。
//
// 示例：
//
//	img.SetFlip(fairygui.FlipTypeHorizontal)  // 水平翻转
func (i *Image) SetFlip(flip FlipType) {
	i.img.SetFlip(widgets.FlipType(flip))
}

// Position 返回图片位置。
func (i *Image) Position() (x, y float64) {
	return i.img.X(), i.img.Y()
}

// SetPosition 设置图片位置。
func (i *Image) SetPosition(x, y float64) {
	i.img.SetPosition(x, y)
}

// Size 返回图片大小。
func (i *Image) Size() (width, height float64) {
	return i.img.Width(), i.img.Height()
}

// SetSize 设置图片大小。
func (i *Image) SetSize(width, height float64) {
	i.img.SetSize(width, height)
}

// Visible 返回图片是否可见。
func (i *Image) Visible() bool {
	return i.img.Visible()
}

// SetVisible 设置图片可见性。
func (i *Image) SetVisible(visible bool) {
	i.img.SetVisible(visible)
}

// Name 返回图片名称。
func (i *Image) Name() string {
	return i.img.Name()
}

// SetName 设置图片名称。
func (i *Image) SetName(name string) {
	i.img.SetName(name)
}

// Alpha 返回图片透明度（0-1）。
func (i *Image) Alpha() float64 {
	return i.img.Alpha()
}

// SetAlpha 设置图片透明度（0-1）。
//
// 示例：
//
//	img.SetAlpha(0.5)  // 半透明
func (i *Image) SetAlpha(alpha float64) {
	i.img.SetAlpha(alpha)
}

// RawImage 返回底层的 widgets.GImage 对象。
//
// 仅在需要访问底层 API 时使用。
func (i *Image) RawImage() *widgets.GImage {
	return i.img
}
