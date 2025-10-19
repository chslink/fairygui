//go:build ebiten

package render

import (
	"errors"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

var labelFont font.Face = basicfont.Face7x13

// SetTextFont overrides the default font used when drawing text-based widgets.
func SetTextFont(face font.Face) {
	if face != nil {
		labelFont = face
	}
}

// DrawComponent traverses the component hierarchy and draws the visible objects onto target.
// It currently supports image widgets; other widget types are skipped silently.
func DrawComponent(target *ebiten.Image, root *core.GComponent, atlas *AtlasManager) error {
	if target == nil {
		return errors.New("render: target image is nil")
	}
	if root == nil {
		return errors.New("render: component is nil")
	}
	if atlas == nil {
		return errors.New("render: atlas manager is nil")
	}

	var geo ebiten.GeoM
	geo.Reset()
	return drawComponent(target, root, atlas, geo, 1)
}

func drawComponent(target *ebiten.Image, comp *core.GComponent, atlas *AtlasManager, parentGeo ebiten.GeoM, parentAlpha float64) error {
	for _, child := range comp.Children() {
		if err := drawObject(target, child, atlas, parentGeo, parentAlpha); err != nil {
			return err
		}
	}
	return nil
}

func drawObject(target *ebiten.Image, obj *core.GObject, atlas *AtlasManager, parentGeo ebiten.GeoM, parentAlpha float64) error {
	if obj == nil || !obj.Visible() {
		return nil
	}
	alpha := parentAlpha * obj.Alpha()
	if alpha <= 0 {
		return nil
	}

	local := ebiten.GeoM{}
	local.Reset()

	w := obj.Width()
	h := obj.Height()
	if w <= 0 {
		if pkgItem, ok := obj.Data().(*assets.PackageItem); ok && pkgItem != nil && pkgItem.Sprite != nil {
			w = float64(pkgItem.Sprite.Rect.Width)
		}
	}
	if h <= 0 {
		if pkgItem, ok := obj.Data().(*assets.PackageItem); ok && pkgItem != nil && pkgItem.Sprite != nil {
			h = float64(pkgItem.Sprite.Rect.Height)
		}
	}

	px, py := obj.Pivot()
	local.Translate(-px*w, -py*h)

	sx, sy := obj.Scale()
	if sx != 0 || sy != 0 {
		local.Scale(sx, sy)
	}

	skewX, skewY := obj.Skew()
	if skewX != 0 || skewY != 0 {
		local.Skew(skewX, skewY)
	}

	rotation := obj.Rotation()
	if rotation != 0 {
		local.Rotate(rotation)
	}

	pos := obj.DisplayObject().Position()
	local.Translate(pos.X, pos.Y)

	combined := parentGeo
	combined.Concat(local)

	switch data := obj.Data().(type) {
	case *assets.PackageItem:
		if err := drawPackageItem(target, data, combined, atlas, alpha); err != nil {
			return err
		}
	case *core.GComponent:
		if err := drawComponent(target, data, atlas, combined, alpha); err != nil {
			return err
		}
	case string:
		if data != "" {
			if err := drawTextImage(target, combined, data, alpha, obj.Width(), obj.Height()); err != nil {
				return err
			}
		}
	case *widgets.GTextField:
		if textValue := data.Text(); textValue != "" {
			if err := drawTextImage(target, combined, textValue, alpha, obj.Width(), obj.Height()); err != nil {
				return err
			}
		}
	case *widgets.GLabel:
		iconItem := data.IconItem()
		textMatrix := combined
		if iconItem != nil {
			iconGeo := combined
			if err := drawPackageItem(target, iconItem, iconGeo, atlas, alpha); err != nil {
				return err
			}
			if iconItem.Sprite != nil {
				shift := ebiten.GeoM{}
				shift.Translate(float64(iconItem.Sprite.Rect.Width)+4, 0)
				shift.Concat(combined)
				textMatrix = shift
			}
		}
		if textValue := data.Title(); textValue != "" {
			if err := drawTextImage(target, textMatrix, textValue, alpha, obj.Width(), obj.Height()); err != nil {
				return err
			}
		}
	default:
		// Unsupported payloads are ignored for now.
	}

	return nil
}

func drawTextImage(target *ebiten.Image, geo ebiten.GeoM, value string, alpha float64, width, height float64) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	face := labelFont
	if face == nil {
		face = basicfont.Face7x13
	}
	lines := strings.Split(value, "\n")
	metrics := face.Metrics()
	lineHeight := metrics.Ascent.Ceil() + metrics.Descent.Ceil()
	if lineHeight <= 0 {
		lineHeight = metrics.Height.Ceil()
	}
	if lineHeight <= 0 {
		lineHeight = 1
	}
	maxWidth := 0
	for _, line := range lines {
		bound := text.BoundString(face, line)
		if w := bound.Dx(); w > maxWidth {
			maxWidth = w
		}
	}
	textHeight := lineHeight * len(lines)
	if len(lines) == 0 {
		textHeight = lineHeight
	}
	if width <= 0 {
		width = float64(maxWidth)
	}
	if height <= 0 {
		height = float64(textHeight)
	}
	imgW := int(math.Ceil(width))
	imgH := int(math.Ceil(height))
	if imgW <= 0 {
		imgW = int(math.Max(1, float64(maxWidth)))
	}
	if imgW <= 0 {
		imgW = 1
	}
	if imgH <= 0 {
		imgH = int(math.Max(1, float64(textHeight)))
	}
	if imgH <= 0 {
		imgH = 1
	}
	textImg := ebiten.NewImage(imgW, imgH)
	ascent := metrics.Ascent.Ceil()
	y := 0
	for _, line := range lines {
		text.Draw(textImg, line, face, 0, y+ascent, color.White)
		y += lineHeight
	}
	opts := &ebiten.DrawImageOptions{GeoM: geo}
	if alpha < 1 {
		opts.ColorM.Scale(1, 1, 1, alpha)
	}
	target.DrawImage(textImg, opts)
	return nil
}

func drawPackageItem(target *ebiten.Image, item *assets.PackageItem, geo ebiten.GeoM, atlas *AtlasManager, alpha float64) error {
	if item == nil {
		return nil
	}
	spriteAny, err := atlas.ResolveSprite(item)
	if err != nil {
		return err
	}
	img, ok := spriteAny.(*ebiten.Image)
	if !ok || img == nil {
		return errors.New("render: atlas returned unexpected sprite type")
	}

	opts := &ebiten.DrawImageOptions{
		GeoM: geo,
	}
	if alpha < 1 {
		opts.ColorM.Scale(1, 1, 1, alpha)
	}
	target.DrawImage(img, opts)
	return nil
}
