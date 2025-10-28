package builder

import (
	"context"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

// BenchmarkBuildComponent_WithCache 测试带缓存的构建性能
func BenchmarkBuildComponent_WithCache(b *testing.B) {
	// 创建测试数据
	pkg := &assets.Package{
		ID:   "test_pkg",
		Name: "TestPackage",
	}

	item := &assets.PackageItem{
		Type: assets.PackageItemTypeComponent,
		ID:   "test_component",
		Name: "TestComponent",
		Component: &assets.ComponentData{
			SourceWidth:  100,
			SourceHeight: 100,
			InitWidth:    100,
			InitHeight:   100,
			Children:     make([]assets.ComponentChild, 10), // 10个子对象
		},
	}

	factory := NewFactory(nil, nil)
	ctx := context.Background()

	// 预热缓存
	_, _ = factory.BuildComponent(ctx, pkg, item)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = factory.BuildComponent(ctx, pkg, item)
	}
}

// BenchmarkEnsurePackageReady 测试包准备性能
func BenchmarkEnsurePackageReady(b *testing.B) {
	pkg := &assets.Package{
		ID:   "test_pkg",
		Name: "TestPackage",
	}

	factory := NewFactory(nil, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = factory.ensurePackageReady(ctx, pkg)
	}
}

// BenchmarkBuildComponent_MultiplePackages 测试多包构建性能
func BenchmarkBuildComponent_MultiplePackages(b *testing.B) {
	// 创建多个包（模拟跨包引用）
	packages := make([]*assets.Package, 5)
	items := make([]*assets.PackageItem, 5)

	for i := 0; i < 5; i++ {
		packages[i] = &assets.Package{
			ID:   "pkg_" + string(rune('0'+i)),
			Name: "Package" + string(rune('0'+i)),
		}

		items[i] = &assets.PackageItem{
			Type: assets.PackageItemTypeComponent,
			ID:   "comp_" + string(rune('0'+i)),
			Name: "Component" + string(rune('0'+i)),
			Component: &assets.ComponentData{
				SourceWidth:  100,
				SourceHeight: 100,
				InitWidth:    100,
				InitHeight:   100,
				Children:     make([]assets.ComponentChild, 5),
			},
		}
	}

	factory := NewFactory(nil, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟跨包构建（缓存会避免重复加载）
		for j := 0; j < 5; j++ {
			_, _ = factory.BuildComponent(ctx, packages[j], items[j])
		}
	}
}
