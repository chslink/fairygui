package builder

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
)

func TestBuildComponentFromRealFUI(t *testing.T) {
	root := filepath.Join("..", "..", "..", "demo", "assets")
	data, err := os.ReadFile(filepath.Join(root, "Bag.fui"))
	if err != nil {
		t.Skipf("demo assets unavailable: %v", err)
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(root, "Bag"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	var component *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Component != nil {
			component = item
			break
		}
	}
	if component == nil {
		t.Fatalf("no component items found")
	}

	factory := NewFactory(nil, nil)
	built, err := factory.BuildComponent(context.Background(), pkg, component)
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	if len(built.Children()) != len(component.Component.Children) {
		t.Fatalf("expected %d children, got %d", len(component.Component.Children), len(built.Children()))
	}

	var imageIndex int = -1
	for idx, meta := range component.Component.Children {
		if meta.Type == assets.ObjectTypeImage && meta.Src != "" {
			imageIndex = idx
			break
		}
	}

	if imageIndex >= 0 {
		builtChild := built.ChildAt(imageIndex)
		if builtChild == nil {
			t.Fatalf("image child missing at index %d", imageIndex)
		}
		meta := component.Component.Children[imageIndex]
		if builtChild.X() != float64(meta.X) || builtChild.Y() != float64(meta.Y) {
			t.Fatalf("image child position mismatch: got (%v,%v) expected (%d,%d)", builtChild.X(), builtChild.Y(), meta.X, meta.Y)
		}
		data, ok := builtChild.Data().(*assets.PackageItem)
		if !ok || data == nil {
			t.Fatalf("expected image child to reference package item")
		}
		expectedWidth := meta.Width
		expectedHeight := meta.Height
		if expectedWidth < 0 && data.Sprite != nil {
			expectedWidth = data.Sprite.Rect.Width
		}
		if expectedHeight < 0 && data.Sprite != nil {
			expectedHeight = data.Sprite.Rect.Height
		}
		if expectedWidth >= 0 && builtChild.Width() != float64(expectedWidth) {
			t.Fatalf("expected width %d, got %v", expectedWidth, builtChild.Width())
		}
		if expectedHeight >= 0 && builtChild.Height() != float64(expectedHeight) {
			t.Fatalf("expected height %d, got %v", expectedHeight, builtChild.Height())
		}
	}

	var textIndex int = -1
	for idx, meta := range component.Component.Children {
		if (meta.Type == assets.ObjectTypeText || meta.Type == assets.ObjectTypeRichText) && meta.Text != "" {
			textIndex = idx
			break
		}
	}
	if textIndex >= 0 {
		builtChild := built.ChildAt(textIndex)
		if builtChild == nil {
			t.Fatalf("text child missing at index %d", textIndex)
		}
		meta := component.Component.Children[textIndex]
		data, ok := builtChild.Data().(string)
		if !ok {
			t.Fatalf("expected text child to store string data")
		}
		if data != meta.Text {
			t.Fatalf("expected text %q, got %q", meta.Text, data)
		}
	}

	var buttonIndex int = -1
	for idx, meta := range component.Component.Children {
		if meta.Type == assets.ObjectTypeButton {
			buttonIndex = idx
			break
		}
	}
	if buttonIndex >= 0 {
		builtChild := built.ChildAt(buttonIndex)
		if builtChild == nil {
			t.Fatalf("button child missing at index %d", buttonIndex)
		}
		if builtChild.Data() == nil {
			t.Fatalf("expected button child to store package item reference")
		}
	}

	var loaderIndex int = -1
	for idx, meta := range component.Component.Children {
		if meta.Type == assets.ObjectTypeLoader {
			loaderIndex = idx
			break
		}
	}
	if loaderIndex >= 0 {
		builtChild := built.ChildAt(loaderIndex)
		if builtChild == nil {
			t.Fatalf("loader child missing at index %d", loaderIndex)
		}
		if builtChild.Data() == nil {
			t.Fatalf("expected loader child to store resource reference")
		}
	}

	if len(component.Component.Controllers) > 0 {
		controllers := built.Controllers()
		if len(controllers) != len(component.Component.Controllers) {
			t.Fatalf("expected %d controllers, got %d", len(component.Component.Controllers), len(controllers))
		}
		if controllers[0].Name != component.Component.Controllers[0].Name {
			t.Fatalf("controller name mismatch")
		}
	}
}

func TestBuildComponentControllers(t *testing.T) {
	root := filepath.Join("..", "..", "..", "demo", "assets")
	data, err := os.ReadFile(filepath.Join(root, "MainMenu.fui"))
	if err != nil {
		t.Skipf("demo assets unavailable: %v", err)
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(root, "MainMenu"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	var mainComponent *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Main" {
			mainComponent = item
			break
		}
	}
	if mainComponent == nil {
		t.Fatalf("Main component not found in MainMenu.fui")
	}
	if len(mainComponent.Component.Controllers) == 0 {
		t.Fatalf("expected component controllers in Main component")
	}

	factory := NewFactory(nil, nil)
	rootComp, err := factory.BuildComponent(context.Background(), pkg, mainComponent)
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}
	controllers := rootComp.Controllers()
	if len(controllers) != len(mainComponent.Component.Controllers) {
		t.Fatalf("expected %d controllers, got %d", len(mainComponent.Component.Controllers), len(controllers))
	}
	if controllers[0].Name != mainComponent.Component.Controllers[0].Name {
		t.Fatalf("controller name mismatch: %s != %s", controllers[0].Name, mainComponent.Component.Controllers[0].Name)
	}
}

