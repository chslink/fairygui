package widgets

import (
	"testing"
	"time"
)

func TestTextInput_CursorPosition(t *testing.T) {
	input := NewTextInput()
	input.SetText("Hello World")

	// 测试设置光标位置
	input.SetCursorPosition(5)
	if got := input.CursorPosition(); got != 5 {
		t.Errorf("CursorPosition() = %d, want 5", got)
	}

	// 测试边界条件
	input.SetCursorPosition(-1)
	if got := input.CursorPosition(); got != 0 {
		t.Errorf("CursorPosition() with negative = %d, want 0", got)
	}

	input.SetCursorPosition(100)
	textLen := len([]rune("Hello World"))
	if got := input.CursorPosition(); got != textLen {
		t.Errorf("CursorPosition() with overflow = %d, want %d", got, textLen)
	}
}

func TestTextInput_Selection(t *testing.T) {
	input := NewTextInput()
	input.SetText("Hello World")

	// 测试设置选择区域
	input.SetSelection(0, 5)
	start, end := input.GetSelection()
	if start != 0 || end != 5 {
		t.Errorf("GetSelection() = (%d, %d), want (0, 5)", start, end)
	}

	// 测试是否有选中
	if !input.HasSelection() {
		t.Error("HasSelection() = false, want true")
	}

	// 测试获取选中文本
	if got := input.GetSelectedText(); got != "Hello" {
		t.Errorf("GetSelectedText() = %q, want %q", got, "Hello")
	}

	// 测试清除选择
	input.ClearSelection()
	if input.HasSelection() {
		t.Error("HasSelection() after ClearSelection() = true, want false")
	}
}

func TestTextInput_SelectAll(t *testing.T) {
	input := NewTextInput()
	input.SetText("Hello World")

	input.SelectAll()
	start, end := input.GetSelection()
	textLen := len([]rune("Hello World"))
	if start != 0 || end != textLen {
		t.Errorf("SelectAll() = (%d, %d), want (0, %d)", start, end, textLen)
	}

	if got := input.GetSelectedText(); got != "Hello World" {
		t.Errorf("GetSelectedText() after SelectAll() = %q, want %q", got, "Hello World")
	}
}

func TestTextInput_Focus(t *testing.T) {
	input := NewTextInput()

	// 初始状态未获得焦点
	if input.IsFocused() {
		t.Error("IsFocused() initially = true, want false")
	}

	// 请求焦点
	input.RequestFocus()
	if !input.IsFocused() {
		t.Error("IsFocused() after RequestFocus() = false, want true")
	}

	if !input.IsCursorVisible() {
		t.Error("IsCursorVisible() after RequestFocus() = false, want true")
	}

	// 失去焦点
	input.LoseFocus()
	if input.IsFocused() {
		t.Error("IsFocused() after LoseFocus() = true, want false")
	}

	if input.IsCursorVisible() {
		t.Error("IsCursorVisible() after LoseFocus() = true, want false")
	}
}

func TestTextInput_CursorBlink(t *testing.T) {
	input := NewTextInput()
	input.RequestFocus()

	// 初始状态光标可见
	if !input.IsCursorVisible() {
		t.Error("IsCursorVisible() initially = false, want true")
	}

	// 模拟时间流逝(超过闪烁间隔)
	time.Sleep(600 * time.Millisecond)
	input.UpdateCursor(0.6)

	// 光标应该切换状态
	visible1 := input.IsCursorVisible()

	// 再次更新
	time.Sleep(600 * time.Millisecond)
	input.UpdateCursor(0.6)

	visible2 := input.IsCursorVisible()

	// 两次状态应该不同
	if visible1 == visible2 {
		t.Errorf("Cursor blink not working: visible1=%v, visible2=%v", visible1, visible2)
	}
}

func TestTextInput_SelectionNormalization(t *testing.T) {
	input := NewTextInput()
	input.SetText("Hello World")

	// 测试反向选择(自动规范化)
	input.SetSelection(5, 0)
	start, end := input.GetSelection()
	if start != 0 || end != 5 {
		t.Errorf("SetSelection(5, 0) = (%d, %d), want (0, 5)", start, end)
	}
}

func TestTextInput_EmptyText(t *testing.T) {
	input := NewTextInput()
	input.SetText("")

	input.SetCursorPosition(5)
	if got := input.CursorPosition(); got != 0 {
		t.Errorf("CursorPosition() on empty text = %d, want 0", got)
	}

	input.SelectAll()
	if input.HasSelection() {
		t.Error("HasSelection() on empty text = true, want false")
	}
}

