package builder

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// TestOverflowFromPackage 测试从实际包中加载 overflow 配置
func TestOverflowFromPackage(t *testing.T) {
	// 加载 Basics 包
	loader := assets.NewFileLoader("../../demo/assets")
	ctx := context.Background()

	data, err := loader.LoadOne(ctx, "Basics.fui", assets.ResourceBinary)
	if err != nil {
		t.Skipf("跳过集成测试：无法加载 Basics.fui: %v", err)
		return
	}

	pkg, err := assets.ParsePackage(data, filepath.Clean("demo/assets/Basics"))
	if err != nil {
		t.Fatalf("解析包失败: %v", err)
	}

	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	// 测试 Component1 (overflow="hidden")
	t.Run("Component1_OverflowHidden", func(t *testing.T) {
		item := pkg.ItemByName("Component1")
		if item == nil {
			t.Skip("Component1 未找到，跳过测试")
			return
		}

		comp, err := factory.BuildComponent(ctx, pkg, item)
		if err != nil {
			t.Fatalf("构建组件失败: %v", err)
		}

		// 验证 overflow 设置
		if comp.Overflow() != core.OverflowHidden {
			t.Errorf("Component1 overflow 应该是 Hidden，实际 %v", comp.Overflow())
		}

		// 验证创建了独立 container
		if comp.Container() == comp.DisplayObject() {
			t.Error("Component1 应该有独立的 container")
		}

		// 验证设置了 scrollRect
		if comp.DisplayObject().ScrollRect() == nil {
			t.Error("Component1 应该设置 scrollRect")
		}
	})

	// 测试 Component8 (overflow="scroll" with margin)
	t.Run("Component8_OverflowScrollWithMargin", func(t *testing.T) {
		item := pkg.ItemByName("Component8")
		if item == nil {
			t.Skip("Component8 未找到，跳过测试")
			return
		}

		comp, err := factory.BuildComponent(ctx, pkg, item)
		if err != nil {
			t.Fatalf("构建组件失败: %v", err)
		}

		// 验证 margin 读取
		margin := comp.Margin()
		expected := core.Margin{Top: 30, Bottom: 30, Left: 30, Right: 30}
		if margin != expected {
			t.Errorf("Component8 margin 应该是 %+v，实际 %+v", expected, margin)
		}

		// Component8 使用 scroll，应该有 ScrollPane
		if comp.ScrollPane() == nil {
			t.Error("Component8 应该有 ScrollPane")
		}
	})

	// 测试 Component7 (overflow="scroll" without margin)
	t.Run("Component7_OverflowScroll", func(t *testing.T) {
		item := pkg.ItemByName("Component7")
		if item == nil {
			t.Skip("Component7 未找到，跳过测试")
			return
		}

		comp, err := factory.BuildComponent(ctx, pkg, item)
		if err != nil {
			t.Fatalf("构建组件失败: %v", err)
		}

		// 验证 overflow 是 scroll
		if comp.Overflow() != core.OverflowScroll {
			t.Errorf("Component7 overflow 应该是 Scroll，实际 %v", comp.Overflow())
		}

		// 应该有 ScrollPane
		if comp.ScrollPane() == nil {
			t.Error("Component7 应该有 ScrollPane")
		}

		// Margin 应该是零（XML 中未指定）
		if !comp.Margin().IsZero() {
			t.Errorf("Component7 margin 应该是零，实际 %+v", comp.Margin())
		}
	})
}

// TestOverflowBuildIntegration 测试 BuildComponent 正确应用 overflow
func TestOverflowBuildIntegration(t *testing.T) {
	// 创建测试包
	pkg := &assets.Package{
		ID:   "test_overflow",
		Name: "TestOverflow",
	}

	// 创建带 overflow hidden 的组件数据
	hiddenComponent := &assets.PackageItem{
		ID:   "hidden_comp",
		Name: "HiddenComponent",
		Type: assets.PackageItemTypeComponent,
		Component: &assets.ComponentData{
			InitWidth:  200,
			InitHeight: 150,
			Margin: assets.Margin{
				Top:    10,
				Bottom: 10,
				Left:   5,
				Right:  5,
			},
			Overflow: assets.OverflowTypeHidden,
		},
	}

	factory := NewFactory(nil, nil)
	ctx := context.Background()

	comp, err := factory.BuildComponent(ctx, pkg, hiddenComponent)
	if err != nil {
		t.Fatalf("构建组件失败: %v", err)
	}

	// 验证 overflow
	if comp.Overflow() != core.OverflowHidden {
		t.Errorf("overflow 应该是 Hidden，实际 %v", comp.Overflow())
	}

	// 验证 margin
	margin := comp.Margin()
	expectedMargin := core.Margin{Top: 10, Bottom: 10, Left: 5, Right: 5}
	if margin != expectedMargin {
		t.Errorf("margin 应该是 %+v，实际 %+v", expectedMargin, margin)
	}

	// 验证独立 container
	if comp.Container() == comp.DisplayObject() {
		t.Error("应该有独立的 container")
	}

	// 验证 scrollRect
	rect := comp.DisplayObject().ScrollRect()
	if rect == nil {
		t.Fatal("应该设置 scrollRect")
	}

	// 验证 scrollRect 尺寸（width - right, height - bottom）
	expectedW := 200.0 - 5.0
	expectedH := 150.0 - 10.0
	if rect.W != expectedW || rect.H != expectedH {
		t.Errorf("scrollRect 尺寸应该是 (%v,%v)，实际 (%v,%v)",
			expectedW, expectedH, rect.W, rect.H)
	}

	// 验证 scrollRect 偏移（left, top）
	if rect.X != 5.0 || rect.Y != 10.0 {
		t.Errorf("scrollRect 偏移应该是 (5,10)，实际 (%v,%v)", rect.X, rect.Y)
	}

	// 验证 container 偏移
	pos := comp.Container().Position()
	if pos.X != 5.0 || pos.Y != 10.0 {
		t.Errorf("container 偏移应该是 (5,10)，实际 (%v,%v)", pos.X, pos.Y)
	}
}
