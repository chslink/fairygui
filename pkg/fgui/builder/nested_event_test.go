package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/gears"
)

// TestNestedComponentEventCapture 测试嵌套组件的事件捕获
// 场景：Main -> container (Component4) -> Demo_Relation -> n5 (Button4)
// 预期：n5 应该能够捕获鼠标事件
func TestNestedComponentEventCapture(t *testing.T) {
	// 禁用 Tween 效果，避免动画延迟
	oldTweenDisabled := gears.DisableAllTweenEffect
	gears.DisableAllTweenEffect = true
	defer func() {
		gears.DisableAllTweenEffect = oldTweenDisabled
	}()

	// 加载 Basics.fui 包
	fuiPath := filepath.Join("..", "..", "..", "demo", "assets", "Basics.fui")
	fuiData, err := os.ReadFile(fuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 .fui 文件: %v", err)
	}

	pkg, err := assets.ParsePackage(fuiData, "demo/assets/Basics")
	if err != nil {
		t.Fatalf("解析 .fui 文件失败: %v", err)
	}

	// 创建测试 Stage 环境
	env := testutil.NewStageEnv(t, 1136, 640)
	stage := env.Stage

	// 构建 Main 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 Main 组件
	var mainItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Main" {
			mainItem = item
			break
		}
	}
	if mainItem == nil {
		t.Fatalf("未找到 Main 组件")
	}

	main, err := factory.BuildComponent(ctx, pkg, mainItem)
	if err != nil {
		t.Fatalf("构建 Main 组件失败: %v", err)
	}

	// 设置控制器到 page 1（在添加到 stage 之前设置）
	c1 := main.ControllerByName("c1")
	if c1 == nil {
		t.Fatalf("未找到 c1 控制器")
	}
	t.Logf("控制器 c1 初始状态: selectedIndex=%d, pageCount=%d", c1.SelectedIndex(), c1.PageCount())
	c1.SetSelectedIndex(1)
	t.Logf("设置控制器 c1 到 page 1, 当前 selectedIndex=%d", c1.SelectedIndex())

	// 将 main 添加到 stage
	stage.AddChild(main.DisplayObject())

	// 查找 container
	container := main.ChildByName("container")
	if container == nil {
		t.Fatalf("未找到 container 对象")
	}

	// 查找 Demo_Relation 组件
	var demoRelationItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_Relation" {
			demoRelationItem = item
			break
		}
	}
	if demoRelationItem == nil {
		t.Fatalf("未找到 Demo_Relation 组件")
	}

	// 构建 Demo_Relation 并添加到 container
	demoRelation, err := factory.BuildComponent(ctx, pkg, demoRelationItem)
	if err != nil {
		t.Fatalf("构建 Demo_Relation 组件失败: %v", err)
	}

	// 将 Demo_Relation 添加到 container
	containerComp, ok := container.Data().(*core.GComponent)
	if !ok || containerComp == nil {
		t.Fatalf("container 不是 GComponent")
	}

	t.Logf("=== 添加 Demo_Relation 前 ===")
	t.Logf("container DisplayObject: %p", container.DisplayObject())
	t.Logf("container Data (GComponent) DisplayObject: %p", containerComp.GObject.DisplayObject())
	t.Logf("container DisplayObject == containerComp DisplayObject: %v",
		container.DisplayObject() == containerComp.GObject.DisplayObject())
	t.Logf("container DisplayObject parent: %v", container.DisplayObject().Parent())
	t.Logf("Demo_Relation DisplayObject: %v", demoRelation.GObject.DisplayObject() != nil)
	t.Logf("Demo_Relation sprite parent (添加前): %v", demoRelation.GObject.DisplayObject().Parent())

	containerComp.AddChild(demoRelation.GObject)

	t.Logf("=== 添加 Demo_Relation 后 ===")
	t.Logf("Demo_Relation GObject parent: %v", demoRelation.GObject.Parent() != nil)
	t.Logf("Demo_Relation sprite parent (添加后): %v", demoRelation.GObject.DisplayObject().Parent())

	// 查找 n5 按钮
	n5 := demoRelation.ChildByName("n5")
	if n5 == nil {
		t.Fatalf("未找到 n5 对象")
	}

	// 打印层级结构和 Gear 信息用于调试
	t.Logf("=== 层级结构 ===")
	mainGlobal := main.GObject.DisplayObject().LocalToGlobal(laya.Point{X: 0, Y: 0})
	t.Logf("Main: local=(%.0f,%.0f) global=(%.0f,%.0f) size=%.0fx%.0f",
		main.GObject.X(), main.GObject.Y(), mainGlobal.X, mainGlobal.Y,
		main.GObject.Width(), main.GObject.Height())

	// 检查 container 的 Gear
	gearXY := container.GetGear(1) // Gear 1 是 GearXY
	t.Logf("  container Gear[1] (XY): %v", gearXY)
	t.Logf("  控制器 c1 当前: selectedIndex=%d, selectedPage=%s",
		c1.SelectedIndex(), c1.SelectedPageID())
	if gearXY != nil {
		// 尝试手动应用 Gear
		gearXY.Apply()
		t.Logf("  手动应用 Gear 后 container pos: (%.0f,%.0f)", container.X(), container.Y())
	}

	containerGlobal := container.DisplayObject().LocalToGlobal(laya.Point{X: 0, Y: 0})
	t.Logf("  container (GComponent): local=(%.0f,%.0f) global=(%.0f,%.0f) size=%.0fx%.0f opaque=%v mouseThrough=%v",
		container.X(), container.Y(), containerGlobal.X, containerGlobal.Y,
		container.Width(), container.Height(),
		containerComp.Opaque(), container.DisplayObject().MouseThrough())
	t.Logf("    container parent: %v, sprite parent: %v",
		container.Parent(), container.DisplayObject().Parent())

	demoRelationGlobal := demoRelation.GObject.DisplayObject().LocalToGlobal(laya.Point{X: 0, Y: 0})
	t.Logf("    Demo_Relation (GComponent): local=(%.0f,%.0f) global=(%.0f,%.0f) size=%.0fx%.0f opaque=%v mouseThrough=%v",
		demoRelation.GObject.X(), demoRelation.GObject.Y(),
		demoRelationGlobal.X, demoRelationGlobal.Y,
		demoRelation.GObject.Width(), demoRelation.GObject.Height(),
		demoRelation.Opaque(), demoRelation.GObject.DisplayObject().MouseThrough())
	t.Logf("      Demo_Relation parent: %v, sprite parent: %v",
		demoRelation.GObject.Parent(), demoRelation.GObject.DisplayObject().Parent())

	n5Global := n5.DisplayObject().LocalToGlobal(laya.Point{X: 0, Y: 0})
	t.Logf("      n5 (GButton): local=(%.0f,%.0f) global=(%.0f,%.0f) size=%.0fx%.0f mouseEnabled=%v",
		n5.X(), n5.Y(), n5Global.X, n5Global.Y,
		n5.Width(), n5.Height(),
		n5.DisplayObject().MouseEnabled())
	t.Logf("        n5 parent: %v, sprite parent: %v",
		n5.Parent(), n5.DisplayObject().Parent())

	// 测试事件捕获
	clickCount := 0
	n5.DisplayObject().Dispatcher().On(laya.EventClick, func(evt laya.Event) {
		clickCount++
		t.Logf("✓ n5 捕获到点击事件")
	})

	// 计算 n5 的全局坐标
	// n5 相对于 Demo_Relation 的位置是 (37, 108)
	// Demo_Relation 相对于 container 的位置是 (0, 0)
	// container 相对于 Main 的位置是 (0, 70) (c1 控制器在 page 1 时)
	// Main 相对于 Stage 的位置是 (0, 0)
	// 所以 n5 的全局坐标应该是 (37, 178)

	clickX := 37.0 + 50.0 // n5.x + 中心偏移
	clickY := 178.0       // 70 (container.y) + 108 (n5.y relative to Demo_Relation)

	t.Logf("=== 测试点击坐标: (%.0f, %.0f) ===", clickX, clickY)

	// 执行点击测试
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: true})
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})

	if clickCount == 0 {
		// 执行命中测试诊断
		t.Logf("=== 命中测试诊断 ===")

		// 计算 n5 的全局坐标
		n5GlobalPos := n5.DisplayObject().LocalToGlobal(laya.Point{X: 0, Y: 0})
		t.Logf("n5 全局左上角: (%.0f, %.0f)", n5GlobalPos.X, n5GlobalPos.Y)
		t.Logf("n5 全局范围: X=[%.0f, %.0f], Y=[%.0f, %.0f]",
			n5GlobalPos.X, n5GlobalPos.X+n5.Width(),
			n5GlobalPos.Y, n5GlobalPos.Y+n5.Height())

		hit := stage.Root().HitTest(laya.Point{X: clickX, Y: clickY})
		if hit == nil {
			t.Errorf("命中测试返回 nil，没有任何对象捕获事件")
		} else {
			hitName := "unknown"
			if hit.Owner() != nil {
				if gobj, ok := hit.Owner().(*core.GObject); ok {
					hitName = gobj.Name()
				}
			}
			t.Logf("命中测试返回: %s (sprite: %p, mouseEnabled: %v, mouseThrough: %v)",
				hitName, hit, hit.MouseEnabled(), hit.MouseThrough())

			// 测试 n5 自己的 HitTest
			n5Hit := n5.DisplayObject().HitTest(laya.Point{X: clickX, Y: clickY})
			if n5Hit == nil {
				t.Logf("n5.HitTest 返回 nil")
				// 测试点在 n5 局部坐标系中的位置
				localPt := n5.DisplayObject().GlobalToLocal(laya.Point{X: clickX, Y: clickY})
				t.Logf("点击坐标在 n5 局部坐标系: (%.0f, %.0f)", localPt.X, localPt.Y)
				t.Logf("n5 尺寸: %.0fx%.0f", n5.Width(), n5.Height())
			} else {
				t.Logf("n5.HitTest 返回: %v", n5Hit == n5.DisplayObject())
			}
		}
		t.Errorf("n5 按钮没有捕获到点击事件")
	} else {
		t.Logf("✓ n5 按钮成功捕获 %d 次点击事件", clickCount)
	}
}

