package examples

import (
	"fmt"
	"time"

	"github.com/chslink/fairygui/pkg/fgui"
)

// AdvancedFeaturesExample 展示 FairyGUI 高级功能的使用
//
// 本文件演示：
// - Tween 补间动画系统
// - Relations 布局关系系统
// - Transitions 过渡动画系统
// - Gears 状态齿轮系统
//
// 这些高级功能已经通过 pkg/fgui 包导出，可以直接使用。

// ============================================================================
// Tween 补间动画示例
// ============================================================================

// TweenBasicExample 演示基础补间动画
func TweenBasicExample() {
	// 创建一个 GObject
	obj := fgui.NewGObject()
	obj.SetPosition(0, 0)

	// 一维补间：从 0 移动到 100（X 轴）
	fgui.TweenTo(0, 100, 1.0). // 1 秒
					SetTarget(obj, "x").
					SetEase(fgui.EaseType(0)). // 线性缓动
					OnUpdate(func(t *fgui.GTweener) {
			val := t.Value()
			obj.SetPosition(val.X, obj.Y())
		}).
		OnComplete(func(t *fgui.GTweener) {
			fmt.Println("移动完成")
		})

	// 在游戏主循环中推进 Tween 系统
	// fgui.TweenAdvance(delta)
}

// TweenMultiDimensionExample 演示多维补间动画
func TweenMultiDimensionExample() {
	obj := fgui.NewGObject()
	obj.SetPosition(0, 0)

	// 二维补间：同时改变 X 和 Y
	fgui.TweenTo2(0, 0, 100, 200, 1.0). // 从 (0,0) 到 (100,200)
							SetTarget(obj).
							OnUpdate(func(t *fgui.GTweener) {
			x, y := t.Value().XY()
			obj.SetPosition(x, y)
		})

	// 三维补间（如果需要）
	fgui.TweenTo3(0, 0, 0, 100, 200, 300, 1.0)
}

// TweenColorExample 演示颜色补间动画
func TweenColorExample() {
	obj := fgui.NewGObject()

	// 颜色补间：从红色渐变到绿色
	fgui.TweenToColor(0xFFFF0000, 0xFF00FF00, 1.0).
		SetTarget(obj).
		OnUpdate(func(t *fgui.GTweener) {
			color := t.Value().Color()
			// 应用颜色到对象
			_ = color
		})
}

// TweenShakeExample 演示抖动效果
func TweenShakeExample() {
	obj := fgui.NewGObject()
	obj.SetPosition(100, 100)

	// 抖动效果：幅度 10，持续 0.5 秒
	fgui.TweenShake(obj.X(), obj.Y(), 10, 0.5).
		SetTarget(obj).
		OnUpdate(func(t *fgui.GTweener) {
			x, y := t.Value().XY()
			obj.SetPosition(x, y)
		})
}

// TweenDelayedCallExample 演示延迟调用
func TweenDelayedCallExample() {
	// 1 秒后执行回调
	fgui.TweenDelayedCall(1.0).
		OnComplete(func(t *fgui.GTweener) {
			fmt.Println("延迟回调被执行")
		})
}

// TweenControlExample 演示补间控制
func TweenControlExample() {
	obj := fgui.NewGObject()

	// 创建补间
	fgui.TweenTo(0, 100, 1.0).SetTarget(obj, "x")

	// 检查是否在补间中
	if fgui.IsTweening(obj, "x") {
		fmt.Println("对象正在补间")
	}

	// 获取补间控制器
	tweener := fgui.GetTween(obj, "x")
	if tweener != nil {
		// 暂停补间
		tweener.SetPaused(true)
	}

	// 终止补间（不完成）
	fgui.KillTween(obj, false, "x")

	// 终止补间（立即完成到结束值）
	fgui.KillTween(obj, true, "x")
}

// ============================================================================
// Relations 布局关系示例
// ============================================================================

// RelationsExample 演示布局关系系统
func RelationsExample() {
	// 创建父容器和子对象
	parent := fgui.NewGComponent()
	parent.SetSize(200, 200)

	child := fgui.NewGObject()
	child.SetSize(50, 50)
	parent.AddChild(child)

	// 获取子对象的关系管理器
	relations := child.Relations()

	// 设置关系：子对象相对父容器居中
	// 参数：目标对象，关系类型，是否使用百分比
	relations.Add(parent.GObject, fgui.RelationType(0), false)

	// 当父容器大小改变时，子对象会自动调整位置
	parent.SetSize(300, 300) // 子对象会自动重新定位
}

