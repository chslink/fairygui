package assets

import (
	"fmt"
	"math"
)

const fixedPointScale = 65536.0

// BitmapFont represents a bitmap font parsed from a FairyGUI package.
type BitmapFont struct {
	Item       *PackageItem
	FontSize   float64
	LineHeight float64
	TTF        bool
	Tint       bool
	AutoScale  bool
	Glyphs     map[rune]*BitmapGlyph
}

// BitmapGlyph stores glyph metrics for a bitmap font.
type BitmapGlyph struct {
	Item    *PackageItem
	OffsetX float64
	OffsetY float64
	Width   float64
	Height  float64
	Advance float64
}

// SpaceAdvance returns the advance used when rendering spaces.
func (f *BitmapFont) SpaceAdvance() float64 {
	if f == nil {
		return 0
	}
	if glyph, ok := f.Glyphs[' ']; ok && glyph != nil && glyph.Advance > 0 {
		return glyph.Advance
	}
	if f.FontSize > 0 {
		return f.FontSize * 0.5
	}
	if f.LineHeight > 0 {
		return f.LineHeight * 0.5
	}
	return 4
}

// BitmapFontData parses (or returns cached) bitmap font metadata for the package item.
func (item *PackageItem) BitmapFontData() (*BitmapFont, error) {
	if item == nil || item.Type != PackageItemTypeFont {
		return nil, fmt.Errorf("assets: item %q is not a font", item.Name)
	}
	item.fontOnce.Do(func() {
		font, err := parseBitmapFont(item)
		if err != nil {
			item.fontErr = err
			return
		}
		item.bitmapFont = font
	})
	return item.bitmapFont, item.fontErr
}

func parseBitmapFont(item *PackageItem) (*BitmapFont, error) {
	if item == nil || item.RawData == nil {
		return nil, fmt.Errorf("assets: font %q missing raw data", item.Name)
	}
	buf := item.RawData
	saved := buf.Pos()
	defer buf.SetPos(saved)
	_ = buf.SetPos(0)

	font := &BitmapFont{
		Item:   item,
		Glyphs: make(map[rune]*BitmapGlyph),
	}

	font.TTF = buf.ReadBool()
	font.Tint = buf.ReadBool()
	font.AutoScale = buf.ReadBool()
	_ = buf.ReadBool() // has channel

	fontSizeRaw := int32(buf.ReadInt32())
	headerAdvanceRaw := int32(buf.ReadInt32())
	lineHeightRaw := int32(buf.ReadInt32())

	font.FontSize = fixedToFloat(fontSizeRaw)
	headerAdvance := fixedToFloat(headerAdvanceRaw)
	lineHeight := fixedToFloat(lineHeightRaw)
	if lineHeight <= 0 {
		lineHeight = font.FontSize
	}
	font.LineHeight = lineHeight

	if !buf.Seek(0, 1) {
		return font, nil
	}

	count := int(buf.ReadInt32())
	maxGlyphHeight := 0.0

	if font.TTF {
		return nil, fmt.Errorf("assets: TTF fonts not supported yet (%s)", item.Name)
	}

	for i := 0; i < count; i++ {
		nextPos := int(buf.ReadInt16()) + buf.Pos()

		r := rune(buf.ReadUint16())
		imgID := stringValue(buf.ReadS())
		_ = buf.ReadInt32() // bx
		_ = buf.ReadInt32() // by
		offsetX := fixedToFloat(buf.ReadInt32())
		offsetY := fixedToFloat(buf.ReadInt32())
		width := int(buf.ReadInt32())
		height := int(buf.ReadInt32())
		advance := int(buf.ReadInt32())
		_ = buf.ReadByte() // channel

		if !font.TTF {
			if imgID == "" {
				buf.SetPos(nextPos)
				continue
			}
			glyphItem := item.Owner.ItemByID(imgID)
			if glyphItem == nil {
				buf.SetPos(nextPos)
				continue
			}

			glyphWidth := float64(width)
			glyphHeight := float64(height)
			if glyphWidth == 0 {
				glyphWidth = float64(glyphItem.Width)
			}
			if glyphHeight == 0 {
				glyphHeight = float64(glyphItem.Height)
			}
			if glyphHeight > maxGlyphHeight {
				maxGlyphHeight = glyphHeight
			}

			adv := float64(advance)
			if adv == 0 {
				if headerAdvance > 0 {
					adv = headerAdvance
				} else if glyphWidth > 0 {
					adv = glyphWidth + offsetX
				}
			}
			if adv == 0 {
				adv = glyphWidth
			}

			font.Glyphs[r] = &BitmapGlyph{
				Item:    glyphItem,
				OffsetX: offsetX,
				OffsetY: offsetY,
				Width:   glyphWidth,
				Height:  glyphHeight,
				Advance: adv,
			}
		}

		buf.SetPos(nextPos)
	}

	if font.LineHeight <= 0 {
		font.LineHeight = math.Max(maxGlyphHeight, 1)
	}
	if font.FontSize <= 0 {
		font.FontSize = font.LineHeight
	}

	return font, nil
}

func fixedToFloat(v int32) float64 {
	if v == 0 {
		return 0
	}
	return float64(v) / fixedPointScale
}
