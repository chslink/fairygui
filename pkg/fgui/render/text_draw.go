//go:build ebiten

package render

import (
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	textutil "github.com/chslink/fairygui/internal/text"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

type renderedTextRun struct {
	text     string
	runes    []rune
	style    textutil.Style
	color    color.NRGBA
	width    float64
	ascent   float64
	descent  float64
	face     font.Face
	bitmap   *assets.BitmapFont
	fontSize int
}

func (r *renderedTextRun) hasGlyphs() bool {
	return len(r.runes) > 0 && (r.bitmap != nil || r.face != nil)
}

type renderedTextLine struct {
	runs     []*renderedTextRun
	width    float64
	ascent   float64
	descent  float64
	height   float64
	hasGlyph bool
}

func drawTextImage(target *ebiten.Image, geo ebiten.GeoM, field *widgets.GTextField, value string, alpha float64, width, height float64, atlas *AtlasManager) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	if field != nil && field.SingleLine() {
		value = strings.ReplaceAll(value, "\n", " ")
	}

	baseStyle, baseColor := deriveBaseStyle(field)
	var segments []textutil.Segment
	if field != nil && field.UBBEnabled() {
		segments = textutil.ParseUBB(value, baseStyle)
	} else {
		segments = []textutil.Segment{{Text: value, Style: baseStyle}}
	}
	if len(segments) == 0 {
		return nil
	}

	lines := splitSegmentsIntoLines(segments)
	letterSpacing := float64(0)
	leading := float64(0)
	align := widgets.TextAlignLeft
	valign := widgets.TextVerticalAlignTop
	if field != nil {
		letterSpacing = float64(field.LetterSpacing())
		leading = float64(field.Leading())
		align = field.Align()
		valign = field.VerticalAlign()
	}

	baseMetrics := resolveBaseMetrics(field)
	renderedLines := make([]*renderedTextLine, 0, len(lines))
	maxLineWidth := 0.0
	textHeight := 0.0

	for idx, parts := range lines {
		line := buildRenderedLine(parts, field, baseColor, baseMetrics, letterSpacing)
		renderedLines = append(renderedLines, line)
		if line.width > maxLineWidth {
			maxLineWidth = line.width
		}
		if idx > 0 {
			textHeight += leading
		}
		textHeight += line.height
	}

	paddingLeft, paddingRight, paddingTop, paddingBottom := computeTextPadding(field, renderedLines)

	contentWidth := maxLineWidth
	contentHeight := textHeight

	finalWidth := width
	if finalWidth <= 0 {
		finalWidth = contentWidth
	}
	requiredWidth := contentWidth + paddingLeft + paddingRight
	if finalWidth < requiredWidth {
		finalWidth = requiredWidth
	}

	finalHeight := height
	if finalHeight <= 0 {
		finalHeight = contentHeight
	}
	requiredHeight := contentHeight + paddingTop + paddingBottom
	if finalHeight < requiredHeight {
		finalHeight = requiredHeight
	}

	imgW := int(math.Ceil(finalWidth))
	imgH := int(math.Ceil(finalHeight))
	if imgW <= 0 {
		imgW = 1
	}
	if imgH <= 0 {
		imgH = 1
	}
	if field != nil {
		field.UpdateLayoutMetrics(finalWidth, finalHeight, contentWidth, contentHeight)
	}
	textImg := ebiten.NewImage(imgW, imgH)

	var strokeColor *color.NRGBA
	strokeSize := 0.0
	if field != nil {
		if c := parseColor(field.StrokeColor()); c != nil {
			cc := *c
			strokeColor = &cc
		}
		strokeSize = field.StrokeSize()
	}

	var shadowColor *color.NRGBA
	shadowOffsetX := 0.0
	shadowOffsetY := 0.0
	if field != nil {
		if c := parseColor(field.ShadowColor()); c != nil {
			cc := *c
			shadowColor = &cc
			shadowOffsetX, shadowOffsetY = field.ShadowOffset()
		}
	}

	availableWidth := finalWidth - paddingLeft - paddingRight
	if availableWidth < 0 {
		availableWidth = 0
	}
	availableHeight := finalHeight - paddingTop - paddingBottom
	if availableHeight < 0 {
		availableHeight = 0
	}

	contentOffsetY := 0.0
	switch valign {
	case widgets.TextVerticalAlignMiddle:
		contentOffsetY = (availableHeight - contentHeight) * 0.5
	case widgets.TextVerticalAlignBottom:
		contentOffsetY = availableHeight - contentHeight
	default:
		contentOffsetY = 0
	}
	if contentOffsetY < 0 {
		contentOffsetY = 0
	}
	cursorY := paddingTop + contentOffsetY

	for lineIndex, line := range renderedLines {
		lineStartX := paddingLeft
		switch align {
		case widgets.TextAlignCenter:
			lineStartX = paddingLeft + (availableWidth-line.width)*0.5
		case widgets.TextAlignRight:
			lineStartX = paddingLeft + (availableWidth - line.width)
		default:
			lineStartX = paddingLeft
		}
		if lineStartX < 0 {
			lineStartX = 0
		}

		lineTop := cursorY
		lineBaseline := lineTop + line.ascent
		cursorX := lineStartX
		prevHadGlyph := false

		for _, run := range line.runs {
			if run == nil {
				continue
			}
			if run.hasGlyphs() {
				if prevHadGlyph && letterSpacing != 0 {
					cursorX += letterSpacing
				}
				if run.bitmap != nil {
					if err := drawBitmapRun(textImg, run, cursorX, lineTop, letterSpacing, atlas); err != nil {
						return err
					}
				} else if run.face != nil {
					renderSystemRun(textImg, run, cursorX, lineBaseline, letterSpacing, strokeColor, strokeSize, shadowColor, shadowOffsetX, shadowOffsetY)
				}
				cursorX += run.width
				prevHadGlyph = true
			}
			if run.style.Underline && run.width > 0 {
				drawUnderline(textImg, cursorX-run.width, lineBaseline, run.width, run.fontSize, run.color)
			}
		}

		cursorY += line.height
		if lineIndex != len(renderedLines)-1 {
			cursorY += leading
		}
	}

	opts := &ebiten.DrawImageOptions{GeoM: geo}
	if alpha < 1 {
		opts.ColorM.Scale(1, 1, 1, alpha)
	}
	target.DrawImage(textImg, opts)
	return nil
}

