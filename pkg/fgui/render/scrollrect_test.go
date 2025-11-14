// +build ebiten

package render

import (
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/core"
)

func TestScrollRectCoordinates(t *testing.T) {
	// 测试 scrollRect 坐标系统
	comp := core.NewGComponent()
	comp.SetSize(200, 150)
	comp.SetMargin(core.Margin{Top: 10, Bottom: 10, Left: 10, Right: 10})

	comp.SetupOverflow(core.OverflowHidden)

	// 验证 scrollRect
	rect := comp.DisplayObject().ScrollRect()
	if rect == nil {
		t.Fatal("scrollRect 应该被设置")
	}

	// scrollRect 应该从 (margin.left, margin.top) 开始
	if rect.X != 10.0 || rect.Y != 10.0 {
		t.Errorf("scrollRect 起点错误: (%.1f, %.1f)，期望 (10, 10)", rect.X, rect.Y)
	}

	// scrollRect 尺寸应该是 (width - margin.right, height - margin.bottom)
	expectedW := 200.0 - 10.0 // width - right
	expectedH := 150.0 - 10.0 // height - bottom
	if rect.W != expectedW || rect.H != expectedH {
		t.Errorf("scrollRect 尺寸错误: (%.1f, %.1f)，期望 (%.1f, %.1f)",
			rect.W, rect.H, expectedW, expectedH)
	}

	// 验证 container 偏移
	containerPos := comp.Container().Position()
	if containerPos.X != 10.0 || containerPos.Y != 10.0 {
		t.Errorf("container 偏移错误: (%.1f, %.1f)，期望 (10, 10)",
			containerPos.X, containerPos.Y)
	}

	t.Log("✓ scrollRect 坐标系统测试通过")
}
