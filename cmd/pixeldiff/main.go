//go:build ebiten

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chslink/fairygui/pkg/fgui"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/builder"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
)

var debugEnabled bool

func debugf(format string, args ...any) {
	if debugEnabled {
		fmt.Printf("[debug] "+format+"\n", args...)
	}
}

func main() {
	var assetsDir = flag.String("assets", "demo/assets", "directory containing FairyGUI assets")
	var packageID = flag.String("package", "", "package id or .fui filename (required)")
	var componentName = flag.String("component", "", "component name inside the package (required)")
	var baselinePath = flag.String("baseline", "", "optional baseline image (.png)")
	var outputPath = flag.String("out", "", "output path for rendered image (defaults to <package>_<component>.png)")
	var diffPath = flag.String("diff", "", "optional path to write diff heatmap image")
	var widthOverride = flag.Int("width", 0, "override render width in pixels")
	var heightOverride = flag.Int("height", 0, "override render height in pixels")
	var debugFlag = flag.Bool("debug", false, "print debug information")
	flag.Parse()

	if *packageID == "" || *componentName == "" {
		flag.Usage()
		os.Exit(2)
	}

	debugEnabled = *debugFlag

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	loader := assets.NewFileLoader(*assetsDir)

	pkgFile := *packageID
	if !strings.HasSuffix(strings.ToLower(pkgFile), ".fui") {
		pkgFile += ".fui"
	}
	data, err := loader.LoadOne(ctx, pkgFile, assets.ResourceBinary)
	if err != nil {
		fatalf("load package data: %v", err)
	}

	resKey := filepath.Join(*assetsDir, strings.TrimSuffix(pkgFile, filepath.Ext(pkgFile)))
	pkg, err := assets.ParsePackage(data, filepath.Clean(resKey))
	if err != nil {
		fatalf("parse package: %v", err)
	}

	atlas := render.NewAtlasManager(loader)
	factory := builder.NewFactoryWithLoader(atlas, loader)
	factory.RegisterPackage(pkg)

	item := findComponent(pkg, *componentName)
	if item == nil {
		fatalf("component %q not found in package %q", *componentName, pkg.Name)
	}

	root, err := factory.BuildComponent(ctx, pkg, item)
	if err != nil {
		fatalf("build component: %v", err)
	}

	if err := atlas.LoadPackage(ctx, pkg); err != nil {
		fatalf("load package textures: %v", err)
	}

	width := chooseDimension(*widthOverride, root.Width(), 800)
	height := chooseDimension(*heightOverride, root.Height(), 600)

	imageRGBA, err := captureComponent(width, height, root, atlas)
	if err != nil {
		fatalf("render component: %v", err)
	}

	actualPath := *outputPath
	if actualPath == "" {
		safePkg := sanitizeName(pkg.Name)
		safeComp := sanitizeName(*componentName)
		actualPath = fmt.Sprintf("%s_%s.png", safePkg, safeComp)
	}

	debugf("canvas size: %dx%d", width, height)

	if err := saveImage(actualPath, imageRGBA); err != nil {
		fatalf("write rendered image: %v", err)
	}
	fmt.Printf("Rendered image saved to %s (%dx%d)\n", actualPath, width, height)

	if *baselinePath == "" {
		return
	}

	diff, err := compareWithBaseline(imageRGBA, *baselinePath, *diffPath)
	if err != nil {
		fatalf("baseline comparison failed: %v", err)
	}

	fmt.Printf("Diff pixels: %d (%.2f%%)  max delta: %d  avg delta: %.2f\n", diff.Count, diff.Ratio*100, diff.MaxDelta, diff.Average)
	if diff.DiffPath != "" {
		fmt.Printf("Diff image saved to %s\n", diff.DiffPath)
	}
}

func findComponent(pkg *assets.Package, name string) *assets.PackageItem {
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Component != nil && item.Name == name {
			return item
		}
	}
	return nil
}

func captureComponent(width, height int, component *core.GComponent, atlas *render.AtlasManager) (*image.RGBA, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid capture size %dx%d", width, height)
	}
	game := newCaptureGame(width, height, component, atlas)

	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowResizable(false)
	ebiten.SetWindowTitle("pixeldiff")

	err := ebiten.RunGame(game)
	if err != nil && !errors.Is(err, ebiten.Termination) {
		return nil, err
	}
	if game.err != nil {
		return nil, game.err
	}
	if !game.rendered {
		return nil, errors.New("capture did not complete")
	}
	return game.image(), nil
}

type captureGame struct {
	width      int
	height     int
	component  *core.GComponent
	atlas      *render.AtlasManager
	root       *core.GRoot
	stage      *fgui.Stage
	rendered   bool
	pixels     []byte
	err        error
	frameCount int
}

