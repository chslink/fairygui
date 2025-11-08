package core

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
)

// TestOverflowHiddenAppliesScrollRect 验证 Overflow: Hidden 设置 scrollRect
func TestOverflowHiddenAppliesScrollRect(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(200, 200)

	// 设置 overflow 为 hidden
	comp.SetupOverflow(OverflowHidden)

	// 验证 scrollRect 是否被设置
	display := comp.DisplayObject()
	if display == nil {
		t.Fatal("DisplayObject is nil")
	}

	// Overflow: hidden 应该设置 scrollRect 在 display 或 container 上
	container := comp.Container()
	var scrollRect *laya.Rect

	if container != nil && container != display {
		scrollRect = container.ScrollRect()
	}
	if scrollRect == nil {
		scrollRect = display.ScrollRect()
	}

	if scrollRect == nil {
		t.Error("Overflow: hidden should set scrollRect, but it's nil")
		return
	}

	// 验证 scrollRect 尺寸应该等于组件尺寸
	if scrollRect.W != 200 || scrollRect.H != 200 {
		t.Errorf("scrollRect size = (%.1f, %.1f), want (200, 200)", scrollRect.W, scrollRect.H)
	}
}

// TestOverflowScrollAppliesScrollRect 验证 Overflow: Scroll 也设置 scrollRect
// 这是关键测试：Free Scroll 应该裁剪内容（设置 scrollRect）
func TestOverflowScrollAppliesScrollRect(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(225, 310)

	// 设置 overflow 为 scroll（对应 Free Scroll 场景）
	comp.EnsureScrollPane(ScrollTypeBoth)

	// 验证 scrollRect 是否被设置
	scrollPane := comp.ScrollPane()
	if scrollPane == nil {
		t.Fatal("ScrollPane should be created for overflow=scroll")
	}

	// 验证 maskContainer 是否有 scrollRect
	// newScrollPane 创建的结构是：display -> maskContainer -> container
	// scrollRect 设置在 maskContainer 上
	display := comp.DisplayObject()
	container := comp.Container()

	var scrollRect *laya.Rect
	if container != nil && container != display {
		// ScrollPane 场景：检查 maskContainer（即 container.Parent()）
		if parent := container.Parent(); parent != nil && parent != display {
			scrollRect = parent.ScrollRect()
		}
		// 回退检查 container 自己
		if scrollRect == nil {
			scrollRect = container.ScrollRect()
		}
	}
	if scrollRect == nil {
		// 回退检查 display
		scrollRect = display.ScrollRect()
	}

	if scrollRect == nil {
		t.Error("Overflow: scroll should set scrollRect for clipping, but it's nil")
		t.Error("This causes the 'Free Scroll' scene to not clip content")
		return
	}

	// 验证 scrollRect 尺寸
	viewSize := scrollPane.viewSize
	if scrollRect.W != viewSize.X || scrollRect.H != viewSize.Y {
		t.Errorf("scrollRect size = (%.1f, %.1f), want (%.1f, %.1f)",
			scrollRect.W, scrollRect.H, viewSize.X, viewSize.Y)
	}
}

// TestOverflowScrollContentClipping 验证滚动时内容被正确裁剪
func TestOverflowScrollContentClipping(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(225, 310)

	// 添加一个超出边界的子对象
	child := NewGObject()
	child.SetPosition(50, 50)
	child.SetSize(300, 400) // 超出父组件尺寸
	comp.AddChild(child)

	// 设置 overflow 为 scroll
	comp.EnsureScrollPane(ScrollTypeBoth)

	scrollPane := comp.ScrollPane()
	if scrollPane == nil {
		t.Fatal("ScrollPane not created")
	}

	// 设置内容尺寸（应该是子对象的边界）
	scrollPane.SetContentSize(350, 450)

	// 验证 scrollRect 是否正确裁剪
	display := comp.DisplayObject()
	container := comp.Container()

	var scrollRect *laya.Rect
	if container != nil && container != display {
		// ScrollPane 场景：检查 maskContainer（即 container.Parent()）
		if parent := container.Parent(); parent != nil && parent != display {
			scrollRect = parent.ScrollRect()
		}
		if scrollRect == nil {
			scrollRect = container.ScrollRect()
		}
	}
	if scrollRect == nil {
		scrollRect = display.ScrollRect()
	}

	if scrollRect == nil {
		t.Fatal("scrollRect not set")
	}

	// scrollRect 应该限制可见区域为组件尺寸
	if scrollRect.W > 225 || scrollRect.H > 310 {
		t.Errorf("scrollRect size (%.1f, %.1f) should not exceed component size (225, 310)",
			scrollRect.W, scrollRect.H)
	}

	// 验证可以滚动
	scrollPane.SetPos(50, 50, false)
	if scrollPane.PosX() != 50 || scrollPane.PosY() != 50 {
		t.Errorf("scroll position = (%.1f, %.1f), want (50, 50)",
			scrollPane.PosX(), scrollPane.PosY())
	}
}

