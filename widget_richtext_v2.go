package fairygui

import (
	"strings"

	textutil "github.com/chslink/fairygui/internal/text"
)

// ============================================================================
// RichTextField - 富文本控件 V2
// ============================================================================

type RichTextField struct {
	*TextField

	// UBB 解析相关
	ubbEnabled bool
	baseStyle  textutil.Style
	segments   []textutil.Segment
}

// NewRichTextField 创建新的富文本控件
func NewRichTextField() *RichTextField {
	rtf := &RichTextField{
		TextField:  NewTextField(),
		ubbEnabled: true, // 富文本默认启用 UBB
		baseStyle: textutil.Style{
			Color:     "#000000",
			FontSize:  12,
			Bold:      false,
			Italic:    false,
			Underline: false,
			Font:      "Arial",
		},
		segments: nil,
	}

	// 默认启用自动换行
	rtf.SetWordWrap(true)

	return rtf
}

// SetUBBEnabled 启用或禁用 UBB 解析
func (rtf *RichTextField) SetUBBEnabled(enabled bool) {
	if rtf.ubbEnabled == enabled {
		return
	}
	rtf.ubbEnabled = enabled
	rtf.parseSegments()
}

// UBBEnabled 返回是否启用了 UBB 解析
func (rtf *RichTextField) UBBEnabled() bool {
	return rtf.ubbEnabled
}

// SetText 设置文本（覆盖父类方法以支持 UBB 解析）
func (rtf *RichTextField) SetText(text string) {
	rtf.TextField.SetText(text)
	rtf.parseSegments()
}

// Text 返回文本（覆盖父类方法）
func (rtf *RichTextField) Text() string {
	return rtf.TextField.Text()
}

// SetBaseStyle 设置基础样式
func (rtf *RichTextField) SetBaseStyle(style textutil.Style) {
	rtf.baseStyle = style
	rtf.parseSegments()
}

// BaseStyle 返回基础样式
func (rtf *RichTextField) BaseStyle() textutil.Style {
	return rtf.baseStyle
}

// GetSegments 返回解析后的富文本段
// 如果 UBB 未启用，返回 nil
func (rtf *RichTextField) GetSegments() []textutil.Segment {
	if !rtf.ubbEnabled {
		return nil
	}
	return rtf.segments
}

// SetColor 设置文本颜色（覆盖父类并更新基础样式）
func (rtf *RichTextField) SetColor(color string) {
	rtf.TextField.SetColor(color)
	rtf.baseStyle.Color = color
	rtf.parseSegments()
}

// SetFontSize 设置字体大小（覆盖父类并更新基础样式）
func (rtf *RichTextField) SetFontSize(size int) {
	rtf.TextField.SetFontSize(size)
	rtf.baseStyle.FontSize = size
	rtf.parseSegments()
}

// SetFont 设置字体（覆盖父类并更新基础样式）
func (rtf *RichTextField) SetFont(font string) {
	rtf.TextField.SetFont(font)
	rtf.baseStyle.Font = font
	rtf.parseSegments()
}

// parseSegments 解析文本为富文本段
// 仅在 UBB 启用时有效
func (rtf *RichTextField) parseSegments() {
	if !rtf.ubbEnabled {
		rtf.segments = nil
		return
	}

	textStr := rtf.TextField.Text()
	if textStr == "" {
		rtf.segments = nil
		return
	}

	// 使用 UBB 解析器解析文本
	rtf.segments = textutil.ParseUBB(textStr, rtf.baseStyle)
}

// HasRichContent 检查是否包含富文本内容
// 如果文本包含 UBB 标签，返回 true
func (rtf *RichTextField) HasRichContent() bool {
	if !rtf.ubbEnabled || rtf.segments == nil {
		return false
	}

	// 如果只有一个段且没有特殊样式，则不是富文本
	if len(rtf.segments) == 0 {
		return false
	}

	// 检查是否有任何段具有特殊样式
	for _, seg := range rtf.segments {
		if seg.Style.Bold || seg.Style.Italic || seg.Style.Underline ||
			seg.Style.Color != rtf.baseStyle.Color ||
			seg.Style.FontSize != rtf.baseStyle.FontSize ||
			seg.Style.Font != rtf.baseStyle.Font ||
			seg.Link != "" || seg.ImageURL != "" {
			return true
		}
	}

	return false
}

// GetPlainText 返回去除 UBB 标签的纯文本
func (rtf *RichTextField) GetPlainText() string {
	if !rtf.ubbEnabled || rtf.segments == nil {
		return rtf.TextField.Text()
	}

	var plain strings.Builder
	for _, seg := range rtf.segments {
		plain.WriteString(seg.Text)
	}
	return plain.String()
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertRichTextField 类型断言
func AssertRichTextField(obj DisplayObject) (*RichTextField, bool) {
	rtf, ok := obj.(*RichTextField)
	return rtf, ok
}

// IsRichTextField 检查是否是 RichTextField
func IsRichTextField(obj DisplayObject) bool {
	_, ok := obj.(*RichTextField)
	return ok
}
