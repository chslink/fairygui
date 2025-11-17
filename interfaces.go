// Package fairygui 提供 FairyGUI 的 Go 实现，基于 Ebiten 游戏引擎。
//
// V2 版本采用接口驱动设计，提供简洁的 API 和高性能的渲染。
package fairygui

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// 核心显示对象接口
// ============================================================================

// Positionable 表示可以设置位置的对象。
type Positionable interface {
	// Position 返回对象的位置（相对于父对象）。
	Position() (x, y float64)

	// SetPosition 设置对象的位置（相对于父对象）。
	SetPosition(x, y float64)

	// GlobalPosition 返回对象的全局位置（相对于舞台）。
	GlobalPosition() (x, y float64)
}

// Sizable 表示可以设置尺寸的对象。
type Sizable interface {
	// Size 返回对象的尺寸。
	Size() (width, height float64)

	// SetSize 设置对象的尺寸。
	SetSize(width, height float64)
}

// Transformable 表示可以进行变换的对象（缩放、旋转、倾斜）。
type Transformable interface {
	// Scale 返回对象的缩放比例。
	Scale() (scaleX, scaleY float64)

	// SetScale 设置对象的缩放比例。
	SetScale(scaleX, scaleY float64)

	// Rotation 返回对象的旋转角度（度）。
	Rotation() float64

	// SetRotation 设置对象的旋转角度（度）。
	SetRotation(rotation float64)

	// Skew 返回对象的倾斜角度（度）。
	Skew() (skewX, skewY float64)

	// SetSkew 设置对象的倾斜角度（度）。
	SetSkew(skewX, skewY float64)

	// Pivot 返回对象的锚点（0-1 归一化坐标）。
	Pivot() (pivotX, pivotY float64)

	// SetPivot 设置对象的锚点（0-1 归一化坐标）。
	SetPivot(pivotX, pivotY float64)
}

// Visible 表示可以控制可见性的对象。
type Visible interface {
	// Visible 返回对象是否可见。
	Visible() bool

	// SetVisible 设置对象是否可见。
	SetVisible(visible bool)

	// Alpha 返回对象的透明度（0-1）。
	Alpha() float64

	// SetAlpha 设置对象的透明度（0-1）。
	SetAlpha(alpha float64)
}

// Drawable 表示可以渲染的对象。
type Drawable interface {
	// Draw 在屏幕上绘制对象。
	Draw(screen *ebiten.Image)
}

// Updatable 表示可以更新的对象（每帧调用）。
type Updatable interface {
	// Update 更新对象状态。
	// delta 是距离上一帧的时间间隔。
	Update(delta time.Duration) error
}

// Hierarchical 表示具有层级关系的对象。
type Hierarchical interface {
	// Parent 返回父对象。
	Parent() DisplayObject

	// Children 返回子对象列表（只读）。
	Children() []DisplayObject

	// AddChild 添加子对象。
	AddChild(child DisplayObject)

	// AddChildAt 在指定索引位置添加子对象。
	AddChildAt(child DisplayObject, index int)

	// RemoveChild 移除子对象。
	RemoveChild(child DisplayObject)

	// RemoveChildAt 移除指定索引位置的子对象。
	RemoveChildAt(index int) DisplayObject

	// ChildCount 返回子对象数量。
	ChildCount() int

	// GetChildAt 返回指定索引位置的子对象。
	GetChildAt(index int) DisplayObject

	// GetChildByName 返回指定名称的子对象。
	GetChildByName(name string) DisplayObject
}

