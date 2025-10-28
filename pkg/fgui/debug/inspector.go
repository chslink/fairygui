package debug

import (
	"fmt"
	"strings"

	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// Inspector 提供对象检查和查找功能
type Inspector struct {
	root *core.GObject
}

// NewInspector 创建对象检查器
func NewInspector(root *core.GObject) *Inspector {
	return &Inspector{root: root}
}

// ObjectInfo 对象信息
type ObjectInfo struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Position   Position               `json:"position"`
	Size       Size                   `json:"size"`
	Visible    bool                   `json:"visible"`
	Alpha      float64                `json:"alpha"`
	Rotation   float64                `json:"rotation"`
	Parent     string                 `json:"parent,omitempty"`
	Children   int                    `json:"children"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Position 位置信息
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Size 尺寸信息
type Size struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// FindByName 按名称查找对象（支持部分匹配）
func (i *Inspector) FindByName(name string) []*core.GObject {
	var results []*core.GObject
	i.walkTree(i.root, func(obj *core.GObject) bool {
		if strings.Contains(obj.Name(), name) {
			results = append(results, obj)
		}
		return true
	})
	return results
}

// FindByType 按类型查找对象
func (i *Inspector) FindByType(typeName string) []*core.GObject {
	var results []*core.GObject
	i.walkTree(i.root, func(obj *core.GObject) bool {
		objType := GetObjectType(obj)
		if strings.Contains(strings.ToLower(objType), strings.ToLower(typeName)) {
			results = append(results, obj)
		}
		return true
	})
	return results
}

// FindByID 按对象指针地址查找（精确匹配）
func (i *Inspector) FindByID(id string) *core.GObject {
	var result *core.GObject
	i.walkTree(i.root, func(obj *core.GObject) bool {
		if fmt.Sprintf("%p", obj) == id {
			result = obj
			return false // 停止遍历
		}
		return true
	})
	return result
}

// FindByPath 按路径查找对象（如：/Scene/Panel/Button）
func (i *Inspector) FindByPath(path string) *core.GObject {
	if path == "" || path == "/" {
		return i.root
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	current := i.root

	for _, part := range parts {
		found := false
		comp, ok := current.Data().(*core.GComponent)
		if !ok {
			return nil
		}

		for _, child := range comp.Children() {
			if child.Name() == part {
				current = child
				found = true
				break
			}
		}

		if !found {
			return nil
		}
	}

	return current
}

// GetInfo 获取对象完整信息
func (i *Inspector) GetInfo(obj *core.GObject) *ObjectInfo {
	if obj == nil {
		return nil
	}

	info := &ObjectInfo{
		ID:       fmt.Sprintf("%p", obj),
		Name:     obj.Name(),
		Type:     GetObjectType(obj),
		Position: Position{X: obj.X(), Y: obj.Y()},
		Size:     Size{Width: obj.Width(), Height: obj.Height()},
		Visible:  obj.Visible(),
		Alpha:    obj.Alpha(),
		Rotation: obj.Rotation(),
	}

	if parent := obj.Parent(); parent != nil {
		info.Parent = parent.Name()
	}

	// 统计子对象数量
	if comp, ok := obj.Data().(*core.GComponent); ok {
		info.Children = len(comp.Children())
	}

	// 添加特定属性
	info.Properties = i.getObjectProperties(obj)

	return info
}

// GetPath 获取对象的完整路径
func (i *Inspector) GetPath(obj *core.GObject) string {
	if obj == nil {
		return ""
	}

	var parts []string
	current := obj

	for current != nil {
		parts = append([]string{current.Name()}, parts...)
		if parent := current.Parent(); parent != nil {
			current = parent.GObject
		} else {
			break
		}
	}

	return "/" + strings.Join(parts, "/")
}

// CountObjects 统计对象数量（可选择性地按类型统计）
func (i *Inspector) CountObjects() map[string]int {
	counts := make(map[string]int)
	counts["total"] = 0
	counts["visible"] = 0
	counts["containers"] = 0

	i.walkTree(i.root, func(obj *core.GObject) bool {
		counts["total"]++
		if obj.Visible() {
			counts["visible"]++
		}

		objType := GetObjectType(obj)
		counts[objType]++

		if _, ok := obj.Data().(*core.GComponent); ok {
			counts["containers"]++
		}

		return true
	})

	return counts
}

// GetChildrenCount 获取对象的子对象数量（递归统计所有后代）
func (i *Inspector) GetChildrenCount(obj *core.GObject, recursive bool) int {
	if obj == nil {
		return 0
	}

	comp, ok := obj.Data().(*core.GComponent)
	if !ok {
		return 0
	}

	children := comp.Children()
	if !recursive {
		return len(children)
	}

	// 递归统计
	count := len(children)
	for _, child := range children {
		count += i.GetChildrenCount(child, true)
	}

	return count
}

// Filter 筛选对象
type Filter struct {
	Name    string // 名称（部分匹配）
	Type    string // 类型（部分匹配）
	Visible *bool  // 可见性
	MinX    *float64
	MaxX    *float64
	MinY    *float64
	MaxY    *float64
}

// FindByFilter 按筛选条件查找对象
func (i *Inspector) FindByFilter(filter Filter) []*core.GObject {
	var results []*core.GObject

	i.walkTree(i.root, func(obj *core.GObject) bool {
		// 名称筛选
		if filter.Name != "" && !strings.Contains(strings.ToLower(obj.Name()), strings.ToLower(filter.Name)) {
			return true
		}

		// 类型筛选
		if filter.Type != "" {
			objType := GetObjectType(obj)
			if !strings.Contains(strings.ToLower(objType), strings.ToLower(filter.Type)) {
				return true
			}
		}

		// 可见性筛选
		if filter.Visible != nil && obj.Visible() != *filter.Visible {
			return true
		}

		// 位置筛选
		if filter.MinX != nil && obj.X() < *filter.MinX {
			return true
		}
		if filter.MaxX != nil && obj.X() > *filter.MaxX {
			return true
		}
		if filter.MinY != nil && obj.Y() < *filter.MinY {
			return true
		}
		if filter.MaxY != nil && obj.Y() > *filter.MaxY {
			return true
		}

		results = append(results, obj)
		return true
	})

	return results
}

// walkTree 遍历对象树
func (i *Inspector) walkTree(obj *core.GObject, fn func(*core.GObject) bool) {
	if obj == nil {
		return
	}

	if !fn(obj) {
		return // 停止遍历
	}

	// 递归遍历子对象
	if comp, ok := obj.Data().(*core.GComponent); ok {
		for _, child := range comp.Children() {
			i.walkTree(child, fn)
		}
	}
}

// getObjectProperties 获取对象特定属性
func (i *Inspector) getObjectProperties(obj *core.GObject) map[string]interface{} {
	props := make(map[string]interface{})

	switch widget := obj.Data().(type) {
	case *widgets.GButton:
		props["title"] = widget.Title()
		props["selected"] = widget.Selected()
		if icon := widget.Icon(); icon != "" {
			props["icon"] = icon
		}
		props["mode"] = widget.Mode()

	case *widgets.GTextField:
		props["text"] = widget.Text()
		props["fontSize"] = widget.FontSize()
		props["color"] = fmt.Sprintf("#%06X", widget.Color())
		props["singleLine"] = widget.SingleLine()
		props["autoSize"] = widget.AutoSize()

	case *widgets.GList:
		props["layout"] = widget.Layout()
		props["lineCount"] = widget.LineCount()
		props["columnCount"] = widget.ColumnCount()
		props["lineGap"] = widget.LineGap()
		props["columnGap"] = widget.ColumnGap()
		props["childrenCount"] = widget.ChildrenCount()
		if widget.IsVirtual() {
			props["virtual"] = true
			props["numItems"] = widget.NumItems()
		}

	case *widgets.GComboBox:
		props["selectedIndex"] = widget.SelectedIndex()
		props["itemCount"] = len(widget.Items())

	case *widgets.GProgressBar:
		props["value"] = widget.Value()
		props["max"] = widget.Max()

	case *widgets.GSlider:
		props["value"] = widget.Value()
		props["max"] = widget.Max()

	case *widgets.GMovieClip:
		props["playing"] = widget.Playing()
		props["frame"] = widget.Frame()

	case *widgets.GTextInput:
		props["text"] = widget.Text()
		props["maxLength"] = widget.MaxLength()
		props["editable"] = widget.Editable()

	case *core.GComponent:
		if scrollPane := widget.ScrollPane(); scrollPane != nil {
			props["scrollable"] = true
			props["scrollX"] = scrollPane.PosX()
			props["scrollY"] = scrollPane.PosY()
		}
	}

	return props
}

// GetObjectType 获取对象类型名称
func GetObjectType(obj *core.GObject) string {
	if obj == nil {
		return "Unknown"
	}

	switch obj.Data().(type) {
	case *core.GComponent:
		return "GComponent"
	case *widgets.GList:
		return "GList"
	case *widgets.GButton:
		return "GButton"
	case *widgets.GTextField:
		return "GTextField"
	case *widgets.GImage:
		return "GImage"
	case *widgets.GGraph:
		return "GGraph"
	case *widgets.GLoader:
		return "GLoader"
	case *widgets.GComboBox:
		return "GComboBox"
	case *widgets.GTree:
		return "GTree"
	case *widgets.GMovieClip:
		return "GMovieClip"
	case *widgets.GProgressBar:
		return "GProgressBar"
	case *widgets.GSlider:
		return "GSlider"
	case *widgets.GScrollBar:
		return "GScrollBar"
	case *widgets.GRichTextField:
		return "GRichTextField"
	case *widgets.GTextInput:
		return "GTextInput"
	case *widgets.GGroup:
		return "GGroup"
	default:
		return fmt.Sprintf("%T", obj.Data())
	}
}
