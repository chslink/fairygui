package fgui_test

import (
	"context"
	"fmt"
	"os"

	"github.com/chslink/fairygui/pkg/fgui"
)

// ExampleFactory demonstrates the basic usage of the Factory API.
func ExampleFactory() {
	// 1. Load a .fui package
	data, err := os.ReadFile("../../demo/assets/MainMenu.fui")
	if err != nil {
		fmt.Println("Failed to load package:", err)
		return
	}

	pkg, err := fgui.ParsePackage(data, "demo/assets/MainMenu")
	if err != nil {
		fmt.Println("Failed to parse package:", err)
		return
	}

	// 2. Create a factory (nil atlas resolver for logic-only builds)
	factory := fgui.NewFactory(nil, nil)

	// 3. Register the package
	factory.RegisterPackage(pkg)

	// 4. Build a component
	item := pkg.ItemByName("Main")
	if item == nil {
		fmt.Println("Component not found")
		return
	}

	component, err := factory.BuildComponent(context.Background(), pkg, item)
	if err != nil {
		fmt.Println("Failed to build component:", err)
		return
	}

	fmt.Printf("Built component: %s\n", component.Name())
}

// ExampleNewFactoryWithLoader demonstrates automatic dependency resolution.
func ExampleNewFactoryWithLoader() {
	// 1. Create a file loader
	loader := fgui.NewFileLoader("../../demo/assets")

	// 2. Create a factory with automatic dependency resolution
	factory := fgui.NewFactoryWithLoader(nil, loader)

	// 3. Load the main package
	data, err := loader.LoadOne(context.Background(), "MainMenu.fui", fgui.ResourceBinary)
	if err != nil {
		fmt.Println("Failed to load package:", err)
		return
	}

	pkg, err := fgui.ParsePackage(data, "demo/assets/MainMenu")
	if err != nil {
		fmt.Println("Failed to parse package:", err)
		return
	}

	factory.RegisterPackage(pkg)

	// 4. Build component - dependencies will be loaded automatically
	item := pkg.ItemByName("Main")
	component, err := factory.BuildComponent(context.Background(), pkg, item)
	if err != nil {
		fmt.Println("Failed to build component:", err)
		return
	}

	fmt.Printf("Built component with dependencies: %s\n", component.Name())
}

// ExampleBuildComponent demonstrates the convenience wrapper.
func ExampleBuildComponent() {
	// Setup
	data, _ := os.ReadFile("../../demo/assets/MainMenu.fui")
	pkg, _ := fgui.ParsePackage(data, "demo/assets/MainMenu")
	factory := fgui.NewFactory(nil, nil)
	factory.RegisterPackage(pkg)
	item := pkg.ItemByName("Main")

	// Build using convenience function
	component, err := fgui.BuildComponent(context.Background(), factory, pkg, item)
	if err != nil {
		fmt.Println("Failed to build:", err)
		return
	}

	fmt.Printf("Component built: %s\n", component.Name())
}
