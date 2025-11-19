package fairygui

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// ============================================================================
// TextField - 文本控件 V2 (简化实现)
// ============================================================================

type TextField struct {
	*Object

	text            string
	color           string
	fontSize        int
	font            string
	bold            bool
	italic          bool
	underline       bool
	strokeColor     string
	strokeSize      int
	shadowColor     string
	shadowOffsetX   float64
	shadowOffsetY   float64
	shadowBlur      float64
	align           TextAlign
	verticalAlign   VerticalAlign
	autoSize        bool
	wordWrap        bool
}

// NewTextField 创建新的文本控件
func NewTextField() *TextField {
	tf := &TextField{
		Object:          NewObject(),
		color:           "#000000",
		font:            "Arial",
		fontSize:        12,
		align:           TextAlignLeft,
		verticalAlign:   VerticalAlignTop,
		autoSize:        true,
		wordWrap:        false,
		shadowBlur:      0,
		shadowOffsetX:   0,
		shadowOffsetY:   0,
	}

	// 默认不拦截事件
	tf.SetTouchable(false)

	return tf
}

// 文本设置
func (tf *TextField) SetText(text string) {
	if tf.text == text {
		return
	}
	tf.text = text
	tf.updateGraphics()
}

func (tf *TextField) Text() string {
	return tf.text
}

// 颜色
func (tf *TextField) SetColor(color string) {
	if tf.color == color {
		return
	}
	tf.color = color
	tf.updateGraphics()
}

func (tf *TextField) Color() string {
	return tf.color
}

// 字体大小
func (tf *TextField) SetFontSize(size int) {
	if tf.fontSize == size {
		return
	}
	tf.fontSize = size
	tf.updateGraphics()
}

func (tf *TextField) FontSize() int {
	return tf.fontSize
}

// 字体
func (tf *TextField) SetFont(font string) {
	if tf.font == font {
		return
	}
	tf.font = font
	tf.updateGraphics()
}

func (tf *TextField) Font() string {
	return tf.font
}

// 粗体
func (tf *TextField) SetBold(bold bool) {
	if tf.bold == bold {
		return
	}
	tf.bold = bold
	tf.updateGraphics()
}

func (tf *TextField) Bold() bool {
	return tf.bold
}

// 斜体
func (tf *TextField) SetItalic(italic bool) {
	if tf.italic == italic {
		return
	}
	tf.italic = italic
	tf.updateGraphics()
}

func (tf *TextField) Italic() bool {
	return tf.italic
}

// 下划线
func (tf *TextField) SetUnderline(underline bool) {
	if tf.underline == underline {
		return
	}
	tf.underline = underline
	tf.updateGraphics()
}

func (tf *TextField) Underline() bool {
	return tf.underline
}

// 描边
func (tf *TextField) SetStrokeColor(color string) {
	if tf.strokeColor == color {
		return
	}
	tf.strokeColor = color
	tf.updateGraphics()
}

func (tf *TextField) StrokeColor() string {
	return tf.strokeColor
}

func (tf *TextField) SetStrokeSize(size int) {
	if tf.strokeSize == size {
		return
	}
	tf.strokeSize = size
	tf.updateGraphics()
}

func (tf *TextField) StrokeSize() int {
	return tf.strokeSize
}

// 阴影
func (tf *TextField) SetShadowColor(color string) {
	if tf.shadowColor == color {
		return
	}
	tf.shadowColor = color
	tf.updateGraphics()
}

func (tf *TextField) ShadowColor() string {
	return tf.shadowColor
}

func (tf *TextField) SetShadowOffset(x, y float64) {
	if tf.shadowOffsetX == x && tf.shadowOffsetY == y {
		return
	}
	tf.shadowOffsetX = x
	tf.shadowOffsetY = y
	tf.updateGraphics()
}

func (tf *TextField) ShadowOffset() (x, y float64) {
	return tf.shadowOffsetX, tf.shadowOffsetY
}

func (tf *TextField) SetShadowBlur(blur float64) {
	if tf.shadowBlur == blur {
		return
	}
	tf.shadowBlur = blur
	tf.updateGraphics()
}

func (tf *TextField) ShadowBlur() float64 {
	return tf.shadowBlur
}

// 对齐
func (tf *TextField) SetAlign(align TextAlign) {
	if tf.align == align {
		return
	}
	tf.align = align
	tf.updateGraphics()
}

func (tf *TextField) Align() TextAlign {
	return tf.align
}

func (tf *TextField) SetVerticalAlign(align VerticalAlign) {
	if tf.verticalAlign == align {
		return
	}
	tf.verticalAlign = align
	tf.updateGraphics()
}

func (tf *TextField) VerticalAlign() VerticalAlign {
	return tf.verticalAlign
}

// 自动大小
func (tf *TextField) SetAutoSize(autoSize bool) {
	if tf.autoSize == autoSize {
		return
	}
	tf.autoSize = autoSize
	tf.updateGraphics()
}

func (tf *TextField) AutoSize() bool {
	return tf.autoSize
}

// 换行
func (tf *TextField) SetWordWrap(wordWrap bool) {
	if tf.wordWrap == wordWrap {
		return
	}
	tf.wordWrap = wordWrap
	tf.updateGraphics()
}

func (tf *TextField) WordWrap() bool {
	return tf.wordWrap
}

// ============================================================================
// Draw 绘制
// ============================================================================

func (tf *TextField) Draw(screen *ebiten.Image) {
	// TODO: 实现实际文本绘制
	// 这里简化，实际应该在渲染层处理
}

// updateGraphics 更新图形
func (tf *TextField) updateGraphics() {
	// 如果自动大小，则计算文本尺寸
	if tf.autoSize {
		// TODO: 计算文本尺寸并设置大小
	}
}

// ============================================================================
// 类型断言辅助函数
// ============================================================================

// AssertTextField 类型断言
func AssertTextField(obj DisplayObject) (*TextField, bool) {
	tf, ok := obj.(*TextField)
	return tf, ok
}

// IsTextField 检查是否是 TextField
func IsTextField(obj DisplayObject) bool {
	_, ok := obj.(*TextField)
	return ok
}

// ============================================================================
// VerticalAlign - 垂直对齐类型
// ============================================================================

type VerticalAlign int

const (
	VerticalAlignTop VerticalAlign = iota
	VerticalAlignMiddle
	VerticalAlignBottom
	VerticalAlignNone // Added VerticalAlignNone to maintain enum completeness
)

func (va VerticalAlign) String() string {
	switch va {
	case VerticalAlignTop:
		return "top"
	case VerticalAlignMiddle:
		return "middle"
	case VerticalAlignBottom:
		return "bottom"
	case VerticalAlignNone:
		return "none"
	default:
		return "none"
	}
}
