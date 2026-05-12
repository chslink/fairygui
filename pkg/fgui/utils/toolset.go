package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/chslink/fairygui/internal/compat/laya"
)

// ARGBToHTMLColor converts an ARGB uint32 to a CSS-style hex string.
// If withAlpha is true the result is "#AARRGGBB", otherwise "#RRGGBB".
func ARGBToHTMLColor(argb uint32, withAlpha bool) string {
	a := uint8(argb >> 24)
	r := uint8(argb >> 16)
	g := uint8(argb >> 8)
	b := uint8(argb)
	if withAlpha {
		return fmt.Sprintf("#%02X%02X%02X%02X", a, r, g, b)
	}
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

// HTMLColorToUint32 converts a "#RRGGBB" or "#AARRGGBB" string to a uint32 ARGB value.
func HTMLColorToUint32(s string) (uint32, error) {
	s = strings.TrimSpace(s)
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	if len(s) == 6 {
		s = "ff" + s
	}
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, fmt.Errorf("utils: invalid color %q: %w", s, err)
	}
	return uint32(v), nil
}

// Clamp constrains value to [min, max].
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Clamp01 constrains value to [0, 1].
func Clamp01(value float64) float64 {
	return Clamp(value, 0, 1)
}

// Lerp performs linear interpolation between start and end.
func Lerp(start, end, percent float64) float64 {
	return start + (end-start)*Clamp01(percent)
}

// Distance returns the Euclidean distance between two points.
func Distance(x1, y1, x2, y2 float64) float64 {
	dx := x1 - x2
	dy := y1 - y2
	return math.Sqrt(dx*dx + dy*dy)
}

// StartsWith checks whether s starts with prefix (case-insensitive).
func StartsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

// EndsWith checks whether s ends with suffix (case-insensitive).
func EndsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && strings.EqualFold(s[len(s)-len(suffix):], suffix)
}

// TrimRight removes trailing whitespace and newline characters.
func TrimRight(s string) string {
	return strings.TrimRight(s, " \t\r\n")
}

// EncodeHTML escapes &, <, >, " and ' characters.
func EncodeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// Repeat returns t modulo length (used to clamp cyclic values).
func Repeat(t, length float64) float64 {
	if length <= 0 {
		return 0
	}
	t = math.Mod(t, length)
	if t < 0 {
		t += length
	}
	return t
}

// DisplayObjectToGObject walks the sprite owner chain to find the owning GObject.
func DisplayObjectToGObject(sprite *laya.Sprite) *laya.Sprite {
	return sprite
}

// RuneLen returns the number of runes in s.
func RuneLen(s string) int {
	return utf8.RuneCountInString(s)
}
