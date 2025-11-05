// +build ebiten

package render

import (
	"context"
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/builder"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestScrollRectClipping(t *testing.T) {
	// 创建一个带 overflow=hidden 的组件
	pkg := &assets.Package{
		ID:   "test_scrollrect",
		Name: "TestScrollRect",
	}

	comp := &assets.PackageItem{
		ID:   "clip_comp",
		Name: "ClipComponent",
		Type: assets.PackageItemTypeComponent,
		Component: &assets.ComponentData{
			InitWidth:  200,
			InitHeight: 150,
			Margin: assets.Margin{
				Top:    10,
				Bottom: 10,
				Left:   10,
				Right:  10,
			},
			Overflow: assets.OverflowTypeHidden,
			Children: []assets.ComponentChild{
				{
					ID:     "child1",
					Name:   "child1",
					Type:   assets.ObjectTypeImage,
					X:      0,
					Y:      0,
					Width:  300, // 超出边界
					Height: 200, // 超出边界
				},
			},
		},
	}

	factory := builder.NewFactory(nil, nil)
	ctx := context.Background()

	root, err := factory.BuildComponent(ctx, pkg, comp)
	if err != nil {
		t.Fatalf("构建组件失败: %v", err)
	}

	// 验证 scrollRect 已设置
	scrollRect := root.DisplayObject().ScrollRect()
	if scrollRect == nil {
		t.Fatal("scrollRect 应该被设置")
	}

	t.Logf("scrollRect: X=%.1f, Y=%.1f, W=%.1f, H=%.1f",
		scrollRect.X, scrollRect.Y, scrollRect.W, scrollRect.H)

	// 验证 container 偏移
	container := root.Container()
	if container == nil {
		t.Fatal("container 不应该为 nil")
	}

	pos := container.Position()
	t.Logf("container position: X=%.1f, Y=%.1f", pos.X, pos.Y)

	// 尝试渲染
	target := ebiten.NewImage(640, 480)
	atlas := NewAtlasManager()

	err = DrawComponent(target, root, atlas)
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}

	t.Log("✓ scrollRect 渲染测试通过")
}

func TestScrollRectCoordinates(t *testing.T) {
	// 测试 scrollRect 坐标系统
	comp := core.NewGComponent()
	comp.SetSize(200, 150)
	comp.SetMargin(core.Margin{Top: 10, Bottom: 10, Left: 10, Right: 10})

	comp.SetupOverflow(core.OverflowHidden)

	// 验证 scrollRect
	rect := comp.DisplayObject().ScrollRect()
	if rect == nil {
		t.Fatal("scrollRect 应该被设置")
	}

	// scrollRect 应该从 (margin.left, margin.top) 开始
	if rect.X != 10.0 || rect.Y != 10.0 {
		t.Errorf("scrollRect 起点错误: (%.1f, %.1f)，期望 (10, 10)", rect.X, rect.Y)
	}

	// scrollRect 尺寸应该是 (width - margin.right, height - margin.bottom)
	expectedW := 200.0 - 10.0 // width - right
	expectedH := 150.0 - 10.0 // height - bottom
	if rect.W != expectedW || rect.H != expectedH {
		t.Errorf("scrollRect 尺寸错误: (%.1f, %.1f)，期望 (%.1f, %.1f)",
			rect.W, rect.H, expectedW, expectedH)
	}

	// 验证 container 偏移
	containerPos := comp.Container().Position()
	if containerPos.X != 10.0 || containerPos.Y != 10.0 {
		t.Errorf("container 偏移错误: (%.1f, %.1f)，期望 (10, 10)",
			containerPos.X, containerPos.Y)
	}

	t.Log("✓ scrollRect 坐标系统测试通过")
}
