# FairyGUI Ebiten
**本项目是纯AI实现**,目前还处于开发阶段
FairyGUI Ebiten 是一个基于 Ebiten 游戏引擎的 FairyGUI UI 框架 Go 语言实现。该项目旨在为 Go 开发者提供强大的 UI 系统，支持丰富的界面组件和交互功能。

## 特性

- 基于 Ebiten 游戏引擎的高效渲染
- 支持 FairyGUI 的 UI 组件系统
- 完整的事件处理机制
- 虚拟列表支持，优化大量数据渲染性能
- 丰富的 UI 组件，包括按钮、列表、滚动条、过渡动画等
- 支持多种文本渲染和字体处理
- 调试工具集，便于开发和调试

## 安装

```bash
go mod init your-project
go get github.com/chslink/fairygui
```

## 快速开始

```go
package main

import (
    "context"
    "log"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/chslink/fairygui/pkg/fgui"
    "github.com/chslink/fairygui/pkg/fgui/core"
)

func main() {
    // 初始化 FairyGUI
    ctx := context.Background()
    
    // 创建 UI 工厂
    factory := fgui.NewFactory(nil, nil)
    
    // 加载 FairyGUI 包
    // data, err := os.ReadFile("path/to/your/ui.fui")
    // if err != nil {
    //     log.Fatal(err)
    // }
    // pkg, err := fgui.ParsePackage(data, "ui-package")
    // if err != nil {
    //     log.Fatal(err)
    // }
    // factory.RegisterPackage(pkg)
    
    // 构建 UI 组件
    // item := pkg.ItemByName("Main")
    // component, err := factory.BuildComponent(ctx, pkg, item)
    // if err != nil {
    //     log.Fatal(err)
    // }
    
    // 创建根容器并添加组件
    root := core.GRoot.Inst()
    // root.GObject.AddChild(component.GComponent.GObject)
    
    // 设置 Ebiten 窗口并运行
    ebiten.SetWindowSize(800, 600)
    ebiten.SetWindowTitle("FairyGUI Ebiten Demo")
    
    // 创建并运行游戏实例
    game := &Game{root: root}
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}

type Game struct {
    root *core.GRoot
}

func (g *Game) Update() error {
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    // FairyGUI 的渲染将自动处理
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return 800, 600
}
```

## 项目结构

- `cmd/` - 命令行工具
- `demo/` - 演示程序
- `docs/` - 文档
- `internal/` - 内部实现
- `pkg/fgui/` - FairyGUI 核心实现
  - `assets/` - 资源加载
  - `builder/` - 组件构建器
  - `core/` - 核心组件
  - `gears/` - 齿轮系统
  - `render/` - 渲染系统
  - `tween/` - 补间动画
  - `utils/` - 工具函数
  - `widgets/` - UI 小部件

## 演示程序

运行内置的演示程序：

```bash
cd demo
go run main.go
```

## 调试工具

项目包含强大的调试工具集，包括：

- Inspector - 对象检查器
- EventSimulator - 事件模拟器
- HTTP 调试服务器 - 提供 Web 界面和 REST API

启动调试服务器：

```go
debugServer := debug.NewServer(root.GObject, stage, 8080)
if err := debugServer.Start(); err != nil {
    log.Printf("调试服务器启动失败: %v", err)
} else {
    log.Printf("调试服务器: %s", debugServer.GetURL())
}
```

## 依赖

- [Ebiten](https://github.com/hajimehoshi/ebiten/v2) - Go 2D 游戏引擎
- [x/image](https://github.com/golang/image) - Go 图像处理库




## 目标

FairyGUI Ebiten 项目的目标是提供一个高性能、功能丰富的 UI 框架，让 Go 开发者能够轻松创建具有复杂用户界面的游戏和应用程序。该项目特别注重与原版 FairyGUI 的兼容性，同时充分利用 Go 语言和 Ebiten 引擎的优势。