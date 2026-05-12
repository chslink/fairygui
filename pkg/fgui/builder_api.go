package fgui

import "github.com/chslink/fairygui/pkg/fgui/widgets"

// ButtonBuilder provides a chainable API for constructing GButton instances.
type ButtonBuilder struct {
	btn     *GButton
	clickID ListenerID
}

// NewButtonBuilder starts a button building chain.
func NewButtonBuilder() *ButtonBuilder {
	return &ButtonBuilder{btn: widgets.NewButton()}
}

// Title sets the button title.
func (b *ButtonBuilder) Title(text string) *ButtonBuilder {
	b.btn.SetTitle(text)
	return b
}

// SelectedTitle sets the selected-state title.
func (b *ButtonBuilder) SelectedTitle(text string) *ButtonBuilder {
	b.btn.SetSelectedTitle(text)
	return b
}

// Icon sets the button icon URL.
func (b *ButtonBuilder) Icon(url string) *ButtonBuilder {
	b.btn.SetIcon(url)
	return b
}

// SelectedIcon sets the selected-state icon.
func (b *ButtonBuilder) SelectedIcon(url string) *ButtonBuilder {
	b.btn.SetSelectedIcon(url)
	return b
}

// Position sets the button position.
func (b *ButtonBuilder) Position(x, y float64) *ButtonBuilder {
	b.btn.SetPosition(x, y)
	return b
}

// Size sets the button size.
func (b *ButtonBuilder) Size(w, h float64) *ButtonBuilder {
	b.btn.SetSize(w, h)
	return b
}

// Alpha sets the button transparency.
func (b *ButtonBuilder) Alpha(a float64) *ButtonBuilder {
	b.btn.SetAlpha(a)
	return b
}

// Visible sets visibility.
func (b *ButtonBuilder) Visible(v bool) *ButtonBuilder {
	b.btn.SetVisible(v)
	return b
}

// Touchable sets interactivity.
func (b *ButtonBuilder) Touchable(t bool) *ButtonBuilder {
	b.btn.SetTouchable(t)
	return b
}

// Mode sets the button mode.
func (b *ButtonBuilder) Mode(m widgets.ButtonMode) *ButtonBuilder {
	b.btn.SetMode(m)
	return b
}

// Selected sets the initial selection state.
func (b *ButtonBuilder) Selected(s bool) *ButtonBuilder {
	b.btn.SetSelected(s)
	return b
}

// Sound sets the click sound.
func (b *ButtonBuilder) Sound(url string) *ButtonBuilder {
	b.btn.SetSound(url)
	return b
}

// OnClick registers a click handler.
func (b *ButtonBuilder) OnClick(fn func()) *ButtonBuilder {
	b.clickID = b.btn.OnClick(fn)
	return b
}

// Build returns the constructed GButton.
func (b *ButtonBuilder) Build() *GButton {
	return b.btn
}

// ListBuilder provides a chainable API for constructing GList instances.
type ListBuilder struct {
	list *GList
}

// NewListBuilder starts a list building chain.
func NewListBuilder() *ListBuilder {
	return &ListBuilder{list: widgets.NewList()}
}

// DefaultItem sets the default item URL.
func (b *ListBuilder) DefaultItem(url string) *ListBuilder {
	b.list.SetDefaultItem(url)
	return b
}

// Virtual enables virtual list mode.
func (b *ListBuilder) Virtual(enabled bool) *ListBuilder {
	b.list.SetVirtual(enabled)
	return b
}

// Loop enables loop mode.
func (b *ListBuilder) Loop(enabled bool) *ListBuilder {
	b.list.SetLoop(enabled)
	return b
}

// NumItems sets the number of data items.
func (b *ListBuilder) NumItems(n int) *ListBuilder {
	b.list.SetNumItems(n)
	return b
}

// ItemRenderer sets the item renderer callback.
func (b *ListBuilder) ItemRenderer(fn func(int, *GObject)) *ListBuilder {
	b.list.SetItemRenderer(fn)
	return b
}

// ItemProvider sets the item provider (for multi-type lists).
func (b *ListBuilder) ItemProvider(fn func(int) string) *ListBuilder {
	b.list.SetItemProvider(fn)
	return b
}

// SelectionMode sets the selection mode.
func (b *ListBuilder) SelectionMode(m widgets.ListSelectionMode) *ListBuilder {
	b.list.SetSelectionMode(m)
	return b
}

// Size sets the list size.
func (b *ListBuilder) Size(w, h float64) *ListBuilder {
	b.list.SetSize(w, h)
	return b
}

// Position sets the list position.
func (b *ListBuilder) Position(x, y float64) *ListBuilder {
	b.list.SetPosition(x, y)
	return b
}

// Build returns the constructed GList.
func (b *ListBuilder) Build() *GList {
	return b.list
}