func deriveBaseStyle(field *widgets.GTextField) (textutil.Style, color.NRGBA) {
	baseColor := color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	if field != nil {
		if c := parseColor(field.Color()); c != nil {
			baseColor = *c
		} else {
			baseColor = color.NRGBA{A: 0xff}
		}
	}
	style := textutil.Style{
		Color:     "#ffffff",
		Bold:      false,
		Italic:    false,
		Underline: false,
		Font:      "",
		FontSize:  12,
	}
	if field != nil {
		style.Color = field.Color()
		style.Bold = field.Bold()
		style.Italic = field.Italic()
		style.Underline = field.Underline()
		style.Font = field.Font()
		if field.FontSize() > 0 {
			style.FontSize = field.FontSize()
		}
	}
	return style, baseColor
}

func splitSegmentsIntoLines(segments []textutil.Segment) [][]textutil.Segment {
	if len(segments) == 0 {
		return [][]textutil.Segment{{}}
	}
	var lines [][]textutil.Segment
	current := make([]textutil.Segment, 0)
	for _, seg := range segments {
		if seg.Text == "" {
			continue
		}
		chunks := strings.Split(seg.Text, "\n")
		for idx, chunk := range chunks {
			if chunk != "" {
				current = append(current, textutil.Segment{Text: chunk, Style: seg.Style})
			}
			if idx != len(chunks)-1 {
				lines = append(lines, current)
				current = make([]textutil.Segment, 0)
			}
		}
	}
	lines = append(lines, current)
	if len(lines) == 0 {
		return [][]textutil.Segment{{}}
	}
	return lines
}

type baseMetrics struct {
	ascent     float64
	descent    float64
	lineHeight float64
}

