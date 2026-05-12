package fgui

import (
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// ComponentConfig holds optional parameters for creating UI components.
type ComponentConfig struct {
	ID        string
	Name      string
	X, Y      float64
	Width     float64
	Height    float64
	ScaleX    float64
	ScaleY    float64
	Alpha     float64
	Rotation  float64
	Visible   *bool
	Touchable *bool
	PivotX    float64
	PivotY    float64
	AsAnchor  bool
	Grayed    bool
	SortOrder int
}

func defaultComponentConfig() *ComponentConfig {
	t := true
	return &ComponentConfig{
		ScaleX:    1,
		ScaleY:    1,
		Alpha:     1,
		Visible:   &t,
		Touchable: &t,
	}
}

// ComponentOption is a functional option for configurable component creation.
type ComponentOption func(*ComponentConfig)

// WithID sets the component ID.
func WithID(id string) ComponentOption { return func(c *ComponentConfig) { c.ID = id } }

// WithName sets the component display name.
func WithName(name string) ComponentOption { return func(c *ComponentConfig) { c.Name = name } }

// WithPosition sets the component position.
func WithPosition(x, y float64) ComponentOption { return func(c *ComponentConfig) { c.X, c.Y = x, y } }

// WithSize sets the component size.
func WithSize(w, h float64) ComponentOption { return func(c *ComponentConfig) { c.Width, c.Height = w, h } }

// WithScale sets the component scale factors.
func WithScale(sx, sy float64) ComponentOption { return func(c *ComponentConfig) { c.ScaleX, c.ScaleY = sx, sy } }

// WithAlpha sets the component transparency.
func WithAlpha(a float64) ComponentOption { return func(c *ComponentConfig) { c.Alpha = a } }

// WithRotation sets the rotation in degrees.
func WithRotation(r float64) ComponentOption { return func(c *ComponentConfig) { c.Rotation = r } }

// WithPivot sets the normalized pivot point.
func WithPivot(px, py float64) ComponentOption { return func(c *ComponentConfig) { c.PivotX, c.PivotY = px, py } }

// WithPivotAnchor sets the pivot and enables anchor mode.
func WithPivotAnchor(px, py float64) ComponentOption {
	return func(c *ComponentConfig) { c.PivotX, c.PivotY = px, py; c.AsAnchor = true }
}

// Hidden makes the component initially invisible.
func Hidden() ComponentOption {
	f := false
	return func(c *ComponentConfig) { c.Visible = &f }
}

// Disabled makes the component non-interactive.
func Disabled() ComponentOption {
	f := false
	return func(c *ComponentConfig) { c.Touchable = &f }
}

// WithGrayed enables the grayed (disabled look) state.
func WithGrayed() ComponentOption { return func(c *ComponentConfig) { c.Grayed = true } }

// WithSortingOrder sets the Z-order for rendering.
func WithSortingOrder(order int) ComponentOption { return func(c *ComponentConfig) { c.SortOrder = order } }

func applyComponentOptions(obj *core.GObject, opts []ComponentOption) {
	cfg := defaultComponentConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.ID != "" {
		obj.SetResourceID(cfg.ID)
	}
	if cfg.Name != "" {
		obj.SetName(cfg.Name)
	}
	obj.SetPosition(cfg.X, cfg.Y)
	if cfg.Width != 0 || cfg.Height != 0 {
		obj.SetSize(cfg.Width, cfg.Height)
	}
	if cfg.ScaleX != 1 || cfg.ScaleY != 1 {
		obj.SetScale(cfg.ScaleX, cfg.ScaleY)
	}
	if cfg.Alpha != 1 {
		obj.SetAlpha(cfg.Alpha)
	}
	if cfg.Rotation != 0 {
		obj.SetRotation(cfg.Rotation)
	}
	if cfg.PivotX != 0 || cfg.PivotY != 0 {
		obj.SetPivotWithAnchor(cfg.PivotX, cfg.PivotY, cfg.AsAnchor)
	}
	if cfg.Visible != nil {
		obj.SetVisible(*cfg.Visible)
	}
	if cfg.Touchable != nil {
		obj.SetTouchable(*cfg.Touchable)
	}
	if cfg.Grayed {
		obj.SetGrayed(true)
	}
	if cfg.SortOrder != 0 {
		obj.SetSortingOrder(cfg.SortOrder)
	}
}

// NewObject creates a GObject with optional configuration.
func NewObject(opts ...ComponentOption) *GObject {
	obj := core.NewGObject()
	applyComponentOptions(obj, opts)
	return obj
}

// NewComponent creates a GComponent with optional configuration.
func NewComponent(opts ...ComponentOption) *GComponent {
	comp := core.NewGComponent()
	applyComponentOptions(comp.GObject, opts)
	return comp
}

// ButtonConfig extends ComponentConfig with button-specific options.
type ButtonConfig struct {
	ComponentConfig
	Title        string
	Icon         string
	Selected     bool
	Mode         string // "common", "check", "radio"
	OnClick      func()
	SoundURL     string
	SoundVolume  float64
}

// ButtonOption is a functional option for button creation.
type ButtonOption func(*ButtonConfig)

func defaultButtonConfig() *ButtonConfig {
	cfg := defaultComponentConfig()
	return &ButtonConfig{
		ComponentConfig: *cfg,
		SoundVolume:     1,
	}
}

// WithTitle sets the button title.
func WithTitle(title string) ButtonOption { return func(c *ButtonConfig) { c.Title = title } }

// WithIcon sets the button icon URL.
func WithIcon(icon string) ButtonOption { return func(c *ButtonConfig) { c.Icon = icon } }

// WithSelected sets the initial selected state.
func WithSelected(sel bool) ButtonOption { return func(c *ButtonConfig) { c.Selected = sel } }

// WithButtonClick sets the click handler.
func WithButtonClick(fn func()) ButtonOption { return func(c *ButtonConfig) { c.OnClick = fn } }

// WithSound sets the button click sound.
func WithSound(url string, vol float64) ButtonOption {
	return func(c *ButtonConfig) { c.SoundURL = url; c.SoundVolume = vol }
}

// CreateButton creates a GButton with optional configuration.
func CreateButton(opts ...ButtonOption) *GButton {
	cfg := defaultButtonConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	btn := widgets.NewButton()
	applyComponentOptions(btn.GObject, []ComponentOption{func(c *ComponentConfig) { *c = cfg.ComponentConfig }})
	if cfg.Title != "" {
		btn.SetTitle(cfg.Title)
	}
	if cfg.Icon != "" {
		btn.SetIcon(cfg.Icon)
	}
	if cfg.Selected {
		btn.SetSelected(true)
	}
	if cfg.OnClick != nil {
		btn.OnClick(cfg.OnClick)
	}
	return btn
}

// LabelConfig extends ComponentConfig with label-specific options.
type LabelConfig struct {
	ComponentConfig
	Text  string
	Icon  string
	Title string
	Color string
	FontSize int
}

type LabelOption func(*LabelConfig)

func WithText(text string) LabelOption  { return func(c *LabelConfig) { c.Text = text } }
func WithLabelIcon(icon string) LabelOption { return func(c *LabelConfig) { c.Icon = icon } }
func WithLabelTitle(title string) LabelOption { return func(c *LabelConfig) { c.Title = title } }
func WithColor(color string) LabelOption { return func(c *LabelConfig) { c.Color = color } }
func WithFontSize(size int) LabelOption { return func(c *LabelConfig) { c.FontSize = size } }

// CreateLabel creates a GLabel with optional configuration.
func CreateLabel(opts ...LabelOption) *GLabel {
	cfg := &LabelConfig{ComponentConfig: *defaultComponentConfig(), FontSize: 12}
	for _, opt := range opts {
		opt(cfg)
	}
	label := widgets.NewLabel()
	applyComponentOptions(label.GObject, []ComponentOption{func(c *ComponentConfig) { *c = cfg.ComponentConfig }})
	if cfg.Text != "" {
		label.SetTitle(cfg.Text)
	}
	if cfg.Icon != "" {
		label.SetIcon(cfg.Icon)
	}
	if cfg.Title != "" {
		label.SetTitle(cfg.Title)
	}
	if cfg.Color != "" {
		label.SetTitleColor(cfg.Color)
	}
	if cfg.FontSize != 0 {
		label.SetTitleFontSize(cfg.FontSize)
	}
	return label
}
