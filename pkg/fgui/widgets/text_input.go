package widgets

import (
	"strings"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// KeyboardType 枚举常用输入类型。
type KeyboardType string

const (
	KeyboardTypeDefault KeyboardType = "text"
	KeyboardTypeNumber  KeyboardType = "number"
	KeyboardTypeURL     KeyboardType = "url"
)

// GTextInput 扩展 GTextField，支持输入配置。
type GTextInput struct {
	*GTextField
	password     bool
	keyboardType KeyboardType
	editable     bool
	maxLength    int
	promptText   string
	restrict     string

	// 光标和选择状态
	cursorPosition   int       // 光标位置(rune 索引)
	selectionStart   int       // 选择起始位置
	selectionEnd     int       // 选择结束位置
	cursorVisible    bool      // 光标是否可见(闪烁状态)
	lastCursorBlink  time.Time // 上次光标闪烁时间
	focused          bool      // 是否获得焦点
	cursorBlinkDelay float64   // 光标闪烁间隔(秒)

	// 实际文本(密码模式下与显示文本不同)
	actualText string
}

// NewTextInput 构建输入框。
func NewTextInput() *GTextInput {
	base := NewText()
	input := &GTextInput{
		GTextField:       base,
		editable:         true,
		keyboardType:     KeyboardTypeDefault,
		maxLength:        0,
		cursorBlinkDelay: 0.5, // 默认 500ms 闪烁间隔
		cursorVisible:    true,
		lastCursorBlink:  time.Now(),
	}
	// 默认为单行输入模式
	base.SetSingleLine(true)
	base.GObject.SetData(input)

	// 启用鼠标交互以支持点击和拖动
	if sprite := base.GObject.DisplayObject(); sprite != nil {
		sprite.SetMouseEnabled(true)
	}

	return input
}

// SetupBeforeAdd 解析输入框额外配置。
func (t *GTextInput) SetupBeforeAdd(ctx *SetupContext, buf *utils.ByteBuffer) {
	if t == nil {
		return
	}
	if t.GTextField != nil {
		t.GTextField.SetupBeforeAdd(ctx, buf)
	}
	if buf == nil {
		return
	}
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()
	if !buf.Seek(0, 4) {
		return
	}
	if prompt := buf.ReadS(); prompt != nil {
		t.SetPromptText(*prompt)
	}
	if restrict := buf.ReadS(); restrict != nil {
		t.SetRestrict(*restrict)
	}
	if buf.Remaining() >= 4 {
		if max := int(buf.ReadInt32()); max > 0 {
			t.SetMaxLength(max)
		}
	}
	if buf.Remaining() >= 4 {
		switch code := int(buf.ReadInt32()); code {
		case 4:
			t.SetKeyboardType(KeyboardTypeNumber)
		case 3:
			t.SetKeyboardType(KeyboardTypeURL)
		case 0:
			// keep default
		default:
			t.SetKeyboardType(KeyboardTypeDefault)
		}
	}
	if buf.Remaining() > 0 && buf.ReadBool() {
		t.SetPassword(true)
	}
}

// SetPassword 切换密码模式。
func (t *GTextInput) SetPassword(enabled bool) {
	t.password = enabled
}

// Password 返回是否密码模式。
func (t *GTextInput) Password() bool {
	return t.password
}

// SetKeyboardType 设置键盘类型。
func (t *GTextInput) SetKeyboardType(value KeyboardType) {
	if value == "" {
		value = KeyboardTypeDefault
	}
	t.keyboardType = value
}

// KeyboardType 返回键盘类型。
func (t *GTextInput) KeyboardType() KeyboardType {
	return t.keyboardType
}

// SetEditable 切换是否可编辑。
func (t *GTextInput) SetEditable(value bool) {
	t.editable = value
}

// Editable 返回是否可编辑。
func (t *GTextInput) Editable() bool {
	return t.editable
}

// SetMaxLength 设置最大字符数。
func (t *GTextInput) SetMaxLength(limit int) {
	if limit < 0 {
		limit = 0
	}
	t.maxLength = limit
}

// MaxLength 返回最大字符数。
func (t *GTextInput) MaxLength() int {
	return t.maxLength
}

// SetPromptText 设置占位文本。
func (t *GTextInput) SetPromptText(text string) {
	t.promptText = strings.TrimSpace(text)
}

// PromptText 返回占位文本。
func (t *GTextInput) PromptText() string {
	return t.promptText
}

// SetRestrict 设置输入过滤字符串。
func (t *GTextInput) SetRestrict(value string) {
	t.restrict = strings.TrimSpace(value)
}

// Restrict 返回输入过滤配置。
func (t *GTextInput) Restrict() string {
	return t.restrict
}

// ========== 光标和选择管理 ==========

// SetCursorPosition 设置光标位置。
func (t *GTextInput) SetCursorPosition(pos int) {
	if t == nil {
		return
	}
	text := t.Text()
	runes := []rune(text)
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}
	t.cursorPosition = pos
	t.cursorVisible = true
	t.lastCursorBlink = time.Now()
	// 清除选择
	t.selectionStart = pos
	t.selectionEnd = pos
}

