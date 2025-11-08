package widgets

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// TestScrollBarVisualDragBehavior 视觉化测试滚动条拖动行为
func TestScrollBarVisualDragBehavior(t *testing.T) {
	_ = testutil.NewStageEnv(t, 100, 100)

	owner := core.NewGComponent()
	owner.SetSize(100, 100)
	pane := owner.EnsureScrollPane(core.ScrollTypeVertical)
	pane.SetContentSize(100, 400)

	sb := NewScrollBar()
	sb.SetSize(20, 100)
	sb.SetTemplateComponent(makeScrollBarTemplate(true))
	sb.SetScrollPane(pane, true)

	display := sb.getContainerDisplayObject()
	track := sb.bar.Height() - sb.grip.Height()

	t.Logf("=== 视觉化拖动测试 ===")
	t.Logf("Bar范围: [%.2f, %.2f], 可移动: %.2f", sb.bar.Y(), sb.bar.Y()+track, track)

	// 测试1：从顶部拖动到中间
	t.Logf("\n--- 测试1：从顶部拖动到bar中间 ---")
	initialGripY1 := sb.grip.Y()
	clickY1 := sb.bar.Y() + 5 // bar顶部上方5像素
	moveY1 := sb.bar.Y() + track/2 // bar中间位置

	localDown1 := display.GlobalToLocal(laya.Point{X: 10, Y: clickY1})
	sb.dragOffset = laya.Point{X: localDown1.X - sb.grip.X(), Y: localDown1.Y - sb.grip.Y()}
	localMove1 := display.GlobalToLocal(laya.Point{X: 10, Y: moveY1})
	curY1 := localMove1.Y - sb.dragOffset.Y
	perc1 := (curY1 - sb.bar.Y()) / track
	pane.SetPercY(perc1, false)
	finalGripY1 := sb.grip.Y()

	t.Logf("鼠标: %.2f -> %.2f (移动%.2f)", clickY1, moveY1, moveY1-clickY1)
	t.Logf("滑块: %.2f -> %.2f (移动%.2f)", initialGripY1, finalGripY1, finalGripY1-initialGripY1)
	t.Logf("跟随度: %.2f%%", (finalGripY1-initialGripY1)/(moveY1-clickY1)*100)

	// 测试2：从中间拖动到底部
	t.Logf("\n--- 测试2：从中间拖动到底部 ---")
	pane.SetPos(0, 0, false) // 重置
	initialGripY2 := sb.grip.Y()
	clickY2 := sb.bar.Y() + track/2 // bar中间位置
	moveY2 := sb.bar.Y() + track - 5 // bar底部上方5像素

	localDown2 := display.GlobalToLocal(laya.Point{X: 10, Y: clickY2})
	sb.dragOffset = laya.Point{X: localDown2.X - sb.grip.X(), Y: localDown2.Y - sb.grip.Y()}
	localMove2 := display.GlobalToLocal(laya.Point{X: 10, Y: moveY2})
	curY2 := localMove2.Y - sb.dragOffset.Y
	perc2 := (curY2 - sb.bar.Y()) / track
	pane.SetPercY(perc2, false)
	finalGripY2 := sb.grip.Y()

	t.Logf("鼠标: %.2f -> %.2f (移动%.2f)", clickY2, moveY2, moveY2-clickY2)
	t.Logf("滑块: %.2f -> %.2f (移动%.2f)", initialGripY2, finalGripY2, finalGripY2-initialGripY2)
	t.Logf("跟随度: %.2f%%", (finalGripY2-initialGripY2)/(moveY2-clickY2)*100)

	// 验证滑块跟随鼠标
	if diff1 := (finalGripY1 - initialGripY1) - (moveY1 - clickY1); diff1 > 1 || diff1 < -1 {
		t.Errorf("测试1滑块移动距离与鼠标不匹配: 期望~%.2f, 实际%.2f", moveY1-clickY1, finalGripY1-initialGripY1)
	}
	if diff2 := (finalGripY2 - initialGripY2) - (moveY2 - clickY2); diff2 > 1 || diff2 < -1 {
		t.Errorf("测试2滑块移动距离与鼠标不匹配: 期望~%.2f, 实际%.2f", moveY2-clickY2, finalGripY2-initialGripY2)
	}
}
