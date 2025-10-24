package render

import (
	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/hajimehoshi/ebiten/v2"
)

var grayscaleMatrix = [20]float64{
	0.299, 0.587, 0.114, 0, 0,
	0.299, 0.587, 0.114, 0, 0,
	0.299, 0.587, 0.114, 0, 0,
	0, 0, 0, 1, 0,
}

func applyColorEffects(opts *ebiten.DrawImageOptions, sprite *laya.Sprite) {
	if opts == nil || sprite == nil {
		return
	}
	gray, enabled, matrix := sprite.ColorEffects()
	if gray {
		cm := colorMatrixToColorM(grayscaleMatrix)
		opts.ColorM.Concat(cm)
	} else if enabled {
		cm := colorMatrixToColorM(matrix)
		opts.ColorM.Concat(cm)
	}
	applyBlendMode(opts, sprite)
}

func applyBlendMode(opts *ebiten.DrawImageOptions, sprite *laya.Sprite) {
	if opts == nil || sprite == nil {
		return
	}
	switch sprite.BlendMode() {
	case laya.BlendModeAdd:
		opts.Blend = ebiten.BlendLighter
	default:
		// ebiten default is source-over, so nothing to do.
	}
}

func colorMatrixToColorM(values [20]float64) ebiten.ColorM {
	var cm ebiten.ColorM
	for row := 0; row < 4; row++ {
		for col := 0; col < 5; col++ {
			index := row*5 + col
			cm.SetElement(row, col, values[index])
		}
	}
	return cm
}