// CursorPosition 返回当前光标位置。
func (t *GTextInput) CursorPosition() int {
	if t == nil {
		return 0
	}
	return t.cursorPosition
}

// SetSelection 设置选择区域。
func (t *GTextInput) SetSelection(start, end int) {
	if t == nil {
		return
	}
	text := t.Text()
	runes := []rune(text)
	maxPos := len(runes)

	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if start > maxPos {
		start = maxPos
	}
	if end > maxPos {
		end = maxPos
	}

	// 确保 start <= end
	if start > end {
		start, end = end, start
	}

	t.selectionStart = start
	t.selectionEnd = end
	t.cursorPosition = end
}

// GetSelection 返回选择区域。
func (t *GTextInput) GetSelection() (start, end int) {
	if t == nil {
		return 0, 0
	}
	return t.selectionStart, t.selectionEnd
}

// HasSelection 返回是否有选中文本。
func (t *GTextInput) HasSelection() bool {
	if t == nil {
		return false
	}
	return t.selectionStart != t.selectionEnd
}

// GetSelectedText 返回选中的文本。
func (t *GTextInput) GetSelectedText() string {
	if t == nil || !t.HasSelection() {
		return ""
	}
	text := t.Text()
	runes := []rune(text)
	if t.selectionStart >= len(runes) || t.selectionEnd > len(runes) {
		return ""
	}
	return string(runes[t.selectionStart:t.selectionEnd])
}

// SelectAll 选择全部文本。
func (t *GTextInput) SelectAll() {
	if t == nil {
		return
	}
	text := t.Text()
	t.SetSelection(0, len([]rune(text)))
}

// ClearSelection 清除选择。
func (t *GTextInput) ClearSelection() {
	if t == nil {
		return
	}
	pos := t.cursorPosition
	t.selectionStart = pos
	t.selectionEnd = pos
}

// RequestFocus 请求焦点。
func (t *GTextInput) RequestFocus() {
	if t == nil {
		return
	}
	t.focused = true
	t.cursorVisible = true
	t.lastCursorBlink = time.Now()

	// 通知 Stage 设置焦点到这个输入框的 DisplayObject
	if t.GTextField != nil && t.GTextField.GObject != nil {
		if sprite := t.GTextField.GObject.DisplayObject(); sprite != nil {
			// 通过 GRoot 获取 Stage
			if root := core.Root(); root != nil {
				if stage := root.Stage(); stage != nil {
					stage.SetFocus(sprite)
				}
			}
		}
	}

	// 注册键盘和鼠标事件监听器
	if t.GTextField != nil && t.GTextField.GObject != nil {
		if sprite := t.GTextField.GObject.DisplayObject(); sprite != nil {
			// 移除旧的监听器(如果有)
			sprite.Dispatcher().Off(laya.EventKeyDown, nil)
			sprite.Dispatcher().Off(laya.EventMouseDown, nil)

			// 添加键盘事件监听器
			sprite.Dispatcher().On(laya.EventKeyDown, func(evt laya.Event) {
				// 从 Event.Data 中获取 KeyboardEvent
				if keyEvt, ok := evt.Data.(laya.KeyboardEvent); ok {
					t.HandleKeyboardEvent(keyEvt)
				}
			})

			// 添加鼠标按下事件监听器
			sprite.Dispatcher().On(laya.EventMouseDown, func(evt laya.Event) {
				// 从 Event.Data 中获取鼠标事件
				if pointerEvt, ok := evt.Data.(laya.PointerEvent); ok {
					// 转换为本地坐标
					localX := pointerEvt.Position.X
					localY := pointerEvt.Position.Y
					if sprite := t.GTextField.GObject.DisplayObject(); sprite != nil {
						local := sprite.GlobalToLocal(laya.Point{X: localX, Y: localY})
						localX = local.X
						localY = local.Y
					}
					t.HandleMouseDown(localX, localY)
				}
			})
		}
	}
}