func newCaptureGame(width, height int, component *core.GComponent, atlas *render.AtlasManager) *captureGame {
	stage := fgui.NewStage(width, height)
	root := core.NewGRoot()
	root.AttachStage(stage)
	root.Resize(width, height)
	if component != nil {
		component.SetPosition(0, 0)
		root.AddChild(component.GObject)
	}
	return &captureGame{
		width:     width,
		height:    height,
		component: component,
		atlas:     atlas,
		root:      root,
		stage:     stage,
	}
}

func (g *captureGame) Update() error {
	if g.err != nil {
		return g.err
	}
	if g.rendered {
		return ebiten.Termination
	}
	g.root.Advance(time.Second/60, fgui.MouseState{})
	g.frameCount++
	return nil
}

func (g *captureGame) Draw(screen *ebiten.Image) {
	if g.err != nil || g.rendered {
		return
	}
	screen.Clear()
	if err := render.DrawComponent(screen, g.root.GComponent, g.atlas); err != nil {
		g.err = err
		return
	}
	w, h := screen.Size()
	debugf("ReadPixels request size: %dx%d", w, h)
	pixels := make([]byte, 4*w*h)
	screen.ReadPixels(pixels)
	g.pixels = pixels
	g.rendered = true
}

func (g *captureGame) Layout(_, _ int) (int, int) {
	return g.width, g.height
}

func (g *captureGame) image() *image.RGBA {
	return &image.RGBA{
		Pix:    g.pixels,
		Stride: 4 * g.width,
		Rect:   image.Rect(0, 0, g.width, g.height),
	}
}

const maxAutoDimension = 8192

func chooseDimension(override int, componentSize float64, fallback int) int {
	if override > 0 {
		return override
	}
	if componentSize > 0 {
		est := int(math.Ceil(componentSize))
		debugf("component requested size %.2f -> %d", componentSize, est)
		if est > 0 && est <= maxAutoDimension {
			return est
		}
		debugf("requested size %d outside range, falling back to %d", est, fallback)
	}
	return fallback
}

func sanitizeName(name string) string {
	if name == "" {
		return "component"
	}
	return strings.ReplaceAll(name, string(os.PathSeparator), "_")
}

func saveImage(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

type diffResult struct {
	Count    int
	Ratio    float64
	MaxDelta uint8
	Average  float64
	DiffPath string
}

func compareWithBaseline(actual *image.RGBA, baselinePath, diffOut string) (*diffResult, error) {
	baselineFile, err := os.Open(baselinePath)
	if err != nil {
		return nil, err
	}
	defer baselineFile.Close()

	baselineImg, err := png.Decode(baselineFile)
	if err != nil {
		return nil, fmt.Errorf("decode baseline: %w", err)
	}
	baselineRGBA := imageToRGBA(baselineImg)

	width := actual.Bounds().Dx()
	height := actual.Bounds().Dy()
	if baselineRGBA.Bounds().Dx() != width || baselineRGBA.Bounds().Dy() != height {
		return nil, fmt.Errorf("baseline dimensions %dx%d do not match rendered %dx%d", baselineRGBA.Bounds().Dx(), baselineRGBA.Bounds().Dy(), width, height)
	}

	diffPixels := 0
	var totalDelta float64
	maxDelta := uint8(0)
	var diffImage *image.RGBA
	if diffOut != "" {
		diffImage = image.NewRGBA(baselineRGBA.Bounds())
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ai := actual.RGBAAt(x, y)
			bi := baselineRGBA.RGBAAt(x, y)
			dr := absDiff(ai.R, bi.R)
			dg := absDiff(ai.G, bi.G)
			db := absDiff(ai.B, bi.B)
			da := absDiff(ai.A, bi.A)
			delta := maxUint8(dr, dg, db, da)
			if delta > 0 {
				diffPixels++
				totalDelta += float64(dr + dg + db + da)
				if delta > maxDelta {
					maxDelta = delta
				}
				if diffImage != nil {
					diffImage.SetRGBA(x, y, color.RGBA{R: dr, G: dg, B: db, A: 0xff})
				}
			} else if diffImage != nil {
				diffImage.SetRGBA(x, y, color.RGBA{A: 0xff})
			}
		}
	}

	if diffImage != nil {
		if err := saveDiffImage(diffOut, diffImage); err != nil {
			return nil, err
		}
	}

	totalPixels := width * height
	ratio := 0.0
	avg := 0.0
	if totalPixels > 0 {
		ratio = float64(diffPixels) / float64(totalPixels)
		avg = totalDelta / float64(totalPixels*4)
	}

	return &diffResult{
		Count:    diffPixels,
		Ratio:    ratio,
		MaxDelta: maxDelta,
		Average:  avg,
		DiffPath: diffOut,
	}, nil
}

func saveDiffImage(path string, img *image.RGBA) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func imageToRGBA(src image.Image) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			dst.SetRGBA(x, y, color.RGBAModel.Convert(src.At(x, y)).(color.RGBA))
		}
	}
	return dst
}

func absDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}

func maxUint8(values ...uint8) uint8 {
	var max uint8
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
