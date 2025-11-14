package render

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"strings"
	"sync"
	"unicode"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/rivo/uniseg"

	textutil "github.com/chslink/fairygui/internal/text"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// textImageCache 缓存渲染后的文本图像，避免每帧重建
var (
	textImageCache   = make(map[string]*ebiten.Image)
	textImageCacheMu sync.RWMutex
)

// graphemeCache 缓存字素集群分析结果，提高性能
var (
	graphemeCacheMu sync.RWMutex
	graphemeCache   = make(map[string][][]byte) // key: string, value: slice of grapheme clusters
)

// getGraphemeClusters 使用 uniseg 分析字符串的字素集群，结果会被缓存
func getGraphemeClusters(s string) [][]byte {
	if s == "" {
		return nil
	}

	graphemeCacheMu.RLock()
	if clusters, ok := graphemeCache[s]; ok {
		graphemeCacheMu.RUnlock()
		return clusters
	}
	graphemeCacheMu.RUnlock()

	var clusters [][]byte
	gr := uniseg.NewGraphemes(s)
	for gr.Next() {
		clusters = append(clusters, append([]byte(nil), gr.Bytes()...))
	}

	graphemeCacheMu.Lock()
	graphemeCache[s] = clusters
	graphemeCacheMu.Unlock()

	return clusters
}

// countGraphemeClusters 返回字符串中的字素集群数量
func countGraphemeClusters(s string) int {
	return uniseg.GraphemeClusterCount(s)
}

// graphemeWidthAt 计算字素集群在指定索引处的宽度
func graphemeWidthAt(r *renderedTextRun, clusterIdx int, clusters [][]byte) float64 {
	if r == nil || clusterIdx < 0 || clusterIdx >= len(clusters) {
		return 0
	}

	// 将集群字节转换为 rune
	clusterRunes := []rune(string(clusters[clusterIdx]))

	// 计算所有字符的总宽度
	width := 0.0
	for _, rn := range clusterRunes {
		width += r.advanceForRune(rn)
	}
	return width
}

// advanceForRune 为单个 rune 计算宽度
func (r *renderedTextRun) advanceForRune(rn rune) float64 {
	if r.bitmap != nil {
		if glyph := r.bitmap.Glyphs[rn]; glyph != nil {
			return glyph.Advance
		}
		return r.bitmap.SpaceAdvance()
	}
	if r.face != nil {
		if adv, ok := r.face.GlyphAdvance(rn); ok {
			return float64(adv) / 64.0
		}
		if rn == ' ' {
			return float64(r.fontSize) * 0.5
		}
		return float64(r.fontSize) * 0.6
	}
	return 0
}

type renderedTextRun struct {
	text      string
	runes     []rune
	style     textutil.Style
	color     color.NRGBA
	link      string
	imageURL  string              // 图片 URL (用于 [img] 标签)
	imageItem *assets.PackageItem // 解析后的图片资源
	advances  []float64
	width     float64
	ascent    float64
	descent   float64
	face      font.Face
	bitmap    *assets.BitmapFont
	fontSize  int
}

func (r *renderedTextRun) hasGlyphs() bool {
	return (len(r.runes) > 0 && (r.bitmap != nil || r.face != nil)) || r.imageItem != nil
}

type renderedTextLine struct {
	runs     []*renderedTextRun
	width    float64
	ascent   float64
	descent  float64
	height   float64
	hasGlyph bool
}

type textPart struct {
	run         *renderedTextRun
	forcedBreak bool
}

func (r *renderedTextRun) spanWidth(start, end int, letterSpacing float64) float64 {
	if r == nil || start >= end || start < 0 || end > len([]rune(r.text)) {
		return 0
	}

	// 使用 uniseg 获取 grapheme 集群
	graphemes := getGraphemeClusters(r.text)
	if len(graphemes) == 0 {
		return 0
	}

	// 确保索引在有效范围内
	if start >= len(graphemes) {
		start = len(graphemes) - 1
	}
	if end > len(graphemes) {
		end = len(graphemes)
	}

	width := 0.0
	for i := start; i < end; i++ {
		width += graphemeWidthAt(r, i, graphemes)
		if i != end-1 {
			width += letterSpacing
		}
	}
	return width
}

// spanRuneWidth 按 rune 索引计算宽度（保留以兼容现有代码）
func (r *renderedTextRun) spanRuneWidth(start, end int, letterSpacing float64) float64 {
	if r == nil || start >= end || start < 0 || end > len(r.runes) {
		return 0
	}
	width := 0.0
	for i := start; i < end; i++ {
		width += r.advanceAt(i)
		if i != end-1 {
			width += letterSpacing
		}
	}
	return width
}

