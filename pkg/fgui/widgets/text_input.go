package widgets

import (
	"strings"

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
}

// NewTextInput 构建输入框。
func NewTextInput() *GTextInput {
	base := NewText()
	input := &GTextInput{
		GTextField:   base,
		editable:     true,
		keyboardType: KeyboardTypeDefault,
		maxLength:    0,
	}
	base.GObject.SetData(input)
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
