package widgets

import "github.com/chslink/fairygui/pkg/fgui/core"

// GTextField is a minimal text widget.
type TextAlign string

const (
	TextAlignLeft   TextAlign = "left"
	TextAlignCenter TextAlign = "center"
	TextAlignRight  TextAlign = "right"
)

type TextVerticalAlign string

const (
	TextVerticalAlignTop    TextVerticalAlign = "top"
	TextVerticalAlignMiddle TextVerticalAlign = "middle"
	TextVerticalAlignBottom TextVerticalAlign = "bottom"
)

type TextAutoSize int

const (
	TextAutoSizeBoth TextAutoSize = iota
	TextAutoSizeHeight
	TextAutoSizeShrink
	TextAutoSizeEllipsis
)

// GTextField is a minimal text widget.
type GTextField struct {
	*core.GObject
	text         string
	color        string
	outlineColor string
	fontSize     int
	font         string
	align        TextAlign
	vertical     TextVerticalAlign
	autoSize     TextAutoSize
	singleLine   bool
	underline    bool
	italic       bool
	bold         bool
	letterSpace  int
	leading      int
	strokeSize   float64
	strokeColor  string
	ubbEnabled   bool
	templateVars bool
}

// NewText creates a new text field widget.
func NewText() *GTextField {
	field := &GTextField{GObject: core.NewGObject()}
	field.color = "#000000"
	field.fontSize = 12
	field.align = TextAlignLeft
	field.vertical = TextVerticalAlignTop
	field.autoSize = TextAutoSizeBoth
	field.GObject.SetData(field)
	return field
}

// SetText updates the displayed text.
func (t *GTextField) SetText(value string) {
	t.text = value
}

// Text returns the current text.
func (t *GTextField) Text() string {
	return t.text
}

// SetColor updates the text colour stored on this widget.
func (t *GTextField) SetColor(value string) {
	t.color = value
}

// Color returns the current text colour.
func (t *GTextField) Color() string {
	return t.color
}

// SetFont stores the requested font face identifier.
func (t *GTextField) SetFont(value string) {
	t.font = value
}

// Font returns the stored font identifier.
func (t *GTextField) Font() string {
	return t.font
}

// SetOutlineColor stores the outline colour value.
func (t *GTextField) SetOutlineColor(value string) {
	t.outlineColor = value
}

// OutlineColor returns the current outline colour.
func (t *GTextField) OutlineColor() string {
	return t.outlineColor
}

// SetFontSize records the font size associated with the text.
func (t *GTextField) SetFontSize(size int) {
	t.fontSize = size
}

// FontSize returns the stored font size.
func (t *GTextField) FontSize() int {
	return t.fontSize
}

// SetAlign updates the horizontal alignment.
func (t *GTextField) SetAlign(value TextAlign) {
	if value == "" {
		value = TextAlignLeft
	}
	t.align = value
}

// Align returns the horizontal alignment.
func (t *GTextField) Align() TextAlign {
	return t.align
}

// SetVerticalAlign updates the vertical alignment.
func (t *GTextField) SetVerticalAlign(value TextVerticalAlign) {
	if value == "" {
		value = TextVerticalAlignTop
	}
	t.vertical = value
}

// VerticalAlign returns the vertical alignment.
func (t *GTextField) VerticalAlign() TextVerticalAlign {
	return t.vertical
}

// SetAutoSize configures the autosize behaviour.
func (t *GTextField) SetAutoSize(value TextAutoSize) {
	t.autoSize = value
}

// AutoSize returns the stored autosize mode.
func (t *GTextField) AutoSize() TextAutoSize {
	return t.autoSize
}

// SetSingleLine toggles single-line mode.
func (t *GTextField) SetSingleLine(value bool) {
	t.singleLine = value
}

// SingleLine reports whether single-line mode is active.
func (t *GTextField) SingleLine() bool {
	return t.singleLine
}

// SetUnderline toggles underline rendering.
func (t *GTextField) SetUnderline(value bool) {
	t.underline = value
}

// Underline reports whether underline should be drawn.
func (t *GTextField) Underline() bool {
	return t.underline
}

// SetItalic toggles italic styling.
func (t *GTextField) SetItalic(value bool) {
	t.italic = value
}

// Italic reports whether italic styling is requested.
func (t *GTextField) Italic() bool {
	return t.italic
}

// SetBold toggles bold styling.
func (t *GTextField) SetBold(value bool) {
	t.bold = value
}

// Bold reports whether bold styling is requested.
func (t *GTextField) Bold() bool {
	return t.bold
}

// SetLetterSpacing stores the additional spacing between characters (pixels).
func (t *GTextField) SetLetterSpacing(value int) {
	t.letterSpace = value
}

// LetterSpacing returns stored spacing between characters.
func (t *GTextField) LetterSpacing() int {
	return t.letterSpace
}

// SetLeading stores the additional line spacing in pixels.
func (t *GTextField) SetLeading(value int) {
	t.leading = value
}

// Leading returns the stored line spacing.
func (t *GTextField) Leading() int {
	return t.leading
}

// SetStrokeSize stores the outline thickness.
func (t *GTextField) SetStrokeSize(value float64) {
	t.strokeSize = value
}

// StrokeSize returns the outline thickness.
func (t *GTextField) StrokeSize() float64 {
	return t.strokeSize
}

// SetStrokeColor configures the outline colour.
func (t *GTextField) SetStrokeColor(value string) {
	t.strokeColor = value
}

// StrokeColor returns the outline colour.
func (t *GTextField) StrokeColor() string {
	return t.strokeColor
}

// SetUBBEnabled toggles UBB formatting support.
func (t *GTextField) SetUBBEnabled(value bool) {
	t.ubbEnabled = value
}

// UBBEnabled reports whether UBB formatting is enabled.
func (t *GTextField) UBBEnabled() bool {
	return t.ubbEnabled
}

// SetTemplateVarsEnabled records whether template variables are active.
func (t *GTextField) SetTemplateVarsEnabled(value bool) {
	t.templateVars = value
}

// TemplateVarsEnabled reports template variable usage.
func (t *GTextField) TemplateVarsEnabled() bool {
	return t.templateVars
}