// DisplayObject 是所有可显示对象的基础接口。
//
// 它组合了多个小接口，提供完整的显示对象功能：
//   - 位置和尺寸 (Positionable, Sizable)
//   - 变换 (Transformable)
//   - 可见性 (Visible)
//   - 层级关系 (Hierarchical)
//   - 渲染 (Drawable)
type DisplayObject interface {
	Positionable
	Sizable
	Transformable
	Visible
	Hierarchical
	Drawable
	EventDispatcher

	// ID 返回对象的唯一标识符。
	ID() string

	// Name 返回对象的名称。
	Name() string

	// SetName 设置对象的名称。
	SetName(name string)

	// Data 返回对象的自定义数据。
	Data() interface{}

	// SetData 设置对象的自定义数据。
	SetData(data interface{})

	// Touchable 返回对象是否可以接收触摸事件。
	Touchable() bool

	// SetTouchable 设置对象是否可以接收触摸事件。
	SetTouchable(touchable bool)

	// DispatchEvent 分发事件（支持冒泡）。
	DispatchEvent(event Event)

	// Dispose 释放对象资源。
	Dispose()
}

// ============================================================================
// 事件系统接口
// ============================================================================

// EventHandler 是事件处理函数类型。
type EventHandler func(event Event)

// EventDispatcher 表示可以分发事件的对象。
type EventDispatcher interface {
	// On 注册事件处理器。
	// 返回一个取消函数，调用后可以移除该处理器。
	On(eventType string, handler EventHandler) (cancel func())

	// Off 移除事件处理器。
	Off(eventType string, handler EventHandler)

	// Once 注册一次性事件处理器（触发后自动移除）。
	Once(eventType string, handler EventHandler)

	// Emit 触发事件。
	Emit(event Event)

	// HasListener 返回是否有指定类型的事件监听器。
	HasListener(eventType string) bool
}

// Event 是所有事件的基础接口。
type Event interface {
	// Type 返回事件类型。
	Type() string

	// Target 返回事件目标对象。
	Target() interface{}

	// CurrentTarget 返回当前处理事件的对象（冒泡过程中）。
	CurrentTarget() interface{}

	// StopPropagation 停止事件传播（冒泡）。
	StopPropagation()

	// IsPropagationStopped 返回事件传播是否已停止。
	IsPropagationStopped() bool

	// IsDefaultPrevented 返回是否已阻止默认行为。
	IsDefaultPrevented() bool

	// PreventDefault 阻止事件的默认行为。
	PreventDefault()
}

// Interactive 表示可交互的对象（支持鼠标/触摸事件）。
type Interactive interface {
	EventDispatcher

	// OnClick 注册点击事件处理器（便捷方法）。
	OnClick(handler func())

	// OnMouseOver 注册鼠标移入事件处理器。
	OnMouseOver(handler func())

	// OnMouseOut 注册鼠标移出事件处理器。
	OnMouseOut(handler func())

	// OnMouseDown 注册鼠标按下事件处理器。
	OnMouseDown(handler func())

	// OnMouseUp 注册鼠标释放事件处理器。
	OnMouseUp(handler func())
}

// ============================================================================
// 渲染系统接口
// ============================================================================

// Renderer 是渲染器接口，负责将显示对象树渲染到屏幕。
type Renderer interface {
	// Draw 渲染显示对象树到屏幕。
	Draw(screen *ebiten.Image, root DisplayObject)

	// DrawText 渲染文本。
	DrawText(screen *ebiten.Image, text string, x, y float64, style TextStyle)

	// DrawTexture 渲染纹理。
	DrawTexture(screen *ebiten.Image, texture *ebiten.Image, options DrawOptions)

	// DrawShape 渲染形状。
	DrawShape(screen *ebiten.Image, shape Shape, options DrawOptions)
}

// TextStyle 定义文本样式。
type TextStyle struct {
	Font      Font           // 字体
	Size      float64        // 字号
	Color     uint32         // 颜色 (RGBA)
	Bold      bool           // 粗体
	Italic    bool           // 斜体
	Underline bool           // 下划线
	Align     TextAlign      // 对齐方式
	Stroke    *StrokeStyle   // 描边样式
	Shadow    *ShadowStyle   // 阴影样式
}

// StrokeStyle 定义描边样式。
type StrokeStyle struct {
	Color     uint32  // 描边颜色
	Thickness float64 // 描边厚度
}

