package render

import (
	"strings"
	"sync"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

var (
	bitmapFonts sync.Map // key -> *assets.BitmapFont
)

// RegisterBitmapFonts registers all bitmap fonts contained in the given package.
func RegisterBitmapFonts(pkg *assets.Package) {
	if pkg == nil || len(pkg.Items) == 0 {
		return
	}
	for _, item := range pkg.Items {
		if item == nil || item.Type != assets.PackageItemTypeFont {
			continue
		}
		font, err := item.BitmapFontData()
		if err != nil || font == nil {
			continue
		}
		aliases := fontAliases(pkg, item)
		for _, alias := range aliases {
			if alias == "" {
				continue
			}
			bitmapFonts.Store(normalizeFontKey(alias), font)
		}
	}
}

func fontAliases(pkg *assets.Package, item *assets.PackageItem) []string {
	if pkg == nil || item == nil {
		return nil
}
	aliases := make([]string, 0, 4)
	if pkg.ID != "" && item.ID != "" {
		aliases = append(aliases, "ui://"+strings.ToLower(pkg.ID+item.ID))
	}
	if pkg.Name != "" && item.Name != "" {
		aliases = append(aliases, "ui://"+strings.ToLower(pkg.Name)+"/"+strings.ToLower(item.Name))
	}
	if item.ID != "" {
		aliases = append(aliases, strings.ToLower(item.ID))
	}
	if item.Name != "" {
		aliases = append(aliases, strings.ToLower(item.Name))
	}
	return aliases
}

func normalizeFontKey(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	return value
}

func lookupBitmapFont(ref string) *assets.BitmapFont {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil
	}
	key := normalizeFontKey(ref)
	if font, ok := bitmapFonts.Load(key); ok {
		return font.(*assets.BitmapFont)
	}
	if strings.HasPrefix(key, "ui://") {
		body := key[len("ui://"):]
		if font, ok := bitmapFonts.Load(body); ok {
			return font.(*assets.BitmapFont)
		}
		if idx := strings.Index(body, "/"); idx >= 0 {
			if font, ok := bitmapFonts.Load(body[idx+1:]); ok {
				return font.(*assets.BitmapFont)
			}
		}
	} else {
		if font, ok := bitmapFonts.Load("ui://"+key); ok {
			return font.(*assets.BitmapFont)
		}
	}
	return nil
}

func TestLookupFont(id string) *assets.BitmapFont {
	return lookupBitmapFont(id)
}
