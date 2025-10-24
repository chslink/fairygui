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
	case assets.ObjectTypeMovieClip:
		return NewMovieClip()
	case assets.ObjectTypeText, assets.ObjectTypeRichText:
		return NewText()
	case assets.ObjectTypeInputText:
		return NewTextInput()
	case assets.ObjectTypeButton:
		return NewButton()
	case assets.ObjectTypeLoader:
		return NewLoader()
	case assets.ObjectTypeGroup:
		return NewGroup()
	case assets.ObjectTypeGraph:
		return NewGraph()
	case assets.ObjectTypeList:
		return NewList()
	case assets.ObjectTypeLabel:
		return NewLabel()
	case assets.ObjectTypeProgressBar:
		return NewProgressBar()
	case assets.ObjectTypeSlider:
		return NewSlider()
	case assets.ObjectTypeScrollBar:
		return NewScrollBar()
	case assets.ObjectTypeComboBox:
		return NewComboBox()
	case assets.ObjectTypeTree:
		return NewTree()
	default:
		return nil
	}
}

// CreateWidgetFromPackage attempts to instantiate a widget based on the package item's object type.
func CreateWidgetFromPackage(item *assets.PackageItem) interface{} {
	if item == nil {
		return nil
	}
	switch item.ObjectType {
	case assets.ObjectTypeButton:
		return NewButton()
	case assets.ObjectTypeList:
		return NewList()
	case assets.ObjectTypeLabel:
		return NewLabel()
	case assets.ObjectTypeLoader:
		return NewLoader()
	case assets.ObjectTypeImage:
		return NewImage()
	case assets.ObjectTypeMovieClip:
		return NewMovieClip()
	case assets.ObjectTypeGraph:
		return NewGraph()
	case assets.ObjectTypeGroup:
		return NewGroup()
	case assets.ObjectTypeText, assets.ObjectTypeRichText:
		return NewText()
	case assets.ObjectTypeInputText:
		return NewTextInput()
	case assets.ObjectTypeProgressBar:
		return NewProgressBar()
	case assets.ObjectTypeSlider:
		return NewSlider()
	case assets.ObjectTypeScrollBar:
		return NewScrollBar()
	case assets.ObjectTypeComboBox:
		return NewComboBox()
	case assets.ObjectTypeTree:
		return NewTree()
	default:
		return nil
	}
}
