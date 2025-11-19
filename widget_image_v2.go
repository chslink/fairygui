package fairygui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// FlipType - 翻转类型定义
// ============================================================================

type FlipType int

const (
	FlipTypeNone FlipType = iota
	FlipTypeHorizontal
	FlipTypeVertical
	FlipTypeBoth
)

// ============================================================================
// Image - 图片控件 V2 (基于新架构)
// ============================================================================

type Image struct {
	*Object

	// 资源项
	pkgItem PackageItem

	// 颜色染色 (RGBA)
	tintColor color.RGBA

	// 翻转
	flip FlipType

	// 填充参数 (用于进度条)
	fillMethod    int
	fillOrigin    int
	fillClockwise bool
	fillAmount    float64

	// 平铺
	scaleByTile    bool
	tileGridIndice int

	// 九宫格
	scale9Grid *Rect
}

func NewImage() *Image {
	img := &Image{
		Object:    NewObject(),
		tintColor: color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}

	// 图片默认不拦截事件
	img.SetTouchable(false)

	return img
}

// SetPackageItem 设置资源项
func (i *Image) SetPackageItem(item PackageItem) {
	if i.pkgItem == item {
		return
	}

	i.pkgItem = item

	// 自动设置尺寸
	if item != nil {
		i.SetSize(float64(item.Width()), float64(item.Height()))
	}

	i.updateGraphics()
}

func (i *Image) PackageItem() PackageItem {
	return i.pkgItem
}

// SetTintColor 设置染色颜色 (RGBA)
func (i *Image) SetTintColor(c color.RGBA) {
	if i.tintColor == c {
		return
	}

	i.tintColor = c
	i.updateGraphics()
}

func (i *Image) TintColor() color.RGBA {
	return i.tintColor
}

// SetFlip 设置翻转
func (i *Image) SetFlip(flip FlipType) {
	if i.flip == flip {
		return
	}

	i.flip = flip
	i.updateGraphics()
}

func (i *Image) Flip() FlipType {
	return i.flip
}

// SetFill 设置填充参数
func (i *Image) SetFill(method, origin int, clockwise bool, amount float64) {
	i.fillMethod = method
	i.fillOrigin = origin
	i.fillClockwise = clockwise
	i.fillAmount = amount

	i.updateGraphics()
}

func (i *Image) Fill() (method, origin int, clockwise bool, amount float64) {
	return i.fillMethod, i.fillOrigin, i.fillClockwise, i.fillAmount
}

// SetScale9Grid 设置九宫格
func (i *Image) SetScale9Grid(grid *Rect) {
	if i.scale9Grid == grid {
		return
	}

	i.scale9Grid = grid
	i.updateGraphics()
}

func (i *Image) Scale9Grid() *Rect {
	return i.scale9Grid
}

// updateGraphics 更新渲染
func (i *Image) updateGraphics() {
	// 简化的实现：只更新纹理
	// TODO: 实现完整的九宫格、平铺、填充

	if i.pkgItem == nil {
		i.SetTexture(nil)
		return
	}

	// TODO: 从资源包加载纹理
	// texture := loadTextureFromPackageItem(i.pkgItem)
	// i.SetTexture(texture)

	// 设置默认尺寸
	i.SetSize(float64(i.pkgItem.Width()), float64(i.pkgItem.Height()))
}

// Draw 自定义绘制
func (i *Image) Draw(screen *ebiten.Image) {
	// 先调用父类绘制
	i.Object.Draw(screen)

	// TODO: 绘制填充效果 (fillAmount)
	// TODO: 绘制遮罩
}

// AssertImage 类型断言
func AssertImage(obj DisplayObject) (*Image, bool) {
	img, ok := obj.(*Image)
	return img, ok
}

// IsImage 检查类型
func IsImage(obj DisplayObject) bool {
	_, ok := obj.(*Image)
	return ok
}
