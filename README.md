# FairyGUI Ebiten

基于 [Ebiten](https://ebiten.org/) 游戏引擎的 FairyGUI UI 框架 Go 语言移植。

## 安装

```bash
go get github.com/chslink/fairygui
```

## 快速开始

```go
package main

import (
	"context"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/chslink/fairygui/pkg/fgui"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
)

type Game struct{ root *core.GRoot }

func (g *Game) Update() error { return nil }
func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 30, 255})
	// 渲染 FGUI 组件树
	render.DrawComponent(screen, g.root.GComponent, nil)
}
func (g *Game) Layout(w, h int) (int, int) { return 800, 600 }

func main() {
	root := core.Inst()
	game := &Game{root: root}
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("FairyGUI Ebiten")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
```

### Options 模式

```go
btn := fgui.CreateButton(
	fgui.WithPosition(100, 200),
	fgui.WithSize(120, 40),
	fgui.WithTitle("确认"),
)
```

### Builder 链式

```go
list := fgui.NewListBuilder().
	Virtual(true).
	NumItems(100).
	ItemRenderer(func(i int, obj *fgui.GObject) {
		// 设置项目内容
	}).
	Build()
```

### 事件系统

```go
cancel := fgui.ListenClick(btn, func() { fmt.Println("clicked") })
defer cancel()
```

## 演示程序

```bash
go run ./demo
```

16 个交互场景：按钮、文本、网格、列表、进度条、窗口、弹窗、拖放、树形视图、引导层、摇杆、虚拟列表、循环列表、过渡动画等。

内建 HTTP 调试服务器（端口 8090），提供组件树、类型过滤、虚拟列表分析。

## 控件覆盖

| 控件 | 状态 |
|------|------|
| GButton | ✅ 完整（Common/Check/Radio 模式） |
| GTextField / GRichTextField | ✅ 完整（UBB 解析） |
| GTextInput | ✅ 完整（控制键/剪贴板，IME 待 Ebiten v2.10） |
| GLabel | ✅ 完整（图标+标题） |
| GImage | ✅ 完整（九宫格/平铺/FillMethod） |
| GLoader | ✅ 完整（异步加载/FillMethod/Scale9） |
| GList / GList(Virtual) | ✅ 完整（虚拟/循环/多选/分页） |
| GTree | ✅ 完整（展开/折叠/自定义节点） |
| GComboBox | ✅ 完整 |
| GProgressBar | ✅ 完整 |
| GSlider | ✅ 完整 |
| GScrollBar | ✅ 完整 |
| GGraph | ✅ 完整（绘图命令） |
| GGroup | ✅ 完整 |
| GMovieClip | ✅ 完整（帧动画/回放控制） |

## 核心能力

| 系统 | 状态 |
|------|------|
| Relations 关联布局 | ✅ |
| Controller 状态机 + Action 链 | ✅ |
| Transition 过渡动画 | ✅ |
| Gears 齿轮系统 (11 种) | ✅ |
| Tween 补间引擎 | ✅ |
| ScrollPane 滚动面板 | ✅ |
| Window 模态窗口 | ✅ |
| PopupMenu 弹出菜单 | ✅ |
| DragDropManager 拖拽 | ✅ |
| ByteBuffer 二进制解析 | ✅ |
| 对象池 | ✅ |
| 音频播放 (WAV/MP3/Ogg) | ✅ |

## 项目结构

```
fairygui/
├── demo/              演示程序 (Ebiten 入口 + 场景)
│   ├── assets/        16 个 .fui 资源包
│   ├── scenes/        交互场景
│   └── debug/         HTTP 调试服务器
├── pkg/fgui/          公开 API
│   ├── core/          核心类型 (GObject/GComponent/GRoot/Window/...)
│   ├── widgets/       UI 控件
│   ├── render/        Ebiten 渲染实现
│   ├── assets/        资源加载与解析
│   ├── builder/       从 .fui 构建组件树
│   ├── gears/         齿轮状态系统
│   ├── tween/         补间动画
│   ├── utils/         工具函数
│   └── audio/         音频播放
├── internal/
│   ├── compat/laya/   LayaAir 兼容层 (Sprite/Event/Stage/Timer)
│   └── text/          UBB 解析
├── docs/              文档
│   ├── architecture.md
│   ├── ime-status.md
│   └── impl-plan/     实施计划文档
└── laya_src/          TypeScript 参考源码
```

## 移植进度

TS 原始模块 72 个，已完整实现 64 个（88.9%）。

**缺失模块**：GLoader3D（3D 骨骼动画，依赖 LayaAir 3D）、TranslationHelper（多语言翻译）、AssetProxy/AsyncOperation（Go 已有替代方案）

详见 `docs/audit-ts-vs-go.md` 完整对比和 `docs/refactor-progress.md` 开发日志。

## 已知限制

- **IME 中文输入**：受 Ebiten 游戏引擎限制，当前不完全稳定。等待 Ebiten v2.10 `Composer` API 发布后彻底解决。详见 `docs/ime-status.md`。
- `cmd/` 调试工具尚未迁移

## License

BSD-3-Clause