// ShadowStyle 定义阴影样式。
type ShadowStyle struct {
	Color   uint32  // 阴影颜色
	OffsetX float64 // X 偏移
	OffsetY float64 // Y 偏移
	Blur    float64 // 模糊半径
}

// TextAlign 定义文本对齐方式。
type TextAlign int

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
)

// DrawOptions 定义绘制选项。
type DrawOptions struct {
	X           float64       // X 坐标
	Y           float64       // Y 坐标
	Width       float64       // 宽度
	Height      float64       // 高度
	ScaleX      float64       // X 缩放
	ScaleY      float64       // Y 缩放
	Rotation    float64       // 旋转角度（弧度）
	Alpha       float64       // 透明度
	Color       *uint32       // 颜色叠加
	BlendMode   BlendMode     // 混合模式
	ClipRect    *Rect         // 裁剪矩形
	NineSlice   *NineSlice    // 九宫格
	Tiling      bool          // 是否平铺
}

// BlendMode 定义混合模式。
type BlendMode int

const (
	BlendModeNormal BlendMode = iota
	BlendModeAdd
	BlendModeMultiply
	BlendModeScreen
)

// Shape 定义可绘制的形状。
type Shape interface {
	// Type 返回形状类型。
	Type() ShapeType
}

// ShapeType 定义形状类型。
type ShapeType int

const (
	ShapeTypeRect ShapeType = iota
	ShapeTypeCircle
	ShapeTypeEllipse
	ShapeTypePolygon
)

// Rect 定义矩形。
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// NineSlice 定义九宫格。
type NineSlice struct {
	Left   float64 // 左边距
	Right  float64 // 右边距
	Top    float64 // 上边距
	Bottom float64 // 下边距
}

// Font 定义字体接口。
type Font interface {
	// Name 返回字体名称。
	Name() string

	// Size 返回字体大小。
	Size() float64

	// Measure 测量文本尺寸。
	Measure(text string) (width, height float64)
}

// ============================================================================
// 资源加载接口
// ============================================================================

// AssetLoader 是资源加载器接口。
type AssetLoader interface {
	// LoadPackage 加载 UI 包。
	LoadPackage(name string) (*Package, error)

	// LoadTexture 加载纹理。
	LoadTexture(url string) (*ebiten.Image, error)

	// LoadAudio 加载音频数据。
	LoadAudio(url string) ([]byte, error)

	// LoadFont 加载字体。
	LoadFont(url string) (Font, error)

	// Exists 检查资源是否存在。
	Exists(url string) bool
}

// Package 表示 UI 包。
type Package interface {
	// ID 返回包的唯一标识符。
	ID() string

	// Name 返回包的名称。
	Name() string

	// GetItem 根据 ID 获取包项。
	GetItem(id string) PackageItem

	// GetItemByName 根据名称获取包项。
	GetItemByName(name string) PackageItem

	// Items 返回所有包项。
	Items() []PackageItem

	// Dependencies 返回依赖的包列表。
	Dependencies() []string
}

// PackageItem 表示包中的资源项。
type PackageItem interface {
	// ID 返回项的唯一标识符。
	ID() string

	// Name 返回项的名称。
	Name() string

	// Type 返回项的类型。
	Type() ResourceType

	// Data 返回项的数据。
	Data() interface{}
}

// ResourceType 定义资源类型。
type ResourceType int

const (
	ResourceTypeUnknown ResourceType = iota
	ResourceTypeComponent
	ResourceTypeImage
	ResourceTypeSound
	ResourceTypeFont
	ResourceTypeAtlas
)

// ============================================================================
// 工厂接口
// ============================================================================

// Factory 是组件工厂接口，用于从包中创建 UI 对象。
type Factory interface {
	// CreateComponent 从包中创建组件。
	CreateComponent(pkg Package, itemName string) (DisplayObject, error)

	// CreateObject 从包中创建对象。
	CreateObject(pkg Package, itemName string) (DisplayObject, error)

	// CreateObjectFromURL 从 URL 创建对象（格式: ui://packageName/itemName）。
	CreateObjectFromURL(url string) (DisplayObject, error)

	// RegisterPackage 注册包到工厂。
	RegisterPackage(pkg Package)

	// GetPackage 获取已注册的包。
	GetPackage(name string) Package
}