// LoseFocus 失去焦点。
func (t *GTextInput) LoseFocus() {
	if t == nil {
		return
	}
	t.focused = false
	t.cursorVisible = false
	t.ClearSelection()

	// 移除键盘和鼠标事件监听器
	if t.GTextField != nil && t.GTextField.GObject != nil {
		if sprite := t.GTextField.GObject.DisplayObject(); sprite != nil {
			sprite.Dispatcher().Off(laya.EventKeyDown, nil)
			sprite.Dispatcher().Off(laya.EventMouseDown, nil)
		}
	}
}

// IsFocused 返回是否获得焦点。
func (t *GTextInput) IsFocused() bool {
	if t == nil {
		return false
	}
	return t.focused
}

// UpdateCursor 更新光标闪烁状态(应在每帧调用)。
func (t *GTextInput) UpdateCursor(deltaTime float64) {
	if t == nil || !t.focused {
		return
	}

	elapsed := time.Since(t.lastCursorBlink).Seconds()
	if elapsed >= t.cursorBlinkDelay {
		t.cursorVisible = !t.cursorVisible
		t.lastCursorBlink = time.Now()
	}
}

// IsCursorVisible 返回光标当前是否可见。
func (t *GTextInput) IsCursorVisible() bool {
	if t == nil {
		return false
	}
	return t.focused && t.cursorVisible
}

// ========== 文本编辑操作 ==========

// InsertText 在光标位置插入文本。
func (t *GTextInput) InsertText(text string) {
	if t == nil || !t.editable || text == "" {
		return
	}

	// 应用字符过滤
	filtered := t.filterText(text)
	if filtered == "" {
		return
	}

	currentText := t.Text()
	runes := []rune(currentText)

	// 如果有选中文本,先删除
	if t.HasSelection() {
		t.DeleteSelection()
		currentText = t.Text()
		runes = []rune(currentText)
	}

	// 插入新文本
	pos := t.cursorPosition
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}

	newRunes := make([]rune, 0, len(runes)+len(filtered))
	newRunes = append(newRunes, runes[:pos]...)
	newRunes = append(newRunes, []rune(filtered)...)
	newRunes = append(newRunes, runes[pos:]...)

	newText := string(newRunes)

	// 应用 maxLength 限制
	if t.maxLength > 0 && len([]rune(newText)) > t.maxLength {
		return
	}

	t.SetText(newText)
	t.SetCursorPosition(pos + len([]rune(filtered)))
}

// DeleteSelection 删除选中的文本。
func (t *GTextInput) DeleteSelection() {
	if t == nil || !t.editable || !t.HasSelection() {
		return
	}

	text := t.Text()
	runes := []rune(text)

	start, end := t.selectionStart, t.selectionEnd
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}

	newRunes := make([]rune, 0, len(runes)-(end-start))
	newRunes = append(newRunes, runes[:start]...)
	newRunes = append(newRunes, runes[end:]...)

	t.SetText(string(newRunes))
	t.SetCursorPosition(start)
}

// Backspace 删除光标前一个字符。
func (t *GTextInput) Backspace() {
	if t == nil || !t.editable {
		return
	}

	// 如果有选中,删除选中内容
	if t.HasSelection() {
		t.DeleteSelection()
		return
	}

	text := t.Text()
	runes := []rune(text)
	pos := t.cursorPosition

	if pos <= 0 || pos > len(runes) {
		return
	}

	newRunes := make([]rune, 0, len(runes)-1)
	newRunes = append(newRunes, runes[:pos-1]...)
	newRunes = append(newRunes, runes[pos:]...)

	t.SetText(string(newRunes))
	t.SetCursorPosition(pos - 1)
}