func (r *renderedTextRun) spanForWidth(start int, maxWidth float64, letterSpacing float64) (int, float64) {
	if r == nil || start < 0 || start >= len(r.runes) {
		return start, 0
	}

	// 使用 uniseg 进行正确的字素集群分割
	graphemes := getGraphemeClusters(r.text[start:])
	if len(graphemes) == 0 {
		return start, 0
	}

	width := 0.0
	lastBreak := -1
	widthAtBreak := 0.0

	for i := 0; i < len(graphemes); i++ {
		clusterBytes := graphemes[i]
		clusterRunes := []rune(string(clusterBytes))

		// 检查是否是换行符
		for _, rn := range clusterRunes {
			if isBreakRune(rn) {
				lastBreak = i
				widthAtBreak = width
				break
			}
		}

		// 计算当前字素集群的宽度
		clusterWidth := 0.0
		for _, rn := range clusterRunes {
			clusterWidth += r.advanceForRune(rn)
		}

		if width > 0 {
			clusterWidth += letterSpacing
		}

		// 检查是否超出最大宽度
		if maxWidth > 0 && width+clusterWidth > maxWidth {
			if lastBreak >= 0 {
				return start + lastBreak + 1, widthAtBreak
			}
			if width == 0 {
				return start + i + 1, width + clusterWidth
			}
			return start + i, width
		}

		width += clusterWidth
	}

	return start + len(graphemes), width
}

func (r *renderedTextRun) advanceAt(idx int) float64 {
	if r == nil || idx < 0 || idx >= len(r.runes) {
		return 0
	}
	if r.advances != nil && idx < len(r.advances) {
		return r.advances[idx]
	}
	if r.bitmap != nil {
		if glyph := r.bitmap.Glyphs[r.runes[idx]]; glyph != nil {
			return glyph.Advance
		}
		return r.bitmap.SpaceAdvance()
	}
	if r.face != nil {
		// Ebiten 的标准方法：优先使用 GlyphAdvance
		if adv, ok := r.face.GlyphAdvance(r.runes[idx]); ok {
			return float64(adv) / 64.0 // 从26.6固定点转换为像素
		}
		// 备选：使用字形边界框
		bounds, _, ok := r.face.GlyphBounds(r.runes[idx])
		if ok {
			return float64(bounds.Max.X-bounds.Min.X) / 64.0
		}
		// 最后备选：对于空格字符，使用字体大小的50%
		if r.runes[idx] == ' ' {
			return float64(r.fontSize) * 0.5
		}
		// 其他字符使用字体大小
		return float64(r.fontSize) * 0.6
	}
	return 0
}

func (r *renderedTextRun) slice(start, end int, letterSpacing float64) *renderedTextRun {
	if r == nil || start >= end || start < 0 {
		return nil
	}

	// 使用 grapheme 集群进行分割
	graphemes := getGraphemeClusters(r.text)
	if len(graphemes) == 0 {
		return nil
	}

	// 确保 end 不超过集群数量
	if end > len(graphemes) {
		end = len(graphemes)
	}

	clone := *r

	// 提取对应的 grapheme 集群
	var selectedGraphemes [][]byte
	var selectedRunes []rune
	for i := start; i < end; i++ {
		selectedGraphemes = append(selectedGraphemes, graphemes[i])
		selectedRunes = append(selectedRunes, []rune(string(graphemes[i]))...)
	}

	// 构建新的文本和 runes
	var sb strings.Builder
	for _, g := range selectedGraphemes {
		sb.Write(g)
	}
	clone.text = sb.String()
	clone.runes = selectedRunes

	if len(r.advances) > 0 {
		// 如果有预计算的 advances，需要重新计算
		clone.advances = nil
	} else {
		clone.advances = nil
	}

	clone.width = r.spanWidth(start, end, letterSpacing)
	return &clone
}

