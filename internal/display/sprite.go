// Package display 提供显示对象的内部实现。
//
// 这个包不应该被外部直接使用，所有公开 API 通过根目录的 fairygui 包暴露。
package display

import (
	"fmt"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
)

var spriteCounter uint64

// Sprite 是显示对象的底层实现。
//
// 它直接基于 Ebiten 设计，不依赖 LayaAir 兼容层。
type Sprite struct {
	// 标识
	id   string
	name string

	// 层级关系
	parent   *Sprite
	children []*Sprite

	// 位置和尺寸
	x      float64
	y      float64
	width  float64
	height float64

	// 变换
	scaleX   float64
	scaleY   float64
	rotation float64 // 弧度
	skewX    float64 // 弧度
	skewY    float64 // 弧度
	pivotX   float64 // 0-1 归一化
	pivotY   float64 // 0-1 归一化

	// 显示属性
	visible bool
	alpha   float64

	// 交互
	touchable bool

	// 渲染数据
	texture  *ebiten.Image             // 纹理（如果有）
	drawOpts *ebiten.DrawImageOptions  // 绘制选项

	// 自定义数据
	data interface{}

	// 所有者（通常是 fairygui.Object）
	owner interface{}

	// 脏标记
	transformDirty bool
}

// NewSprite 创建一个新的 Sprite。
func NewSprite() *Sprite {
	counter := atomic.AddUint64(&spriteCounter, 1)
	return &Sprite{
		id:       fmt.Sprintf("sprite-%d", counter),
		visible:  true,
		alpha:    1.0,
		scaleX:   1.0,
		scaleY:   1.0,
		touchable: true,
		transformDirty: true,
	}
}

// ============================================================================
// 标识
// ============================================================================

// ID 返回 Sprite 的唯一标识符。
func (s *Sprite) ID() string {
	return s.id
}

// Name 返回 Sprite 的名称。
func (s *Sprite) Name() string {
	return s.name
}

// SetName 设置 Sprite 的名称。
func (s *Sprite) SetName(name string) {
	s.name = name
}

// ============================================================================
// 位置和尺寸
// ============================================================================

// Position 返回 Sprite 的位置（相对于父对象）。
func (s *Sprite) Position() (x, y float64) {
	return s.x, s.y
}

// SetPosition 设置 Sprite 的位置（相对于父对象）。
func (s *Sprite) SetPosition(x, y float64) {
	if s.x != x || s.y != y {
		s.x = x
		s.y = y
		s.transformDirty = true
	}
}

// Size 返回 Sprite 的尺寸。
func (s *Sprite) Size() (width, height float64) {
	return s.width, s.height
}

// SetSize 设置 Sprite 的尺寸。
func (s *Sprite) SetSize(width, height float64) {
	s.width = width
	s.height = height
}

// ============================================================================
// 变换
// ============================================================================

// Scale 返回 Sprite 的缩放比例。
func (s *Sprite) Scale() (scaleX, scaleY float64) {
	return s.scaleX, s.scaleY
}

// SetScale 设置 Sprite 的缩放比例。
func (s *Sprite) SetScale(scaleX, scaleY float64) {
	if s.scaleX != scaleX || s.scaleY != scaleY {
		s.scaleX = scaleX
		s.scaleY = scaleY
		s.transformDirty = true
	}
}

// Rotation 返回 Sprite 的旋转角度（弧度）。
func (s *Sprite) Rotation() float64 {
	return s.rotation
}

// SetRotation 设置 Sprite 的旋转角度（弧度）。
func (s *Sprite) SetRotation(rotation float64) {
	if s.rotation != rotation {
		s.rotation = rotation
		s.transformDirty = true
	}
}

// Skew 返回 Sprite 的倾斜角度（弧度）。
func (s *Sprite) Skew() (skewX, skewY float64) {
	return s.skewX, s.skewY
}

// SetSkew 设置 Sprite 的倾斜角度（弧度）。
func (s *Sprite) SetSkew(skewX, skewY float64) {
	if s.skewX != skewX || s.skewY != skewY {
		s.skewX = skewX
		s.skewY = skewY
		s.transformDirty = true
	}
}

// Pivot 返回 Sprite 的锚点（0-1 归一化坐标）。
func (s *Sprite) Pivot() (pivotX, pivotY float64) {
	return s.pivotX, s.pivotY
}

