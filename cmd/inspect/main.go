package main

import (
	"context"
	"fmt"
	"log"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

func main() {
	ctx := context.Background()
	loader := assets.NewFileLoader("demo/assets")
	data, err := loader.LoadOne(ctx, "MainMenu.fui", assets.ResourceBinary)
	if err != nil {
		log.Fatal(err)
	}
	pkg, err := assets.ParsePackage(data, "MainMenu")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("package name=%s id=%s version=%d\n", pkg.Name, pkg.ID, pkg.Version)
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent {
			fmt.Printf("component id=%s name=%s size=%dx%d\n", item.ID, item.Name, item.Width, item.Height)
		}
	}
}
