package widgets

import (
	"testing"
	"time"
)

func TestTextInput_CursorPosition(t *testing.T) {
	input := NewTextInput()
	input.SetText("Hello World")

	input.SetCursorPosition(5)
	if got := input.CursorPosition(); got != 5 {
		t.Errorf("CursorPosition() = %d, want 5", got)
	}
}

func TestTextInput_Selection(t *testing.T) {
	input := NewTextInput()
	input.SetText("Hello World")

	input.SetSelection(0, 5)
	start, end := input.GetSelection()
	if start != 0 || end != 5 {
		t.Errorf("GetSelection() = (%d, %d), want (0, 5)", start, end)
	}
	if !input.HasSelection() {
		t.Error("HasSelection() = false, want true")
	}
	if got := input.GetSelectedText(); got != "Hello" {
		t.Errorf("GetSelectedText() = %q, want 'Hello'", got)
	}
	input.ClearSelection()
	if input.HasSelection() {
		t.Error("HasSelection() after ClearSelection() = true")
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
		t.Errorf("GetSelectedText() = %q", got)
	}
}

func TestTextInput_Focus(t *testing.T) {
	input := NewTextInput()

	if input.IsFocused() {
		t.Error("IsFocused() initially = true")
	}
	input.RequestFocus()
	if !input.IsFocused() {
		t.Error("IsFocused() after RequestFocus() = false")
	}
	if !input.IsCursorVisible() {
		t.Error("IsCursorVisible() after RequestFocus() = false")
	}
	input.LoseFocus()
	if input.IsFocused() {
		t.Error("IsFocused() after LoseFocus() = true")
	}
}

func TestTextInput_CursorBlink(t *testing.T) {
	input := NewTextInput()
	input.RequestFocus()

	if !input.IsCursorVisible() {
		t.Error("cursor initially not visible")
	}
	time.Sleep(600 * time.Millisecond)
	input.UpdateCursor(0.6)

	v1 := input.IsCursorVisible()
	time.Sleep(600 * time.Millisecond)
	input.UpdateCursor(0.6)

	v2 := input.IsCursorVisible()
	if v1 == v2 {
		t.Errorf("blink not working: v1=%v v2=%v", v1, v2)
	}
}

func TestTextInput_SelectionNormalization(t *testing.T) {
	input := NewTextInput()
	input.SetText("Hello World")

	input.SetSelection(5, 0)
	start, end := input.GetSelection()
	if start != 0 || end != 5 {
		t.Errorf("reverse selection = (%d, %d), want (0, 5)", start, end)
	}
}

func TestTextInput_EmptyText(t *testing.T) {
	input := NewTextInput()
	input.SetText("")

	input.SetCursorPosition(5)
	if got := input.CursorPosition(); got != 0 {
		t.Errorf("CursorPosition() on empty text after clamp = %d, want 0", got)
	}
	input.SelectAll()
}

func TestTextInput_MouseDown(t *testing.T) {
	input := NewTextInput()
	input.SetText("Hello World")
	input.SetEditable(true)

	handled := input.HandleMouseDown(0, 0)
	if !handled {
		t.Error("HandleMouseDown should return true for editable input")
	}
	if !input.IsFocused() {
		t.Error("HandleMouseDown should set focus")
	}
}
