package widgets

import (
	"strings"
	"time"

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

// GTextInput is a single-line text input widget.
type GTextInput struct {
	*GTextField
	password     bool
	keyboardType KeyboardType
	editable     bool
	maxLength    int
	promptText   string
	restrict     string

	// cursor / selection state
	cursorPos       int
	selStart        int
	selEnd          int
	cursorVisible   bool
	lastCursorBlink time.Time
	blinkDelay      float64
	focused         bool

	actualText string
}

func NewTextInput() *GTextInput {
	base := NewText()
	input := &GTextInput{
		GTextField:  base,
		editable:    true,
		blinkDelay:  0.5,
		cursorVisible: true,
		lastCursorBlink: time.Now(),
	}
	base.SetSingleLine(true)
	base.GObject.SetData(input)
	if sprite := base.GObject.DisplayObject(); sprite != nil {
		sprite.SetMouseEnabled(true)
	}
	return input
}

func (t *GTextInput) SetupBeforeAdd(buf *utils.ByteBuffer, beginPos int) {
	if t == nil { return }
	if t.GTextField != nil { t.GTextField.SetupBeforeAdd(buf, beginPos) }
	if buf == nil { return }
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(beginPos, 4) { return }
	if prompt := buf.ReadS(); prompt != nil { t.SetPromptText(*prompt) }
	if restrict := buf.ReadS(); restrict != nil { t.SetRestrict(*restrict) }
	if buf.Remaining() >= 4 { if max := int(buf.ReadInt32()); max > 0 { t.SetMaxLength(max) } }
	if buf.Remaining() >= 4 {
		switch code := int(buf.ReadInt32()); code {
		case 4: t.SetKeyboardType(KeyboardTypeNumber)
		case 3: t.SetKeyboardType(KeyboardTypeURL)
		default: t.SetKeyboardType(KeyboardTypeDefault)
		}
	}
	if buf.Remaining() > 0 && buf.ReadBool() { t.SetPassword(true) }
}

// --- simple accessors ---

func (t *GTextInput) SetPassword(v bool) { t.password = v }
func (t *GTextInput) Password() bool     { return t.password }
func (t *GTextInput) SetKeyboardType(v KeyboardType) {
	if v == "" { v = KeyboardTypeDefault }; t.keyboardType = v
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

// --- cursor & selection ---

func (t *GTextInput) runes() []rune { return []rune(t.Text()) }

func (t *GTextInput) clampCursor() {
	r := t.runes()
	if t.cursorPos < 0 { t.cursorPos = 0 }
	if t.cursorPos > len(r) { t.cursorPos = len(r) }
}

func (t *GTextInput) CursorPosition() int {
	if t == nil { return 0 }
	t.clampCursor()
	return t.cursorPos
}

func (t *GTextInput) SetCursorPosition(pos int) {
	if t == nil { return }
	t.cursorPos = pos
	t.clampCursor()
	t.selStart = t.cursorPos
	t.selEnd = t.cursorPos
	t.cursorVisible = true
	t.lastCursorBlink = time.Now()
}

func (t *GTextInput) GetSelection() (start, end int) {
	if t == nil { return 0, 0 }
	return t.selStart, t.selEnd
}

func (t *GTextInput) SetSelection(start, end int) {
	if t == nil { return }
	r := t.runes()
	max := len(r)
	if start < 0 { start = 0 }
	if end < 0 { end = 0 }
	if start > max { start = max }
	if end > max { end = max }
	if start > end { start, end = end, start }
	t.selStart = start
	t.selEnd = end
	t.cursorPos = end
}

func (t *GTextInput) HasSelection() bool {
	if t == nil { return false }
	return t.selStart != t.selEnd
}

func (t *GTextInput) SelectedText() string {
	r := t.runes()
	s, e := t.selStart, t.selEnd
	if s < 0 { s = 0 }
	if e > len(r) { e = len(r) }
	if s >= e { return "" }
	return string(r[s:e])
}

func (t *GTextInput) GetSelectedText() string { return t.SelectedText() }

func (t *GTextInput) SelectAll() {
	r := t.runes()
	t.selStart = 0
	t.selEnd = len(r)
	t.cursorPos = len(r)
}

func (t *GTextInput) ClearSelection() {
	t.selStart = t.cursorPos
	t.selEnd = t.cursorPos
}

func (t *GTextInput) textLen() int { return len(t.runes()) }

// --- focus ---

var focusedInput *GTextInput

// InputChar delivers a character to the currently focused GTextInput.
// Call from the Ebiten game loop with characters from ebiten.AppendInputChars.
func InputChar(s string) {
	if focusedInput != nil && focusedInput.focused && focusedInput.editable {
		focusedInput.InsertChars(s)
	}
}

func (t *GTextInput) RequestFocus() {
	if t == nil || !t.editable { return }
	if focusedInput != nil && focusedInput != t { focusedInput.LoseFocus() }
	t.focused = true
	focusedInput = t
	t.cursorVisible = true
	t.lastCursorBlink = time.Now()

	if sprite := t.GTextField.GObject.DisplayObject(); sprite != nil {
		if root := core.Root(); root != nil {
			if stage := root.Stage(); stage != nil { stage.SetFocus(sprite) }
		}
	}
	t.registerListeners()
}

func (t *GTextInput) LoseFocus() {
	if t == nil { return }
	t.focused = false
	if focusedInput == t { focusedInput = nil }
	t.cursorVisible = false
	t.ClearSelection()
	sprite := t.GTextField.GObject.DisplayObject()
	if sprite != nil {
		sprite.Dispatcher().Off(laya.EventKeyDown, nil)
		sprite.Dispatcher().Off(laya.EventMouseDown, nil)
	}
}

func (t *GTextInput) IsFocused() bool       { return t != nil && t.focused }
func (t *GTextInput) IsCursorVisible() bool { return t != nil && t.focused && t.cursorVisible }

func (t *GTextInput) UpdateCursor(dt float64) {
	if t == nil || !t.focused { return }
	if time.Since(t.lastCursorBlink).Seconds() >= t.blinkDelay {
		t.cursorVisible = !t.cursorVisible
		t.lastCursorBlink = time.Now()
	}
}

func (t *GTextInput) registerListeners() {
	sprite := t.GTextField.GObject.DisplayObject()
	if sprite == nil { return }
	sprite.Dispatcher().Off(laya.EventKeyDown, nil)
	sprite.Dispatcher().Off(laya.EventMouseDown, nil)

	sprite.Dispatcher().On(laya.EventKeyDown, func(evt *laya.Event) {
		if ke, ok := evt.Data.(laya.KeyboardEvent); ok {
			t.HandleKeyboardEvent(ke)
		}
	})
	sprite.Dispatcher().On(laya.EventMouseDown, func(evt *laya.Event) {
		if pe, ok := evt.Data.(laya.PointerEvent); ok {
			local := sprite.GlobalToLocal(pe.Position)
			t.HandleMouseDown(local.X, local.Y)
		}
	})
}

// --- mouse ---

func (t *GTextInput) HandleMouseDown(x, y float64) bool {
	if t == nil || !t.editable { return false }
	if !t.focused { t.RequestFocus() }

	// Estimate cursor position from x coordinate
	fontSize := float64(t.FontSize())
	charW := fontSize * 0.6
	pos := int(x / charW)
	r := t.runes()
	if pos < 0 { pos = 0 }
	if pos > len(r) { pos = len(r) }
	t.cursorPos = pos
	t.selStart = pos
	t.selEnd = pos
	t.cursorVisible = true
	t.lastCursorBlink = time.Now()
	return true
}

// --- keyboard (control keys only; characters come via InsertChars) ---

func (t *GTextInput) HandleKeyboardEvent(event laya.KeyboardEvent) bool {
	if t == nil || !t.focused || !t.editable || !event.Down { return false }

	// Shortcuts
	if event.Modifiers.Ctrl || event.Modifiers.Meta { return t.handleShortcut(event) }

	switch event.Code {
	case laya.KeyCodeBackspace:
		if t.HasSelection() { t.deleteSelection() } else { t.backspace() }
		return true
	case laya.KeyCodeDelete:
		if t.HasSelection() { t.deleteSelection() } else { t.del() }
		return true
	case laya.KeyCodeLeft:
		if event.Modifiers.Shift { t.extendSelLeft() } else { t.moveCursor(-1) }
		return true
	case laya.KeyCodeRight:
		if event.Modifiers.Shift { t.extendSelRight() } else { t.moveCursor(1) }
		return true
	case laya.KeyCodeHome:
		if event.Modifiers.Shift { t.extendSelTo(0) } else { t.moveCursorTo(0) }
		return true
	case laya.KeyCodeEnd:
		if event.Modifiers.Shift { t.extendSelTo(t.textLen()) } else { t.moveCursorTo(t.textLen()) }
		return true
	case laya.KeyCodeEnter:
		if t.SingleLine() { return true }
		t.insertChars("\n")
		return true
	case laya.KeyCodeTab:
		if t.SingleLine() { return false }
		t.insertChars("\t")
		return true
	default:
		// Only process rune chars when they come through InsertChars.
		// event.Rune may be stale; ignore it here.
		return false
	}
}

// InsertChars inserts a string at the current cursor position, replacing any selection.
// Called from the game loop with characters collected by ebiten.AppendInputChars.
func (t *GTextInput) InsertChars(s string) {
	if t == nil || !t.focused || !t.editable || s == "" { return }
	t.insertChars(s)
}

func (t *GTextInput) insertChars(s string) {
	if t.maxLength > 0 && t.textLen() >= t.maxLength { return }
	r := t.runes()
	a, b := t.selStart, t.selEnd
	if a > b { a, b = b, a }
	// Replace selection
	insert := []rune(s)
	r = append(append(r[:a], insert...), r[b:]...)
	t.GTextField.SetText(string(r))
	t.cursorPos = a + len(insert)
	t.selStart = t.cursorPos
	t.selEnd = t.cursorPos
	t.cursorVisible = true
	t.lastCursorBlink = time.Now()
}

func (t *GTextInput) backspace() {
	if t.cursorPos <= 0 { return }
	r := t.runes()
	r = append(r[:t.cursorPos-1], r[t.cursorPos:]...)
	t.GTextField.SetText(string(r))
	t.cursorPos--
	t.selStart = t.cursorPos
	t.selEnd = t.cursorPos
}

func (t *GTextInput) del() {
	r := t.runes()
	if t.cursorPos >= len(r) { return }
	r = append(r[:t.cursorPos], r[t.cursorPos+1:]...)
	t.GTextField.SetText(string(r))
}

func (t *GTextInput) deleteSelection() {
	r := t.runes()
	a, b := t.selStart, t.selEnd
	if a > b { a, b = b, a }
	r = append(r[:a], r[b:]...)
	t.GTextField.SetText(string(r))
	t.cursorPos = a
	t.selStart = a
	t.selEnd = a
}

func (t *GTextInput) moveCursor(delta int) {
	t.cursorPos += delta
	t.clampCursor()
	t.selStart = t.cursorPos
	t.selEnd = t.cursorPos
}

func (t *GTextInput) moveCursorTo(pos int) {
	t.cursorPos = pos
	t.clampCursor()
	t.selStart = t.cursorPos
	t.selEnd = t.cursorPos
}

func (t *GTextInput) extendSelLeft() {
	if !t.HasSelection() { t.selStart, t.selEnd = t.cursorPos, t.cursorPos }
	t.cursorPos--
	t.clampCursor()
	t.selEnd = t.cursorPos
	if t.selEnd < t.selStart { t.selStart, t.selEnd = t.selEnd, t.selStart }
}

func (t *GTextInput) extendSelRight() {
	if !t.HasSelection() { t.selStart, t.selEnd = t.cursorPos, t.cursorPos }
	t.cursorPos++
	t.clampCursor()
	t.selEnd = t.cursorPos
	if t.selEnd < t.selStart { t.selStart, t.selEnd = t.selEnd, t.selStart }
}

func (t *GTextInput) extendSelTo(pos int) {
	if !t.HasSelection() { t.selStart, t.selEnd = t.cursorPos, t.cursorPos }
	t.cursorPos = pos
	t.clampCursor()
	t.selEnd = t.cursorPos
	if t.selEnd < t.selStart { t.selStart, t.selEnd = t.selEnd, t.selStart }
}

// --- shortcuts ---

func (t *GTextInput) handleShortcut(event laya.KeyboardEvent) bool {
	switch event.Code {
	case laya.KeyCodeA: t.SelectAll(); return true
	case laya.KeyCodeC:
		if s := t.SelectedText(); s != "" { internalClipboard = s }
		return t.HasSelection()
	case laya.KeyCodeX:
		if s := t.SelectedText(); s != "" { internalClipboard = s }
		t.deleteSelection(); return true
	case laya.KeyCodeV:
		if internalClipboard != "" { t.insertChars(internalClipboard) }
		return true
	}
	return false
}

// internalClipboard is an in-process clipboard.
var internalClipboard string
