package builder

import (
	"context"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// TestScrollPaneSetup 测试 ScrollPane 的基本设置
func TestScrollPaneSetup(t *testing.T) {
	// 加载 Basics 包
	loader := assets.NewFileLoader("../../../demo/assets")
	ctx := context.Background()

	data, err := loader.LoadOne(ctx, "Basics.fui", assets.ResourceBinary)
	if err != nil {
		t.Skipf("跳过集成测试：无法加载 Basics.fui: %v", err)
		return
	}

	pkg, err := assets.ParsePackage(data, "demo/assets/Basics")
	if err != nil {
		t.Fatalf("解析包失败: %v", err)
	}

	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	// 测试 Component7 (overflow="scroll" scroll="both")
	t.Run("Component7_ScrollPane", func(t *testing.T) {
		item := pkg.ItemByName("Component7")
		if item == nil {
			t.Skip("Component7 未找到")
			return
		}

		t.Logf("Component7 metadata:")
		if item.Component != nil {
			t.Logf("  InitWidth=%d, InitHeight=%d", item.Component.InitWidth, item.Component.InitHeight)
			t.Logf("  Overflow=%d", item.Component.Overflow)
			t.Logf("  Children count=%d", len(item.Component.Children))
			if len(item.Component.Children) > 0 {
				child := item.Component.Children[0]
				t.Logf("  Child[0]: ID=%s, Type=%d, Size=(%d,%d)", child.ID, child.Type, child.Width, child.Height)
			}
		}

		comp, err := factory.BuildComponent(ctx, pkg, item)
		if err != nil {
			t.Fatalf("构建组件失败: %v", err)
		}

		// 验证 ScrollPane 创建
		pane := comp.ScrollPane()
		if pane == nil {
			t.Fatal("Component7 应该有 ScrollPane")
		}

		t.Logf("ScrollPane info:")

		viewWidth := pane.ViewWidth()
		viewHeight := pane.ViewHeight()
		t.Logf("  ViewSize=(%.1f, %.1f)", viewWidth, viewHeight)

		contentSize := pane.ContentSize()
		t.Logf("  ContentSize=(%.1f, %.1f)", contentSize.X, contentSize.Y)

		t.Logf("  VtScrollBarURL=%s", pane.VtScrollBarURL())
		t.Logf("  HzScrollBarURL=%s", pane.HzScrollBarURL())

		// 检查内容是否超出视口（需要滚动）
		needsScroll := contentSize.X > viewWidth || contentSize.Y > viewHeight
		t.Logf("  NeedsScroll=%v", needsScroll)

		// 检查子对象数量
		childCount := len(comp.Children())
		t.Logf("  Component children count=%d", childCount)

		// Component7 应该有：
		// 1. 原始的图片子对象 (n1)
		// 2. 垂直滚动条 (如果需要)
		// 3. 水平滚动条 (如果需要)

		// 验证滚动条可见性
		if needsScroll {
			// 检查垂直滚动条
			vtBar := comp.ChildByName("vt_scrollbar")
			if vtBar != nil {
				t.Logf("  VtScrollBar visible=%v", vtBar.Visible())
				if !vtBar.Visible() {
					t.Error("垂直滚动条应该可见")
				}
			}

			// 检查水平滚动条
			hzBar := comp.ChildByName("hz_scrollbar")
			if hzBar != nil {
				t.Logf("  HzScrollBar visible=%v", hzBar.Visible())
				if !hzBar.Visible() {
					t.Error("水平滚动条应该可见")
				}
			}
		}

		if viewWidth <= 0 || viewHeight <= 0 {
			t.Error("ViewSize 应该 > 0")
		}
	})

	// 测试 Component8 (overflow="scroll" with margin)
	t.Run("Component8_ScrollPaneWithMargin", func(t *testing.T) {
		item := pkg.ItemByName("Component8")
		if item == nil {
			t.Skip("Component8 未找到")
			return
		}

		t.Logf("Component8 metadata:")
		if item.Component != nil {
			t.Logf("  InitWidth=%d, InitHeight=%d", item.Component.InitWidth, item.Component.InitHeight)
			t.Logf("  Margin=%+v", item.Component.Margin)
			t.Logf("  Overflow=%d", item.Component.Overflow)
		}

		comp, err := factory.BuildComponent(ctx, pkg, item)
		if err != nil {
			t.Fatalf("构建组件失败: %v", err)
		}

		pane := comp.ScrollPane()
		if pane == nil {
			t.Fatal("Component8 应该有 ScrollPane")
		}

		// 验证 margin
		margin := comp.Margin()
		expectedMargin := core.Margin{Top: 30, Bottom: 30, Left: 30, Right: 30}
		if margin != expectedMargin {
			t.Errorf("Margin 应该是 %+v，实际 %+v", expectedMargin, margin)
		}

		// ViewSize 应该考虑 margin 和滚动条尺寸
		viewWidth := pane.ViewWidth()
		viewHeight := pane.ViewHeight()
		t.Logf("ViewSize with margin: (%.1f, %.1f)", viewWidth, viewHeight)

		// 预期 ViewSize = InitSize - 滚动条尺寸 - margin
		// 假设滚动条宽度为 17 (标准尺寸)
		const scrollBarSize = 17.0
		expectedViewWidth := float64(item.Component.InitWidth) - scrollBarSize - float64(margin.Left) - float64(margin.Right)
		expectedViewHeight := float64(item.Component.InitHeight) - scrollBarSize - float64(margin.Top) - float64(margin.Bottom)

		// 允许一些浮点误差
		if viewWidth < expectedViewWidth-1 || viewWidth > expectedViewWidth+1 {
			t.Errorf("ViewWidth 应该约为 %.1f（InitWidth %.0f - scrollBar %.0f - margin %d），实际 %.1f",
				expectedViewWidth, float64(item.Component.InitWidth), scrollBarSize, margin.Left+margin.Right, viewWidth)
		}
		if viewHeight < expectedViewHeight-1 || viewHeight > expectedViewHeight+1 {
			t.Errorf("ViewHeight 应该约为 %.1f（InitHeight %.0f - scrollBar %.0f - margin %d），实际 %.1f",
				expectedViewHeight, float64(item.Component.InitHeight), scrollBarSize, margin.Top+margin.Bottom, viewHeight)
		}
	})
}

// TestScrollBarCreation 测试滚动条的创建
func TestScrollBarCreation(t *testing.T) {
	// 创建一个需要滚动的组件
	pkg := &assets.Package{
		ID:   "test_scrollbar",
		Name: "TestScrollBar",
	}

	// 创建子图片项（大于容器）
	imgItem := &assets.PackageItem{
		ID:   "big_image",
		Name: "BigImage",
		Type: assets.PackageItemTypeImage,
		Sprite: &assets.AtlasSprite{
			Rect: assets.Rect{
				Width:  500,
				Height: 500,
			},
		},
	}
	pkg.Items = append(pkg.Items, imgItem)

	// 创建带滚动的组件
	scrollComp := &assets.PackageItem{
		ID:   "scroll_comp",
		Name: "ScrollComponent",
		Type: assets.PackageItemTypeComponent,
		Component: &assets.ComponentData{
			InitWidth:  200,
			InitHeight: 200,
			Overflow:   assets.OverflowTypeScroll,
			Children: []assets.ComponentChild{
				{
					ID:     "child_img",
					Name:   "child_img",
					Type:   assets.ObjectTypeImage,
					Src:    "big_image",
					X:      0,
					Y:      0,
					Width:  500,
					Height: 500,
				},
			},
		},
	}

	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)
	ctx := context.Background()

	comp, err := factory.BuildComponent(ctx, pkg, scrollComp)
	if err != nil {
		t.Fatalf("构建组件失败: %v", err)
	}

	pane := comp.ScrollPane()
	if pane == nil {
		t.Fatal("应该有 ScrollPane")
	}

	viewWidth := pane.ViewWidth()
	viewHeight := pane.ViewHeight()
	t.Logf("ViewSize: (%.1f, %.1f)", viewWidth, viewHeight)

	contentSize := pane.ContentSize()
	t.Logf("ContentSize: (%.1f, %.1f)", contentSize.X, contentSize.Y)

	// 内容应该比视口大
	if contentSize.X <= viewWidth && contentSize.Y <= viewHeight {
		t.Error("内容应该比视口大，需要滚动")
	}

	// 检查滚动条 URL
	vtURL := pane.VtScrollBarURL()
	hzURL := pane.HzScrollBarURL()
	t.Logf("ScrollBar URLs: vt=%s, hz=%s", vtURL, hzURL)

	// 注意：滚动条的实际创建需要依赖包中的滚动条资源
	// 这个测试只验证 ScrollPane 的基本设置
}