// Delete 删除光标后一个字符。
func (t *GTextInput) Delete() {
	if t == nil || !t.editable {
		return
	}

	// 如果有选中,删除选中内容
	if t.HasSelection() {
		t.DeleteSelection()
		return
	}

	text := t.Text()
	runes := []rune(text)
	pos := t.cursorPosition

	if pos < 0 || pos >= len(runes) {
		return
	}

	newRunes := make([]rune, 0, len(runes)-1)
	newRunes = append(newRunes, runes[:pos]...)
	newRunes = append(newRunes, runes[pos+1:]...)

	t.SetText(string(newRunes))
	// 光标位置不变
}

// MoveCursor 移动光标。
func (t *GTextInput) MoveCursor(delta int) {
	if t == nil {
		return
	}

	newPos := t.cursorPosition + delta
	t.SetCursorPosition(newPos)
}

// MoveCursorToStart 移动光标到开头。
func (t *GTextInput) MoveCursorToStart() {
	if t == nil {
		return
	}
	t.SetCursorPosition(0)
}

// MoveCursorToEnd 移动光标到末尾。
func (t *GTextInput) MoveCursorToEnd() {
	if t == nil {
		return
	}
	text := t.Text()
	t.SetCursorPosition(len([]rune(text)))
}

// filterText 应用 restrict 过滤规则。
func (t *GTextInput) filterText(text string) string {
	if t == nil || t.restrict == "" {
		return text
	}

	// 简化版过滤:只保留 restrict 中包含的字符
	// 完整实现应该支持范围(a-z)和否定(^)等
	allowed := []rune(t.restrict)
	allowedMap := make(map[rune]bool)
	for _, r := range allowed {
		allowedMap[r] = true
	}

	filtered := make([]rune, 0, len(text))
	for _, r := range text {
		if allowedMap[r] {
			filtered = append(filtered, r)
		}
	}

	return string(filtered)
}

// ========== 键盘事件处理 ==========

// HandleKeyboardEvent 处理键盘事件。
func (t *GTextInput) HandleKeyboardEvent(event laya.KeyboardEvent) bool {
	if t == nil || !t.focused || !t.editable {
		return false
	}

	// 只处理按键按下事件
	if !event.Down {
		return false
	}

	// 处理快捷键
	if event.Modifiers.Ctrl || event.Modifiers.Meta {
		return t.handleShortcut(event)
	}

	// 处理特殊键
	switch event.Code {
	case laya.KeyCodeBackspace:
		t.Backspace()
		return true

	case laya.KeyCodeDelete:
		t.Delete()
		return true

	case laya.KeyCodeLeft:
		if event.Modifiers.Shift {
			// Shift+Left: 向左扩展选择
			t.extendSelectionLeft()
		} else {
			t.MoveCursor(-1)
		}
		return true

	case laya.KeyCodeRight:
		if event.Modifiers.Shift {
			// Shift+Right: 向右扩展选择
			t.extendSelectionRight()
		} else {
			t.MoveCursor(1)
		}
		return true

	case laya.KeyCodeHome:
		if event.Modifiers.Shift {
			// Shift+Home: 选择到行首
			t.extendSelectionToStart()
		} else {
			t.MoveCursorToStart()
		}
		return true

	case laya.KeyCodeEnd:
		if event.Modifiers.Shift {
			// Shift+End: 选择到行尾
			t.extendSelectionToEnd()
		} else {
			t.MoveCursorToEnd()
		}
		return true

	case laya.KeyCodeEnter:
		// 单行模式忽略回车
		if t.SingleLine() {
			return true
		}
		t.InsertText("\n")
		return true

	case laya.KeyCodeTab:
		// Tab 键插入制表符(如果允许)
		if !t.SingleLine() {
			t.InsertText("\t")
			return true
		}
		return false // 单行模式下 Tab 用于焦点切换

	default:
		// 处理普通字符输入（包括中文等 Unicode 字符）
		// 排除控制字符（< 32），但允许所有可打印字符
		if event.Rune != 0 && !isControlChar(event.Rune) {
			t.InsertText(string(event.Rune))
			return true
		}
	}

	return false
}