func resolveBaseMetrics(field *widgets.GTextField) baseMetrics {
	size := 12
	if field != nil && field.FontSize() > 0 {
		size = field.FontSize()
	}
	face := fontFaceForSize(size)
	if face == nil {
		face = selectFontFace(field)
	}
	metrics := face.Metrics()
	ascent := float64(metrics.Ascent.Ceil())
	descent := float64(metrics.Descent.Ceil())
	if ascent <= 0 && descent <= 0 {
		ascent = float64(metrics.Height.Ceil())
	}
	if ascent <= 0 {
		ascent = 1
	}
	if descent < 0 {
		descent = 0
	}
	lineHeight := ascent + descent
	if lineHeight <= 0 {
		lineHeight = float64(metrics.Height.Ceil())
		if lineHeight <= 0 {
			lineHeight = 1
		}
	}
	return baseMetrics{
		ascent:     ascent,
		descent:    descent,
		lineHeight: lineHeight,
	}
}

func buildRenderedLine(segments []textutil.Segment, field *widgets.GTextField, baseColor color.NRGBA, base baseMetrics, letterSpacing float64) *renderedTextLine {
	line := &renderedTextLine{
		runs: make([]*renderedTextRun, 0, len(segments)),
	}
	prevHadGlyph := false
	for _, seg := range segments {
		run := buildRenderedRun(seg, field, baseColor, base, letterSpacing)
		if run == nil {
			continue
		}
		line.runs = append(line.runs, run)
		if run.hasGlyphs() {
			if prevHadGlyph && letterSpacing != 0 {
				line.width += letterSpacing
			}
			line.width += run.width
			prevHadGlyph = true
			line.hasGlyph = true
		}
		if run.ascent > line.ascent {
			line.ascent = run.ascent
		}
		if run.descent > line.descent {
			line.descent = run.descent
		}
	}
	if !line.hasGlyph {
		line.ascent = math.Max(line.ascent, base.ascent)
		line.descent = math.Max(line.descent, base.descent)
	}
	line.height = line.ascent + line.descent
	if line.height <= 0 {
		line.height = base.lineHeight
	}
	return line
}

func buildRenderedRun(seg textutil.Segment, field *widgets.GTextField, baseColor color.NRGBA, base baseMetrics, letterSpacing float64) *renderedTextRun {
	run := &renderedTextRun{
		text:  seg.Text,
		runes: []rune(seg.Text),
		style: seg.Style,
		color: baseColor,
	}
	if seg.Style.Color != "" {
		if c := parseColor(seg.Style.Color); c != nil {
			run.color = *c
		}
	}

	size := seg.Style.FontSize
	if size <= 0 && field != nil && field.FontSize() > 0 {
		size = field.FontSize()
	}
	if size <= 0 {
		size = 12
	}
	run.fontSize = size

	fontRef := seg.Style.Font
	if fontRef == "" && field != nil {
		fontRef = field.Font()
	}
	if font := lookupBitmapFont(fontRef); font != nil {
		run.bitmap = font
		run.width = bitmapLineWidth(run.runes, font, letterSpacing)
		run.ascent = font.LineHeight
		if run.ascent <= 0 {
			run.ascent = float64(run.fontSize)
		}
		run.descent = 0
		return run
	}

	face := fontFaceForSize(size)
	if face == nil {
		face = selectFontFace(field)
	}
	run.face = face
	metrics := face.Metrics()
	run.ascent = float64(metrics.Ascent.Ceil())
	run.descent = float64(metrics.Descent.Ceil())
	if run.ascent <= 0 && run.descent <= 0 {
		run.ascent = base.ascent
		run.descent = base.descent
	}
	if run.ascent <= 0 {
		run.ascent = base.ascent
	}
	if run.descent < 0 {
		run.descent = 0
	}

	if len(run.runes) == 0 {
		run.width = 0
		return run
	}

	width := 0.0
	for idx, r := range run.runes {
		if adv, ok := face.GlyphAdvance(r); ok {
			width += float64(adv) / 64.0
		} else {
			bounds, _, ok := face.GlyphBounds(r)
			if ok {
				width += float64(bounds.Max.X-bounds.Min.X) / 64.0
			} else {
				width += float64(text.BoundString(face, string(r)).Dx())
			}
		}
		if idx != len(run.runes)-1 {
			width += letterSpacing
		}
	}
	run.width = width
	return run
}

