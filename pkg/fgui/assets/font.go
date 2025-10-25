package assets

import (
	"fmt"
	"log"
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
	// Atlas 纹理模式字段 (用于 BMFont .fnt 格式)
	AtlasX      int32 // 在 atlas 纹理中的 x 坐标 (相对于 font sprite rect)
	AtlasY      int32 // 在 atlas 纹理中的 y 坐标 (相对于 font sprite rect)
	SpriteRectX int   // font item 的 sprite rect 偏移 X
	SpriteRectY int   // font item 的 sprite rect 偏移 Y
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

	log.Printf("📖 解析位图字体 %s: TTF=%v, Tint=%v, AutoScale=%v", item.Name, font.TTF, font.Tint, font.AutoScale)

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

	for i := 0; i < count; i++ {
		nextPos := int(buf.ReadInt16()) + buf.Pos()

		r := rune(buf.ReadUint16())
		imgID := stringValue(buf.ReadS())
		bx := buf.ReadInt32()
		by := buf.ReadInt32()
		offsetX := fixedToFloat(buf.ReadInt32())
		offsetY := fixedToFloat(buf.ReadInt32())
		width := int(buf.ReadInt32())
		height := int(buf.ReadInt32())
		advance := int(buf.ReadInt32())
		_ = buf.ReadByte() // channel

		log.Printf("  字符 U+%04X: imgID=%q, bx=%d, by=%d, offset=(%.1f,%.1f), size=(%d,%d), advance=%d",
			r, imgID, bx, by, offsetX, offsetY, width, height, advance)

		// 参考 LayaAir UIPackage.ts:783-804 的逻辑
		// 根据是否使用 atlas 纹理决定如何处理字形
		useAtlas := font.TTF || (imgID == "" && (bx != 0 || by != 0))

		var glyphItem *PackageItem
		var glyphWidth, glyphHeight float64

		if useAtlas {
			// Atlas 模式：使用 atlas 纹理和 bx, by 坐标
			// 参考 LayaAir UIPackage.ts:756-759, 783-786
			// mainSprite = this._sprites[item.id]
			// mainTexture = this.getItemAsset(mainSprite.atlas)
			// bg.texture = Laya.Texture.create(mainTexture,
			//     bx + mainSprite.rect.x, by + mainSprite.rect.y, bg.width, bg.height)

			if item.Sprite == nil || item.Sprite.Atlas == nil {
				log.Printf("⚠️ 位图字体 %s: 字符 U+%04X 缺少 sprite 或 atlas 引用", item.Name, r)
				buf.SetPos(nextPos)
				continue
			}

			// 字形图片就是 font item 的 Sprite.Atlas (即 texture 属性指向的图片)
			glyphItem = item.Sprite.Atlas

			glyphWidth = float64(width)
			glyphHeight = float64(height)
			if glyphWidth <= 0 || glyphHeight <= 0 {
				log.Printf("⚠️ 位图字体 %s: 字符 U+%04X 的尺寸无效 (%d x %d)", item.Name, r, width, height)
				buf.SetPos(nextPos)
				continue
			}

			// 创建字形,记录 atlas 坐标和 sprite rect 偏移
			adv := float64(advance)
			if adv == 0 {
				if headerAdvance > 0 {
					adv = headerAdvance
				} else {
					adv = glyphWidth + offsetX
				}
			}
			if adv == 0 {
				adv = glyphWidth
			}

			// 存储 sprite rect 偏移 (用于渲染时计算正确的 atlas 坐标)
			spriteRectX := 0
			spriteRectY := 0
			if item.Sprite != nil {
				spriteRectX = item.Sprite.Rect.X
				spriteRectY = item.Sprite.Rect.Y
			}

			font.Glyphs[r] = &BitmapGlyph{
				Item:        glyphItem,
				OffsetX:     offsetX,
				OffsetY:     offsetY,
				Width:       glyphWidth,
				Height:      glyphHeight,
				Advance:     adv,
				AtlasX:      bx,
				AtlasY:      by,
				SpriteRectX: spriteRectX,
				SpriteRectY: spriteRectY,
			}

			if glyphHeight > maxGlyphHeight {
				maxGlyphHeight = glyphHeight
			}
		} else {
			// 独立图片模式：每个字形有独立的 PackageItem
			if imgID == "" {
				log.Printf("⚠️ 位图字体 %s: 字符 U+%04X 缺少图片ID且无 atlas 坐标", item.Name, r)
				buf.SetPos(nextPos)
				continue
			}

			glyphItem = item.Owner.ItemByID(imgID)
			if glyphItem == nil {
				log.Printf("⚠️ 位图字体 %s: 字符 U+%04X 的图片 %s 未找到", item.Name, r, imgID)
				buf.SetPos(nextPos)
				continue
			}

			glyphWidth = float64(width)
			glyphHeight = float64(height)
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
				// 独立图片模式不需要 AtlasX, AtlasY
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
