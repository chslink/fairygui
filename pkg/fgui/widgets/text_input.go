package widgets

import (
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2/exp/textinput"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

type KeyboardType string

const (
	KeyboardTypeDefault KeyboardType = "text"
	KeyboardTypeNumber  KeyboardType = "number"
	KeyboardTypeURL     KeyboardType = "url"
)

var (
	// activeField is the currently focused textinput.Field (one per application).
	activeField   *textinput.Field
	activeInput   *GTextInput
	fieldLastText string
)

// UpdateInputField must be called each frame from the Ebiten game loop.
// It synchronizes the native IME text state with the active GTextInput.
func UpdateInputField(mx, my int) {
	if activeField == nil {
		return
	}
	func() {
		defer func() { recover() }()
		activeField.HandleInput(mx, my)
	}()
	newText := activeField.Text()
	if newText != fieldLastText {
		fieldLastText = newText
		if activeInput != nil {
			activeInput.GTextField.SetText(newText)
			activeInput.syncStateFromField()
		}
	}
}

// GTextInput is a single-line text input widget with native IME support.
type GTextInput struct {
	*GTextField
	password     bool
	keyboardType KeyboardType
	editable     bool
	maxLength    int
	promptText   string
	restrict     string

	field            *textinput.Field
	focused          bool
	cursorPosition   int
	selectionStart   int
	selectionEnd     int
	cursorVisible    bool
	lastCursorBlink  time.Time
	cursorBlinkDelay float64

	actualText string
}

func NewTextInput() *GTextInput {
	base := NewText()
	input := &GTextInput{
		GTextField:       base,
		editable:         true,
		keyboardType:     KeyboardTypeDefault,
		cursorBlinkDelay: 0.5,
		cursorVisible:    true,
		lastCursorBlink:  time.Now(),
		field:            &textinput.Field{},
	}
	base.SetSingleLine(true)
	base.GObject.SetData(input)
	if sprite := base.GObject.DisplayObject(); sprite != nil {
		sprite.SetMouseEnabled(true)
	}
	return input
}

func (t *GTextInput) SetupBeforeAdd(buf *utils.ByteBuffer, beginPos int) {
	if t == nil {
		return
	}
	if t.GTextField != nil {
		t.GTextField.SetupBeforeAdd(buf, beginPos)
	}
	if buf == nil {
		return
	}
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(beginPos, 4) {
		return
	}
	if prompt := buf.ReadS(); prompt != nil {
		t.SetPromptText(*prompt)
	}
	if restrict := buf.ReadS(); restrict != nil {
		t.SetRestrict(*restrict)
	}
	if buf.Remaining() >= 4 {
		if max := int(buf.ReadInt32()); max > 0 {
			t.SetMaxLength(max)
		}
	}
	if buf.Remaining() >= 4 {
		switch code := int(buf.ReadInt32()); code {
		case 4:
			t.SetKeyboardType(KeyboardTypeNumber)
		case 3:
			t.SetKeyboardType(KeyboardTypeURL)
		default:
			t.SetKeyboardType(KeyboardTypeDefault)
		}
	}
	if buf.Remaining() > 0 && buf.ReadBool() {
		t.SetPassword(true)
	}
}

func (t *GTextInput) SetPassword(enabled bool)  { t.password = enabled }
func (t *GTextInput) Password() bool             { return t.password }
func (t *GTextInput) SetKeyboardType(v KeyboardType) {
	if v == "" { v = KeyboardTypeDefault }
	t.keyboardType = v
}
func (t *GTextInput) KeyboardType() KeyboardType { return t.keyboardType }
func (t *GTextInput) SetEditable(v bool)         { t.editable = v }
func (t *GTextInput) Editable() bool             { return t.editable }
func (t *GTextInput) SetMaxLength(limit int)     { t.maxLength = limit }
func (t *GTextInput) MaxLength() int             { return t.maxLength }
func (t *GTextInput) SetPromptText(text string)  { t.promptText = strings.TrimSpace(text) }
func (t *GTextInput) PromptText() string         { return t.promptText }
func (t *GTextInput) SetRestrict(v string)       { t.restrict = strings.TrimSpace(v) }
func (t *GTextInput) Restrict() string           { return t.restrict }

// --- Cursor & selection (proxy to textinput.Field) ---

func (t *GTextInput) syncStateFromField() {
	if t == nil || t.field == nil {
		return
	}
	// Sync text from Field to GTextField (for rendering)
	text := t.field.Text()
	if t.GTextField.Text() != text {
		t.GTextField.SetText(text)
	}
	// Sync cursor/selection from bytes to runes
	runes := []rune(text)
	selStart, selEnd := t.field.Selection()
	t.selectionStart = byteToRuneIndex(runes, selStart)
	t.selectionEnd = byteToRuneIndex(runes, selEnd)
	t.cursorPosition = t.selectionEnd
}

func byteToRuneIndex(runes []rune, byteIdx int) int {
	count := 0
	for i := range runes {
		if count >= byteIdx {
			return i
		}
		count += len(string(runes[i]))
	}
	return len(runes)
}

func (t *GTextInput) CursorPosition() int {
	if t == nil { return 0 }
	return t.cursorPosition
}

func (t *GTextInput) SetCursorPosition(pos int) {
	if t == nil { return }
	t.cursorPosition = pos
	t.selectionStart = pos
	t.selectionEnd = pos
}

func (t *GTextInput) SelectedText() string {
	if t == nil || !t.HasSelection() { return "" }
	runes := []rune(t.Text())
	s, e := t.selectionStart, t.selectionEnd
	if s < 0 { s = 0 }
	if e > len(runes) { e = len(runes) }
	if s >= e { return "" }
	return string(runes[s:e])
}

// GetSelectedText is an alias for SelectedText (backward compat).
func (t *GTextInput) GetSelectedText() string { return t.SelectedText() }

// HandleMouseDown forwards the click coordinates to the native textinput.Field
// for cursor positioning and selection.
func (t *GTextInput) HandleMouseDown(x, y float64) bool {
	if t == nil || !t.editable { return false }
	t.focused = true
	t.cursorVisible = true
	t.lastCursorBlink = time.Now()

	// Try to use native IME Field; fall back to simple cursor tracking if Ebiten is not running (test env).
	func() {
		defer func() { recover() }()
		if activeField != t.field {
			t.field.Focus()
			t.field.SetTextAndSelection(t.Text(), 0, len(t.Text()))
			activeField = t.field
			activeInput = t
			fieldLastText = t.Text()
			t.registerEventListeners()
		}
		t.field.HandleInput(int(x), int(y))
		t.syncStateFromField()
	}()
	return true
}

func (t *GTextInput) SetSelection(start, end int) {
	if t == nil { return }
	runes := []rune(t.Text())
	max := len(runes)
	if start < 0 { start = 0 }
	if end < 0 { end = 0 }
	if start > max { start = max }
	if end > max { end = max }
	if start > end { start, end = end, start }
	t.selectionStart = start
	t.selectionEnd = end
	// Also update the native Field if it's focused
	if activeField == t.field {
		byteStart := runeToByteIndex(runes, start)
		byteEnd := runeToByteIndex(runes, end)
		t.field.SetSelection(byteStart, byteEnd)
	}
}

func runeToByteIndex(runes []rune, idx int) int {
	count := 0
	for i := 0; i < idx && i < len(runes); i++ {
		count += len(string(runes[i]))
	}
	return count
}

func (t *GTextInput) GetSelection() (int, int) {
	if t == nil { return 0, 0 }
	return t.selectionStart, t.selectionEnd
}

func (t *GTextInput) HasSelection() bool {
	if t == nil { return false }
	return t.selectionStart != t.selectionEnd
}

func (t *GTextInput) SelectAll() {
	text := t.Text()
	l := len([]rune(text))
	t.selectionStart = 0
	t.selectionEnd = l
	if activeField == t.field {
		t.field.SetSelection(0, len(text))
	}
}

func (t *GTextInput) ClearSelection() {
	p := t.cursorPosition
	t.selectionStart = p
	t.selectionEnd = p
}

// --- Focus management ---

func (t *GTextInput) RequestFocus() {
	if t == nil || !t.editable { return }
	t.focused = true
	t.cursorVisible = true
	t.lastCursorBlink = time.Now()

	if sprite := t.GTextField.GObject.DisplayObject(); sprite != nil {
		if root := core.Root(); root != nil {
			if stage := root.Stage(); stage != nil {
				stage.SetFocus(sprite)
			}
		}
	}

	func() {
		defer func() { recover() }()
		if activeField != t.field {
			if activeField != nil {
				activeField.Blur()
			}
		}
		t.field.SetTextAndSelection(t.GTextField.Text(), 0, len(t.GTextField.Text()))
		t.field.Focus()
		activeField = t.field
		activeInput = t
		fieldLastText = t.GTextField.Text()
	}()
	t.registerEventListeners()
}

func (t *GTextInput) registerEventListeners() {
	sprite := t.GTextField.GObject.DisplayObject()
	if sprite == nil { return }
	sprite.Dispatcher().Off(laya.EventKeyDown, nil)
	sprite.Dispatcher().Off(laya.EventMouseDown, nil)

	sprite.Dispatcher().On(laya.EventKeyDown, func(evt *laya.Event) {
		if keyEvt, ok := evt.Data.(laya.KeyboardEvent); ok {
			t.handleNativeKey(keyEvt)
		}
	})
	sprite.Dispatcher().On(laya.EventMouseDown, func(evt *laya.Event) {
		if pe, ok := evt.Data.(laya.PointerEvent); ok {
			local := sprite.GlobalToLocal(pe.Position)
			t.field.HandleInput(int(local.X), int(local.Y))
			t.syncStateFromField()
		}
	})
}

func (t *GTextInput) LoseFocus() {
	if t == nil { return }
	t.focused = false
	t.cursorVisible = false
	t.ClearSelection()

	sprite := t.GTextField.GObject.DisplayObject()
	if sprite != nil {
		sprite.Dispatcher().Off(laya.EventKeyDown, nil)
		sprite.Dispatcher().Off(laya.EventMouseDown, nil)
	}
	if activeField == t.field {
		t.field.Blur()
		activeField = nil
		activeInput = nil
	}
}

func (t *GTextInput) IsFocused() bool { return t != nil && t.focused }
func (t *GTextInput) IsCursorVisible() bool { return t != nil && t.focused && t.cursorVisible }

func (t *GTextInput) UpdateCursor(deltaTime float64) {
	if t == nil || !t.focused { return }
	if time.Since(t.lastCursorBlink).Seconds() >= t.cursorBlinkDelay {
		t.cursorVisible = !t.cursorVisible
		t.lastCursorBlink = time.Now()
	}
	if activeField == t.field {
		t.syncStateFromField()
	}
}

// --- Keyboard events (only for keys the textinput.Field doesn't handle) ---

func (t *GTextInput) handleNativeKey(event laya.KeyboardEvent) bool {
	if t == nil || !t.focused || !t.editable || !event.Down { return false }

	// Enter: submit (Field may or may not handle this; we handle it explicitly)
	if event.Code == laya.KeyCodeEnter {
		if t.SingleLine() { return true }
		t.insert("\n")
		return true
	}
	// Tab: allow framework to switch focus
	if event.Code == laya.KeyCodeTab {
		if t.SingleLine() { return false }
		t.insert("\t")
		return true
	}
	// Escape: blur
	if event.Code == laya.KeyCodeEscape {
		t.LoseFocus()
		return true
	}
	// All other keys are handled natively by textinput.Field
	return false
}

func (t *GTextInput) insert(s string) {
	cur := t.field.Text()
	pos := t.cursorPosition
	runes := []rune(cur)
	if t.HasSelection() {
		s0, e0 := t.selectionStart, t.selectionEnd
		if s0 > e0 { s0, e0 = e0, s0 }
		if s0 < 0 { s0 = 0 }
		if e0 > len(runes) { e0 = len(runes) }
		newRunes := append(append([]rune{}, runes[:s0]...), []rune(s)...)
		newRunes = append(newRunes, runes[e0:]...)
		pos = s0 + len([]rune(s))
		runes = newRunes
	} else {
		if pos < 0 { pos = 0 }
		if pos > len(runes) { pos = len(runes) }
		newRunes := append(append([]rune{}, runes[:pos]...), []rune(s)...)
		newRunes = append(newRunes, runes[pos:]...)
		pos += len([]rune(s))
		runes = newRunes
	}
	t.GTextField.SetText(string(runes))
	t.field.SetTextAndSelection(string(runes), pos, pos)
	t.syncStateFromField()
}
