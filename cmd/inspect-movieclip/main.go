package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: inspect-movieclip <.fui file>")
		os.Exit(1)
	}

	fuiPath := os.Args[1]

	// Load the package
	data, err := os.ReadFile(fuiPath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	pkg, err := assets.ParsePackage(data, filepath.Base(fuiPath))
	if err != nil {
		log.Fatalf("Failed to parse package: %v", err)
	}

	fmt.Printf("Package: %s\n", pkg.ID)
	fmt.Printf("Items: %d\n", len(pkg.Items))
	fmt.Println()

	// Find all MovieClip items
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeMovieClip {
			inspectMovieClip(item)
		}
	}
}

func inspectMovieClip(item *assets.PackageItem) {
	fmt.Printf("MovieClip: %s (%s)\n", item.Name, item.ID)
	fmt.Printf("  Size: %dx%d\n", item.Width, item.Height)
	fmt.Printf("  Interval: %dms\n", item.Interval)
	fmt.Printf("  Swing: %v\n", item.Swing)
	fmt.Printf("  RepeatDelay: %dms\n", item.RepeatDelay)
	fmt.Printf("  Frames: %d\n", len(item.Frames))

	if len(item.Frames) > 0 {
		fmt.Println("  Frame details:")
		for i, frame := range item.Frames {
			spriteInfo := "no sprite"
			if frame.Sprite != nil {
				spriteInfo = fmt.Sprintf("sprite %s (%dx%d offset:%.1f,%.1f)",
					frame.SpriteID,
					frame.Sprite.Rect.Width, frame.Sprite.Rect.Height,
					frame.Sprite.Offset.X, frame.Sprite.Offset.Y)
			}
			fmt.Printf("    Frame %d: %dx%d offset:(%d,%d) addDelay:%dms %s\n",
				i, frame.Width, frame.Height, frame.OffsetX, frame.OffsetY, frame.AddDelay, spriteInfo)
		}
	}
	fmt.Println()
}