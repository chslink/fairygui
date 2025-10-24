//go:build ebiten

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chslink/fairygui/cmd/internal/capture"
	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/render"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
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
	var imageName = flag.String("image", "", "image name or id inside the package (required)")
	var outputPath = flag.String("out", "", "output path for rendered image (defaults to <package>_<image>.png)")
	var widthOverride = flag.Int("width", 0, "override render width in pixels")
	var heightOverride = flag.Int("height", 0, "override render height in pixels")
	var debugFlag = flag.Bool("debug", false, "enable debug logging")
	flag.Parse()

	if *packageID == "" || *imageName == "" {
		flag.Usage()
		os.Exit(2)
	}
	debugEnabled = *debugFlag

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
	if err := atlas.LoadPackage(ctx, pkg); err != nil {
		fatalf("load package textures: %v", err)
	}

	item := findImage(pkg, *imageName)
	if item == nil {
		fatalf("image %q not found in package %q", *imageName, pkg.Name)
	}
	if _, err := atlas.ResolveSprite(item); err != nil {
		fatalf("resolve sprite: %v", err)
	}

	root := core.NewGComponent()

	imgWidget := widgets.NewImage()
	obj := imgWidget.GObject
	obj.SetData(item)
	w, h := spriteSize(item)
	if item.PixelHitTest != nil {
		render.ApplyPixelHitTest(obj.DisplayObject(), item.PixelHitTest)
	}
	if w > 0 && h > 0 {
		obj.SetSize(w, h)
	}
	root.AddChild(obj)

	width := chooseDimension(*widthOverride, w, 256)
	height := chooseDimension(*heightOverride, h, 256)
	root.SetSize(float64(width), float64(height))

	debugFn := func(format string, args ...any) {
		debugf(format, args...)
	}
	if !debugEnabled {
		debugFn = nil
	}
	imageRGBA, err := capture.CaptureComponent(width, height, root, atlas, debugFn)
	if err != nil {
		fatalf("render image: %v", err)
	}

	actualPath := *outputPath
	if actualPath == "" {
		safePkg := sanitizeName(pkg.Name)
		safeImg := sanitizeName(*imageName)
		actualPath = fmt.Sprintf("%s_%s.png", safePkg, safeImg)
	}
	debugf("canvas size: %dx%d", width, height)

	if err := saveImage(actualPath, imageRGBA); err != nil {
		fatalf("write rendered image: %v", err)
	}
	fmt.Printf("Rendered image saved to %s (%dx%d)\n", actualPath, width, height)
}

func findImage(pkg *assets.Package, name string) *assets.PackageItem {
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeImage && item.Name == name {
			return item
		}
		if item.Type == assets.PackageItemTypeImage && item.ID == name {
			return item
		}
	}
	return nil
}

func spriteSize(item *assets.PackageItem) (float64, float64) {
	if item == nil {
		return 0, 0
	}
	if item.Sprite != nil {
		return float64(item.Sprite.Rect.Width), float64(item.Sprite.Rect.Height)
	}
	if item.Width > 0 || item.Height > 0 {
		return float64(item.Width), float64(item.Height)
	}
	return 0, 0
}

const maxAutoDimension = 8192

func chooseDimension(override int, contentSize float64, fallback int) int {
	if override > 0 {
		return override
	}
	if contentSize > 0 {
		est := int(math.Ceil(contentSize))
		debugf("content size %.2f -> %d", contentSize, est)
		if est > 0 && est <= maxAutoDimension {
			return est
		}
		debugf("requested size %d outside range, falling back to %d", est, fallback)
	}
	return fallback
}

func sanitizeName(name string) string {
	if name == "" {
		return "image"
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

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
