// Package types 提供内部使用的类型定义
//
// 这个包定义了 internal 包之间共享的基本类型，避免循环依赖
package types

import (
	"image"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// DisplayObject 表示可显示的对象（在 interfaces.go 中有完整定义）
// 这里只定义内部使用所需的最小接口
type DisplayObject interface {
	Position() (x, y float64)
	SetPosition(x, y float64)
	Size() (width, height float64)
	SetSize(width, height float64)
	Visible() bool
	SetVisible(visible bool)
	Alpha() float64
	SetAlpha(alpha float64)
	Scale() (scaleX, scaleY float64)
	SetScale(scaleX, scaleY float64)
	Rotation() float64
	SetRotation(rotation float64)
	Parent() DisplayObject
	Children() []DisplayObject
	AddChild(child DisplayObject)
	RemoveChild(child DisplayObject)
	Draw(screen *ebiten.Image)
	Bounds() image.Rectangle
	GlobalBounds() image.Rectangle
	ID() string
	Name() string
	SetName(name string)
	Data() interface{}
	SetData(data interface{})
	Touchable() bool
	SetTouchable(touchable bool)
	Dispose()
}

// EventHandler 定义事件处理器类型
type EventHandler func(event Event)

// Event 定义事件接口
type Event interface {
	Type() string
	Target() DisplayObject
	CurrentTarget() DisplayObject
	Data() interface{}
	StopPropagation()
	PreventDefault()
	IsPropagationStopped() bool
	IsDefaultPrevented() bool
}

// EventDispatcher 定义事件分发器接口
type EventDispatcher interface {
	On(eventType string, handler EventHandler) (cancel func())
	Off(eventType string, handler EventHandler)
	Once(eventType string, handler EventHandler)
	Emit(event Event)
	HasListener(eventType string) bool
	RemoveAllListeners()
	RemoveListenersForType(eventType string)
}

// Renderer 是渲染器接口
type Renderer interface {
	Draw(screen *ebiten.Image, root DisplayObject)
	DrawText(screen *ebiten.Image, text string, x, y float64, style TextStyle)
	DrawTexture(screen *ebiten.Image, texture *ebiten.Image, dst Rect, opts DrawOptions)
	DrawRect(screen *ebiten.Image, x, y, width, height float64, color Color)
	DrawRoundedRect(screen *ebiten.Image, x, y, width, height, radius float64, color Color)
	DrawCircle(screen *ebiten.Image, x, y, radius float64, color Color)
	DrawLine(screen *ebiten.Image, x1, y1, x2, y2, width float64, color Color)
	PushTransform()
	PopTransform()
	PushClip(rect Rect)
	PopClip()
	SetColorMatrix(matrix *ColorMatrix)
	ResetColorMatrix()
	SetBlendMode(mode BlendMode)
	ResetBlendMode()
}

// Positionable 定义位置接口
type Positionable interface {
	Position() (x, y float64)
	SetPosition(x, y float64)
}

// Sizable 定义尺寸接口
type Sizable interface {
	Size() (width, height float64)
	SetSize(width, height float64)
}

// Transformable 定义变换接口
type Transformable interface {
	Scale() (scaleX, scaleY float64)
	SetScale(scaleX, scaleY float64)
	Rotation() float64
	SetRotation(rotation float64)
	Skew() (skewX, skewY float64)
	SetSkew(skewX, skewY float64)
	Pivot() (pivotX, pivotY float64)
	SetPivot(pivotX, pivotY float64)
}

// Visible 定义可见性接口
type Visible interface {
	Visible() bool
	SetVisible(visible bool)
	Alpha() float64
	SetAlpha(alpha float64)
}

// Hierarchical 定义层级关系接口
type Hierarchical interface {
	Parent() DisplayObject
	Children() []DisplayObject
	AddChild(child DisplayObject)
	AddChildAt(child DisplayObject, index int)
	RemoveChild(child DisplayObject)
	RemoveChildAt(index int) DisplayObject
	ChildCount() int
	GetChildAt(index int) DisplayObject
	GetChildByName(name string) DisplayObject
}

// Drawable 定义渲染接口
type Drawable interface {
	Draw(screen *ebiten.Image)
	Bounds() image.Rectangle
	GlobalBounds() image.Rectangle
}

// Updatable 定义更新接口
type Updatable interface {
	Update(delta time.Duration) error
}

// Interactive 定义交互接口
type Interactive interface {
	Touchable() bool
	SetTouchable(touchable bool)
	OnClick(handler func())
	OnMouseOver(handler func())
	OnMouseOut(handler func())
	OnMouseDown(handler func())
	OnMouseUp(handler func())
}

// Component 定义容器组件接口
type Component interface {
	DisplayObject
	Interactive
	Controllers() []Controller
	GetController(name string) Controller
	AddController(controller Controller)
}

// Controller 定义控制器接口
type Controller interface {
	Name() string
	SelectedIndex() int
	SetSelectedIndex(index int)
	SelectedPage() string
	SetSelectedPage(name string)
	PageCount() int
	OnChanged(handler func())
}

// Factory 定义组件工厂接口
type Factory interface {
	CreateComponent(pkg Package, itemName string) (DisplayObject, error)
	CreateObject(pkg Package, itemName string) (DisplayObject, error)
	CreateObjectFromURL(url string) (DisplayObject, error)
	RegisterPackage(pkg Package)
	GetPackage(name string) Package
}

// Package 定义资源包接口
type Package interface {
	ID() string
	Name() string
	GetItem(id string) PackageItem
	GetItemByName(name string) PackageItem
	Items() []PackageItem
	Dependencies() []string
}

// PackageItem 定义资源项接口
type PackageItem interface {
	ID() string
	Name() string
	Type() ResourceType
	Data() interface{}
}

// Tween 定义补间动画接口
type Tween interface {
	To(target interface{}, duration time.Duration, properties map[string]interface{}) Tween
	From(target interface{}, duration time.Duration, properties map[string]interface{}) Tween
	Ease(easeFunc EaseFunc) Tween
	OnComplete(callback func()) Tween
	OnUpdate(callback func()) Tween
	Start()
	Stop()
	Pause()
	Resume()
}

// Transition 定义过渡动画接口
type Transition interface {
	Play(onComplete func())
	Stop()
	Pause()
	Resume()
	IsPlaying() bool
	SetDuration(duration time.Duration)
}

// Root 定义根对象接口
type Root interface {
	Component
	Updatable
	Resize(width, height int)
	SetRenderer(renderer Renderer)
	Renderer() Renderer
	SetInputManager(input InputManager)
	InputManager() InputManager
}

// InputManager 定义输入管理器接口
type InputManager interface {
	Update()
	MousePosition() (x, y int)
	IsMouseButtonPressed(button MouseButton) bool
	IsKeyPressed(key Key) bool
	TouchIDs() []int
	TouchPosition(id int) (x, y int)
}

// ResourceType 定义资源类型
type ResourceType int

const (
	ResourceTypeUnknown ResourceType = iota
	ResourceTypeComponent
	ResourceTypeImage
	ResourceTypeSound
	ResourceTypeFont
	ResourceTypeAtlas
)

// BlendMode 定义混合模式
type BlendMode int

const (
	BlendModeNormal BlendMode = iota
	BlendModeAdd
	BlendModeMultiply
	BlendModeScreen
)

// MouseButton 定义鼠标按钮
type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonRight
	MouseButtonMiddle
)