// SetPivot 设置 Sprite 的锚点（0-1 归一化坐标）。
func (s *Sprite) SetPivot(pivotX, pivotY float64) {
	if s.pivotX != pivotX || s.pivotY != pivotY {
		s.pivotX = pivotX
		s.pivotY = pivotY
		s.transformDirty = true
	}
}

// ============================================================================
// 显示属性
// ============================================================================

// Visible 返回 Sprite 是否可见。
func (s *Sprite) Visible() bool {
	return s.visible
}

// SetVisible 设置 Sprite 是否可见。
func (s *Sprite) SetVisible(visible bool) {
	s.visible = visible
}

// Alpha 返回 Sprite 的透明度（0-1）。
func (s *Sprite) Alpha() float64 {
	return s.alpha
}

// SetAlpha 设置 Sprite 的透明度（0-1）。
func (s *Sprite) SetAlpha(alpha float64) {
	if alpha < 0 {
		alpha = 0
	} else if alpha > 1 {
		alpha = 1
	}
	s.alpha = alpha
}

// ============================================================================
// 交互
// ============================================================================

// Touchable 返回 Sprite 是否可以接收触摸事件。
func (s *Sprite) Touchable() bool {
	return s.touchable
}

// SetTouchable 设置 Sprite 是否可以接收触摸事件。
func (s *Sprite) SetTouchable(touchable bool) {
	s.touchable = touchable
}

// ============================================================================
// 层级关系
// ============================================================================

// Parent 返回父 Sprite。
func (s *Sprite) Parent() *Sprite {
	return s.parent
}

// Children 返回子 Sprite 列表（只读副本）。
func (s *Sprite) Children() []*Sprite {
	result := make([]*Sprite, len(s.children))
	copy(result, s.children)
	return result
}

// ChildCount 返回子对象数量。
func (s *Sprite) ChildCount() int {
	return len(s.children)
}

// AddChild 添加子 Sprite。
func (s *Sprite) AddChild(child *Sprite) {
	s.AddChildAt(child, len(s.children))
}

// AddChildAt 在指定索引位置添加子 Sprite。
func (s *Sprite) AddChildAt(child *Sprite, index int) {
	if child == nil || child == s {
		return
	}

	// 如果已经有父对象，先移除
	if child.parent != nil {
		child.parent.RemoveChild(child)
	}

	// 规范化索引
	if index < 0 {
		index = 0
	} else if index > len(s.children) {
		index = len(s.children)
	}

	// 插入子对象
	s.children = append(s.children, nil)
	copy(s.children[index+1:], s.children[index:])
	s.children[index] = child
	child.parent = s
}

// RemoveChild 移除子 Sprite。
func (s *Sprite) RemoveChild(child *Sprite) bool {
	for i, c := range s.children {
		if c == child {
			s.RemoveChildAt(i)
			return true
		}
	}
	return false
}

// RemoveChildAt 移除指定索引位置的子 Sprite。
func (s *Sprite) RemoveChildAt(index int) *Sprite {
	if index < 0 || index >= len(s.children) {
		return nil
	}

	child := s.children[index]
	child.parent = nil

	// 移除子对象
	s.children = append(s.children[:index], s.children[index+1:]...)

	return child
}

// RemoveAllChildren 移除所有子对象。
func (s *Sprite) RemoveAllChildren() {
	for _, child := range s.children {
		child.parent = nil
	}
	s.children = s.children[:0]
}

// GetChildAt 返回指定索引位置的子 Sprite。
func (s *Sprite) GetChildAt(index int) *Sprite {
	if index < 0 || index >= len(s.children) {
		return nil
	}
	return s.children[index]
}

// GetChildByName 返回指定名称的子 Sprite（只查找直接子对象）。
func (s *Sprite) GetChildByName(name string) *Sprite {
	for _, child := range s.children {
		if child.name == name {
			return child
		}
	}
	return nil
}

// GetChildIndex 返回子对象的索引，如果不是子对象返回 -1。
func (s *Sprite) GetChildIndex(child *Sprite) int {
	for i, c := range s.children {
		if c == child {
			return i
		}
	}
	return -1
}

