package core

import (
	"testing"
)

// TestSetBoundsPreservesContentSize 验证 SetBounds 不会错误覆盖 contentSize
// 这个测试重现并验证修复的问题：
// 1. 初始设置合理的 contentSize（大于 viewSize）
// 2. UpdateBounds 计算出更小的 contentSize（因为子对象未完全创建）
// 3. SetBounds 应该拒绝这个更小的值，保留原有的 contentSize
func TestSetBoundsPreservesContentSize(t *testing.T) {
	// 创建组件和 ScrollPane
	comp := NewGComponent()
	comp.SetSize(425, 120)

	// 创建 ScrollPane 并设置 viewSize
	scrollPane := newScrollPane(comp)
	scrollPane.SetViewSize(408, 103) // viewSize 小于初始 contentSize

	// 设置初始 contentSize（模拟从布局文件读取的正确值）
	scrollPane.SetContentSize(425, 120)

	currentSize := scrollPane.ContentSize()
	if currentSize.Y != 120 {
		t.Fatalf("Initial contentSize.Y = %.1f, want 120", currentSize.Y)
	}

	// 模拟 UpdateBounds 计算出错误的更小值
	// 这种情况发生在 GList 虚拟化或子对象未完全创建时
	comp.SetBounds(0, 0, 914, 93) // 高度从 120 降到 93

	// 验证 SetBounds 保护逻辑是否生效
	newSize := scrollPane.ContentSize()
	if newSize.Y != 120 {
		t.Errorf("contentSize.Y = %.1f after SetBounds, want 120 (should be preserved)", newSize.Y)
		t.Errorf("SetBounds incorrectly overwrote contentSize from 120 to %.1f", newSize.Y)
	}

	// 验证宽度被正确更新（因为宽度没有异常缩小）
	if newSize.X != 914 {
		t.Errorf("contentSize.X = %.1f, want 914", newSize.X)
	}
}

// TestSetBoundsAllowsNormalUpdates 验证 SetBounds 在正常情况下仍然工作
func TestSetBoundsAllowsNormalUpdates(t *testing.T) {
	// 创建组件和 ScrollPane
	comp := NewGComponent()
	comp.SetSize(425, 120)

	scrollPane := newScrollPane(comp)
	scrollPane.SetViewSize(408, 103)
	scrollPane.SetContentSize(425, 120)

	// 正常情况1：内容变大（应该允许）
	comp.SetBounds(0, 0, 1000, 200)
	newSize := scrollPane.ContentSize()
	if newSize.Y != 200 {
		t.Errorf("contentSize.Y = %.1f, want 200 (normal increase should be allowed)", newSize.Y)
	}

	// 正常情况2：内容适度缩小（>= viewSize，减少不超过 20%）
	scrollPane.SetContentSize(425, 150) // 重置
	comp.SetBounds(0, 0, 400, 130)      // 缩小到 130（150 * 0.867）
	newSize = scrollPane.ContentSize()
	if newSize.Y != 130 {
		t.Errorf("contentSize.Y = %.1f, want 130 (moderate decrease should be allowed)", newSize.Y)
	}
}

// TestSetBoundsWithoutScrollPane 验证没有 scrollPane 时不会崩溃
func TestSetBoundsWithoutScrollPane(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(100, 100)

	// 没有 scrollPane，SetBounds 应该正常工作
	comp.SetBounds(0, 0, 200, 150)

	// 不应该崩溃
}

// TestUpdateBoundsCalculation 验证 UpdateBounds 的基本计算
func TestUpdateBoundsCalculation(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(200, 200)

	// 添加子对象
	child1 := NewGObject()
	child1.SetPosition(10, 10)
	child1.SetSize(50, 50)
	comp.AddChild(child1)

	child2 := NewGObject()
	child2.SetPosition(100, 100)
	child2.SetSize(80, 80)
	comp.AddChild(child2)

	// 创建 ScrollPane
	scrollPane := newScrollPane(comp)
	scrollPane.SetViewSize(150, 150)

	// 调用 UpdateBounds
	comp.UpdateBounds()

	// 验证 contentSize 应该是子对象的边界框
	// ax = min(10, 100) = 10
	// ay = min(10, 100) = 10
	// ar = max(10+50, 100+80) = 180
	// ab = max(10+50, 100+80) = 180
	// contentSize = (180, 180)
	contentSize := scrollPane.ContentSize()
	if contentSize.X != 180 {
		t.Errorf("contentSize.X = %.1f, want 180", contentSize.X)
	}
	if contentSize.Y != 180 {
		t.Errorf("contentSize.Y = %.1f, want 180", contentSize.Y)
	}
}

// TestSetBoundsPreservationConditions 验证保护条件的边界情况
func TestSetBoundsPreservationConditions(t *testing.T) {
	tests := []struct {
		name           string
		currentHeight  float64
		newHeight      float64
		viewHeight     float64
		shouldPreserve bool
	}{
		{
			name:           "Should preserve: large drop below viewHeight",
			currentHeight:  120,
			newHeight:      93,
			viewHeight:     103,
			shouldPreserve: true, // 93 < 103 && 93 < 120*0.8
		},
		{
			name:           "Should NOT preserve: small drop",
			currentHeight:  120,
			newHeight:      100,
			viewHeight:     95,
			shouldPreserve: false, // 100 >= 120*0.8
		},
		{
			name:           "Should NOT preserve: increase",
			currentHeight:  120,
			newHeight:      150,
			viewHeight:     100,
			shouldPreserve: false, // newHeight > currentHeight
		},
		{
			name:           "Should NOT preserve: current below viewHeight",
			currentHeight:  80,
			newHeight:      60,
			viewHeight:     100,
			shouldPreserve: false, // currentHeight < viewHeight
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewGComponent()
			comp.SetSize(200, 200)

			scrollPane := newScrollPane(comp)
			scrollPane.SetViewSize(200, tt.viewHeight)
			scrollPane.SetContentSize(200, tt.currentHeight)

			comp.SetBounds(0, 0, 200, tt.newHeight)

			actualHeight := scrollPane.ContentSize().Y
			if tt.shouldPreserve {
				if actualHeight != tt.currentHeight {
					t.Errorf("Expected contentSize to be preserved at %.1f, got %.1f",
						tt.currentHeight, actualHeight)
				}
			} else {
				if actualHeight != tt.newHeight {
					t.Errorf("Expected contentSize to be updated to %.1f, got %.1f",
						tt.newHeight, actualHeight)
				}
			}
		})
	}
}
