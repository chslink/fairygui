# 文本输入 IME 支持 — 调研与状态

> 日期: 2026-05-13
> 分支: `fix/text-rendering-issues`

## 当前实现路径

### 字符输入
```
ebiten.AppendInputChars(nil)
  → 每帧收集已提交字符（含 IME 输出）
  → widgets.InputChar(string)
  → GTextInput.InsertChars(s)
```

### 控制键（自实现）
```
keyboardEvents() → IsKeyPressed 采样
  → HandleKeyboardEvent
  → Backspace/Delete/←↑↓→/Home/End/Shift+方向键/Ctrl+C/V/X/A
```

### IME 模式启用（混合）
```
syncIMEField() → textinput.Field.Focus()/Blur()
  → 通知 Ebiten/GLFW 启用 IME 文本输入模式
  → 不调用 Field.HandleInput()（避免消费控制键）
```

### Windows IME 上下文
```
enableIME() → FindWindowW("Ebiten") → ImmAssociateContext(hwnd, hIMC)
  → 首次 Update 执行一次
```

---

## 已解决

| 问题 | 状态 | 方案 |
|------|------|------|
| 描边/阴影过厚 | ✅ | padding +1→+2 |
| 斜体左边缘裁剪 | ✅ | computeTextPadding 加 italic shear 补偿 |
| 缓存 key 无 italic/bold | ✅ | 追加到 key 中 |
| 链接 hover/点击无效 | ✅ | attachLinkHandler 时 mouseThrough=false |
| Placeholder 渲染 | ✅ | drawPlaceholderImage 强制 UBB 解析 |
| 控制键（Backspace/Delete/箭头） | ✅ | HandleKeyboardEvent 自实现 |
| Ctrl+C/V/X 剪贴板 | ✅ | 内部 clipboard 变量 |
| 光标位置 | ✅ | font.Face.GlyphAdvance 实测字形宽度 |

---

## 未完全解决 — IME 中文输入

### 现象
- 点击输入框后，系统输入法**可切换**
- 但切换后 IME 组合窗口**不稳定**：有时能弹出示，有时自动切回英文
- `ebiten.AppendInputChars` 在 IME 激活时能收到已提交字符
- IME 组合过程中的字符（预编辑串）无法渲染

### 根因分析

**Ebiten 是游戏引擎，不是原生 GUI 框架。** Windows IME 需要完整的消息循环支持：

| 所需支持 | Ebiten 状态 |
|---------|------------|
| `WM_IME_SETCONTEXT` | GLFW 部分处理 |
| `WM_IME_STARTCOMPOSITION` | GLFW 支持 |
| `WM_IME_COMPOSITION` | GLFW 支持 |
| `WM_IME_ENDCOMPOSITION` | GLFW 支持 |
| `ImmAssociateContext` | 我们自行调用 ✅ |
| IME 候选窗渲染 | 依赖系统 IME 窗口 |
| 预编辑串渲染 | `textinput.Field.TextForRendering()` |
| `glfwSetCharCallback` | Ebiten 内部使用 |
| `glfwSetCharModsCallback` | 不明确 |

**核心矛盾**：
1. `textinput.Field` 需要调用 `Field.HandleInput(x, y)` 每帧处理 IME 事件，但这会**消费所有键盘输入**（包括控制键）
2. 不调用 `HandleInput` 则 IME 组合事件无法被正确处理
3. Ebiten 的 `IsKeyPressed` 轮询方式在 IME 激活时可能产生干扰

### 尝试过的方案

| 方案 | 结果 |
|------|------|
| `textinput.Field` 全权管理 | ❌ 控制键失效 |
| `textinput.Field` 仅 Focus/Blur（当前） | ⚠️ IME 不完全稳定 |
| 纯 `AppendInputChars` 无 Field | ❌ IME 无法切换 |
| Windows `ImmAssociateContext` | ⚠️ 有改善但不彻底 |

### 可能的未来方向

1. **Ebiten v2.10+ 的 `Composer` API** — 正在开发中的新 IME API，可能彻底解决
2. **通过 CGO 直接调用 Windows IME API** — 完全绕过 Ebiten 的 IME 层
3. **使用 `textinput.Field` + 过滤控制键** — 在 `HandleInput` 后检查文字变化只同步文本，不解码控制键
4. **平台原生方案** — 在 Windows 上使用 `windows.SetWindowLongPtr` 子类化窗口过程直接处理 WM_IME_* 消息

---

## 修改文件清单

| 文件 | 修改 |
|------|------|
| `pkg/fgui/widgets/text_input.go` | **完全重写** — 自实现控制键 + AppendInputChars 字符 |
| `pkg/fgui/render/text_draw.go` | 缓存 key + 斜体补偿 + 描边 padding + drawPlaceholderImage |
| `pkg/fgui/widgets/text.go` | 链接时 mouseThrough=false |
| `pkg/fgui/render/draw_ebiten.go` | 光标渲染用 GlyphAdvance + placeholder UBB |
| `demo/main.go` | 移除 keyboardEvents 中的 AppendInputChars + syncIMEField + enableIME |
| `demo/ime_windows.go` | **新建** — Windows ImmAssociateContext 启用 IME |
| `demo/ime_other.go` | **新建** — 非 Windows 空桩 |
| `pkg/fgui/widgets/text_input_test.go` | 简化为 8 个测试 |