// isControlChar 判断是否为控制字符
func isControlChar(r rune) bool {
	// ASCII 控制字符 (0-31) 和 DEL (127)
	if r < 32 || r == 127 {
		return true
	}
	// Unicode 控制字符范围
	// 0x80-0x9F: C1 控制字符
	// 0x200B-0x200F: 零宽度字符
	// 0x202A-0x202E: 双向文本控制字符
	if (r >= 0x80 && r <= 0x9F) ||
		(r >= 0x200B && r <= 0x200F) ||
		(r >= 0x202A && r <= 0x202E) {
		return true
	}
	return false
}

// handleShortcut 处理快捷键。
func (t *GTextInput) handleShortcut(event laya.KeyboardEvent) bool {
	switch event.Code {
	case laya.KeyCodeA:
		// Ctrl+A: 全选
		t.SelectAll()
		return true

	case laya.KeyCodeC:
		// Ctrl+C: 复制 (暂时只返回 true,实际复制需要剪贴板支持)
		// TODO: 实现剪贴板复制
		return t.HasSelection()

	case laya.KeyCodeX:
		// Ctrl+X: 剪切
		if t.HasSelection() {
			// TODO: 复制到剪贴板
			t.DeleteSelection()
			return true
		}
		return false

	case laya.KeyCodeV:
		// Ctrl+V: 粘贴
		// TODO: 从剪贴板粘贴
		return true

	case laya.KeyCodeZ:
		// Ctrl+Z: 撤销
		// TODO: 实现撤销/重做
		return true
	}

	return false
}

// extendSelectionLeft 向左扩展选择。
func (t *GTextInput) extendSelectionLeft() {
	if t == nil {
		return
	}

	if !t.HasSelection() {
		// 开始新选择
		t.selectionStart = t.cursorPosition
		t.selectionEnd = t.cursorPosition
	}

	newPos := t.cursorPosition - 1
	if newPos < 0 {
		newPos = 0
	}

	// 更新选择范围
	if t.cursorPosition == t.selectionEnd {
		// 光标在选择末尾,向左移动末尾
		t.selectionEnd = newPos
	} else {
		// 光标在选择开头,向左移动开头
		t.selectionStart = newPos
	}

	t.cursorPosition = newPos

	// 规范化选择
	if t.selectionStart > t.selectionEnd {
		t.selectionStart, t.selectionEnd = t.selectionEnd, t.selectionStart
	}
}

// extendSelectionRight 向右扩展选择。
func (t *GTextInput) extendSelectionRight() {
	if t == nil {
		return
	}

	text := t.Text()
	maxPos := len([]rune(text))

	if !t.HasSelection() {
		t.selectionStart = t.cursorPosition
		t.selectionEnd = t.cursorPosition
	}

	newPos := t.cursorPosition + 1
	if newPos > maxPos {
		newPos = maxPos
	}

	if t.cursorPosition == t.selectionEnd {
		t.selectionEnd = newPos
	} else {
		t.selectionStart = newPos
	}

	t.cursorPosition = newPos

	if t.selectionStart > t.selectionEnd {
		t.selectionStart, t.selectionEnd = t.selectionEnd, t.selectionStart
	}
}

// extendSelectionToStart 扩展选择到开头。
func (t *GTextInput) extendSelectionToStart() {
	if t == nil {
		return
	}

	if !t.HasSelection() {
		t.selectionEnd = t.cursorPosition
	}

	t.selectionStart = 0
	t.cursorPosition = 0
}

// extendSelectionToEnd 扩展选择到末尾。
func (t *GTextInput) extendSelectionToEnd() {
	if t == nil {
		return
	}

	text := t.Text()
	maxPos := len([]rune(text))

	if !t.HasSelection() {
		t.selectionStart = t.cursorPosition
	}

	t.selectionEnd = maxPos
	t.cursorPosition = maxPos
}

// ========== 鼠标交互 ==========

// CharPosInfo 描述字符位置信息。
type CharPosInfo struct {
	CharIndex int     // 字符索引(rune)
	X         float64 // 字符起始 X 坐标
	Y         float64 // 字符起始 Y 坐标
	Width     float64 // 字符宽度
	Height    float64 // 字符高度
}