// ============================================================================
// 组件接口
// ============================================================================

// Component 是容器组件接口，可以包含其他对象。
type Component interface {
	DisplayObject
	Interactive

	// Controllers 返回组件的所有控制器。
	Controllers() []Controller

	// GetController 根据名称获取控制器。
	GetController(name string) Controller

	// AddController 添加控制器。
	AddController(controller Controller)
}

// Controller 是控制器接口，用于管理组件状态。
type Controller interface {
	// Name 返回控制器名称。
	Name() string

	// SelectedIndex 返回当前选中的索引。
	SelectedIndex() int

	// SetSelectedIndex 设置选中的索引。
	SetSelectedIndex(index int)

	// SelectedPage 返回当前选中的页面名称。
	SelectedPage() string

	// SetSelectedPage 设置选中的页面名称。
	SetSelectedPage(name string)

	// PageCount 返回页面数量。
	PageCount() int

	// OnChanged 注册状态改变事件处理器。
	OnChanged(handler func())
}

// ============================================================================
// 动画接口
// ============================================================================

// Tween 是补间动画接口。
type Tween interface {
	// To 设置目标值。
	To(target interface{}, duration time.Duration, properties map[string]interface{}) Tween

	// From 设置起始值。
	From(target interface{}, duration time.Duration, properties map[string]interface{}) Tween

	// Ease 设置缓动函数。
	Ease(easeFunc EaseFunc) Tween

	// OnComplete 设置完成回调。
	OnComplete(callback func()) Tween

	// OnUpdate 设置更新回调。
	OnUpdate(callback func()) Tween

	// Start 开始动画。
	Start()

	// Stop 停止动画。
	Stop()

	// Pause 暂停动画。
	Pause()

	// Resume 恢复动画。
	Resume()
}

// EaseFunc 是缓动函数类型。
type EaseFunc func(t float64) float64

// Transition 是过渡动画接口。
type Transition interface {
	// Play 播放过渡动画。
	Play(onComplete func())

	// Stop 停止过渡动画。
	Stop()

	// Pause 暂停过渡动画。
	Pause()

	// Resume 恢复过渡动画。
	Resume()

	// IsPlaying 返回是否正在播放。
	IsPlaying() bool

	// SetDuration 设置播放时长。
	SetDuration(duration time.Duration)
}

// ============================================================================
// 输入接口
// ============================================================================

// InputManager 是输入管理器接口。
type InputManager interface {
	// Update 更新输入状态（每帧调用）。
	Update()

	// MousePosition 返回鼠标位置。
	MousePosition() (x, y int)

	// IsMouseButtonPressed 返回鼠标按钮是否按下。
	IsMouseButtonPressed(button MouseButton) bool

	// IsKeyPressed 返回键盘按键是否按下。
	IsKeyPressed(key Key) bool

	// TouchIDs 返回当前所有触摸点的 ID。
	TouchIDs() []int

	// TouchPosition 返回指定触摸点的位置。
	TouchPosition(id int) (x, y int)
}

// MouseButton 定义鼠标按钮。
type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonRight
	MouseButtonMiddle
)

// Key 是键盘按键别名（使用 Ebiten 的定义）。
type Key = ebiten.Key

// ============================================================================
// 根对象接口
// ============================================================================

// Root 是根对象接口，管理整个 UI 树。
type Root interface {
	Component
	Updatable

	// Resize 调整舞台尺寸。
	Resize(width, height int)

	// SetRenderer 设置渲染器。
	SetRenderer(renderer Renderer)

	// Renderer 返回当前渲染器。
	Renderer() Renderer

	// SetInputManager 设置输入管理器。
	SetInputManager(input InputManager)

	// InputManager 返回输入管理器。
	InputManager() InputManager
}