func TestTextInput_MouseInteraction(t *testing.T) {
	input := NewTextInput()
	input.SetText("Hello World")
	input.SetEditable(true)

	// 测试鼠标点击定位光标
	t.Run("HandleMouseDown", func(t *testing.T) {
		// 点击文本开头
		handled := input.HandleMouseDown(0, 0)
		if !handled {
			t.Error("HandleMouseDown should return true for valid click")
		}
		if !input.IsFocused() {
			t.Error("HandleMouseDown should request focus")
		}
		pos := input.CursorPosition()
		if pos != 0 {
			t.Errorf("Click at start: cursor position = %d, want 0", pos)
		}

		// 点击文本中间(近似位置)
		fontSize := float64(12)
		approxMidX := fontSize * 0.6 * 5 // 近似第5个字符位置
		input.HandleMouseDown(approxMidX, 0)
		pos = input.CursorPosition()
		if pos < 4 || pos > 6 {
			t.Errorf("Click at middle: cursor position = %d, want around 5", pos)
		}

		// 点击文本结尾后
		input.HandleMouseDown(1000, 0)
		pos = input.CursorPosition()
		textLen := len([]rune("Hello World"))
		if pos != textLen {
			t.Errorf("Click at end: cursor position = %d, want %d", pos, textLen)
		}
	})

	// 测试鼠标拖动选择
	t.Run("HandleMouseDrag", func(t *testing.T) {
		// 先点击设置起始位置
		input.HandleMouseDown(0, 0)

		// 拖动到中间
		fontSize := float64(12)
		approxMidX := fontSize * 0.6 * 5
		handled := input.HandleMouseDrag(approxMidX, 0)
		if !handled {
			t.Error("HandleMouseDrag should return true for valid drag")
		}

		if !input.HasSelection() {
			t.Error("HandleMouseDrag should create selection")
		}

		start, end := input.GetSelection()
		if start == end {
			t.Errorf("Drag selection: start=%d, end=%d should be different", start, end)
		}
	})

	// 测试双击选择单词
	t.Run("HandleMouseDoubleClick", func(t *testing.T) {
		input.SetText("Hello World Test")

		// 双击 "World" 中间
		fontSize := float64(12)
		approxWorldX := fontSize * 0.6 * 7 // "Hello W" 的近似位置
		handled := input.HandleMouseDoubleClick(approxWorldX, 0)
		if !handled {
			t.Error("HandleMouseDoubleClick should return true")
		}

		if !input.HasSelection() {
			t.Error("Double click should create selection")
		}

		selected := input.GetSelectedText()
		// 由于是近似计算，选中的可能是 "Hello" 或 "World"
		if selected != "Hello" && selected != "World" {
			t.Logf("Double click selected: %q (approximate position)", selected)
		}
	})
}

func TestTextInput_GetCharPositionAtPoint(t *testing.T) {
	input := NewTextInput()
	input.SetText("Hello")

	// 测试空文本
	t.Run("EmptyText", func(t *testing.T) {
		emptyInput := NewTextInput()
		emptyInput.SetText("")
		pos, inBounds := emptyInput.GetCharPositionAtPoint(0, 0)
		if pos != 0 || !inBounds {
			t.Errorf("Empty text: pos=%d, inBounds=%v, want 0, true", pos, inBounds)
		}
	})

	// 测试单行文本
	t.Run("SingleLine", func(t *testing.T) {
		// 点击开头
		pos, inBounds := input.GetCharPositionAtPoint(0, 0)
		if !inBounds {
			t.Error("Point at start should be in bounds")
		}
		if pos != 0 {
			t.Errorf("Position at start = %d, want 0", pos)
		}

		// 点击末尾后
		pos, inBounds = input.GetCharPositionAtPoint(1000, 0)
		if !inBounds {
			t.Error("Point far right should be in bounds")
		}
		textLen := len([]rune("Hello"))
		if pos != textLen {
			t.Errorf("Position at far right = %d, want %d", pos, textLen)
		}
	})

	// 测试不可编辑状态
	t.Run("NotEditable", func(t *testing.T) {
		input.SetEditable(false)
		handled := input.HandleMouseDown(0, 0)
		if handled {
			t.Error("Non-editable input should not handle mouse down")
		}
		input.SetEditable(true) // 恢复状态
	})
}

func TestTextInput_SelectWordAt(t *testing.T) {
	input := NewTextInput()
	input.SetText("Hello World Test")

	// 测试选择第一个单词
	input.selectWordAt(2) // "Hello" 中间
	selected := input.GetSelectedText()
	if selected != "Hello" {
		t.Errorf("Select word at 2: got %q, want %q", selected, "Hello")
	}

	// 测试选择第二个单词
	input.selectWordAt(7) // "World" 中间
	selected = input.GetSelectedText()
	if selected != "World" {
		t.Errorf("Select word at 7: got %q, want %q", selected, "World")
	}

	// 测试选择最后一个单词
	input.selectWordAt(13) // "Test" 中间
	selected = input.GetSelectedText()
	if selected != "Test" {
		t.Errorf("Select word at 13: got %q, want %q", selected, "Test")
	}

	// 测试边界条件
	input.selectWordAt(5) // 空格位置
	selected = input.GetSelectedText()
	if selected != "" {
		t.Logf("Select at space position selected: %q", selected)
	}
}
