package debug

import (
	"fmt"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// EventSimulator 事件模拟器，用于调试和测试
type EventSimulator struct {
	stage *laya.Stage
}

// NewEventSimulator 创建事件模拟器
func NewEventSimulator(stage *laya.Stage) *EventSimulator {
	return &EventSimulator{stage: stage}
}

// ClickObject 模拟点击对象
func (s *EventSimulator) ClickObject(obj *core.GObject) error {
	if obj == nil {
		return fmt.Errorf("对象为空")
	}

	if !obj.Visible() {
		return fmt.Errorf("对象不可见: %s", obj.Name())
	}

	// 计算对象中心点
	x := obj.X() + obj.Width()/2
	y := obj.Y() + obj.Height()/2

	// 模拟点击
	return s.Click(x, y)
}

// Click 模拟在指定坐标点击
func (s *EventSimulator) Click(x, y float64) error {
	if s.stage == nil {
		return fmt.Errorf("stage未设置")
	}

	// 查找目标对象
	pt := laya.Point{X: x, Y: y}
	target := s.stage.HitTest(pt)
	if target == nil {
		return fmt.Errorf("坐标 (%.0f, %.0f) 处无可点击对象", x, y)
	}

	// 创建并触发事件序列
	downEvt := laya.Event{Type: laya.EventMouseDown, Data: pt}
	upEvt := laya.Event{Type: laya.EventMouseUp, Data: pt}
	clickEvt := laya.Event{Type: laya.EventClick, Data: pt}

	target.Emit(laya.EventMouseDown, downEvt)
	target.Emit(laya.EventMouseUp, upEvt)
	target.Emit(laya.EventClick, clickEvt)

	return nil
}

// ClickByPath 通过对象路径模拟点击
func (s *EventSimulator) ClickByPath(inspector *Inspector, path string) error {
	obj := inspector.FindByPath(path)
	if obj == nil {
		return fmt.Errorf("未找到对象: %s", path)
	}
	return s.ClickObject(obj)
}

// ClickByName 通过对象名称模拟点击（点击第一个匹配的对象）
func (s *EventSimulator) ClickByName(inspector *Inspector, name string) error {
	objs := inspector.FindByName(name)
	if len(objs) == 0 {
		return fmt.Errorf("未找到对象: %s", name)
	}
	return s.ClickObject(objs[0])
}

// TouchObject 模拟触摸对象（支持多点触控）
func (s *EventSimulator) TouchObject(obj *core.GObject, touchID int) error {
	if obj == nil {
		return fmt.Errorf("对象为空")
	}

	x := obj.X() + obj.Width()/2
	y := obj.Y() + obj.Height()/2

	return s.Touch(x, y, touchID)
}

// Touch 模拟触摸指定坐标
func (s *EventSimulator) Touch(x, y float64, touchID int) error {
	if s.stage == nil {
		return fmt.Errorf("stage未设置")
	}

	// 查找目标对象
	pt := laya.Point{X: x, Y: y}
	target := s.stage.HitTest(pt)
	if target == nil {
		return fmt.Errorf("坐标 (%.0f, %.0f) 处无可触摸对象", x, y)
	}

	// 创建触摸事件（使用与TouchInput相关的数据结构）
	touchData := map[string]interface{}{
		"x":       x,
		"y":       y,
		"touchID": touchID,
	}

	touchBeginEvt := laya.Event{Type: laya.EventTouchBegin, Data: touchData}
	touchEndEvt := laya.Event{Type: laya.EventTouchEnd, Data: touchData}

	target.Emit(laya.EventTouchBegin, touchBeginEvt)
	target.Emit(laya.EventTouchEnd, touchEndEvt)

	return nil
}

// DragObject 模拟拖拽对象
func (s *EventSimulator) DragObject(obj *core.GObject, fromX, fromY, toX, toY float64) error {
	if obj == nil {
		return fmt.Errorf("对象为空")
	}

	if s.stage == nil {
		return fmt.Errorf("stage未设置")
	}

	displayObj := obj.DisplayObject()
	if displayObj == nil {
		return fmt.Errorf("对象无显示对象")
	}

	// 创建拖拽事件序列
	fromPt := laya.Point{X: fromX, Y: fromY}
	toPt := laya.Point{X: toX, Y: toY}

	downEvt := laya.Event{Type: laya.EventMouseDown, Data: fromPt}
	moveEvt := laya.Event{Type: laya.EventMouseMove, Data: toPt}
	upEvt := laya.Event{Type: laya.EventMouseUp, Data: toPt}

	displayObj.Emit(laya.EventMouseDown, downEvt)
	displayObj.Emit(laya.EventMouseMove, moveEvt)
	displayObj.Emit(laya.EventMouseUp, upEvt)

	return nil
}

// SendCustomEvent 发送自定义事件
func (s *EventSimulator) SendCustomEvent(obj *core.GObject, eventType string, data interface{}) error {
	if obj == nil {
		return fmt.Errorf("对象为空")
	}

	displayObj := obj.DisplayObject()
	if displayObj == nil {
		return fmt.Errorf("对象无显示对象")
	}

	evt := laya.Event{Type: laya.EventType(eventType), Data: data}
	displayObj.Emit(laya.EventType(eventType), evt)

	return nil
}

// GetObjectAt 获取指定坐标处的对象
func (s *EventSimulator) GetObjectAt(x, y float64) *laya.Sprite {
	if s.stage == nil {
		return nil
	}
	pt := laya.Point{X: x, Y: y}
	return s.stage.HitTest(pt)
}

// IsPointInObject 检查坐标是否在对象范围内
func (s *EventSimulator) IsPointInObject(x, y float64, obj *core.GObject) bool {
	if obj == nil {
		return false
	}

	objX := obj.X()
	objY := obj.Y()
	objW := obj.Width()
	objH := obj.Height()

	return x >= objX && x <= objX+objW && y >= objY && y <= objY+objH
}

// SimulateClickResult 点击模拟结果
type SimulateClickResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Target    string `json:"target,omitempty"`
	TargetID  string `json:"target_id,omitempty"`
	Timestamp string `json:"timestamp"`
}
