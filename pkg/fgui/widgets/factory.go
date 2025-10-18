package widgets

import "github.com/chslink/fairygui/pkg/fgui/assets"

// CreateWidget creates a widget for the given component child metadata.
func CreateWidget(meta *assets.ComponentChild) interface{} {
	if meta == nil {
		return nil
	}
	switch meta.Type {
	case assets.ObjectTypeImage:
		return NewImage()
	case assets.ObjectTypeText:
		return NewText()
	default:
		return NewText()
	}
}