func TestBuildComponentAppliesTransforms(t *testing.T) {
	component := &assets.PackageItem{
		Type: assets.PackageItemTypeComponent,
		Component: &assets.ComponentData{
			InitWidth:   200,
			InitHeight:  100,
			PivotX:      0.5,
			PivotY:      0.5,
			PivotAnchor: true,
			Children: []assets.ComponentChild{
				{
					Name:        "transformChild",
					Type:        assets.ObjectTypeGraph,
					X:           20,
					Y:           30,
					Width:       60,
					Height:      40,
					ScaleX:      1.5,
					ScaleY:      0.75,
					Rotation:    45,
					SkewX:       10,
					SkewY:       -5,
					PivotX:      0.25,
					PivotY:      0.75,
					PivotAnchor: true,
					Alpha:       0.8,
				},
			},
		},
	}

	factory := NewFactory(nil, nil)
	root, err := factory.BuildComponent(context.Background(), &assets.Package{}, component)
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	px, py := root.Pivot()
	if px != 0.5 || py != 0.5 {
		t.Fatalf("expected root pivot (0.5,0.5), got (%v,%v)", px, py)
	}
	if !root.PivotAsAnchor() {
		t.Fatalf("expected root pivot to act as anchor")
	}

	child := root.ChildAt(0)
	if child == nil {
		t.Fatalf("expected child component")
	}
	sx, sy := child.Scale()
	if sx != 1.5 || sy != 0.75 {
		t.Fatalf("expected child scale (1.5,0.75), got (%v,%v)", sx, sy)
	}
	const epsilon = 1e-6
	if diff := child.Rotation() - (45 * math.Pi / 180); diff > epsilon || diff < -epsilon {
		t.Fatalf("expected rotation %v radians, got %v", 45*math.Pi/180, child.Rotation())
	}
	cpx, cpy := child.Pivot()
	if cpx != 0.25 || cpy != 0.75 {
		t.Fatalf("expected child pivot (0.25,0.75), got (%v,%v)", cpx, cpy)
	}
	if diff := math.Abs(child.Alpha() - 0.8); diff > epsilon {
		t.Fatalf("expected alpha 0.8, got %v", child.Alpha())
	}
	pos := child.DisplayObject().Position()
	rawX := 20.0 - 0.25*60.0
	rawY := 30.0 - 0.75*40.0
	offsetX, offsetY := computePivotOffset(60, 40, 0.25, 0.75, 45*math.Pi/180, 10*math.Pi/180, -5*math.Pi/180, 1.5, 0.75)
	expectedX := rawX + offsetX
	expectedY := rawY + offsetY
	if math.Abs(pos.X-expectedX) > epsilon || math.Abs(pos.Y-expectedY) > epsilon {
		t.Fatalf("expected anchored sprite position (%v,%v), got (%v,%v)", expectedX, expectedY, pos.X, pos.Y)
	}
	cskx, csky := child.Skew()
	if math.Abs(cskx-10*math.Pi/180) > epsilon || math.Abs(csky-(-5)*math.Pi/180) > epsilon {
		t.Fatalf("expected skew (10,-5) degrees -> (%v,%v) radians, got (%v,%v)", 10*math.Pi/180, -5*math.Pi/180, cskx, csky)
	}
}

func TestBuildComponentCrossPackageReference(t *testing.T) {
	root := filepath.Join("..", "..", "..", "demo", "assets")
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Skipf("demo assets unavailable: %v", err)
	}

	packages := make(map[string]*assets.Package)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".fui" {
			continue
		}
		name := entry.Name()
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		pkg, err := assets.ParsePackage(data, filepath.Join(root, base))
		if err != nil {
			continue
		}
		packages[pkg.ID] = pkg
	}
	if len(packages) == 0 {
		t.Skip("no packages parsed from demo assets")
	}

	ctx := context.Background()
	found := false
	for _, pkg := range packages {
		for _, item := range pkg.Items {
			if item.Type != assets.PackageItemTypeComponent || item.Component == nil {
				continue
			}
			for _, child := range item.Component.Children {
				if child.PackageID == "" || child.PackageID == pkg.ID {
					continue
				}
				dep := packages[child.PackageID]
				if dep == nil {
					for _, candidate := range packages {
						if candidate.Name == child.PackageID {
							dep = candidate
							break
						}
					}
				}
				if dep == nil || dep.ItemByID(child.Src) == nil {
					continue
				}
				loader := assets.NewFileLoader(root)
				factory := NewFactoryWithLoader(nil, loader)
				factory.RegisterPackage(pkg)
				if _, err := factory.BuildComponent(ctx, pkg, item); err != nil {
					t.Fatalf("BuildComponent failed for cross-package reference: %v", err)
				}
				found = true
				break
			}
			if found {
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Skip("no cross-package references found in demo assets")
	}
}

func computePivotOffset(width, height float64, pivotX, pivotY float64, rotation, skewX, skewY float64, scaleX, scaleY float64) (float64, float64) {
	px := pivotX * width
	py := pivotY * height
	cosY := math.Cos(rotation + skewY)
	sinY := math.Sin(rotation + skewY)
	cosX := math.Cos(rotation - skewX)
	sinX := math.Sin(rotation - skewX)
	a := cosY * scaleX
	b := sinY * scaleX
	c := -sinX * scaleY
	d := cosX * scaleY
	transformedX := a*px + c*py
	transformedY := b*px + d*py
	return px - transformedX, py - transformedY
}