func drawTextImage(target *ebiten.Image, geo ebiten.GeoM, field *widgets.GTextField, value string, alpha float64, width, height float64, atlas *AtlasManager, sprite *laya.Sprite) error {
	var linkRegions []widgets.TextLinkRegion
	if field != nil {
		defer func() {
			field.SetLinkRegions(linkRegions)
		}()
	}
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

	letterSpacing := float64(0)
	leading := float64(0)
	align := widgets.TextAlignLeft
	valign := widgets.TextVerticalAlignTop
	allowWrap := false
	if field != nil {
		letterSpacing = float64(field.LetterSpacing())
		leading = float64(field.Leading())
		align = field.Align()
		valign = field.VerticalAlign()

		// Laya 行为：是否换行仅由 widthAutoSize / singleLine 决定
		allowWrap = !field.WidthAutoSize() && !field.SingleLine()
	}

	baseMetrics := resolveBaseMetrics(field)
	parts := buildTextParts(segments, field, baseColor, baseMetrics, letterSpacing)
	wrapWidth := width
	if wrapWidth <= 0 && field != nil {
		wrapWidth = field.Width()
	}
	if wrapWidth > 0 {
		padLeft, padRight := estimateHorizontalPadding(field)
		wrapWidth -= padLeft + padRight
		if wrapWidth < 0 {
			wrapWidth = 0
		}
	}

	wrapped := wrapRenderedRuns(parts, wrapWidth, letterSpacing, allowWrap)
	renderedLines := make([]*renderedTextLine, 0, len(wrapped))
	maxLineWidth := 0.0
	textHeight := 0.0

	for idx, runs := range wrapped {
		line := buildRenderedLineFromRuns(runs, baseMetrics, letterSpacing)
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

	// 生成缓存键：基于文本内容、样式和尺寸
	var strokeColorStr, shadowColorStr string
	var strokeSizeVal, shadowOffXVal, shadowOffYVal float64
	if field != nil {
		strokeColorStr = field.StrokeColor()
		strokeSizeVal = field.StrokeSize()
		shadowColorStr = field.ShadowColor()
		shadowOffXVal, shadowOffYVal = field.ShadowOffset()
	}
	cacheKey := fmt.Sprintf("text_%s_%s_%s_%d_%.0fx%.0f_%.1f_%.1f_%v_%v_%.1f_%s_%.1f_%s_%.1f_%.1f",
		value, baseStyle.Font, baseStyle.Color, baseStyle.FontSize,
		finalWidth, finalHeight, letterSpacing, leading,
		align, valign,
		strokeSizeVal, strokeColorStr,
		shadowOffXVal+shadowOffYVal, shadowColorStr,
		paddingLeft+paddingRight, paddingTop+paddingBottom)

	// 尝试从缓存获取
	textImageCacheMu.RLock()
	textImg, cached := textImageCache[cacheKey]
	textImageCacheMu.RUnlock()

	if !cached {
		// 缓存未命中，创建新图像
		textImg = ebiten.NewImage(imgW, imgH)

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
				runStartX := cursorX
				if run.hasGlyphs() {
					if prevHadGlyph && letterSpacing != 0 {
						cursorX += letterSpacing
						runStartX = cursorX
					}
					if run.imageItem != nil {
						// 绘制图片
						local := ebiten.GeoM{}
						local.Translate(cursorX, lineTop)
						if err := drawPackageItem(textImg, run.imageItem, local, atlas, 1, nil); err != nil {
							log.Printf("⚠️ 绘制图片失败 %s: %v", run.imageURL, err)
						}
					} else if run.bitmap != nil {
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
				if run.link != "" && run.width > 0 {
					linkRegions = append(linkRegions, widgets.TextLinkRegion{
						Target: run.link,
						Bounds: laya.Rect{
							X: runStartX,
							Y: lineTop,
							W: run.width,
							H: line.height,
						},
					})
				}
			}

			cursorY += line.height
			if lineIndex != len(renderedLines)-1 {
				cursorY += leading
			}
		}

		// 存入缓存
		textImageCacheMu.Lock()
		textImageCache[cacheKey] = textImg
		textImageCacheMu.Unlock()
	}

	opts := &ebiten.DrawImageOptions{GeoM: geo}
	if alpha < 1 {
		opts.ColorM.Scale(1, 1, 1, alpha)
	}
	applyColorEffects(opts, sprite)
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
				current = append(current, textutil.Segment{Text: chunk, Style: seg.Style, Link: seg.Link})
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

	// 使用原始度量值，避免过度调整
	ascent := float64(metrics.Ascent) / 64.0 // 从固定点转换为像素
	descent := float64(metrics.Descent) / 64.0
	height := float64(metrics.Height) / 64.0

	// 处理异常值
	if ascent <= 0 {
		ascent = float64(size) * 0.8 // 字体大小的80%作为上升部
	}
	if descent <= 0 {
		descent = float64(size) * 0.2 // 字体大小的20%作为下降部
	}
	if height <= 0 {
		height = float64(size)
	}

	// LayaAir 通常使用固定的行高比例，但间距要更紧凑
	lineHeight := ascent + descent + float64(size)*0.15 // 减少额外间距以更接近 LayaAir

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
		text:     seg.Text,
		runes:    []rune(seg.Text),
		style:    seg.Style,
		color:    baseColor,
		link:     seg.Link,
		imageURL: seg.ImageURL,
	}

	// 处理图片标签
	if seg.ImageURL != "" {
		// 解析图片 URL 获取 PackageItem
		// URL 格式: ui://package_id/item_id 或 ui://package_name/item_name
		item := assets.GetItemByURL(seg.ImageURL)
		if item != nil {
			run.imageItem = item
			// 使用图片的尺寸
			run.width = float64(item.Width)
			run.ascent = float64(item.Height) * 0.8 // 图片的基线位置
			run.descent = float64(item.Height) * 0.2
			run.fontSize = item.Height
			return run
		}
		// 如果图片未找到,仍然返回占位符文本
		log.Printf("⚠️ 图片未找到: %s", seg.ImageURL)
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
	if fontRef != "" {
		if font := lookupBitmapFont(fontRef); font != nil {
			run.bitmap = font
			run.advances = make([]float64, len(run.runes))
			width := 0.0
			for idx, r := range run.runes {
				advance := font.SpaceAdvance()
				if glyph := font.Glyphs[r]; glyph != nil {
					advance = glyph.Advance
				}
				run.advances[idx] = advance
				width += advance
				if idx != len(run.runes)-1 {
					width += letterSpacing
				}
			}
			run.width = width
			run.ascent = font.LineHeight
			if run.ascent <= 0 {
				run.ascent = float64(run.fontSize)
			}
			run.descent = 0
			return run
		}
	}

	face := fontFaceForSize(size)
	if face == nil {
		face = selectFontFace(field)
	}
	run.face = face
	metrics := face.Metrics()
	// 直接使用固定点数值，避免取整导致的精度损失
	run.ascent = float64(metrics.Ascent) / 64.0
	run.descent = float64(metrics.Descent) / 64.0
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

	run.advances = make([]float64, len(run.runes))
	width := 0.0
	fontSize := run.fontSize // 保存字体大小用于备选计算
	for idx, r := range run.runes {
		advance := 0.0
		if adv, ok := face.GlyphAdvance(r); ok {
			advance = float64(adv) / 64.0
		} else {
			bounds, _, ok := face.GlyphBounds(r)
			if ok {
				advance = float64(bounds.Max.X-bounds.Min.X) / 64.0
			} else {
				// 备选方案：使用字体大小的比例估算
				advance = float64(fontSize) * 0.6
			}
		}
		run.advances[idx] = advance
		width += advance
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

func estimateHorizontalPadding(field *widgets.GTextField) (left, right float64) {
	if field == nil {
		return 0, 0
	}
	if stroke := field.StrokeSize(); stroke > 0 {
		left = math.Max(left, stroke)
		right = math.Max(right, stroke)
	}
	if c := parseColor(field.ShadowColor()); c != nil {
		offX, _ := field.ShadowOffset()
		if offX < 0 {
			left = math.Max(left, -offX)
		} else if offX > 0 {
			right = math.Max(right, offX)
		}
	}
	return
}

func buildTextParts(segments []textutil.Segment, field *widgets.GTextField, baseColor color.NRGBA, base baseMetrics, letterSpacing float64) []textPart {
	var parts []textPart
	for _, seg := range segments {
		chunks := strings.Split(seg.Text, "\n")
		if len(chunks) == 0 {
			parts = append(parts, textPart{forcedBreak: true})
			continue
		}
		for idx, chunk := range chunks {
			if chunk != "" {
				run := buildRenderedRun(textutil.Segment{Text: chunk, Style: seg.Style, Link: seg.Link, ImageURL: seg.ImageURL}, field, baseColor, base, letterSpacing)
				if run != nil && len(run.runes) > 0 {
					parts = append(parts, textPart{run: run})
				}
			}
			if idx != len(chunks)-1 {
				parts = append(parts, textPart{forcedBreak: true})
			}
		}
	}
	if len(parts) == 0 {
		parts = append(parts, textPart{forcedBreak: true})
	}
	return parts
}

func wrapRenderedRuns(parts []textPart, wrapWidth float64, letterSpacing float64, allowWrap bool) [][]*renderedTextRun {
	lines := make([][]*renderedTextRun, 0)
	current := make([]*renderedTextRun, 0)
	currentWidth := 0.0

	flush := func() {
		lines = append(lines, current)
		current = make([]*renderedTextRun, 0)
		currentWidth = 0
	}

	for _, part := range parts {
		if part.forcedBreak {
			flush()
			continue
		}
		run := part.run
		if run == nil || len(run.runes) == 0 {
			continue
		}

		// 图片 run 作为整体处理,不切分
		if run.imageItem != nil {
			if allowWrap && wrapWidth > 0 && currentWidth+run.width > wrapWidth && len(current) > 0 {
				flush()
			}
			current, currentWidth = appendRun(current, currentWidth, run, letterSpacing)
			continue
		}

		if !allowWrap || wrapWidth <= 0 {
			current, currentWidth = appendRun(current, currentWidth, run, letterSpacing)
			continue
		}
		start := 0
		for start < len(run.runes) {
			if wrapWidth > 0 && currentWidth >= wrapWidth && len(current) > 0 {
				flush()
				continue
			}
			remaining := wrapWidth - currentWidth
			if remaining <= 0 && len(current) > 0 {
				flush()
				continue
			}
			end, _ := run.spanForWidth(start, remaining, letterSpacing)
			if end == start {
				if len(current) > 0 {
					flush()
					continue
				}
				end = start + 1
			}
			chunk := run.slice(start, end, letterSpacing)
			if chunk != nil && len(chunk.runes) > 0 {
				current, currentWidth = appendRun(current, currentWidth, chunk, letterSpacing)
			}
			start = end
			// 跳过空白 grapheme 集群
			start = skipWhitespaceGraphemes(run.text, start)
			if start < len(run.runes) {
				flush()
			}
		}
	}
	lines = append(lines, current)
	if len(lines) == 0 {
		lines = append(lines, nil)
	}
	return lines
}

func appendRun(line []*renderedTextRun, currentWidth float64, run *renderedTextRun, letterSpacing float64) ([]*renderedTextRun, float64) {
	if run == nil {
		return line, currentWidth
	}
	// 图片 run 可以没有 runes,但必须有 imageItem
	if len(run.runes) == 0 && run.imageItem == nil {
		return line, currentWidth
	}
	if len(line) > 0 && letterSpacing != 0 && run.hasGlyphs() {
		currentWidth += letterSpacing
	}
	line = append(line, run)
	currentWidth += run.width
	return line, currentWidth
}

func buildRenderedLineFromRuns(runs []*renderedTextRun, base baseMetrics, letterSpacing float64) *renderedTextLine {
	line := &renderedTextLine{
		runs: make([]*renderedTextRun, 0, len(runs)),
	}
	prevHadGlyph := false

	// 找到行中的主要字体大小，用于确定统一的基线
	dominantSize := 0
	dominantAscent := 0.0
	dominantDescent := 0.0

	for _, run := range runs {
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

			// 选择最大字体大小作为主要字体
			if run.fontSize > dominantSize {
				dominantSize = run.fontSize
				dominantAscent = run.ascent
				dominantDescent = run.descent
			}
		}
	}

	// 使用统一的基线：基于主要字体大小
	if line.hasGlyph {
		line.ascent = dominantAscent
		line.descent = dominantDescent
	} else {
		line.ascent = base.ascent
		line.descent = base.descent
	}

	// 确保行高足够容纳所有字符
	// 但保持基线固定，这是关键修复
	line.height = line.ascent + line.descent
	if line.height <= 0 {
		line.height = base.lineHeight
	}

	return line
}

// isBreakRune 使用 Unicode 标准判断是否为换行符
// 包括空格、换行符、制表符等
func isBreakRune(r rune) bool {
	// 使用 unicode.IsSpace 判断所有空格类字符
	// 包括: 空格, 换行符, 制表符, 回车符, 垂直制表符, 换页符等
	return unicode.IsSpace(r)
}

// hasGraphemeBreak 检查两个 rune 之间是否有字素边界
func hasGraphemeBreak(s string, index int) bool {
	if index <= 0 || index >= len([]rune(s)) {
		return false
	}

	// 使用 uniseg 判断边界
	gr := uniseg.NewGraphemes(s)
	pos := 0
	for gr.Next() {
		clusterBytes := gr.Bytes()
		clusterLen := len([]rune(string(clusterBytes)))

		// 如果当前位置包含目标索引，则有边界
		if pos+clusterLen > index {
			return pos < index
		}
		pos += clusterLen
	}
	return false
}

// skipWhitespaceGraphemes 跳过开头的空白 grapheme 集群，返回新的起始位置
func skipWhitespaceGraphemes(text string, start int) int {
	graphemes := getGraphemeClusters(text)
	if len(graphemes) == 0 || start >= len(graphemes) {
		return start
	}

	newStart := start
	for newStart < len(graphemes) {
		clusterRunes := []rune(string(graphemes[newStart]))
		isWhitespace := true
		for _, rn := range clusterRunes {
			if !isBreakRune(rn) {
				isWhitespace = false
				break
			}
		}
		if !isWhitespace {
			break
		}
		newStart++
	}
	return newStart
}

func drawBitmapRun(dst *ebiten.Image, run *renderedTextRun, startX float64, lineTop float64, letterSpacing float64, atlas *AtlasManager) error {
	if run.bitmap == nil || len(run.runes) == 0 {
		return nil
	}
	font := run.bitmap
	cursor := startX
	renderedCount := 0
	missingGlyphs := []rune{}

	for idx, r := range run.runes {
		glyph := font.Glyphs[r]
		advance := font.SpaceAdvance()
		if run.advances != nil && idx < len(run.advances) {
			advance = run.advances[idx]
		} else if glyph != nil {
			advance = glyph.Advance
		}
		if glyph != nil && glyph.Item != nil {
			local := ebiten.GeoM{}
			local.Translate(cursor+glyph.OffsetX, lineTop+glyph.OffsetY)

			// 检查是否使用 atlas 纹理模式 (BMFont .fnt 格式)
			if glyph.AtlasX != 0 || glyph.AtlasY != 0 {
				// Atlas 模式：从 atlas 纹理中提取子区域
				atlasImage, err := atlas.GetAtlasImage(glyph.Item)
				if err != nil {
					log.Printf("⚠️ 无法加载 atlas 纹理 %s: %v", glyph.Item.ID, err)
					missingGlyphs = append(missingGlyphs, r)
				} else if atlasImage != nil {
					// 创建子图像 (从 atlas 中截取字形区域)
					bounds := atlasImage.Bounds()

					// 参考 LayaAir UIPackage.ts:783-786
					// bg.texture = Laya.Texture.create(mainTexture,
					//     bx + mainSprite.rect.x, by + mainSprite.rect.y, bg.width, bg.height);
					// bx, by 是相对于 font sprite rect 的坐标，需要加上 rect 偏移
					x0 := int(glyph.AtlasX) + glyph.SpriteRectX
					y0 := int(glyph.AtlasY) + glyph.SpriteRectY
					x1 := x0 + int(glyph.Width)
					y1 := y0 + int(glyph.Height)

					// 边界检查
					if x0 >= 0 && y0 >= 0 && x1 <= bounds.Dx() && y1 <= bounds.Dy() {
						subImg := atlasImage.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)
						opts := &ebiten.DrawImageOptions{GeoM: local}
						// 应用文本颜色
						opts.ColorScale.ScaleWithColor(run.color)
						dst.DrawImage(subImg, opts)
						renderedCount++
					} else {
						log.Printf("⚠️ 字形 U+%04X 的 atlas 坐标越界: (%d,%d)-(%d,%d), atlas 尺寸: %dx%d",
							r, x0, y0, x1, y1, bounds.Dx(), bounds.Dy())
						missingGlyphs = append(missingGlyphs, r)
					}
				}
			} else {
				// 独立图片模式：直接绘制 PackageItem
				if err := drawPackageItem(dst, glyph.Item, local, atlas, 1, nil); err != nil {
					return err
				}
				renderedCount++
			}
		} else {
			missingGlyphs = append(missingGlyphs, r)
		}
		if idx != len(run.runes)-1 {
			cursor += advance + letterSpacing
		} else {
			cursor += advance
		}
	}

	if len(missingGlyphs) > 0 {
		log.Printf("⚠️ 位图字体缺失字形: %v (已渲染: %d/%d)", missingGlyphs, renderedCount, len(run.runes))
	}

	return nil
}

func renderSystemRun(dst *ebiten.Image, run *renderedTextRun, startX float64, baseline float64, letterSpacing float64, strokeColor *color.NRGBA, strokeSize float64, shadowColor *color.NRGBA, shadowOffsetX, shadowOffsetY float64) {
	if run.face == nil || len(run.runes) == 0 {
		return
	}
	if run.style.Italic {
		renderItalicSystemRun(dst, run, startX, baseline, letterSpacing, strokeColor, strokeSize, shadowColor, shadowOffsetX, shadowOffsetY)
		return
	}

	// 使用新的 text/v2 库渲染，基线计算更准确
	textFace := textv2.NewGoXFace(run.face)

	// 计算正确的渲染位置
	renderY := baseline - run.ascent

	// 如果有描边，使用高质量描边方案（临时图像 + alpha 膨胀）
	if strokeColor != nil && strokeSize > 0 {
		renderTextWithStroke(dst, run.text, run.face, startX, renderY, run.color, *strokeColor, strokeSize, run.style.Bold)

		// 渲染阴影（在描边后）
		if shadowColor != nil && (shadowOffsetX != 0 || shadowOffsetY != 0) {
			renderTextWithStroke(dst, run.text, run.face, startX+shadowOffsetX, renderY+shadowOffsetY, *shadowColor, *shadowColor, 0, run.style.Bold)
		}
		return
	}

	// 无描边的正常渲染
	opts := &textv2.DrawOptions{
		LayoutOptions: textv2.LayoutOptions{
			PrimaryAlign:   textv2.AlignStart,
			SecondaryAlign: textv2.AlignStart,
			LineSpacing:    0,
		},
	}
	opts.GeoM.Translate(startX, renderY)
	opts.ColorScale.ScaleWithColor(run.color)

	// 渲染阴影
	if shadowColor != nil && (shadowOffsetX != 0 || shadowOffsetY != 0) {
		shadowOpts := *opts
		shadowOpts.ColorScale.ScaleWithColor(*shadowColor)
		shadowOpts.GeoM.Translate(shadowOffsetX, shadowOffsetY)
		textv2.Draw(dst, run.text, textFace, &shadowOpts)
	}

	// 渲染主文本
	textv2.Draw(dst, run.text, textFace, opts)

	// 渲染粗体效果
	if run.style.Bold {
		boldOpts := *opts
		boldOpts.GeoM.Translate(0.6, 0)
		textv2.Draw(dst, run.text, textFace, &boldOpts)
	}
}

// renderTextWithStroke 使用高质量算法渲染带描边的文本
// 通过 alpha 通道膨胀避免多次绘制导致的锯齿感
// 参数 x, y 是文本渲染区域的左上角位置（不是基线位置！）
func renderTextWithStroke(dst *ebiten.Image, text string, fontFace font.Face, x, y float64, textColor, strokeColor color.NRGBA, strokeSize float64, bold bool) {
	// 使用 textv2 测量文本边界
	textFace := textv2.NewGoXFace(fontFace)
	width, height := textv2.Measure(text, textFace, 0)

	// 计算临时图像尺寸（需要包含描边空间）
	padding := math.Ceil(strokeSize) * 2
	tempWidth := int(width) + int(padding*2)
	tempHeight := int(height) + int(padding*2)

	if tempWidth <= 0 || tempHeight <= 0 {
		return
	}

	// 创建临时图像用于渲染文本 alpha
	temp := ebiten.NewImage(tempWidth, tempHeight)
	defer temp.Dispose()

	// 在临时图像上渲染白色文本（只需要 alpha 通道）
	tempOpts := &textv2.DrawOptions{
		LayoutOptions: textv2.LayoutOptions{
			PrimaryAlign:   textv2.AlignStart,
			SecondaryAlign: textv2.AlignStart,
		},
	}
	// textv2.Draw 的坐标是渲染区域左上角
	// 在临时图像上，渲染区域左上角在 (padding, padding)
	tempOpts.GeoM.Translate(padding, padding)
	tempOpts.ColorScale.ScaleWithColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	textv2.Draw(temp, text, textFace, tempOpts)

	// 如果需要粗体，额外绘制一次偏移
	if bold {
		boldOpts := *tempOpts
		boldOpts.GeoM.Translate(0.6, 0)
		textv2.Draw(temp, text, textFace, &boldOpts)
	}

	// 创建描边图像（通过膨胀 alpha 通道）
	if strokeSize > 0 {
		stroke := ebiten.NewImage(tempWidth, tempHeight)
		defer stroke.Dispose()

		// 使用简单的膨胀：在多个方向偏移绘制原始 alpha
		// 这比 CPU 端的形态学膨胀更高效
		iStrokeSize := int(math.Ceil(strokeSize))
		for dy := -iStrokeSize; dy <= iStrokeSize; dy++ {
			for dx := -iStrokeSize; dx <= iStrokeSize; dx++ {
				// 圆形膨胀核
				if dx*dx+dy*dy <= iStrokeSize*iStrokeSize {
					drawOpts := &ebiten.DrawImageOptions{}
					drawOpts.GeoM.Translate(float64(dx), float64(dy))
					stroke.DrawImage(temp, drawOpts)
				}
			}
		}

		// 将描边绘制到目标图像（使用描边颜色）
		// 临时图像上，渲染区域左上角在 (padding, padding)
		// 目标上，渲染区域左上角在 (x, y)
		// 所以临时图像的 (padding, padding) 应该对应目标的 (x, y)
		// 因此临时图像的 (0, 0) 应该对应目标的 (x - padding, y - padding)
		strokeDrawOpts := &ebiten.DrawImageOptions{}
		strokeDrawOpts.GeoM.Translate(x-padding, y-padding)
		strokeDrawOpts.ColorScale.ScaleWithColor(strokeColor)
		dst.DrawImage(stroke, strokeDrawOpts)
	}

	// 在描边上绘制原始文本（使用文本颜色）
	textDrawOpts := &ebiten.DrawImageOptions{}
	textDrawOpts.GeoM.Translate(x-padding, y-padding)
	textDrawOpts.ColorScale.ScaleWithColor(textColor)
	dst.DrawImage(temp, textDrawOpts)
}

func drawSystemGlyphs(dst *ebiten.Image, run *renderedTextRun, startX, baseline float64, letterSpacing float64, col color.NRGBA) {
	if run == nil || run.face == nil || len(run.runes) == 0 {
		return
	}
	src := image.NewUniform(col)
	drawer := font.Drawer{
		Dst:  dst,
		Src:  src,
		Face: run.face,
	}
	x := startX
	// 使用更高精度的基线计算，确保文本定位准确
	// Ebiten 的 font.Drawer 期望基线坐标以固定点格式提供
	for idx, r := range run.runes {
		// 将浮点坐标精确转换为固定点坐标
		drawer.Dot = fixed.Point26_6{
			X: fixed.Int26_6(math.Round(x*64) + 0.5), // 添加0.5以改善舍入精度
			Y: fixed.Int26_6(math.Round(baseline*64) + 0.5),
		}
		drawer.DrawString(string(r))
		advance := 0.0
		if run.advances != nil && idx < len(run.advances) {
			advance = run.advances[idx]
		} else if adv, ok := run.face.GlyphAdvance(r); ok {
			advance = float64(adv) / 64.0
		} else {
			bounds, _, ok := run.face.GlyphBounds(r)
			if ok {
				advance = float64(bounds.Max.X-bounds.Min.X) / 64.0
			} else {
				// 备选方案：使用字体大小的比例估算
				advance = float64(run.fontSize) * 0.6
			}
		}
		x += advance
		if idx != len(run.runes)-1 {
			x += letterSpacing
		}
	}
}

func renderItalicSystemRun(dst *ebiten.Image, run *renderedTextRun, startX float64, baseline float64, letterSpacing float64, strokeColor *color.NRGBA, strokeSize float64, shadowColor *color.NRGBA, shadowOffsetX, shadowOffsetY float64) {
	// 斜体文本渲染（与正常文本相同，只是添加 Skew 变换）
	// 使用 FilterLinear 优化变换时的插值效果，减少锯齿
	const italicShear = -0.25

	textFace := textv2.NewGoXFace(run.face)
	renderY := baseline - run.ascent

	// 如果有描边，使用高质量描边方案
	if strokeColor != nil && strokeSize > 0 {
		renderTextWithStrokeAndSkew(dst, run.text, run.face, startX, renderY, run.color, *strokeColor, strokeSize, run.style.Bold, italicShear)

		// 渲染阴影（在描边后）
		if shadowColor != nil && (shadowOffsetX != 0 || shadowOffsetY != 0) {
			renderTextWithStrokeAndSkew(dst, run.text, run.face, startX+shadowOffsetX, renderY+shadowOffsetY, *shadowColor, *shadowColor, 0, run.style.Bold, italicShear)
		}
		return
	}

	// 无描边的正常渲染
	opts := &textv2.DrawOptions{
		LayoutOptions: textv2.LayoutOptions{
			PrimaryAlign:   textv2.AlignStart,
			SecondaryAlign: textv2.AlignStart,
			LineSpacing:    0,
		},
	}
	opts.GeoM.Skew(italicShear, 0)
	opts.GeoM.Translate(startX, renderY)
	opts.ColorScale.ScaleWithColor(run.color)
	// 使用线性插值优化斜体变换时的锯齿问题
	opts.Filter = ebiten.FilterLinear

	// 渲染阴影
	if shadowColor != nil && (shadowOffsetX != 0 || shadowOffsetY != 0) {
		shadowOpts := *opts
		shadowOpts.ColorScale.ScaleWithColor(*shadowColor)
		shadowOpts.GeoM.Translate(shadowOffsetX, shadowOffsetY)
		textv2.Draw(dst, run.text, textFace, &shadowOpts)
	}

	// 渲染主文本
	textv2.Draw(dst, run.text, textFace, opts)

	// 渲染粗体效果
	if run.style.Bold {
		boldOpts := *opts
		boldOpts.GeoM.Translate(0.6, 0)
		textv2.Draw(dst, run.text, textFace, &boldOpts)
	}
}

// renderTextWithStrokeAndSkew 使用高质量算法渲染带描边和斜体变换的文本
// 参数 x, y 是文本渲染区域的左上角位置（不是基线位置！）
// 使用超采样技术减少锯齿效果
func renderTextWithStrokeAndSkew(dst *ebiten.Image, text string, fontFace font.Face, x, y float64, textColor, strokeColor color.NRGBA, strokeSize float64, bold bool, skew float64) {
	// 使用超采样技术优化斜体渲染，减少锯齿
	const supersampleScale = 2.0 // 超采样比例：2x 分辨率

	// 使用 textv2 测量文本边界
	textFace := textv2.NewGoXFace(fontFace)
	width, height := textv2.Measure(text, textFace, 0)

	// 计算临时图像尺寸（需要包含描边空间和斜体偏移）
	// 注意：这里是超采样尺寸，所以要乘以 2
	padding := (math.Ceil(strokeSize) + 2.0) * supersampleScale // 额外增加缓冲区
	skewOffset := math.Abs(skew * height) * supersampleScale
	tempWidth := int(width*supersampleScale) + int(skewOffset+padding*2)
	tempHeight := int(height*supersampleScale) + int(padding*2)

	if tempWidth <= 0 || tempHeight <= 0 {
		return
	}

	// 创建临时图像用于渲染文本 alpha（超采样）
	temp := ebiten.NewImage(tempWidth, tempHeight)
	defer temp.Dispose()

	// 在临时图像上渲染白色文本（应用斜体变换）
	tempOpts := &textv2.DrawOptions{
		LayoutOptions: textv2.LayoutOptions{
			PrimaryAlign:   textv2.AlignStart,
			SecondaryAlign: textv2.AlignStart,
		},
	}
	// 先缩放再斜体变换，确保超采样效果
	tempOpts.GeoM.Scale(supersampleScale, supersampleScale)
	tempOpts.GeoM.Skew(skew, 0)
	// textv2.Draw 的坐标是渲染区域左上角
	// 斜体需要额外的水平空间：skewOffset
	// 在临时图像上，渲染区域左上角在 (padding + skewOffset/supersampleScale, padding/supersampleScale)
	tempOpts.GeoM.Translate((padding/supersampleScale)+(skewOffset/supersampleScale), padding/supersampleScale)
	tempOpts.ColorScale.ScaleWithColor(color.NRGBA{R: 255, G: 255, B: 255, A: 255})
	textv2.Draw(temp, text, textFace, tempOpts)

	// 如果需要粗体，额外绘制一次偏移
	if bold {
		boldOpts := *tempOpts
		boldOpts.GeoM.Translate(0.6*supersampleScale, 0)
		textv2.Draw(temp, text, textFace, &boldOpts)
	}

	// 创建描边图像（通过膨胀 alpha 通道）
	if strokeSize > 0 {
		stroke := ebiten.NewImage(tempWidth, tempHeight)
		defer stroke.Dispose()

		// 使用圆形膨胀核（注意：描边尺寸也需要考虑超采样）
		iStrokeSize := int(math.Ceil(strokeSize * supersampleScale))
		for dy := -iStrokeSize; dy <= iStrokeSize; dy++ {
			for dx := -iStrokeSize; dx <= iStrokeSize; dx++ {
				if dx*dx+dy*dy <= iStrokeSize*iStrokeSize {
					drawOpts := &ebiten.DrawImageOptions{}
					drawOpts.GeoM.Translate(float64(dx), float64(dy))
					stroke.DrawImage(temp, drawOpts)
				}
			}
		}

		// 将描边绘制到目标图像
		// 临时图像的坐标系是超采样的，所以坐标需要除以 supersampleScale
		// 最终图像会自动缩放回原始尺寸
		strokeDrawOpts := &ebiten.DrawImageOptions{}
		strokeDrawOpts.GeoM.Translate((x/supersampleScale)-(padding/supersampleScale)-(skewOffset/supersampleScale), (y/supersampleScale)-(padding/supersampleScale))
		strokeDrawOpts.ColorScale.ScaleWithColor(strokeColor)
		// 使用线性插值优化缩放时的锯齿问题
		strokeDrawOpts.Filter = ebiten.FilterLinear
		dst.DrawImage(stroke, strokeDrawOpts)
	}

	// 在描边上绘制原始文本
	textDrawOpts := &ebiten.DrawImageOptions{}
	textDrawOpts.GeoM.Translate((x/supersampleScale)-(padding/supersampleScale)-(skewOffset/supersampleScale), (y/supersampleScale)-(padding/supersampleScale))
	textDrawOpts.ColorScale.ScaleWithColor(textColor)
	// 使用线性插值优化缩放时的锯齿问题
	textDrawOpts.Filter = ebiten.FilterLinear
	dst.DrawImage(temp, textDrawOpts)
}

func drawUnderline(dst *ebiten.Image, startX, baseline, width float64, fontSize int, col color.NRGBA) {
	if width <= 0 {
		return
	}
	// LayaAir 的下划线通常很细，约为字体大小的 1/20 左右
	// 但至少要有 1 像素可见
	thickness := math.Max(1, float64(fontSize)/20)
	// 下划线位置：baseline 下方约 1-2 像素
	y := baseline + 2.0
	vector.DrawFilledRect(dst, float32(startX), float32(y), float32(width), float32(thickness), col, true)
}
