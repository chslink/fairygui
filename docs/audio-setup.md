# FairyGUI 音频播放功能使用指南

## 概述

FairyGUI Go 版本现在支持按钮点击音效播放功能。基于 Ebiten 音频系统实现，支持 MP3、Wav、Ogg 等多种音频格式。

## 功能特性

- ✅ 支持按钮点击音效（从 UIConfig 继承默认设置）
- ✅ 支持多种音频格式（MP3、Wav、Ogg）
- ✅ 支持独立按钮音效覆盖
- ✅ 音量控制（0-1 范围）
- ✅ 异步播放，不阻塞 UI 渲染

## 快速开始

### 1. 在游戏初始化时设置音频系统

```go
package main

import (
    "github.com/chslink/fairygui/pkg/fgui"
)

func main() {
    // 1. 初始化音频系统
    fgui.InitAudio(48000) // 使用48000采样率

    // 2. 注册音频播放器为按钮音效播放器
    fgui.RegisterButtonSoundPlayer()

    // 3. 注册音效资源
    clickData, _ := os.ReadFile("button_click.wav")
    fgui.RegisterAudio("button_click", clickData)

    // 4. 设置全局默认按钮音效（可选）
    fgui.GetUIConfig().ButtonSound = "button_click"
    fgui.GetUIConfig().ButtonSoundVolumeScale = 0.8

    // ... 其他初始化代码
}
```

### 2. 创建带音效的按钮

#### 方式1：使用全局默认音效

```go
// 如果设置了全局默认音效，所有按钮都会使用
fgui.GetUIConfig().ButtonSound = "button_click"

// 按钮会自动使用全局音效
button := widgets.NewButton()
```

#### 方式2：为单个按钮设置音效

```go
button := widgets.NewButton()
button.SetSound("button_click")
button.SetSoundVolumeScale(0.5) // 50%音量
```

#### 方式3：从 UI 包中加载音效

```go
// 在 .fui 文件中为按钮配置音效，Go 版本会自动解析
button := widgets.NewButton()
// 通过 SetupBeforeAdd 读取音效配置
button.SetupBeforeAdd(buf, pos)
```

## 完整示例

```go
package main

import (
    "os"
    "log"

    "github.com/chslink/fairygui/pkg/fgui"
)

func setupAudio() {
    // 初始化音频系统
    fgui.InitAudio(48000)

    // 注册按钮音效播放器
    fgui.RegisterButtonSoundPlayer()

    // 加载并注册音效
    clickWav, err := os.ReadFile("assets/sounds/button_click.wav")
    if err != nil {
        log.Printf("Failed to load button click sound: %v", err)
        return
    }
    fgui.RegisterAudio("button_click", clickWav)

    // 设置全局默认按钮音效
    fgui.GetUIConfig().ButtonSound = "button_click"
    fgui.GetUIConfig().ButtonSoundVolumeScale = 0.8

    log.Println("Audio system initialized successfully")
}

func main() {
    // 初始化音频
    setupAudio()

    // ... 其他游戏初始化代码
}
```

## API 参考

### 核心 API

#### `fgui.InitAudio(sampleRate int)`
初始化音频系统。

**参数：**
- `sampleRate`: 采样率（默认 48000）

**示例：**
```go
fgui.InitAudio(48000)
```

#### `fgui.RegisterButtonSoundPlayer()`
将音频播放器注册为按钮音效播放器，使所有按钮点击都播放音效。

**示例：**
```go
fgui.RegisterButtonSoundPlayer()
```

#### `fgui.RegisterAudio(name string, data []byte)`
注册音频资源。

**参数：**
- `name`: 音频资源名称
- `data`: 音频字节数据（支持 MP3、Wav、Ogg 格式）

**示例：**
```go
data, _ := os.ReadFile("click.wav")
fgui.RegisterAudio("button_click", data)
```