// TestSimpleButtonEventCapture 测试简单按钮的事件捕获（对照组）
func TestSimpleButtonEventCapture(t *testing.T) {
	fuiPath := filepath.Join("..", "..", "..", "demo", "assets", "Basics.fui")
	fuiData, err := os.ReadFile(fuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 .fui 文件: %v", err)
	}

	pkg, err := assets.ParsePackage(fuiData, "demo/assets/Basics")
	if err != nil {
		t.Fatalf("解析 .fui 文件失败: %v", err)
	}

	env := testutil.NewStageEnv(t, 1136, 640)
	stage := env.Stage

	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	var mainItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Main" {
			mainItem = item
			break
		}
	}
	if mainItem == nil {
		t.Fatalf("未找到 Main 组件")
	}

	main, err := factory.BuildComponent(ctx, pkg, mainItem)
	if err != nil {
		t.Fatalf("构建 Main 组件失败: %v", err)
	}

	stage.AddChild(main.DisplayObject())

	// 测试 btn_Back（在 Main 层级的按钮）
	btnBack := main.ChildByName("btn_Back")
	if btnBack == nil {
		t.Fatalf("未找到 btn_Back 对象")
	}

	clickCount := 0
	btnBack.DisplayObject().Dispatcher().On(laya.EventClick, func(evt laya.Event) {
		clickCount++
	})

	// btn_Back 位置是 (963, -1)，点击中心点
	clickX := 963.0 + 80.0
	clickY := -1.0 + 35.0

	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: true})
	env.Advance(time.Millisecond*16, laya.MouseState{X: clickX, Y: clickY, Primary: false})

	if clickCount == 0 {
		t.Errorf("btn_Back 按钮没有捕获到点击事件")
	} else {
		t.Logf("✓ btn_Back 按钮成功捕获 %d 次点击事件", clickCount)
	}
}
