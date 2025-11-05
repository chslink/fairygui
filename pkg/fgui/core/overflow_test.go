package core

import (
	"testing"
)

func TestOverflowHiddenCreatesContainer(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(200, 200)

	// 初始状态：container 应该等于 display
	if comp.container != comp.display {
		t.Fatal("初始状态 container 应该等于 display")
	}

	// 设置 overflow hidden
	comp.SetupOverflow(OverflowHidden)

	// 现在应该有独立的 container
	if comp.container == comp.display {
		t.Fatal("OverflowHidden 应该创建独立的 container")
	}

	// container 应该是 display 的子对象
	children := comp.display.Children()
	if len(children) != 1 {
		t.Fatalf("display 应该有 1 个子对象，实际有 %d", len(children))
	}
	if children[0] != comp.container {
		t.Fatal("container 应该是 display 的子对象")
	}
}

func TestOverflowHiddenSetsScrollRect(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(200, 200)
	comp.SetMargin(Margin{Top: 10, Bottom: 10, Left: 10, Right: 10})

	comp.SetupOverflow(OverflowHidden)

	// 应该设置 scrollRect
	rect := comp.display.ScrollRect()
	if rect == nil {
		t.Fatal("OverflowHidden 应该设置 scrollRect")
	}

	// 验证 scrollRect 尺寸（考虑 margin）
	expectedW := 200.0 - 10.0 // width - right margin
	expectedH := 200.0 - 10.0 // height - bottom margin
	if rect.W != expectedW || rect.H != expectedH {
		t.Fatalf("scrollRect 尺寸错误，期望 (%v,%v)，实际 (%v,%v)",
			expectedW, expectedH, rect.W, rect.H)
	}

	// 验证 scrollRect 偏移（margin）
	if rect.X != 10.0 || rect.Y != 10.0 {
		t.Fatalf("scrollRect 偏移错误，期望 (10,10)，实际 (%v,%v)",
			rect.X, rect.Y)
	}
}

func TestOverflowMarginOffset(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(200, 200)
	comp.SetMargin(Margin{Top: 20, Bottom: 15, Left: 10, Right: 5})

	comp.SetupOverflow(OverflowHidden)

	// container 应该偏移 (left, top)
	pos := comp.container.Position()
	if pos.X != 10.0 || pos.Y != 20.0 {
		t.Fatalf("container 偏移错误，期望 (10,20)，实际 (%v,%v)", pos.X, pos.Y)
	}
}

func TestOverflowVisibleNoContainer(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(200, 200)

	comp.SetupOverflow(OverflowVisible)

	// OverflowVisible 不应该创建独立 container（除非有 margin）
	if comp.container != comp.display {
		t.Fatal("OverflowVisible 不应该创建独立 container")
	}

	// 不应该设置 scrollRect
	if comp.display.ScrollRect() != nil {
		t.Fatal("OverflowVisible 不应该设置 scrollRect")
	}
}

func TestOverflowVisibleWithMargin(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(200, 200)
	comp.SetMargin(Margin{Top: 10, Bottom: 10, Left: 10, Right: 10})

	comp.SetupOverflow(OverflowVisible)

	// 即使是 OverflowVisible，有 margin 时也应该创建独立 container
	if comp.container == comp.display {
		t.Fatal("OverflowVisible 有 margin 时应该创建独立 container")
	}

	// 但不应该设置 scrollRect
	if comp.display.ScrollRect() != nil {
		t.Fatal("OverflowVisible 不应该设置 scrollRect")
	}

	// container 应该偏移
	pos := comp.container.Position()
	if pos.X != 10.0 || pos.Y != 10.0 {
		t.Fatalf("container 偏移错误，期望 (10,10)，实际 (%v,%v)", pos.X, pos.Y)
	}
}

func TestUpdateMaskOnSizeChange(t *testing.T) {
	comp := NewGComponent()
	comp.SetSize(200, 200)
	comp.SetMargin(Margin{Top: 10, Bottom: 10, Left: 10, Right: 10})
	comp.SetupOverflow(OverflowHidden)

	// 初始 scrollRect
	rect := comp.display.ScrollRect()
	if rect == nil {
		t.Fatal("scrollRect 应该已设置")
	}
	if rect.W != 190.0 || rect.H != 190.0 {
		t.Fatalf("初始 scrollRect 尺寸错误：(%v,%v)", rect.W, rect.H)
	}

	// 改变尺寸
	comp.SetSize(300, 250)

	// scrollRect 应该更新
	rect = comp.display.ScrollRect()
	if rect == nil {
		t.Fatal("scrollRect 应该保持设置")
	}
	expectedW := 300.0 - 10.0
	expectedH := 250.0 - 10.0
	if rect.W != expectedW || rect.H != expectedH {
		t.Fatalf("scrollRect 应该更新为 (%v,%v)，实际 (%v,%v)",
			expectedW, expectedH, rect.W, rect.H)
	}
}

func TestMarginIsZero(t *testing.T) {
	tests := []struct {
		name     string
		margin   Margin
		expected bool
	}{
		{"全零", Margin{}, true},
		{"有 Top", Margin{Top: 1}, false},
		{"有 Bottom", Margin{Bottom: 1}, false},
		{"有 Left", Margin{Left: 1}, false},
		{"有 Right", Margin{Right: 1}, false},
		{"全不为零", Margin{Top: 1, Bottom: 2, Left: 3, Right: 4}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.margin.IsZero(); got != tt.expected {
				t.Errorf("IsZero() = %v, 期望 %v", got, tt.expected)
			}
		})
	}
}

func TestOverflowAccessors(t *testing.T) {
	comp := NewGComponent()

	// 默认值
	if comp.Overflow() != OverflowVisible {
		t.Errorf("默认 overflow 应该是 Visible，实际 %v", comp.Overflow())
	}

	// 设置 overflow
	comp.SetupOverflow(OverflowHidden)
	if comp.Overflow() != OverflowHidden {
		t.Errorf("overflow 应该是 Hidden，实际 %v", comp.Overflow())
	}

	// Margin 访问器
	margin := Margin{Top: 5, Bottom: 10, Left: 15, Right: 20}
	comp.SetMargin(margin)
	got := comp.Margin()
	if got != margin {
		t.Errorf("Margin() = %+v, 期望 %+v", got, margin)
	}
}
