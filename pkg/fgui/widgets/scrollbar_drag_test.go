package widgets

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// TestScrollBarDragFollowMouse 验证修复后滑块跟随鼠标
func TestScrollBarDragFollowMouse(t *testing.T) {
	_ = testutil.NewStageEnv(t, 100, 100)

	owner := core.NewGComponent()
	owner.SetSize(100, 100)
	pane := owner.EnsureScrollPane(core.ScrollTypeVertical)
	pane.SetContentSize(100, 400)

	sb := NewScrollBar()
	sb.SetSize(20, 100)
	sb.SetTemplateComponent(makeScrollBarTemplate(true))
	sb.SetScrollPane(pane, true)

	// 初始化
	track := sb.bar.Height() - sb.grip.Height()
	t.Logf("=== 修复后测试 ===")
	t.Logf("可移动范围: %.2f", track)

	// 场景1：点击滑块顶部并向下拖动
	t.Logf("\n--- 场景1：点击滑块顶部(Y=5)，向下拖动20像素 ---")
	display := sb.getContainerDisplayObject()
	clickPos1 := laya.Point{X: 10, Y: 5} // 滑块顶部
	localDown1 := display.GlobalToLocal(clickPos1)
	sb.dragOffset = laya.Point{X: localDown1.X - sb.grip.X(), Y: localDown1.Y - sb.grip.Y()}
	t.Logf("点击位置: (%.2f, %.2f)", localDown1.X, localDown1.Y)
	t.Logf("dragOffset: (%.2f, %.2f)", sb.dragOffset.X, sb.dragOffset.Y)

	// 向下移动20像素
	movePos1 := laya.Point{X: 10, Y: 25} // 移动到Y=25
	localMove1 := display.GlobalToLocal(movePos1)
	curY1 := localMove1.Y - sb.dragOffset.Y
	perc1 := (curY1 - sb.bar.Y()) / track
	t.Logf("移动后鼠标位置: (%.2f, %.2f)", localMove1.X, localMove1.Y)
	t.Logf("curY = %.2f - %.2f = %.2f", localMove1.Y, sb.dragOffset.Y, curY1)
	t.Logf("perc = %.4f", perc1)
	pane.SetPercY(perc1, false)
	t.Logf("滑块新位置: Y=%.2f", sb.grip.Y())
	t.Logf("期望: 滑块应该跟随鼠标向下移动约20像素")

	// 场景2：点击滑块中间并向下拖动
	t.Logf("\n--- 场景2：点击滑块中间(Y=15)，向下拖动20像素 ---")
	pane.SetPos(0, 0, false) // 重置
	clickPos2 := laya.Point{X: 10, Y: 15} // 滑块中间
	localDown2 := display.GlobalToLocal(clickPos2)
	sb.dragOffset = laya.Point{X: localDown2.X - sb.grip.X(), Y: localDown2.Y - sb.grip.Y()}
	t.Logf("点击位置: (%.2f, %.2f)", localDown2.X, localDown2.Y)
	t.Logf("dragOffset: (%.2f, %.2f)", sb.dragOffset.X, sb.dragOffset.Y)

	// 向下移动20像素
	movePos2 := laya.Point{X: 10, Y: 35}
	localMove2 := display.GlobalToLocal(movePos2)
	curY2 := localMove2.Y - sb.dragOffset.Y
	perc2 := (curY2 - sb.bar.Y()) / track
	t.Logf("移动后鼠标位置: (%.2f, %.2f)", localMove2.X, localMove2.Y)
	t.Logf("curY = %.2f - %.2f = %.2f", localMove2.Y, sb.dragOffset.Y, curY2)
	t.Logf("perc = %.4f", perc2)
	pane.SetPercY(perc2, false)
	t.Logf("滑块新位置: Y=%.2f", sb.grip.Y())
	t.Logf("期望: 滑块应该跟随鼠标向下移动约20像素")

	// 验证：两个场景中，相同距离的拖动应该产生相似的结果
	if diff := perc2 - perc1; diff > 0.01 || diff < -0.01 {
		t.Errorf("两个场景的perc差异过大: %.4f", diff)
	}
}
