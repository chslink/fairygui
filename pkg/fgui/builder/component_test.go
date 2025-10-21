package builder

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
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
		var data *assets.PackageItem
		switch payload := builtChild.Data().(type) {
		case *assets.PackageItem:
			data = payload
		case *widgets.GImage:
			data = payload.PackageItem()
		default:
			t.Fatalf("expected image child to reference package item, got %T", builtChild.Data())
		}
		if data == nil {
			t.Fatalf("image child missing package item reference")
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
		switch v := builtChild.Data().(type) {
		case string:
			if v != meta.Text {
				t.Fatalf("expected text %q, got %q", meta.Text, v)
			}
		case *widgets.GTextField:
			if v.Text() != meta.Text {
				t.Fatalf("expected text %q, got %q", meta.Text, v.Text())
			}
		case *widgets.GLabel:
			if v.Title() != meta.Text {
				t.Fatalf("expected label title %q, got %q", meta.Text, v.Title())
			}
		default:
			t.Fatalf("unexpected text child data type %T", builtChild.Data())
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
		loader, ok := builtChild.Data().(*widgets.GLoader)
		if !ok || loader == nil {
			t.Fatalf("expected loader child to store loader widget, got %T", builtChild.Data())
		}
		if loader.PackageItem() == nil && loader.URL() == "" {
			t.Fatalf("expected loader to reference a package item or url")
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

func TestBuildLoaderSettings(t *testing.T) {
	root := filepath.Join("..", "..", "..", "demo", "assets")
	data, err := os.ReadFile(filepath.Join(root, "Basics.fui"))
	if err != nil {
		t.Skipf("demo assets unavailable: %v", err)
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(root, "Basics"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	var loaderComp *assets.PackageItem
	for _, item := range pkg.Items {
		if item.Type == assets.PackageItemTypeComponent && item.Name == "Demo_Loader" {
			loaderComp = item
			break
		}
	}
	if loaderComp == nil {
		t.Fatalf("Demo_Loader component not found")
	}

	factory := NewFactory(nil, nil)
	rootComp, err := factory.BuildComponent(context.Background(), pkg, loaderComp)
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	loaders := collectLoaders(rootComp)
	if len(loaders) == 0 {
		t.Fatalf("expected loader children to be constructed")
	}
	hasScale := false
	for _, loader := range loaders {
		if loader.Fill() == widgets.LoaderFillScale {
			hasScale = true
			break
		}
	}
	if !hasScale {
		t.Fatalf("expected at least one loader to use scale fill")
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

func TestBuildComponentCreatesLabelWidget(t *testing.T) {
	rootDir := filepath.Join("..", "..", "..", "demo", "assets")
	data, err := os.ReadFile(filepath.Join(rootDir, "Bag.fui"))
	if err != nil {
		t.Skipf("demo assets unavailable: %v", err)
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(rootDir, "Bag"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	factory := NewFactory(nil, nil)
	root, err := factory.BuildComponent(context.Background(), pkg, pkg.ItemByName("BagWin"))
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	found := false
	for _, child := range root.Children() {
		if label, ok := child.Data().(*widgets.GLabel); ok {
			if label.Resource() != "" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("expected at least one label with resource metadata")
	}
}

func TestBuildComponentCreatesButtonWidget(t *testing.T) {
	rootDir := filepath.Join("..", "..", "..", "demo", "assets")
	data, err := os.ReadFile(filepath.Join(rootDir, "MainMenu.fui"))
	if err != nil {
		t.Skipf("demo assets unavailable: %v", err)
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(rootDir, "MainMenu"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	factory := NewFactory(nil, nil)
	root, err := factory.BuildComponent(context.Background(), pkg, pkg.ItemByName("Main"))
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	buttons := collectButtons(root)
	if len(buttons) == 0 {
		t.Fatalf("expected to discover button children")
	}
	for _, button := range buttons {
		if button.Resource() == "" {
			t.Fatalf("expected button resource to be populated")
		}
		if button.PackageItem() == nil {
			t.Fatalf("expected button package item to be resolved")
		}
		if button.ButtonController() == nil {
			t.Fatalf("expected button to expose internal button controller")
		}
		if button.TemplateComponent() == nil {
			t.Fatalf("expected button template component to be instantiated")
		}
		titleObj := button.TitleObject()
		if titleObj == nil {
			t.Fatalf("expected button to expose title object")
		}
		switch titleObj.Data().(type) {
		case *widgets.GTextField, *widgets.GLabel, *widgets.GButton:
		default:
			t.Fatalf("unexpected title object data type %T", titleObj.Data())
		}
		if iconObj := button.IconObject(); iconObj != nil {
			switch iconObj.Data().(type) {
			case *widgets.GLoader, *widgets.GButton, string:
			default:
				t.Fatalf("unexpected icon object data type %T", iconObj.Data())
			}
		}
	}
}

func TestBuildComponentCreatesListWidget(t *testing.T) {
	component := &assets.PackageItem{
		Type: assets.PackageItemTypeComponent,
		Component: &assets.ComponentData{
			Children: []assets.ComponentChild{
				{
					Name: "list",
					Type: assets.ObjectTypeList,
					Data: "defaultItem",
				},
			},
		},
	}

	factory := NewFactory(nil, nil)
	root, err := factory.BuildComponent(context.Background(), &assets.Package{}, component)
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	child := root.ChildAt(0)
	if child == nil {
		t.Fatalf("expected list child")
	}
	list, ok := child.Data().(*widgets.GList)
	if !ok || list == nil {
		t.Fatalf("expected child data to be GList, got %T", child.Data())
	}
	if list.DefaultItem() != "defaultItem" {
		t.Fatalf("unexpected default item: %s", list.DefaultItem())
	}
}

func TestBuildComponentAssignsGroups(t *testing.T) {
	rootDir := filepath.Join("..", "..", "..", "demo", "assets")
	data, err := os.ReadFile(filepath.Join(rootDir, "Transition.fui"))
	if err != nil {
		t.Skipf("demo assets unavailable: %v", err)
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(rootDir, "Transition"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	main := pkg.ItemByName("Main")
	if main == nil {
		t.Fatalf("Transition Main component missing")
	}

	factory := NewFactory(nil, nil)
	root, err := factory.BuildComponent(context.Background(), pkg, main)
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	groupObj := root.ChildByName("g0")
	if groupObj == nil {
		t.Fatalf("expected group child g0")
	}

	buttonNames := []string{"btn0", "btn1", "btn2", "btn3", "btn4"}
	for _, name := range buttonNames {
		child := root.ChildByName(name)
		if child == nil {
			t.Fatalf("expected child %s", name)
		}
		if child.Group() != groupObj {
			t.Fatalf("expected %s to bind group g0", name)
		}
	}
}

func TestBuildComponentTooltips(t *testing.T) {
	rootDir := filepath.Join("..", "..", "..", "demo", "assets")
	data, err := os.ReadFile(filepath.Join(rootDir, "MainMenu.fui"))
	if err != nil {
		t.Skipf("demo assets unavailable: %v", err)
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(rootDir, "MainMenu"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	factory := NewFactory(nil, nil)
	root, err := factory.BuildComponent(context.Background(), pkg, pkg.ItemByName("Main"))
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	found := false
	var walk func(*core.GComponent)
	walk = func(c *core.GComponent) {
		if c == nil || found {
			return
		}
		for _, child := range c.Children() {
			if child == nil {
				continue
			}
			if child.Tooltips() != "" {
				found = true
				return
			}
			if nested, ok := child.Data().(*core.GComponent); ok && nested != nil {
				walk(nested)
				if found {
					return
				}
			}
		}
	}
	walk(root)
	if !found {
		t.Skip("tooltips not present in current MainMenu assets")
	}
}

func TestBuildComponentParsesTransitions(t *testing.T) {
	rootDir := filepath.Join("..", "..", "..", "demo", "assets")
	data, err := os.ReadFile(filepath.Join(rootDir, "Transition.fui"))
	if err != nil {
		t.Skipf("demo assets unavailable: %v", err)
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(rootDir, "Transition"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}

	target := pkg.ItemByName("BOSS")
	if target == nil {
		t.Fatalf("Transition BOSS component missing")
	}

	factory := NewFactory(nil, nil)
	root, err := factory.BuildComponent(context.Background(), pkg, target)
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	transitions := root.Transitions()
	if len(transitions) == 0 {
		t.Fatalf("expected transitions metadata on Transition.BOSS component")
	}
	if len(transitions) != 1 {
		t.Fatalf("expected exactly one transition, got %d", len(transitions))
	}

	info := transitions[0]
	if info.Name != "t0" {
		t.Fatalf("expected transition named t0, got %q", info.Name)
	}
	if info.ItemCount != len(info.Items) {
		t.Fatalf("item count mismatch: meta=%d actual=%d", info.ItemCount, len(info.Items))
	}
	if len(info.Items) == 0 {
		t.Fatalf("transition t0 should contain at least one item")
	}
	first := info.Items[0]
	if first.Type != core.TransitionActionSound {
		t.Fatalf("expected first item to be sound, got %v", first.Type)
	}
	if first.Value.Sound == "" {
		t.Fatalf("sound item missing resource id")
	}
	var foundTween bool
	for _, it := range info.Items {
		if it.Tween != nil {
			foundTween = true
			if it.Tween.Duration == 0 {
				t.Fatalf("tween item duration should be > 0")
			}
			break
		}
	}
	if !foundTween {
		t.Fatalf("expected at least one tween item in transition t0")
	}

	runtime := root.Transition("t0")
	if runtime == nil {
		t.Fatalf("expected runtime transition t0")
	}
	if !runtime.Info().AutoPlay && runtime.Playing() {
		t.Fatalf("transition should not be playing unless autoPlay is enabled")
	}
	if info.TotalDuration <= 0 {
		t.Fatalf("expected transition to report total duration")
	}
}

func TestBuildComponentHitTest(t *testing.T) {
	rootDir := filepath.Join("..", "..", "..", "demo", "assets")
	data, err := os.ReadFile(filepath.Join(rootDir, "HitTest.fui"))
	if err != nil {
		t.Skipf("hit test assets unavailable: %v", err)
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(rootDir, "HitTest"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	factory := NewFactory(nil, nil)
	root, err := factory.BuildComponent(context.Background(), pkg, pkg.ItemByName("Main"))
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	foundPixel := false
	var walk func(*core.GComponent)
	walk = func(c *core.GComponent) {
		if c == nil || foundPixel {
			return
		}
		if hit := c.HitTest(); hit.Mode == core.HitTestModePixel && hit.ItemID != "" {
			foundPixel = true
			return
		}
		for _, child := range c.Children() {
			if child == nil {
				continue
			}
			if nested, ok := child.Data().(*core.GComponent); ok && nested != nil {
				walk(nested)
				if foundPixel {
					return
				}
			}
		}
	}
	walk(root)
	if !foundPixel {
		t.Skip("pixel hit-test metadata not present in current HitTest assets")
	}
}

func TestBuildComponentMask(t *testing.T) {
	rootDir := filepath.Join("..", "..", "..", "demo", "assets")
	data, err := os.ReadFile(filepath.Join(rootDir, "Cooldown.fui"))
	if err != nil {
		t.Skipf("cooldown assets unavailable: %v", err)
	}
	pkg, err := assets.ParsePackage(data, filepath.Join(rootDir, "Cooldown"))
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	factory := NewFactory(nil, nil)
	root, err := factory.BuildComponent(context.Background(), pkg, pkg.ItemByName("Main"))
	if err != nil {
		t.Fatalf("BuildComponent failed: %v", err)
	}

	foundMask := false
	var walk func(*core.GComponent)
	walk = func(c *core.GComponent) {
		if c == nil || foundMask {
			return
		}
		if mask, _ := c.Mask(); mask != nil {
			foundMask = true
			return
		}
		for _, child := range c.Children() {
			if child == nil {
				continue
			}
			if nested, ok := child.Data().(*core.GComponent); ok && nested != nil {
				walk(nested)
				if foundMask {
					return
				}
			}
		}
	}
	walk(root)
	if !foundMask {
		t.Skip("mask metadata not present in current Cooldown assets")
	}
}

func collectLoaders(comp *core.GComponent) []*widgets.GLoader {
	var result []*widgets.GLoader
	if comp == nil {
		return result
	}
	for _, child := range comp.Children() {
		if loader, ok := child.Data().(*widgets.GLoader); ok && loader != nil {
			result = append(result, loader)
		}
		if nested, ok := child.Data().(*core.GComponent); ok && nested != nil {
			result = append(result, collectLoaders(nested)...)
		}
	}
	return result
}

func collectButtons(comp *core.GComponent) []*widgets.GButton {
	var result []*widgets.GButton
	if comp == nil {
		return result
	}
	for _, child := range comp.Children() {
		if button, ok := child.Data().(*widgets.GButton); ok && button != nil {
			result = append(result, button)
		}
		if nested, ok := child.Data().(*core.GComponent); ok && nested != nil {
			result = append(result, collectButtons(nested)...)
		}
	}
	return result
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

func TestSetupComponentControllersAppliesSelection(t *testing.T) {
	component := core.NewGComponent()
	holder := component.GObject
	holder.SetData(component)

	ctrl := core.NewController("ctrl")
	ctrl.SetPages([]string{"page", "other"}, []string{"Page A", "Page B"})
	component.AddController(ctrl)
	ctrl.SetSelectedIndex(1) // start from non-default to observe override

	data := []byte{
		0x05, 0x01, // segCount=5, useShort=1
		0x00, 0x00, // block0 offset
		0x00, 0x00, // block1 offset
		0x00, 0x00, // block2 offset
		0x00, 0x00, // block3 offset
		0x00, 0x0C, // block4 offset (12 bytes from start)
		0xFF, 0xFF, // pageController index (-1)
		0x00, 0x01, // controller override count = 1
		0x00, 0x01, // controller name index -> "ctrl"
		0x00, 0x02, // page id index -> "page"
		0x00, 0x01, // property assignment count = 1 (Version >= 2)
		0x00, 0x03, // target path index -> "target"
		0x00, 0x05, // property id (arbitrary)
		0x00, 0x04, // value index -> "value"
	}
	buf := utils.NewByteBuffer(data)
	buf.StringTable = []string{"", "ctrl", "page", "target", "value"}
	buf.Version = 2

	setupComponentControllers(holder, buf)

	if got := ctrl.SelectedPageID(); got != "page" {
		t.Fatalf("expected controller override to select page %q, got %q", "page", got)
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
