package fairygui

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// ============================================================================
// FileLoader - 简化的资源加载器
// ============================================================================

// FileLoader 是简化的文件资源加载器，封装了 pkg/fgui/assets 的功能。
type FileLoader struct {
	root   string
	loader *assets.FileLoader

	// 已加载的包缓存
	mu       sync.RWMutex
	packages map[string]*PackageWrapper
}

// NewFileLoader 创建一个新的文件加载器。
//
// root 参数指定资源根目录，例如 "./assets"。
//
// 示例：
//
//	loader := fairygui.NewFileLoader("./assets")
//	pkg, err := loader.LoadPackage("Main")
func NewFileLoader(root string) *FileLoader {
	return &FileLoader{
		root:     root,
		loader:   assets.NewFileLoader(root),
		packages: make(map[string]*PackageWrapper),
	}
}

// LoadPackage 加载一个 UI 包。
//
// 包名应该是不带扩展名的文件名，例如 "Main" 对应 "Main.fui" 文件。
// 该方法会自动处理依赖包的加载。
//
// 返回的 Package 对象可用于创建 UI 对象。
func (l *FileLoader) LoadPackage(name string) (*PackageWrapper, error) {
	// 检查是否已加载
	l.mu.RLock()
	if pkg, ok := l.packages[name]; ok {
		l.mu.RUnlock()
		return pkg, nil
	}
	l.mu.RUnlock()

	// 加载包文件
	packagePath := filepath.Join(l.root, name+".fui")
	data, err := l.loader.LoadOne(context.Background(), packagePath, assets.ResourceBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to load package %s: %w", name, err)
	}

	// 解析包
	pkg, err := assets.ParsePackage(data, name)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package %s: %w", name, err)
	}

	// 注意：ParsePackage 已经加载了所有资源项

	// 封装包
	wrapper := &PackageWrapper{
		pkg:    pkg,
		loader: l,
	}

	// 缓存包
	l.mu.Lock()
	l.packages[name] = wrapper
	l.mu.Unlock()

	// 自动加载依赖
	for _, dep := range pkg.Dependencies {
		if _, err := l.LoadPackage(dep.Name); err != nil {
			return nil, fmt.Errorf("failed to load dependency %s: %w", dep.Name, err)
		}
	}

	return wrapper, nil
}

// GetPackage 获取已加载的包。
//
// 如果包未加载，返回 nil。
func (l *FileLoader) GetPackage(name string) *PackageWrapper {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.packages[name]
}

// LoadTexture 加载纹理图片。
//
// url 参数可以是相对路径（相对于 root）或绝对路径。
func (l *FileLoader) LoadTexture(url string) (*ebiten.Image, error) {
	path := url
	if !filepath.IsAbs(path) {
		path = filepath.Join(l.root, path)
	}

	img, _, err := ebitenutil.NewImageFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load texture %s: %w", url, err)
	}

	return img, nil
}

// LoadAudio 加载音频数据。
func (l *FileLoader) LoadAudio(url string) ([]byte, error) {
	path := url
	if !filepath.IsAbs(path) {
		path = filepath.Join(l.root, path)
	}

	data, err := l.loader.LoadOne(context.Background(), path, assets.ResourceSound)
	if err != nil {
		return nil, fmt.Errorf("failed to load audio %s: %w", url, err)
	}

	return data, nil
}

// LoadFont 加载字体。
//
// 注意：字体加载功能尚未完全实现。
func (l *FileLoader) LoadFont(url string) (Font, error) {
	return nil, fmt.Errorf("font loading not implemented yet")
}

// Exists 检查资源是否存在。
func (l *FileLoader) Exists(url string) bool {
	path := url
	if !filepath.IsAbs(path) {
		path = filepath.Join(l.root, path)
	}

	data, err := l.loader.LoadOne(context.Background(), path, assets.ResourceBinary)
	return err == nil && len(data) > 0
}

// ============================================================================
// PackageWrapper - 包装 assets.Package
// ============================================================================

// PackageWrapper 封装了 pkg/fgui/assets.Package，提供简化的 API。
type PackageWrapper struct {
	pkg    *assets.Package
	loader *FileLoader
}

// ID 返回包的唯一标识符。
func (p *PackageWrapper) ID() string {
	return p.pkg.ID
}

// Name 返回包的名称。
func (p *PackageWrapper) Name() string {
	return p.pkg.Name
}

// GetItem 根据 ID 获取包项。
func (p *PackageWrapper) GetItem(id string) PackageItem {
	// 遍历 Items 查找匹配的 ID
	for _, item := range p.pkg.Items {
		if item.ID == id {
			return &PackageItemWrapper{item: item, pkg: p}
		}
	}
	return nil
}

// GetItemByName 根据名称获取包项。
func (p *PackageWrapper) GetItemByName(name string) PackageItem {
	// 遍历 Items 查找匹配的 Name
	for _, item := range p.pkg.Items {
		if item.Name == name {
			return &PackageItemWrapper{item: item, pkg: p}
		}
	}
	return nil
}

// Items 返回所有包项。
func (p *PackageWrapper) Items() []PackageItem {
	items := p.pkg.Items
	wrappers := make([]PackageItem, len(items))
	for i, item := range items {
		wrappers[i] = &PackageItemWrapper{item: item, pkg: p}
	}
	return wrappers
}

// Dependencies 返回依赖的包列表。
func (p *PackageWrapper) Dependencies() []string {
	deps := make([]string, len(p.pkg.Dependencies))
	for i, dep := range p.pkg.Dependencies {
		deps[i] = dep.Name
	}
	return deps
}

// RawPackage 返回底层的 assets.Package 对象。
//
// 仅在需要访问底层 API 时使用。
func (p *PackageWrapper) RawPackage() *assets.Package {
	return p.pkg
}

// ============================================================================
// PackageItemWrapper - 包装 assets.PackageItem
// ============================================================================

// PackageItemWrapper 封装了 pkg/fgui/assets.PackageItem。
type PackageItemWrapper struct {
	item *assets.PackageItem
	pkg  *PackageWrapper
}

// ID 返回项的唯一标识符。
func (i *PackageItemWrapper) ID() string {
	return i.item.ID
}

// Name 返回项的名称。
func (i *PackageItemWrapper) Name() string {
	return i.item.Name
}

// Type 返回项的类型。
func (i *PackageItemWrapper) Type() ResourceType {
	return convertPackageItemType(i.item.Type)
}

// Data 返回项的数据。
func (i *PackageItemWrapper) Data() interface{} {
	return i.item.RawData
}

// RawItem 返回底层的 assets.PackageItem 对象。
func (i *PackageItemWrapper) RawItem() *assets.PackageItem {
	return i.item
}

// Package 返回该项所属的包。
func (i *PackageItemWrapper) Package() *PackageWrapper {
	return i.pkg
}

// ============================================================================
// 辅助函数
// ============================================================================

// convertPackageItemType 将 assets.PackageItemType 转换为 ResourceType。
func convertPackageItemType(itemType assets.PackageItemType) ResourceType {
	switch itemType {
	case assets.PackageItemTypeComponent:
		return ResourceTypeComponent
	case assets.PackageItemTypeImage, assets.PackageItemTypeMovieClip:
		return ResourceTypeImage
	case assets.PackageItemTypeSound:
		return ResourceTypeSound
	case assets.PackageItemTypeFont:
		return ResourceTypeFont
	case assets.PackageItemTypeAtlas:
		return ResourceTypeAtlas
	default:
		return ResourceTypeUnknown
	}
}
