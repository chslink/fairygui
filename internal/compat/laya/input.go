package laya

// MouseButtons describes primary mouse buttons state.
type MouseButtons struct {
	Left   bool
	Right  bool
	Middle bool
}

// KeyModifiers mirrors common keyboard modifier state.
type KeyModifiers struct {
	Shift bool
	Ctrl  bool
	Alt   bool
	Meta  bool
}

// KeyCode represents a platform-agnostic key identifier.
type KeyCode int

// 常用键码常量 (映射到 Ebiten 的键码)
const (
	KeyCodeBackspace KeyCode = 8
	KeyCodeTab       KeyCode = 9
	KeyCodeEnter     KeyCode = 13
	KeyCodeEscape    KeyCode = 27
	KeyCodeSpace     KeyCode = 32
	KeyCodeLeft      KeyCode = 37
	KeyCodeUp        KeyCode = 38
	KeyCodeRight     KeyCode = 39
	KeyCodeDown      KeyCode = 40
	KeyCodeDelete    KeyCode = 46
	KeyCodeHome      KeyCode = 36
	KeyCodeEnd       KeyCode = 35
	KeyCodeA         KeyCode = 65
	KeyCodeC         KeyCode = 67
	KeyCodeV         KeyCode = 86
	KeyCodeX         KeyCode = 88
	KeyCodeZ         KeyCode = 90
)

// KeyboardEvent carries keyboard input details dispatched by the stage.
type KeyboardEvent struct {
	Code      KeyCode
	Rune      rune
	Down      bool
	Repeat    bool
	Modifiers KeyModifiers
}

// TouchPhase enumerates touch lifecycle stages.
type TouchPhase int

const (
	TouchPhaseBegin TouchPhase = iota
	TouchPhaseMove
	TouchPhaseEnd
	TouchPhaseCancel
)

// TouchInput describes a single touch update fed into the stage.
type TouchInput struct {
	ID       int
	Position Point
	Phase    TouchPhase
	Primary  bool
}

// InputState bundles mouse, touch, and keyboard input for a single frame.
type InputState struct {
	Mouse   MouseState
	Touches []TouchInput
	Keys    []KeyboardEvent
}