// Key 定义键盘按键（使用 Ebiten）
type Key = ebiten.Key

// EaseFunc 定义缓动函数
type EaseFunc func(t float64) float64

// ============================================================================
// 数据结构定义
// ============================================================================

// Rect 定义矩形
type Rect struct {
	X, Y, Width, Height float64
}

// Color 定义颜色
type Color struct {
	R, G, B, A float64
}

// ColorMatrix 定义颜色矩阵
type ColorMatrix struct {
	M [4][5]float64 // 4x5 颜色矩阵
}

// DrawOptions 定义绘制选项
type DrawOptions struct {
	X           float64
	Y           float64
	Width       float64
	Height      float64
	ScaleX      float64
	ScaleY      float64
	Rotation    float64
	Alpha       float64
	Color       *Color
	BlendMode   BlendMode
	ClipRect    *Rect
	NineSlice   *NineSlice
	Tiling      bool
	SourceRect  *image.Rectangle
}

// NineSlice 定义九宫格
type NineSlice struct {
	Left   float64
	Right  float64
	Top    float64
	Bottom float64
}

// TextStyle 定义文本样式
type TextStyle struct {
	Font        Font
	Size        float64
	Color       Color
	Bold        bool
	Italic      bool
	Underline   bool
	Align       TextAlign
	Stroke      *StrokeStyle
	Shadow      *ShadowStyle
	LineSpacing float64
}

// StrokeStyle 定义描边样式
type StrokeStyle struct {
	Color     Color
	Thickness float64
}

// ShadowStyle 定义阴影样式
type ShadowStyle struct {
	Color   Color
	OffsetX float64
	OffsetY float64
	Blur    float64
}

// TextAlign 定义文本对齐
type TextAlign int

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
)

// Font 定义字体接口
type Font interface {
	Name() string
	Size() float64
	Measure(text string) (width, height float64)
}

// Transform 定义变换接口
type Transform interface {
	Matrix() Matrix
	Translate(x, y float64)
	Scale(x, y float64)
	Rotate(angle float64)
	Concat(other Transform)
}

// Matrix 定义 3x3 变换矩阵
type Matrix struct {
	A, B, C, D, Tx, Ty float64
}

// Shape 定义形状接口
type Shape interface {
	Type() ShapeType
}

// ShapeType 定义形状类型
type ShapeType int

const (
	ShapeTypeRect ShapeType = iota
	ShapeTypeCircle
	ShapeTypeEllipse
	ShapeTypePolygon
)
