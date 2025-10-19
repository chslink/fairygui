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
	case assets.ObjectTypeText, assets.ObjectTypeRichText:
		return NewText()
	case assets.ObjectTypeButton:
		return NewButton()
	case assets.ObjectTypeLoader:
		return NewLoader()
	case assets.ObjectTypeGroup:
		return NewGroup()
	case assets.ObjectTypeGraph:
		return NewGraph()
	default:
		return nil
	}
}
