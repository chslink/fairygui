package display

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// 坐标变换
// ============================================================================

// LocalToGlobal 将局部坐标转换为全局坐标（相对于舞台）。
func (s *Sprite) LocalToGlobal(localX, localY float64) (globalX, globalY float64) {
	// 构建从局部到全局的变换矩阵
	geoM := s.getLocalToGlobalMatrix()

	// 应用变换
	globalX, globalY = geoM.Apply(localX, localY)
	return
}

// GlobalToLocal 将全局坐标转换为局部坐标（相对于当前对象）。
func (s *Sprite) GlobalToLocal(globalX, globalY float64) (localX, localY float64) {
	// 构建从局部到全局的变换矩阵
	geoM := s.getLocalToGlobalMatrix()

	// 求逆矩阵
	invGeoM := geoM
	invGeoM.Invert()

	// 应用逆变换
	localX, localY = invGeoM.Apply(globalX, globalY)
	return
}

// getLocalToGlobalMatrix 获取从局部到全局的变换矩阵。
func (s *Sprite) getLocalToGlobalMatrix() ebiten.GeoM {
	var geoM ebiten.GeoM

	// 从当前对象开始，向上遍历到根对象
	sprite := s
	for sprite != nil {
		// 应用当前对象的变换
		m := sprite.getLocalTransformMatrix()
		m.Concat(geoM)
		geoM = m

		sprite = sprite.parent
	}

	return geoM
}

// getLocalTransformMatrix 获取局部变换矩阵（不包括父对象的变换）。
func (s *Sprite) getLocalTransformMatrix() ebiten.GeoM {
	var geoM ebiten.GeoM

	// 1. 应用锚点偏移
	if s.pivotX != 0 || s.pivotY != 0 {
		geoM.Translate(-s.width*s.pivotX, -s.height*s.pivotY)
	}

	// 2. 应用倾斜
	if s.skewX != 0 || s.skewY != 0 {
		if s.skewX != 0 {
			geoM.Skew(0, s.skewX)
		}
		if s.skewY != 0 {
			geoM.Skew(s.skewY, 0)
		}
	}

	// 3. 应用缩放
	if s.scaleX != 1.0 || s.scaleY != 1.0 {
		geoM.Scale(s.scaleX, s.scaleY)
	}

	// 4. 应用旋转
	if s.rotation != 0 {
		geoM.Rotate(s.rotation)
	}

	// 5. 应用位置
	geoM.Translate(s.x, s.y)

	return geoM
}

// GlobalPosition 返回全局位置（相对于舞台）。
func (s *Sprite) GlobalPosition() (x, y float64) {
	return s.LocalToGlobal(0, 0)
}

// Bounds 返回全局边界矩形（相对于舞台）。
func (s *Sprite) Bounds() (x, y, width, height float64) {
	// 获取四个角的全局坐标
	tlX, tlY := s.LocalToGlobal(0, 0)
	trX, trY := s.LocalToGlobal(s.width, 0)
	blX, blY := s.LocalToGlobal(0, s.height)
	brX, brY := s.LocalToGlobal(s.width, s.height)

	// 找出边界
	minX := math.Min(math.Min(tlX, trX), math.Min(blX, brX))
	maxX := math.Max(math.Max(tlX, trX), math.Max(blX, brX))
	minY := math.Min(math.Min(tlY, trY), math.Min(blY, brY))
	maxY := math.Max(math.Max(tlY, trY), math.Max(blY, brY))

	return minX, minY, maxX - minX, maxY - minY
}

// ============================================================================
// 碰撞检测
// ============================================================================

// HitTest 检测点是否在对象内部（局部坐标）。
func (s *Sprite) HitTest(localX, localY float64) bool {
	return localX >= 0 && localX <= s.width && localY >= 0 && localY <= s.height
}

// HitTestGlobal 检测点是否在对象内部（全局坐标）。
func (s *Sprite) HitTestGlobal(globalX, globalY float64) bool {
	localX, localY := s.GlobalToLocal(globalX, globalY)
	return s.HitTest(localX, localY)
}

// HitTestChildren 在子对象中查找包含指定点的对象（全局坐标）。
// 返回最上层的（最后绘制的）包含该点的子对象。
func (s *Sprite) HitTestChildren(globalX, globalY float64) *Sprite {
	// 从后向前遍历（最上层的先检测）
	for i := len(s.children) - 1; i >= 0; i-- {
		child := s.children[i]

		if !child.visible || !child.touchable {
			continue
		}

		// 先检查子对象的子对象
		if hit := child.HitTestChildren(globalX, globalY); hit != nil {
			return hit
		}

		// 再检查子对象本身
		if child.HitTestGlobal(globalX, globalY) {
			return child
		}
	}

	return nil
}