// GetCharPositionAtPoint 根据屏幕坐标获取最接近的字符位置。
// 返回字符索引(用于定位光标)和是否在文本区域内。
func (t *GTextInput) GetCharPositionAtPoint(x, y float64) (int, bool) {
	if t == nil || t.GTextField == nil {
		return 0, false
	}

	text := t.Text()
	if text == "" {
		return 0, true
	}

	runes := []rune(text)

	// 简化实现：仅支持单行文本的光标定位
	// 对于单行文本，Y 坐标不影响字符定位
	if !t.SingleLine() {
		// 多行文本暂不支持精确定位，返回最接近的开头或结尾
		if x < 0 {
			return 0, true
		}
		return len(runes), true
	}

	// 处理点击在文本开头之前的情况
	if x <= 0 {
		return 0, true
	}

	// 获取文本度量信息
	// 注意：这是一个简化实现，假设等宽字体或近似计算
	// 真实实现需要访问渲染层的字形位置信息
	letterSpacing := float64(t.LetterSpacing())

	// 估算字符宽度(基于字体大小)
	fontSize := float64(t.FontSize())
	if fontSize <= 0 {
		fontSize = 12
	}
	avgCharWidth := fontSize * 0.6 // 近似平均字符宽度

	// 遍历字符，找到最接近的位置
	currentX := 0.0
	for i := 0; i < len(runes); i++ {
		charWidth := avgCharWidth
		// 空格字符通常较窄
		if runes[i] == ' ' {
			charWidth = fontSize * 0.3
		}

		// 计算字符中点位置
		midPoint := currentX + charWidth/2

		// 如果点击位置在当前字符的左半部分，光标应在字符前
		if x < midPoint {
			return i, true
		}

		// 移动到下一个字符位置
		currentX += charWidth
		if i < len(runes)-1 {
			currentX += letterSpacing
		}
	}

	// 点击在文本结束后，光标应在末尾
	return len(runes), true
}

// HandleMouseDown 处理鼠标按下事件。
// 用于定位光标和开始文本选择。
func (t *GTextInput) HandleMouseDown(x, y float64) bool {
	if t == nil || !t.editable {
		return false
	}

	// 请求焦点
	t.RequestFocus()

	// 获取点击位置对应的字符索引
	charIndex, inBounds := t.GetCharPositionAtPoint(x, y)
	if !inBounds {
		return false
	}

	// 设置光标位置
	t.SetCursorPosition(charIndex)

	// 清除之前的选择
	t.ClearSelection()

	return true
}

// HandleMouseDrag 处理鼠标拖动事件。
// 用于拖动选择文本。
func (t *GTextInput) HandleMouseDrag(x, y float64) bool {
	if t == nil || !t.focused || !t.editable {
		return false
	}

	// 获取拖动位置对应的字符索引
	charIndex, inBounds := t.GetCharPositionAtPoint(x, y)
	if !inBounds {
		return false
	}

	// 如果还没有选择，从当前光标位置开始
	if !t.HasSelection() {
		t.selectionStart = t.cursorPosition
	}

	// 更新选择结束位置
	t.selectionEnd = charIndex
	t.cursorPosition = charIndex

	// 规范化选择(确保 start <= end)
	if t.selectionStart > t.selectionEnd {
		t.selectionStart, t.selectionEnd = t.selectionEnd, t.selectionStart
	}

	return true
}

// HandleMouseDoubleClick 处理鼠标双击事件。
// 双击选择整个单词。
func (t *GTextInput) HandleMouseDoubleClick(x, y float64) bool {
	if t == nil || !t.editable {
		return false
	}

	// 请求焦点
	t.RequestFocus()

	// 获取点击位置对应的字符索引
	charIndex, inBounds := t.GetCharPositionAtPoint(x, y)
	if !inBounds {
		return false
	}

	// 选择点击位置的单词
	t.selectWordAt(charIndex)
	return true
}

// selectWordAt 选择指定位置的单词。
// 单词边界由空白字符定义。
func (t *GTextInput) selectWordAt(pos int) {
	if t == nil {
		return
	}

	text := t.Text()
	runes := []rune(text)

	if pos < 0 || pos >= len(runes) {
		return
	}

	// 向前查找单词边界
	start := pos
	for start > 0 && !isWordBoundary(runes[start-1]) {
		start--
	}

	// 向后查找单词边界
	end := pos
	for end < len(runes) && !isWordBoundary(runes[end]) {
		end++
	}

	// 设置选择
	if start < end {
		t.SetSelection(start, end)
	}
}

// isWordBoundary 判断字符是否为单词边界(空白字符)。
func isWordBoundary(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
