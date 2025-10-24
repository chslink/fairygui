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