#### `fgui.GetAudioPlayer() *audio.AudioPlayer`
获取音频播放器单例。

**返回值：**
- `*audio.AudioPlayer`: 音频播放器实例

**示例：**
```go
player := fgui.GetAudioPlayer()
player.RegisterAudioData("my_sound", soundData)
```

### UIConfig API

#### `GetUIConfig() *core.UIConfig`
获取全局 UI 配置实例。

**返回值：**
- `*core.UIConfig`: UI 配置实例

**示例：**
```go
config := fgui.GetUIConfig()
config.ButtonSound = "button_click"           // 设置默认按钮音效
config.ButtonSoundVolumeScale = 0.8          // 设置默认音量
```

### 按钮 API

#### `NewButton() *widgets.GButton`
创建新按钮。

**返回值：**
- `*widgets.GButton`: 按钮实例

**示例：**
```go
button := widgets.NewButton()
```

#### `SetSound(sound string)`
设置按钮音效。

**参数：**
- `sound`: 音效资源名称

**示例：**
```go
button.SetSound("button_click")
```

#### `SetSoundVolumeScale(volume float64)`
设置按钮音效音量。

**参数：**
- `volume`: 音量（0-1 范围）

**示例：**
```go
button.SetSoundVolumeScale(0.5) // 50% 音量
```

## 音频格式支持

支持的音频格式：
- **Wav**: 无损格式，播放速度快，文件较大
- **MP3**: 有损格式，压缩率高，文件小
- **Ogg**: 开源有损格式，压缩率高，文件小

## 最佳实践

### 1. 音效文件大小
- 按钮点击音效建议使用短音效（< 1秒）
- 使用 MP3 或 Ogg 格式以减小文件大小
- 保持采样率在 22050-48000 之间平衡质量和大小

### 2. 音量设置
- 默认音量建议设置为 0.3-0.5，避免过大
- 可以为不同按钮设置不同音量
- 避免所有音效都使用相同音量

### 3. 性能优化
- 预加载所有常用音效
- 音频缓存会自动管理内存
- 避免频繁播放长音效

## 故障排除

### 1. 没有声音
- 检查 `fgui.RegisterButtonSoundPlayer()` 是否调用
- 检查音频数据是否正确加载
- 检查系统音量是否开启
- 验证音频格式是否支持

### 2. 编译错误
- 确保 `go.mod` 中包含 `github.com/hajimehoshi/ebiten/v2`
- 运行 `go mod tidy` 更新依赖
- 检查是否导入了 `github.com/chslink/fairygui/pkg/fgui/audio`

### 3. 音频播放卡顿
- 检查音频文件大小是否过大
- 减少音频采样率
- 确保音频数据已预加载

## 集成到 Demo

在 `demo/main.go` 中已经集成了音频系统：

```go
func newGame(ctx context.Context) (*game, error) {
    // 初始化音频系统
    fgui.InitAudio(48000)
    fgui.RegisterButtonSoundPlayer()

    // 注册示例音效（需要加载真实的音频文件）
    // data, _ := os.ReadFile("path/to/click.wav")
    // fgui.RegisterAudio("button_click", data)

    // ... 其他代码
}
```

## 示例音效文件

在 `demo/assets/sounds/` 目录中放置音效文件：
```
demo/
  assets/
    sounds/
      button_click.wav    # 按钮点击音效
      button_hover.wav    # 按钮悬停音效（可选）
      ui_confirm.wav      # 确认音效（可选）
```

加载示例：
```go
clickData, _ := os.ReadFile("demo/assets/sounds/button_click.wav")
fgui.RegisterAudio("button_click", clickData)
```

## 许可证

音频播放功能基于以下开源项目：
- [Ebitengine](https://github.com/hajimehoshi/ebiten)
- [go-mp3](https://github.com/hajimehoshi/go-mp3)
- [go-oggvorbis](https://github.com/hajimehoshi/go-oggvorbis)
