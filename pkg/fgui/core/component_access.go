package core

import "strings"

// ComponentAccessor is implemented by widget wrappers that expose an underlying component.
type ComponentAccessor interface {
	ComponentRoot() *GComponent
}

func componentFromObject(obj *GObject) *GComponent {
	if obj == nil {
		return nil
	}
	switch data := obj.Data().(type) {
	case *GComponent:
		return data
	case ComponentAccessor:
		return data.ComponentRoot()
	default:
		return nil
	}
}

// ComponentFrom returns the component associated with the object, if any.
func ComponentFrom(obj *GObject) *GComponent {
	return componentFromObject(obj)
}

// FindChildByPath resolves nested children using dot-separated names.
func FindChildByPath(comp *GComponent, path string) *GObject {
	if comp == nil || path == "" {
		return nil
	}
	segments := strings.Split(path, ".")
	current := comp
	var obj *GObject
	for idx, segment := range segments {
		if segment == "" {
			continue
		}
		obj = current.ChildByName(segment)
		if obj == nil {
			return nil
		}
		if idx == len(segments)-1 {
			return obj
		}
		nested := componentFromObject(obj)
		if nested == nil {
			return nil
		}
		current = nested
	}
	return obj
}
