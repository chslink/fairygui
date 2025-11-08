package widgets

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// TestScrollBarDebugDrag 调试拖动问题
func TestScrollBarDebugDrag(t *testing.T) {
	_ = testutil.NewStageEnv(t, 100, 100)
	
	owner := core.NewGComponent()
	owner.SetSize(100, 100)
	pane := owner.EnsureScrollPane(core.ScrollTypeVertical)
	pane.SetContentSize(100, 400)
	
	sb := NewScrollBar()
	sb.SetSize(20, 100)
	sb.SetTemplateComponent(makeScrollBarTemplate(true))
	sb.SetScrollPane(pane, true)
	
	t.Logf("=== 调试拖动问题 ===")
	t.Logf("ScrollBar GObject: %p", sb.GComponent.GObject)
	t.Logf("Template GObject: %p", sb.template.GObject)
	t.Logf("Bar GObject: %p", sb.bar)
	t.Logf("Grip GObject: %p", sb.grip)
	
	display := sb.getContainerDisplayObject()
	if display == nil {
		t.Fatal("display is nil")
	}
	
	t.Logf("Container DisplayObject: %p", display)
	t.Logf("Grip在template中的位置: (%.2f, %.2f)", sb.grip.X(), sb.grip.Y())
	
	// 场景1：鼠标点击滑块顶部
	t.Logf("\n--- 场景1：点击滑块顶部 ---")
	clickY1 := float64(5) // 滑块顶部
	downPos1 := laya.Point{X: 10, Y: clickY1}
	localDown1 := display.GlobalToLocal(downPos1)
	
	t.Logf("鼠标全局位置: (%.2f, %.2f)", downPos1.X, downPos1.Y)
	t.Logf("鼠标在container中的位置: (%.2f, %.2f)", localDown1.X, localDown1.Y)
	t.Logf("滑块在container中的位置: (%.2f, %.2f)", sb.grip.X(), sb.grip.Y())
	
	// 当前修复的计算
	dragOffset1 := laya.Point{X: localDown1.X - sb.grip.X(), Y: localDown1.Y - sb.grip.Y()}
	t.Logf("dragOffset(当前修复): (%.2f, %.2f)", dragOffset1.X, dragOffset1.Y)
	
	// TypeScript的计算方式
	tsDragOffset := laya.Point{X: downPos1.X - sb.grip.X(), Y: downPos1.Y - sb.grip.Y()}
	t.Logf("dragOffset(TS风格): (%.2f, %.2f)", tsDragOffset.X, tsDragOffset.Y)
	
	// 鼠标向下移动20像素
	t.Logf("\n--- 鼠标向下移动20像素 ---")
	moveY := clickY1 + 20
	movePos := laya.Point{X: 10, Y: moveY}
	localMove := display.GlobalToLocal(movePos)
	
	t.Logf("鼠标全局位置: (%.2f, %.2f)", movePos.X, movePos.Y)
	t.Logf("鼠标在container中的位置: (%.2f, %.2f)", localMove.X, localMove.Y)
	
	// 使用当前修复的计算
	curY1 := localMove.Y - dragOffset1.Y
	track := sb.bar.Height() - sb.grip.Height()
	perc1 := (curY1 - sb.bar.Y()) / track
	t.Logf("当前修复: curY=%.2f, perc=%.4f", curY1, perc1)
	
	// 使用TS风格的计算
	curY2 := movePos.Y - tsDragOffset.Y
	perc2 := (curY2 - sb.bar.Y()) / track
	t.Logf("TS风格: curY=%.2f, perc=%.4f", curY2, perc2)
	
	// 验证哪个是对的
	if dragOffset1.Y != tsDragOffset.Y {
		t.Logf("\n❌ 发现问题：两种dragOffset计算结果不同！")
		t.Logf("差异: (%.2f, %.2f)", dragOffset1.Y - tsDragOffset.Y, 0.0)

		// 问题可能是grip的位置不是全局位置
		t.Logf("\n问题分析：")
		t.Logf("- grip.X()返回的是: %.2f", sb.grip.X())
		t.Logf("- grip.X()是在哪个坐标系？")
	}
}