func computeTextPadding(field *widgets.GTextField, lines []*renderedTextLine) (left, right, top, bottom float64) {
	left, right, top, bottom = 0, 0, 0, 0
	if field == nil {
		return
	}
	if stroke := field.StrokeSize(); stroke > 0 {
		left = math.Max(left, stroke)
		right = math.Max(right, stroke)
		top = math.Max(top, stroke)
		bottom = math.Max(bottom, stroke)
	}
	if c := parseColor(field.ShadowColor()); c != nil {
		offX, offY := field.ShadowOffset()
		if offX < 0 {
			left = math.Max(left, -offX)
		} else if offX > 0 {
			right = math.Max(right, offX)
		}
		if offY < 0 {
			top = math.Max(top, -offY)
		} else if offY > 0 {
			bottom = math.Max(bottom, offY)
		}
	}
	return
}

func drawBitmapRun(dst *ebiten.Image, run *renderedTextRun, startX float64, lineTop float64, letterSpacing float64, atlas *AtlasManager) error {
	if run.bitmap == nil || len(run.runes) == 0 {
		return nil
	}
	font := run.bitmap
	cursor := startX
	for idx, r := range run.runes {
		glyph := font.Glyphs[r]
		advance := font.SpaceAdvance()
		if glyph != nil {
			advance = glyph.Advance
			if glyph.Item != nil {
				local := ebiten.GeoM{}
				local.Translate(cursor+glyph.OffsetX, lineTop+glyph.OffsetY)
				if err := drawPackageItem(dst, glyph.Item, local, atlas, 1); err != nil {
					return err
				}
			}
		}
		if idx != len(run.runes)-1 {
			cursor += advance + letterSpacing
		} else {
			cursor += advance
		}
	}
	return nil
}

func renderSystemRun(dst *ebiten.Image, run *renderedTextRun, startX float64, baseline float64, letterSpacing float64, strokeColor *color.NRGBA, strokeSize float64, shadowColor *color.NRGBA, shadowOffsetX, shadowOffsetY float64) {
	if run.face == nil || len(run.runes) == 0 {
		return
	}
	drawGlyphs := func(col color.NRGBA, offsetX, offsetY float64) {
		src := image.NewUniform(col)
		drawer := font.Drawer{
			Dst:  dst,
			Src:  src,
			Face: run.face,
		}
		x := startX + offsetX
		y := baseline + offsetY
		for idx, r := range run.runes {
			drawer.Dot = fixed.Point26_6{
				X: fixed.Int26_6(x * 64),
				Y: fixed.Int26_6(y * 64),
			}
			drawer.DrawString(string(r))
			if adv, ok := run.face.GlyphAdvance(r); ok {
				x += float64(adv) / 64.0
			} else {
				bounds, _, ok := run.face.GlyphBounds(r)
				if ok {
					x += float64(bounds.Max.X-bounds.Min.X) / 64.0
				} else {
					x += float64(text.BoundString(run.face, string(r)).Dx())
				}
			}
			if idx != len(run.runes)-1 {
				x += letterSpacing
			}
		}
	}

	if strokeColor != nil && strokeSize > 0 {
		radius := int(math.Ceil(strokeSize))
		for dx := -radius; dx <= radius; dx++ {
			for dy := -radius; dy <= radius; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				if math.Hypot(float64(dx), float64(dy)) > strokeSize+0.1 {
					continue
				}
				drawGlyphs(*strokeColor, float64(dx), float64(dy))
			}
		}
	}

	if shadowColor != nil && (shadowOffsetX != 0 || shadowOffsetY != 0) {
		drawGlyphs(*shadowColor, shadowOffsetX, shadowOffsetY)
	}

	drawGlyphs(run.color, 0, 0)
	if run.style.Bold {
		drawGlyphs(run.color, 0.6, 0)
	}
}

func drawUnderline(dst *ebiten.Image, startX, baseline, width float64, fontSize int, col color.NRGBA) {
	if width <= 0 {
		return
	}
	thickness := math.Max(1, float64(fontSize)/14)
	y := baseline + thickness
	vector.DrawFilledRect(dst, float32(startX), float32(y), float32(width), float32(thickness), col, true)
}

func bitmapLineWidth(runes []rune, font *assets.BitmapFont, letterSpacing float64) float64 {
	if len(runes) == 0 || font == nil {
		return 0
	}
	width := 0.0
	for idx, r := range runes {
		adv := font.SpaceAdvance()
		if glyph := font.Glyphs[r]; glyph != nil {
			adv = glyph.Advance
		}
		width += adv
		if idx != len(runes)-1 {
			width += letterSpacing
		}
	}
	return width
}