// ============================================================================
// Transitions 过渡动画示例
// ============================================================================

// TransitionsExample 演示过渡动画系统
func TransitionsExample() {
	// 创建组件
	comp := fgui.NewGComponent()

	// 通常 Transition 是从 FUI 包中加载的
	// 这里演示如何手动添加和使用

	// 添加过渡信息（通常由构建器自动完成）
	info := fgui.TransitionInfo{
		Name: "fadeIn",
		// ... 其他配置
	}
	comp.AddTransition(info)

	// 获取并播放过渡动画
	trans := comp.Transition("fadeIn")
	if trans != nil {
		// Play 参数：播放次数（-1 表示无限循环），延迟时间
		trans.Play(1, 0)
	}

	// 停止过渡动画
	if trans != nil {
		// Stop 参数：是否立即完成到结束状态
		trans.Stop(false)
	}
}

// ============================================================================
// Gears 状态齿轮示例
// ============================================================================

// GearsExample 演示状态齿轮系统
func GearsExample() {
	// 创建组件
	comp := fgui.NewGComponent()

	// 通常 Controller 是从 FUI 包中加载的
	// 这里演示如何获取和使用控制器

	// 假设组件已经有一个名为 "state" 的控制器（通常由构建器添加）
	ctrl := comp.GetController("state")
	if ctrl != nil {
		// 切换状态
		ctrl.SetSelectedIndex(1)

		// 获取当前状态
		index := ctrl.SelectedIndex()
		_ = index
	}

	// 齿轮系统会根据控制器状态自动调整对象属性
	// 这些关联通常在 FairyGUI 编辑器中配置
}

// ============================================================================
// 综合示例：制作一个简单的弹出动画
// ============================================================================

// PopupAnimationExample 综合示例：弹出窗口动画
func PopupAnimationExample() {
	// 创建弹窗
	popup := fgui.NewGComponent()
	popup.SetSize(300, 200)
	popup.SetPosition(200, 150)

	// 初始状态：缩小 + 透明
	popup.SetScale(0.5, 0.5)
	popup.SetAlpha(0)

	// 添加到舞台
	fgui.Root().AddChild(popup.GObject)

	// 弹出动画：放大 + 淡入
	// 注意：这里需要创建一个结构体来存储 scale 值
	scaleTarget := &struct{ x, y float64 }{x: 0.5, y: 0.5}

	// 缩放动画
	fgui.TweenTo2(0.5, 0.5, 1.0, 1.0, 0.3).
		SetTarget(scaleTarget).
		SetEase(fgui.EaseType(0)).
		OnUpdate(func(t *fgui.GTweener) {
			x, y := t.Value().XY()
			popup.SetScale(x, y)
		})

	// 透明度动画
	fgui.TweenTo(0, 1, 0.3).
		SetTarget(popup.GObject, "alpha").
		OnUpdate(func(t *fgui.GTweener) {
			popup.SetAlpha(t.Value().X)
		})
}

// ============================================================================
// 主循环集成
// ============================================================================

// MainLoopExample 展示如何在主循环中集成高级功能
func MainLoopExample() {
	// 在游戏主循环的 Update 方法中
	delta := 16 * time.Millisecond // 60 FPS

	// 推进 Tween 系统
	fgui.TweenAdvance(delta)

	// 推进根节点（包含 Transitions 等）
	fgui.Advance(delta, fgui.MouseState{})

	// 或使用带输入状态的版本
	input := fgui.InputState{
		// ... 输入数据
	}
	fgui.AdvanceInput(delta, input)
}

// ============================================================================
// 使用建议
// ============================================================================

/*
高级功能使用建议：

1. Tween 补间动画：
   - 适用于简单的属性动画（位置、大小、颜色等）
   - 支持缓动函数、延迟、重复等
   - 记得在主循环调用 TweenAdvance()

2. Relations 布局关系：
   - 用于动态布局，自动调整子对象位置
   - 适合响应式 UI 设计
   - 通常在 FairyGUI 编辑器中配置

3. Transitions 过渡动画：
   - 用于复杂的时间轴动画
   - 支持多个对象、多个属性的组合动画
   - 通常在 FairyGUI 编辑器中制作

4. Gears 状态齿轮：
   - 用于根据控制器状态切换对象属性
   - 适合多状态 UI（正常/高亮/禁用等）
   - 通常在 FairyGUI 编辑器中配置

5. 性能优化：
   - 避免同时创建大量补间
   - 及时清理不需要的补间（KillTween）
   - 使用对象池复用对象
*/
