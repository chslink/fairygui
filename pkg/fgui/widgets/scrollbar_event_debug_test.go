package widgets

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// TestScrollBarEventDebug 调试滚动条事件
func TestScrollBarEventDebug(t *testing.T) {
	env := testutil.NewStageEnv(t, 200, 200)
	_ = env
	
	// 创建List（这是用户最常用的场景）
	owner := core.NewGComponent()
	owner.SetSize(200, 400)
	pane := owner.EnsureScrollPane(core.ScrollTypeVertical)
	pane.SetContentSize(200, 1000)
	
	sb := NewScrollBar()
	sb.SetSize(20, 400)
	sb.SetTemplateComponent(makeScrollBarTemplate(true))
	sb.SetScrollPane(pane, true)
	
	t.Logf("=== 事件绑定检查 ===")
	
	// 检查grip是否绑定了事件
	if sb.grip == nil || sb.grip.DisplayObject() == nil {
		t.Fatal("grip not initialized")
	}
	
	t.Logf("✅ Grip已初始化")
	
	// 手动调用onGripMouseDown来测试
	t.Logf("\n--- 测试onGripMouseDown调用 ---")
	
	// 创建一个测试事件
	pe := laya.PointerEvent{
		Position: laya.Point{X: 10, Y: 15},
	}
	
	// 模拟onGripMouseDown
	t.Logf("调用onGripMouseDown前:")
	t.Logf("  sb.dragging = %v", sb.dragging)
	t.Logf("  sb.dragOffset = (%.2f, %.2f)", sb.dragOffset.X, sb.dragOffset.Y)
	
	// 手动执行onGripMouseDown的逻辑
	display := sb.GComponent.GObject.DisplayObject()
	if display != nil {
		local := display.GlobalToLocal(pe.Position)
		sb.dragging = true
		sb.dragOffset = laya.Point{X: local.X - sb.grip.X(), Y: local.Y - sb.grip.Y()}
		
		t.Logf("调用onGripMouseDown后:")
		t.Logf("  sb.dragging = %v", sb.dragging)
		t.Logf("  sb.dragOffset = (%.2f, %.2f)", sb.dragOffset.X, sb.dragOffset.Y)
	}
	
	// 模拟鼠标移动到滑块上
	t.Logf("\n--- 模拟拖动 ---")
	newPos := laya.Point{X: 10, Y: 45}
	localNew := display.GlobalToLocal(newPos)
	curY := localNew.Y - sb.dragOffset.Y
	track := sb.bar.Height() - sb.grip.Height()
	perc := (curY - sb.bar.Y()) / track
	
	t.Logf("鼠标新位置: (%.2f, %.2f)", newPos.X, newPos.Y)
	t.Logf("计算得perc: %.4f", perc)
	
	// 检查事件是否会冒泡到owner
	ownerDO := owner.GObject.DisplayObject()
	if ownerDO != nil {
		t.Logf("\nOwner DisplayObject存在: %T", ownerDO)
	}
}
