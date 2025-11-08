package fgui

import (
	"context"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/audio"
	"github.com/chslink/fairygui/pkg/fgui/builder"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// Public aliases to mirror the TypeScript API surface.
type (
	// Core types
	Stage          = laya.Stage
	Scheduler      = laya.Scheduler
	MouseState     = laya.MouseState
	InputState     = laya.InputState
	TouchInput     = laya.TouchInput
	TouchPhase     = laya.TouchPhase
	MouseButtons   = laya.MouseButtons
	KeyModifiers   = laya.KeyModifiers
	KeyCode        = laya.KeyCode
	KeyboardEvent  = laya.KeyboardEvent
	PointerEvent   = laya.PointerEvent
	EventType      = laya.EventType
	GRoot          = core.GRoot
	GComponent     = core.GComponent
	GObject        = core.GObject
	PopupDirection = core.PopupDirection

	// Asset types
	Package        = assets.Package
	PackageItem    = assets.PackageItem
	Loader         = assets.Loader
	FileLoader     = assets.FileLoader
	ResourceType   = assets.ResourceType
)

const (
	// PopupDirectionAuto positions the popup below the target when possible.
	PopupDirectionAuto = core.PopupDirectionAuto
	// PopupDirectionUp positions the popup above the target.
	PopupDirectionUp = core.PopupDirectionUp
	// PopupDirectionDown positions the popup below the target.
	PopupDirectionDown = core.PopupDirectionDown

	// Resource type constants
	ResourceBinary = assets.ResourceBinary
	ResourceImage  = assets.ResourceImage
	ResourceSound  = assets.ResourceSound
)

// NewStage constructs a compat stage suitable for attaching to the root.
func NewStage(width, height int) *Stage {
	return laya.NewStage(width, height)
}

// NewGObject creates a bare UI object backed by a compat sprite.
func NewGObject() *core.GObject {
	return core.NewGObject()
}

// NewGComponent constructs an empty component container.
func NewGComponent() *core.GComponent {
	return core.NewGComponent()
}

// Root returns the singleton GRoot instance (alias of core.Root()).
func Root() *core.GRoot {
	return core.Root()
}

// Instance is an alias to Root for parity with the TypeScript API.
func Instance() *core.GRoot {
	return core.Inst()
}

// AttachStage binds the singleton root to the provided stage.
func AttachStage(stage *Stage) *core.GRoot {
	root := core.Root()
	root.AttachStage(stage)
	return root
}

// CurrentStage returns the compat stage currently attached to the root, if any.
func CurrentStage() *Stage {
	return core.Root().Stage()
}

// Advance ticks the singleton root and underlying stage scheduler.
func Advance(delta time.Duration, mouse MouseState) {
	core.Root().Advance(delta, mouse)
}

// AdvanceInput ticks the singleton root using a full input state payload.
func AdvanceInput(delta time.Duration, input InputState) {
	core.Root().AdvanceInput(delta, input)
}

// CurrentScheduler exposes the stage scheduler for timer integrations.
func CurrentScheduler() *Scheduler {
	return core.Root().Scheduler()
}

// ShowPopup displays the popup using the singleton root.
func ShowPopup(popup, target *core.GObject, dir PopupDirection) {
	core.Root().ShowPopup(popup, target, dir)
}

// HidePopup hides the specified popup via the singleton root.
func HidePopup(popup *core.GObject) {
	core.Root().HidePopup(popup)
}

// HideAllPopups hides all active popups on the singleton root.
func HideAllPopups() {
	core.Root().HideAllPopups()
}

// TogglePopup toggles the popup on the singleton root.
func TogglePopup(popup, target *core.GObject, dir PopupDirection) {
	core.Root().TogglePopup(popup, target, dir)
}

// HasAnyPopup reports whether the singleton root currently has visible popups.
func HasAnyPopup() bool {
	return core.Root().HasAnyPopup()
}

// Resize updates both root and stage dimensions for the singleton root.
func Resize(width, height int) {
	core.Root().Resize(width, height)
}

// ContentScale reports the current content scale level.
func ContentScale() int {
	return core.ContentScaleLevel
}

// ────────────────────────────────────────────────────────────────────────────
// Factory & Builder API
// ────────────────────────────────────────────────────────────────────────────

// Factory builds runtime components from parsed package metadata.
// This is the primary entry point for creating FairyGUI UI from .fui packages.
type Factory = builder.Factory

// AtlasResolver loads textures and sprites for rendering.
// Build-tagged implementations typically live in the render package.
type AtlasResolver = builder.AtlasResolver

// PackageResolver resolves cross-package dependencies by ID or name.
type PackageResolver = builder.PackageResolver

// NewFactory creates a new factory for building UI components.
//
// Parameters:
//   - resolver: Handles texture/sprite loading (can be nil for logic-only builds)
//   - pkgResolver: Resolves cross-package dependencies (can be nil for single-package apps)
//
// Example:
//   factory := fgui.NewFactory(atlasManager, nil)
//   factory.RegisterPackage(pkg)
//   component, err := factory.BuildComponent(ctx, pkg, item)
func NewFactory(resolver AtlasResolver, pkgResolver PackageResolver) *Factory {
	return builder.NewFactory(resolver, pkgResolver)
}

// NewFactoryWithLoader creates a factory with automatic dependency resolution.
// The loader will be used to load dependent packages on-demand.
//
// Parameters:
//   - resolver: Handles texture/sprite loading
//   - loader: Asset loader for automatic dependency resolution
//
// Example:
//   loader := assets.NewFileLoader("./assets")
//   factory := fgui.NewFactoryWithLoader(atlasManager, loader)
func NewFactoryWithLoader(resolver AtlasResolver, loader assets.Loader) *Factory {
	return builder.NewFactoryWithLoader(resolver, loader)
}

// BuildComponent is a convenience wrapper for Factory.BuildComponent.
// Requires a factory to be created first via NewFactory or NewFactoryWithLoader.
func BuildComponent(ctx context.Context, factory *Factory, pkg *assets.Package, item *assets.PackageItem) (*core.GComponent, error) {
	return factory.BuildComponent(ctx, pkg, item)
}

// ────────────────────────────────────────────────────────────────────────────
// Asset Loading API
// ────────────────────────────────────────────────────────────────────────────

// ParsePackage parses a FairyGUI package from raw bytes.
//
// Parameters:
//   - data: Raw .fui file bytes
//   - resKey: Resource key for the package (typically the file path without extension)
//
// Example:
//   data, _ := os.ReadFile("assets/MainMenu.fui")
//   pkg, err := fgui.ParsePackage(data, "assets/MainMenu")
func ParsePackage(data []byte, resKey string) (*assets.Package, error) {
	return assets.ParsePackage(data, resKey)
}

// NewFileLoader creates a loader that reads assets from the filesystem.
//
// Parameters:
//   - root: Root directory containing .fui files
//
// Example:
//   loader := fgui.NewFileLoader("./assets")
func NewFileLoader(root string) *assets.FileLoader {
	return assets.NewFileLoader(root)
}

// GetPackageByName 通过包名获取包
// 对应 TypeScript 版本的 UIPackage.getByName
func GetPackageByName(name string) *assets.Package {
	return assets.GetPackageByName(name)
}

// GetPackageByID 通过包ID获取包
// 对应 TypeScript 版本的 UIPackage.getItemByID
func GetPackageByID(id string) *assets.Package {
	return assets.GetPackageByID(id)
}

// GetItemByURL 通过URL获取资源项
// 对应 TypeScript 版本的 UIPackage.getItemByURL
func GetItemByURL(url string) *assets.PackageItem {
	return assets.GetItemByURL(url)
}

// CreateObject 从包中创建对象
// 对应 TypeScript 版本的 UIPackage.createObject
func CreateObject(pkgName, resName string) *core.GObject {
	pkg := assets.GetPackageByName(pkgName)
	if pkg == nil {
		return nil
	}
	item := pkg.ItemByName(resName)
	if item == nil {
		return nil
	}
	// 使用默认工厂创建
	factory := builder.NewFactory(nil, nil)
	ctx := context.Background()
	comp, err := factory.BuildComponent(ctx, pkg, item)
	if err != nil {
		return nil
	}
	return comp.GObject
}

// ────────────────────────────────────────────────────────────────────────────
// UIConfig API
// ────────────────────────────────────────────────────────────────────────────

// SetDefaultScrollBars 设置全局默认滚动条资源URL
// 对应 TypeScript 版本的 UIConfig.verticalScrollBar 和 UIConfig.horizontalScrollBar
//
// Parameters:
//   - vertical: 垂直滚动条的资源URL (格式: ui://packageId/itemId)
//   - horizontal: 水平滚动条的资源URL (格式: ui://packageId/itemId)
//
// Example:
//   // 在加载 Basics 包后设置默认滚动条
//   fgui.SetDefaultScrollBars("ui://9leh0eyf/i3s65w", "ui://9leh0eyf/i3s65i")
func SetDefaultScrollBars(vertical, horizontal string) {
	config := core.GetUIConfig()
	if vertical != "" {
		config.VerticalScrollBar = vertical
	}
	if horizontal != "" {
		config.HorizontalScrollBar = horizontal
	}
}

// SetDefaultButtonSound 设置全局默认按钮点击音效
// 对应 TypeScript 版本的 UIConfig.buttonSound
//
// Parameters:
//   - soundURL: 按钮点击音效的资源URL (格式: ui://packageId/itemId)
//
// Example:
//   fgui.SetDefaultButtonSound("ui://Basics/click")
func SetDefaultButtonSound(soundURL string) {
	config := core.GetUIConfig()
	config.ButtonSound = soundURL
}

// SetDefaultPopupMenu 设置全局默认右键菜单资源
// 对应 TypeScript 版本的 UIConfig.popupMenu
//
// Parameters:
//   - menuURL: 右键菜单的资源URL (格式: ui://packageId/itemId)
//
// Example:
//   fgui.SetDefaultPopupMenu("ui://Basics/PopupMenu")
func SetDefaultPopupMenu(menuURL string) {
	config := core.GetUIConfig()
	config.PopupMenu = menuURL
}

// ────────────────────────────────────────────────────────────────────────────
// Audio API
// ────────────────────────────────────────────────────────────────────────────

// GetAudioPlayer 获取音频播放器单例
func GetAudioPlayer() *audio.AudioPlayer {
	return audio.GetInstance()
}

// InitAudio 初始化音频系统
// 必须在游戏开始时调用一次
//
// Parameters:
//   - sampleRate: 采样率，默认48000
func InitAudio(sampleRate int) {
	audio.GetInstance().Init(sampleRate)
}

// RegisterButtonSoundPlayer 注册音频播放器为按钮音效播放器
// 这使得所有按钮点击都会播放配置的音效
func RegisterButtonSoundPlayer() {
	audio.RegisterAsDefaultButtonSoundPlayer()
}

// RegisterAudio 注册音频资源
// 将音频数据注册到音频播放器中，后续可以通过名称播放
//
// Parameters:
//   - name: 音频资源名称
//   - data: 音频字节数据（支持MP3、Wav、Ogg格式）
//
// Example:
//   data, _ := os.ReadFile("click.wav")
//   fgui.RegisterAudio("button_click", data)
func RegisterAudio(name string, data []byte) {
	audio.RegisterAudio(name, data)
}

// SetAudioLoader 设置音频系统的资源加载器
// 用于自动从FUI包中加载音效数据
//
// Parameters:
//   - loader: 资源加载器（通常使用 fgui.NewFileLoader）
//
// Example:
//   loader := fgui.NewFileLoader("./assets")
//   fgui.SetAudioLoader(loader)
func SetAudioLoader(loader assets.Loader) {
	audio.SetLoader(loader)
}