// SetChildIndex 设置子对象的索引（调整层级顺序）。
func (s *Sprite) SetChildIndex(child *Sprite, index int) {
	oldIndex := s.GetChildIndex(child)
	if oldIndex == -1 || oldIndex == index {
		return
	}

	// 移除旧位置
	s.children = append(s.children[:oldIndex], s.children[oldIndex+1:]...)

	// 插入新位置
	if index < 0 {
		index = 0
	} else if index > len(s.children) {
		index = len(s.children)
	}

	s.children = append(s.children, nil)
	copy(s.children[index+1:], s.children[index:])
	s.children[index] = child
}

// ============================================================================
// 渲染
// ============================================================================

// Texture 返回 Sprite 的纹理。
func (s *Sprite) Texture() *ebiten.Image {
	return s.texture
}

// SetTexture 设置 Sprite 的纹理。
func (s *Sprite) SetTexture(texture *ebiten.Image) {
	s.texture = texture
	if texture != nil && s.width == 0 && s.height == 0 {
		// 如果没有设置尺寸，使用纹理尺寸
		bounds := texture.Bounds()
		s.width = float64(bounds.Dx())
		s.height = float64(bounds.Dy())
	}
}

// DrawOptions 返回绘制选项（懒加载）。
func (s *Sprite) DrawOptions() *ebiten.DrawImageOptions {
	if s.drawOpts == nil {
		s.drawOpts = &ebiten.DrawImageOptions{}
	}
	return s.drawOpts
}

// UpdateDrawOptions 更新绘制选项（根据当前变换）。
func (s *Sprite) UpdateDrawOptions() {
	if !s.transformDirty {
		return
	}

	opts := s.DrawOptions()
	opts.GeoM.Reset()

	// 1. 应用锚点偏移
	if s.pivotX != 0 || s.pivotY != 0 {
		opts.GeoM.Translate(-s.width*s.pivotX, -s.height*s.pivotY)
	}

	// 2. 应用倾斜
	if s.skewX != 0 || s.skewY != 0 {
		// 倾斜变换矩阵
		// [1, tan(skewY), 0]
		// [tan(skewX), 1, 0]
		// [0, 0, 1]
		m := opts.GeoM
		if s.skewX != 0 {
			m.Skew(0, s.skewX)
		}
		if s.skewY != 0 {
			m.Skew(s.skewY, 0)
		}
	}

	// 3. 应用缩放
	if s.scaleX != 1.0 || s.scaleY != 1.0 {
		opts.GeoM.Scale(s.scaleX, s.scaleY)
	}

	// 4. 应用旋转
	if s.rotation != 0 {
		opts.GeoM.Rotate(s.rotation)
	}

	// 5. 应用位置
	opts.GeoM.Translate(s.x, s.y)

	// 6. 应用透明度
	opts.ColorScale.Reset()
	opts.ColorScale.ScaleAlpha(float32(s.alpha))

	s.transformDirty = false
}

// Draw 在屏幕上绘制 Sprite。
func (s *Sprite) Draw(screen *ebiten.Image) {
	if !s.visible || s.alpha <= 0 {
		return
	}

	// 更新绘制选项
	s.UpdateDrawOptions()

	// 如果有纹理，绘制纹理
	if s.texture != nil {
		screen.DrawImage(s.texture, s.drawOpts)
	}

	// 绘制子对象
	for _, child := range s.children {
		child.Draw(screen)
	}
}

// ============================================================================
// 自定义数据
// ============================================================================

// Data 返回自定义数据。
func (s *Sprite) Data() interface{} {
	return s.data
}

// SetData 设置自定义数据。
func (s *Sprite) SetData(data interface{}) {
	s.data = data
}

// Owner 返回所有者对象。
func (s *Sprite) Owner() interface{} {
	return s.owner
}

// SetOwner 设置所有者对象。
func (s *Sprite) SetOwner(owner interface{}) {
	s.owner = owner
}

// ============================================================================
// 辅助方法
// ============================================================================

// Dispose 释放资源。
func (s *Sprite) Dispose() {
	// 移除所有子对象
	s.RemoveAllChildren()

	// 从父对象移除
	if s.parent != nil {
		s.parent.RemoveChild(s)
	}

	// 清空引用
	s.texture = nil
	s.drawOpts = nil
	s.data = nil
	s.owner = nil
}

// String 返回 Sprite 的字符串表示（用于调试）。
func (s *Sprite) String() string {
	if s.name != "" {
		return fmt.Sprintf("Sprite(%s: %s)", s.id, s.name)
	}
	return fmt.Sprintf("Sprite(%s)", s.id)
}
