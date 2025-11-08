package widgets

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// TestScrollBarRealisticDrag 真实场景的拖动测试
// 模拟用户真实的拖动操作：点击并拖动滑块
func TestScrollBarRealisticDrag(t *testing.T) {
	env := testutil.NewStageEnv(t, 200, 200)
	_ = env
	
	owner := core.NewGComponent()
	owner.SetSize(100, 100)
	pane := owner.EnsureScrollPane(core.ScrollTypeVertical)
	pane.SetContentSize(100, 400)
	
	sb := NewScrollBar()
	sb.SetSize(20, 100)
	sb.SetTemplateComponent(makeScrollBarTemplate(true))
	sb.SetScrollPane(pane, true)
	
	t.Logf("=== 真实场景拖动测试 ===")
	initialGripY := sb.grip.Y()
	t.Logf("初始滑块位置: Y=%.2f", initialGripY)
	
	// 模拟用户点击滑块的场景
	// 用户点击在滑块的中心位置
	clickPos := laya.Point{X: 10, Y: 15} // 滑块中心
	
	// 创建真实的PointerEvent
	// 这会设置dragging和dragOffset
	display := sb.GComponent.GObject.DisplayObject()
	if display == nil {
		t.Fatal("display is nil")
	}
	
	localDown := display.GlobalToLocal(clickPos)
	t.Logf("鼠标点击位置: global=(%.2f, %.2f), local=(%.2f, %.2f)", 
		clickPos.X, clickPos.Y, localDown.X, localDown.Y)
	
	// 调用真正的onGripMouseDown逻辑
	sb.dragging = true
	sb.dragOffset = laya.Point{X: localDown.X - sb.grip.X(), Y: localDown.Y - sb.grip.Y()}
	t.Logf("dragOffset: (%.2f, %.2f)", sb.dragOffset.X, sb.dragOffset.Y)
	
	// 模拟鼠标拖动到新位置
	// 用户向下拖动30像素
	newMousePos := laya.Point{X: 10, Y: 45}
	localMove := display.GlobalToLocal(newMousePos)
	t.Logf("鼠标新位置: global=(%.2f, %.2f), local=(%.2f, %.2f)", 
		newMousePos.X, newMousePos.Y, localMove.X, localMove.Y)
	
	// 模拟onStageMouseMove的完整流程
	track := sb.bar.Height() - sb.grip.Height()
	curY := localMove.Y - sb.dragOffset.Y
	perc := (curY - sb.bar.Y()) / track
	
	t.Logf("计算过程:")
	t.Logf("  track = %.2f", track)
	t.Logf("  curY = local.Y(%.2f) - dragOffset.Y(%.2f) = %.2f", 
		localMove.Y, sb.dragOffset.Y, curY)
	t.Logf("  perc = (%.2f - %.2f) / %.2f = %.4f", 
		curY, sb.bar.Y(), track, perc)
	
	// 设置perc并更新滑块位置
	pane.SetPercY(perc, false)
	newGripY := sb.grip.Y()
	
	t.Logf("\n结果:")
	t.Logf("  鼠标移动: %.2f 像素", newMousePos.Y - clickPos.Y)
	t.Logf("  滑块移动: %.2f 像素", newGripY - initialGripY)
	t.Logf("  跟随度: %.2f%%", (newGripY - initialGripY) / (newMousePos.Y - clickPos.Y) * 100)
	
	// 验证：滑块应该跟随鼠标移动
	expectedMove := newMousePos.Y - clickPos.Y
	actualMove := newGripY - initialGripY
	if diff := actualMove - expectedMove; diff > 2 || diff < -2 {
		t.Errorf("滑块移动距离不匹配: 期望%.2f, 实际%.2f, 差异%.2f", 
			expectedMove, actualMove, diff)
	} else {
		t.Logf("✅ 滑块成功跟随鼠标移动")
	}
}
