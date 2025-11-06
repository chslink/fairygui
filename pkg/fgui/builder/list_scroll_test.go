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
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// TestListScroll 测试List组件滚动功能
// 场景：Demo_List.xml 中的n0列表组件
// 预期：设置滚动位置后，列表项应该正确更新位置
func TestListScroll(t *testing.T) {
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

	// 构建 Demo_List 组件
	factory := NewFactory(nil, nil)
	factory.RegisterPackage(pkg)

	ctx := context.Background()

	// 查找 Demo_List 组件
	var demoItem *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_List" {
			demoItem = item
			break
		}
	}
	if demoItem == nil {
		t.Fatalf("未找到 Demo_List 组件")
	}

	demo, err := factory.BuildComponent(ctx, pkg, demoItem)
	if err != nil {
		t.Fatalf("构建 Demo_List 组件失败: %v", err)
	}

	// 将 demo 添加到 stage
	stage.AddChild(demo.DisplayObject())

	// 获取 n0 列表组件
	n0 := demo.ChildByName("n0")
	if n0 == nil {
		t.Fatalf("未找到 n0 对象")
	}

	// 确认是 GList 类型
	list, ok := n0.Data().(*widgets.GList)
	if !ok || list == nil {
		t.Fatalf("n0 不是 GList")
	}

	t.Logf("列表 n0 初始状态: 项目数量=%d, 子对象数量=%d, 虚拟模式=%v",
		list.NumItems(), list.ChildrenCount(), list.IsVirtual())

	// 已经确认list是*widgets.GList类型
	// 从GList获取GComponent
	listComp := list.GComponent
	if listComp == nil {
		t.Fatalf("GList的GComponent为nil")
	}

	// 获取滚动面板
	scrollPane := listComp.ScrollPane()
	if scrollPane == nil {
		t.Fatalf("n0 没有 ScrollPane")
	}

	// 记录初始滚动位置和内容尺寸
	initialPosY := scrollPane.PosY()
	contentHeight := scrollPane.ContentSize().Y
	viewHeight := scrollPane.ViewHeight()
	t.Logf("初始滚动位置: Y=%.0f, 内容高度=%.0f, 视图高度=%.0f", initialPosY, contentHeight, viewHeight)

	// 从GComponent获取子元素
	children := listComp.Children()
	if len(children) == 0 {
		t.Fatalf("列表没有子项")
	}
	t.Logf("列表初始可见子项数量: %d", len(children))

	// 记录第一个子项的初始位置
	firstChild := children[0]
	initialChildY := firstChild.Y()
	t.Logf("第一个子项初始位置: Y=%.0f", initialChildY)

	// 尝试滚动到中间位置
	targetScrollPos := (contentHeight - viewHeight) / 2
	t.Logf("尝试滚动到位置: Y=%.0f", targetScrollPos)

	// 设置滚动位置
	scrollPane.SetPos(0, targetScrollPos, false)

	// 推进一帧以确保更新
	env.Advance(time.Millisecond*16, laya.MouseState{X: 100, Y: 100, Primary: false})

	// 检查滚动位置是否已更新
	newScrollPos := scrollPane.PosY()
	t.Logf("滚动后位置: Y=%.0f", newScrollPos)

	if newScrollPos == initialPosY {
		t.Errorf("滚动位置没有变化，滚动失败")
	}

	// 检查第一个子项的位置是否已更新
	newChildren := listComp.Children()
	if len(newChildren) == 0 {
		t.Fatalf("滚动后列表没有子项")
	}

	// 检查子项位置 - 对于非虚拟列表，子项的绝对Y坐标不会改变，但滚动位置应该改变
	t.Logf("滚动后可见子项数量: %d", len(newChildren))
	for i, child := range newChildren {
		currentY := child.Y()
		t.Logf("子项[%d] 当前位置: Y=%.0f", i, currentY)
	}

	// 简化测试，专注于检查滚动位置是否变化
	t.Logf("滚动前第一个子项位置: Y=%.0f", initialChildY)
	if len(newChildren) > 0 {
		newFirstChildY := newChildren[0].Y()
		t.Logf("滚动后第一个子项位置: Y=%.0f", newFirstChildY)

		// 对于非虚拟列表，子项的绝对位置不会改变，但滚动位置应该改变
		// 检查滚动位置是否有效改变（考虑浮点精度误差）
		if newScrollPos <= initialPosY+1 && newScrollPos >= initialPosY-1 {
			t.Errorf("滚动位置没有有效改变: 预期 %.0f, 实际 %.0f", targetScrollPos, newScrollPos)
		}
	}

	// 验证滚动容器的位置是否正确应用了负坐标偏移（这是滚动的核心机制）
	container := list.Container()
	if container != nil {
		// 在ScrollPane中，容器位置应该是负的滚动位置
		// 这确保内容在视觉上向上移动
		containerY := container.Position().Y
		t.Logf("滚动容器位置: Y=%.0f", containerY)
		// 检查容器位置是否正确反映了滚动偏移（考虑浮点精度误差）
		expectedContainerY := -newScrollPos
		if containerY <= expectedContainerY+1 && containerY >= expectedContainerY-1 {
			t.Logf("滚动容器位置验证通过，正确应用了负坐标偏移")
		} else {
			t.Errorf("容器位置未完全反映滚动偏移: 期望=%.0f, 实际=%.0f", expectedContainerY, containerY)
		}
	} else {
		t.Logf("警告: 未找到列表的容器组件")
	}

	// 测试滚动到底部
	bottomPos := contentHeight - viewHeight
	if bottomPos < 0 {
		bottomPos = 0
	}
	scrollPane.SetPos(0, bottomPos, false)
	env.Advance(time.Millisecond*16, laya.MouseState{X: 100, Y: 100, Primary: false})

	bottomScrollPos := scrollPane.PosY()
	t.Logf("滚动到底部后位置: Y=%.0f", bottomScrollPos)

	if bottomScrollPos != bottomPos {
		t.Errorf("滚动到底部失败，期望: %.0f, 实际: %.0f", bottomPos, bottomScrollPos)
	}
}
