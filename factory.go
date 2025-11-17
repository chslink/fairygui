package fairygui

import (
	"context"
	"fmt"
	"strings"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/builder"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
)

// ============================================================================
// ComponentFactory - 简化的组件工厂
// ============================================================================

// ComponentFactory 是简化的组件工厂，封装了 pkg/fgui/builder.Factory 的功能。
type ComponentFactory struct {
	factory *builder.Factory
	loader  *FileLoader
}

// NewComponentFactory 创建一个新的组件工厂。
//
// loader 参数指定资源加载器，用于加载包和纹理。
// atlasManager 参数指定图集管理器，如果为 nil 则使用默认的 render.NewAtlasManager()。
//
// 示例：
//
//	loader := fairygui.NewFileLoader("./assets")
//	factory := fairygui.NewComponentFactory(loader, nil)
func NewComponentFactory(loader *FileLoader, atlasManager builder.AtlasResolver) *ComponentFactory {
	if atlasManager == nil {
		atlasManager = render.NewAtlasManager(loader.loader)
	}

	builderFactory := builder.NewFactoryWithLoader(atlasManager, loader.loader)

	return &ComponentFactory{
		factory: builderFactory,
		loader:  loader,
	}
}

// RegisterPackage 注册包到工厂，使其可以被 CreateObjectFromURL 使用。
func (f *ComponentFactory) RegisterPackage(pkg Package) {
	if wrapper, ok := pkg.(*PackageWrapper); ok {
		f.factory.RegisterPackage(wrapper.pkg)
	}
}

// GetPackage 获取已注册的包。
func (f *ComponentFactory) GetPackage(name string) Package {
	// 尝试从 FileLoader 获取已加载的包
	if f.loader != nil {
		pkg := f.loader.GetPackage(name)
		if pkg != nil {
			return pkg
		}
	}
	return nil
}

// CreateComponent 从包中创建组件，返回底层的 GObject。
//
// pkg 参数指定包，itemName 参数指定项的名称或 ID。
//
// 示例：
//
//	obj, err := factory.CreateComponent(pkg, "MainPanel")
func (f *ComponentFactory) CreateComponent(pkg Package, itemName string) (*core.GObject, error) {
	wrapper, ok := pkg.(*PackageWrapper)
	if !ok {
		return nil, fmt.Errorf("invalid package type")
	}

	// 查找包项
	item := f.findItem(wrapper.pkg, itemName)
	if item == nil {
		return nil, fmt.Errorf("item not found: %s", itemName)
	}

	// 构建组件
	comp, err := f.factory.BuildComponent(context.Background(), wrapper.pkg, item)
	if err != nil {
		return nil, fmt.Errorf("failed to build component: %w", err)
	}

	return comp.GObject, nil
}

// CreateObject 从包中创建对象（与 CreateComponent 相同）。
func (f *ComponentFactory) CreateObject(pkg Package, itemName string) (*core.GObject, error) {
	return f.CreateComponent(pkg, itemName)
}

// CreateObjectFromURL 从 URL 创建对象。
//
// URL 格式: ui://packageName/itemName
//
// 示例：
//
//	obj, err := factory.CreateObjectFromURL("ui://Main/Button")
func (f *ComponentFactory) CreateObjectFromURL(url string) (*core.GObject, error) {
	if !strings.HasPrefix(url, "ui://") {
		return nil, fmt.Errorf("invalid URL format, expected ui://packageName/itemName")
	}

	// 解析 URL: ui://packageName/itemName
	parts := strings.SplitN(url[5:], "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid URL format, expected ui://packageName/itemName")
	}

	packageName := parts[0]
	itemName := parts[1]

	// 获取包
	pkg := f.GetPackage(packageName)
	if pkg == nil {
		// 尝试加载包
		if f.loader != nil {
			loadedPkg, err := f.loader.LoadPackage(packageName)
			if err != nil {
				return nil, fmt.Errorf("failed to load package %s: %w", packageName, err)
			}
			pkg = loadedPkg
			// 注册包
			f.RegisterPackage(pkg)
		} else {
			return nil, fmt.Errorf("package not found: %s", packageName)
		}
	}

	// 创建对象
	return f.CreateObject(pkg, itemName)
}

// findItem 查找包项，支持通过名称或 ID 查找。
func (f *ComponentFactory) findItem(pkg *assets.Package, nameOrID string) *assets.PackageItem {
	// 尝试通过名称查找
	if item := pkg.ItemByName(nameOrID); item != nil {
		return item
	}

	// 尝试通过 ID 查找
	if item := pkg.ItemByID(nameOrID); item != nil {
		return item
	}

	return nil
}

// RawFactory 返回底层的 builder.Factory 对象。
//
// 仅在需要访问底层 API 时使用。
func (f *ComponentFactory) RawFactory() *builder.Factory {
	return f.factory
}
