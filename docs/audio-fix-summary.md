# 音效系统自动加载修复总结

## 问题分析

在检查 TypeScript 版本代码后，我们发现：

### TypeScript 版本的工作流程（GButton.ts:502-509）
```typescript
private __click(evt: Laya.Event): void {
    if (this._sound) {
        var pi: PackageItem = UIPackage.getItemByURL(this._sound);
        if (pi)
            GRoot.inst.playOneShotSound(pi.file);  // 使用 pi.file（文件路径）
        else
            GRoot.inst.playOneShotSound(this._sound);
    }
}
```

**关键点**：
1. 使用 `UIPackage.getItemByURL(this._sound)` 获取 PackageItem
2. 如果找到 PackageItem，使用 `pi.file`（文件路径，不是 URL）调用 `playOneShotSound`
3. `pi.file` 是 FUI 包中定义的音频文件路径（如 `Basics_gojg7u.wav`）

### 原始 Go 版本的问题
- `PlayOneShotSound` 接收文件路径（`Basics_gojg7u.wav`）
- `audio.Play()` 试图将文件路径当作 FairyGUI URL（`ui://...`）处理
- 导致无法正确加载和播放音效

## 修复方案

### 1. 添加全局资源加载器支持
**文件**: `pkg/fgui/audio/audio.go`

添加了：
- 全局资源加载器 `globalLoader`
- `SetLoader()` 函数
- 支持自动从包中加载音效数据

### 2. 修改 Play() 方法支持两种加载方式
**文件**: `pkg/fgui/audio/audio.go`

修改 `Play()` 方法，现在支持：
- **文件路径加载**（如 `Basics_gojg7u.wav`）
  - 直接使用文件路径加载音频文件
  - 用于按钮点击音效等场景
- **URL 加载**（如 `ui://9leh0eyf/gojg7u`）
  - 从 FUI 包中查找 PackageItem
  - 使用 PackageItem 的 File 字段加载音频文件
  - 用于自定义场景

新增方法：
- `tryDirectLoad()`: 直接使用文件路径加载
- `tryAutoLoadAndPlay()`: 从包中自动加载并播放

### 3. 导出公共 API
**文件**: `pkg/fgui/api.go`

导出 `SetAudioLoader()` 函数：
```go
func SetAudioLoader(loader assets.Loader)
```

### 4. 集成到 Demo
**文件**: `demo/main.go`

在创建 loader 后设置音频加载器：
```go
loader := fgui.NewFileLoader(assetsDir)
fgui.SetAudioLoader(loader)
```

## 测试验证

### 创建测试文件
**文件**: `pkg/fgui/audio/audio_autoload_test.go`

测试覆盖：
1. `TestAudioAutoLoad`: 测试音效自动加载功能
2. `TestAudioPlayFromPackage`: 测试从包中播放音效
3. `TestAudioPlayWithURL`: 测试使用 URL 播放音效

### 测试结果
```
=== RUN   TestAudioAutoLoad
--- PASS: TestAudioAutoLoad
=== RUN   TestAudioPlayFromPackage
--- PASS: TestAudioPlayFromPackage
=== RUN   TestAudioPlayWithURL
--- PASS: TestAudioPlayWithURL
PASS
```

### 发现的音效资源
- ID=gojg7u, Name=tabswitch, File=Basics_gojg7u.wav
- ID=o4lt7w, Name=click, File=Basics_o4lt7w.wav
- 音效 URL: ui://9leh0eyf/gojg7u

## 修复后的工作流程

### 场景 1: 按钮点击音效（GButton）
1. GButton 获取 UIConfig 中的 `buttonSound`（URL）
2. 使用 `GetItemByURL()` 查找 PackageItem
3. 使用 `pi.File`（文件路径）调用 `PlayOneShotSound()`
4. `audio.Play()` 识别文件路径，直接加载并播放
5. ✅ 音效正常播放

### 场景 2: 自定义音效播放
1. 用户直接调用 `player.Play("ui://package/item", 1.0)`
2. `audio.Play()` 识别为 FairyGUI URL
3. 从包中查找 PackageItem
4. 使用 PackageItem 的 File 字段加载音频文件
5. ✅ 音效正常播放

## 总结

通过这次修复，Go 版本的音效系统现在与 TypeScript 版本的行为完全一致：

- ✅ 按钮点击音效自动工作
- ✅ 支持文件路径和 URL 两种加载方式
- ✅ 自动从 FUI 包中加载音效数据
- ✅ 与 TypeScript 版本保持 API 兼容

## 重要文件修改

1. `pkg/fgui/audio/audio.go` - 核心音效加载逻辑
2. `pkg/fgui/api.go` - 公共 API 导出
3. `demo/main.go` - 集成到演示应用
4. `pkg/fgui/audio/audio_autoload_test.go` - 测试验证

## 使用示例

### 方式 1: 使用 UIConfig（推荐）
```go
// 在初始化时设置
fgui.SetDefaultButtonSound("ui://Basics/click")

// 设置音频加载器
fgui.SetAudioLoader(loader)

// 启动音频系统
fgui.InitAudio(48000)
fgui.RegisterButtonSoundPlayer()
```

### 方式 2: 手动注册（备用）
```go
// 读取音频文件
data, _ := os.ReadFile("demo/assets/Basics_o4lt7w.wav")

// 注册音频数据
fgui.RegisterAudio("ui://Basics/click", data)
```

方式 1 是推荐方式，因为它与 TypeScript 版本的行为一致，且更简洁。