// TestOverflowTypes 验证不同 overflow 类型的行为
func TestOverflowTypes(t *testing.T) {
	tests := []struct {
		name            string
		overflow        OverflowType
		shouldClip      bool
		shouldScroll    bool
		setupFunc       func(*GComponent)
	}{
		{
			name:       "Overflow: visible",
			overflow:   OverflowVisible,
			shouldClip: false,
			shouldScroll: false,
			setupFunc: func(c *GComponent) {
				// Visible 不需要特殊设置
			},
		},
		{
			name:       "Overflow: hidden",
			overflow:   OverflowHidden,
			shouldClip: true,
			shouldScroll: false,
			setupFunc: func(c *GComponent) {
				c.SetupOverflow(OverflowHidden)
			},
		},
		{
			name:       "Overflow: scroll (free scroll)",
			overflow:   OverflowScroll,
			shouldClip: true,
			shouldScroll: true,
			setupFunc: func(c *GComponent) {
				c.EnsureScrollPane(ScrollTypeBoth)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewGComponent()
			comp.SetSize(200, 200)

			tt.setupFunc(comp)

			// 检查裁剪
			display := comp.DisplayObject()
			container := comp.Container()
			var scrollRect *laya.Rect
			if container != nil && container != display {
				// ScrollPane 场景：检查 maskContainer（即 container.Parent()）
				if parent := container.Parent(); parent != nil && parent != display {
					scrollRect = parent.ScrollRect()
				}
				if scrollRect == nil {
					scrollRect = container.ScrollRect()
				}
			}
			if scrollRect == nil {
				scrollRect = display.ScrollRect()
			}

			hasScrollRect := (scrollRect != nil)
			if hasScrollRect != tt.shouldClip {
				t.Errorf("hasScrollRect = %v, want %v", hasScrollRect, tt.shouldClip)
			}

			// 检查滚动
			hasScrollPane := (comp.ScrollPane() != nil)
			if hasScrollPane != tt.shouldScroll {
				t.Errorf("hasScrollPane = %v, want %v", hasScrollPane, tt.shouldScroll)
			}
		})
	}
}

// TestScrollPaneWithMargin 测试带 margin 的 ScrollPane
func TestScrollPaneWithMargin(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(225, 343)
	
	// 设置 margin
	margin := Margin{Left: 30, Top: 30, Right: 30, Bottom: 30}
	comp.SetMargin(margin)
	
	// 创建 ScrollPane
	scrollPane := comp.EnsureScrollPane(ScrollTypeBoth)
	
	// 验证 container 的位置
	container := comp.Container()
	if container != nil && container.Parent() != nil {
		// container 应该在 maskContainer 下
		maskContainer := container.Parent()
		pos := container.Position()
		
		t.Logf("Container position: (%.1f, %.1f)", pos.X, pos.Y)
		t.Logf("Expected position: (%.1f, %.1f)", float64(margin.Left), float64(margin.Top))
		
		if pos.X != float64(margin.Left) {
			t.Errorf("Container X position = %.1f, want %.1f", pos.X, float64(margin.Left))
		}
		if pos.Y != float64(margin.Top) {
			t.Errorf("Container Y position = %.1f, want %.1f", pos.Y, float64(margin.Top))
		}
		
		// 验证 maskContainer 的 scrollRect
		scrollRect := maskContainer.ScrollRect()
		if scrollRect != nil {
			t.Logf("ScrollRect: X=%.1f, Y=%.1f, W=%.1f, H=%.1f", 
				scrollRect.X, scrollRect.Y, scrollRect.W, scrollRect.H)
			
			// scrollRect 的位置应该考虑 margin
			expectedX := float64(margin.Left)
			expectedY := float64(margin.Top)
			expectedW := 225 - float64(margin.Left + margin.Right)
			expectedH := 343 - float64(margin.Top + margin.Bottom)
			
			t.Logf("Expected ScrollRect: X=%.1f, Y=%.1f, W=%.1f, H=%.1f",
				expectedX, expectedY, expectedW, expectedH)
		}
	}
	
	// 验证 viewSize 是否扣除了 margin
	viewWidth := scrollPane.ViewWidth()
	viewHeight := scrollPane.ViewHeight()
	
	expectedViewWidth := 225 - float64(margin.Left + margin.Right)
	expectedViewHeight := 343 - float64(margin.Top + margin.Bottom)
	
	t.Logf("ViewSize: (%.1f, %.1f)", viewWidth, viewHeight)
	t.Logf("Expected ViewSize: (%.1f, %.1f)", expectedViewWidth, expectedViewHeight)
	
	if viewWidth != expectedViewWidth {
		t.Errorf("ViewWidth = %.1f, want %.1f", viewWidth, expectedViewWidth)
	}
	if viewHeight != expectedViewHeight {
		t.Errorf("ViewHeight = %.1f, want %.1f", viewHeight, expectedViewHeight)
	}
}
