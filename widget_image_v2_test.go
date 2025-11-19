package fairygui

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// mockPackageItem 是 PackageItem 的模拟实现
type mockPackageItem struct {
	id     string
	name   string
	width  int
	height int
}

func (m *mockPackageItem) ID() string           { return m.id }
func (m *mockPackageItem) Name() string         { return m.name }
func (m *mockPackageItem) Type() ResourceType   { return ResourceTypeImage }
func (m *mockPackageItem) Data() interface{}    { return nil }
func (m *mockPackageItem) Width() int           { return m.width }
func (m *mockPackageItem) Height() int          { return m.height }

func TestNewImage(t *testing.T) {
	img := NewImage()
	if img == nil {
		t.Fatal("NewImage() returned nil")
	}

	if img.Object == nil {
		t.Error("Image.Object is nil")
	}

	// 检查默认属性
	expTint := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	if img.TintColor() != expTint {
		t.Errorf("默认染色颜色不正确: got %+v, want %+v", img.TintColor(), expTint)
	}

	if img.Flip() != FlipTypeNone {
		t.Errorf("默认翻转类型不正确: got %v, want %v", img.Flip(), FlipTypeNone)
	}

	if img.Touchable() {
		t.Error("Image 应该默认不拦截事件")
	}
}

func TestImage_SetPackageItem(t *testing.T) {
	img := NewImage()
	mockItem := &mockPackageItem{
		id:     "test123",
		name:   "testImage",
		width:  100,
		height: 200,
	}

	img.SetPackageItem(mockItem)

	if img.PackageItem() != mockItem {
		t.Error("PackageItem() 返回的值不正确")
	}

	// 检查尺寸是否自动设置
	x, y := img.Position()
	width, height := img.Size()

	if x != 0 || y != 0 {
		t.Errorf("位置不正确: got (%.1f, %.1f), want (0, 0)", x, y)
	}

	if width != 100 || height != 200 {
		t.Errorf("尺寸不正确: got (%.1f, %.1f), want (100, 200)", width, height)
	}
}

func TestImage_SetPackageItem_Nil(t *testing.T) {
	img := NewImage()
	img.SetPackageItem(nil)

	if img.PackageItem() != nil {
		t.Error("PackageItem 应该为 nil")
	}
}

func TestImage_SetTintColor(t *testing.T) {
	img := NewImage()
	c := color.RGBA{R: 255, G: 0, B: 0, A: 128}

	img.SetTintColor(c)

	if img.TintColor() != c {
		t.Errorf("染色颜色不正确: got %+v, want %+v", img.TintColor(), c)
	}
}

func TestImage_SetFlip(t *testing.T) {
	img := NewImage()

	flips := []FlipType{
		FlipTypeNone,
		FlipTypeHorizontal,
		FlipTypeVertical,
		FlipTypeBoth,
	}

	for _, flip := range flips {
		img.SetFlip(flip)
		if img.Flip() != flip {
			t.Errorf("翻转类型不正确: got %v, want %v", img.Flip(), flip)
		}
	}
}

func TestImage_SetFill(t *testing.T) {
	img := NewImage()

	method := 1
	origin := 2
	clockwise := true
	amount := 0.75

	img.SetFill(method, origin, clockwise, amount)

	gotMethod, gotOrigin, gotClockwise, gotAmount := img.Fill()

	if gotMethod != method {
		t.Errorf("fillMethod 不正确: got %d, want %d", gotMethod, method)
	}

	if gotOrigin != origin {
		t.Errorf("fillOrigin 不正确: got %d, want %d", gotOrigin, origin)
	}

	if gotClockwise != clockwise {
		t.Errorf("fillClockwise 不正确: got %v, want %v", gotClockwise, clockwise)
	}

	if gotAmount != amount {
		t.Errorf("fillAmount 不正确: got %.2f, want %.2f", gotAmount, amount)
	}
}

func TestImage_Scale9Grid(t *testing.T) {
	img := NewImage()

	// 初始状态应该是 nil
	if img.Scale9Grid() != nil {
		t.Error("初始 Scale9Grid 应该是 nil")
	}

	grid := &Rect{
		X:      10,
		Y:      20,
		Width:  30,
		Height: 40,
	}

	img.SetScale9Grid(grid)

	got := img.Scale9Grid()
	if got == nil {
		t.Fatal("Scale9Grid 不应该为 nil")
	}

	if *got != *grid {
		t.Errorf("Scale9Grid 不正确: got %+v, want %+v", got, grid)
	}

	// 测试设置为 nil
	img.SetScale9Grid(nil)
	if img.Scale9Grid() != nil {
		t.Error("Scale9Grid 应该为 nil")
	}
}

func TestImage_PackageItemAutoSize(t *testing.T) {
	tests := []struct {
		name        string
		itemWidth   int
		itemHeight  int
		wantWidth   float64
		wantHeight  float64
	}{
		{"小尺寸", 50, 60, 50, 60},
		{"大尺寸", 500, 600, 500, 600},
		{"正方形", 100, 100, 100, 100},
		{"宽度为0", 0, 100, 0, 100},
		{"高度为0", 100, 0, 100, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := NewImage()
			mockItem := &mockPackageItem{
				id:     "test",
				name:   "test",
				width:  tt.itemWidth,
				height: tt.itemHeight,
			}

			img.SetPackageItem(mockItem)

			width, height := img.Size()
			if width != tt.wantWidth || height != tt.wantHeight {
				t.Errorf("尺寸不正确: got (%.1f, %.1f), want (%.1f, %.1f)",
					width, height, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestImage_Chaining(t *testing.T) {
	img := NewImage()

	mockItem := &mockPackageItem{
		id:     "test",
		name:   "test",
		width:  100,
		height: 100,
	}

	// 测试方法链式调用
	img.SetPackageItem(mockItem)
	img.SetTintColor(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	img.SetFlip(FlipTypeHorizontal)

	// 验证所有设置都生效
	if img.PackageItem() != mockItem {
		t.Error("PackageItem 设置失败")
	}

	if img.Flip() != FlipTypeHorizontal {
		t.Error("Flip 设置失败")
	}
}

func TestImage_ObjectInterface(t *testing.T) {
	img := NewImage()

	// 测试 Object 接口的兼容性
	// 确保 Image 可以被视为 DisplayObject
	var obj DisplayObject = img
	if obj == nil {
		t.Fatal("Image 应该实现 DisplayObject 接口")
	}

	// 测试一些基本的 Object 方法
	obj.SetPosition(100, 200)
	x, y := obj.Position()
	if x != 100 || y != 200 {
		t.Errorf("Position 不正确: got (%.1f, %.1f), want (100, 200)", x, y)
	}

	obj.SetAlpha(0.5)
	if obj.Alpha() != 0.5 {
		t.Errorf("Alpha 不正确: got %.1f, want 0.5", obj.Alpha())
	}
}

func TestImage_Draw(t *testing.T) {
	img := NewImage()

	// 创建一个离屏图像用于测试绘制
	screen := ebiten.NewImage(800, 600)

	// 设置一些属性
	mockItem := &mockPackageItem{
		id:     "test",
		name:   "test",
		width:  100,
		height: 100,
	}
	img.SetPackageItem(mockItem)
	img.SetPosition(50, 50)
	img.SetVisible(true)

	// 绘制（不应该 panic）
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Draw 方法 panic: %v", r)
		}
	}()

	img.Draw(screen)
}

func TestImage_Events(t *testing.T) {
	img := NewImage()

	// 测试点击事件
	clicked := false
	img.OnClick(func() {
		clicked = true
	})

	if !img.HasListener(EventClick) {
		t.Error("应该有点击事件监听器")
	}

	// 触发点击事件
	event := NewUIEvent(EventClick, img, nil)
	img.DispatchEvent(event)

	if !clicked {
		t.Error("点击事件处理器没有被调用")
	}
}

func TestImage_Visible(t *testing.T) {
	img := NewImage()

	if !img.Visible() {
		t.Error("新创建的 Image 应该默认可见")
	}

	img.SetVisible(false)
	if img.Visible() {
		t.Error("SetVisible(false) 失败")
	}

	img.SetVisible(true)
	if !img.Visible() {
		t.Error("SetVisible(true) 失败")
	}
}

func TestImage_Alpha(t *testing.T) {
	img := NewImage()

	alphas := []float64{0, 0.25, 0.5, 0.75, 1.0}

	for _, alpha := range alphas {
		img.SetAlpha(alpha)
		if img.Alpha() != alpha {
			t.Errorf("Alpha 设置失败: got %.2f, want %.2f", img.Alpha(), alpha)
		}
	}
}

func TestImage_Transform(t *testing.T) {
	img := NewImage()

	// 测试缩放
	img.SetScale(2.0, 3.0)
	sx, sy := img.Scale()
	if sx != 2.0 || sy != 3.0 {
		t.Errorf("Scale 不正确: got (%.1f, %.1f), want (2.0, 3.0)", sx, sy)
	}

	// 测试旋转（角度）
	img.SetRotation(90)
	rot := img.Rotation()
	if rot != 90 {
		t.Errorf("Rotation 不正确: got %.1f, want 90", rot)
	}

	// 测试倾斜（角度）
	img.SetSkew(45, 30)
	skewX, skewY := img.Skew()
	if skewX < 44.9 || skewX > 45.1 || skewY < 29.9 || skewY > 30.1 {
		t.Errorf("Skew 不正确: got (%.1f, %.1f), want (45, 30)", skewX, skewY)
	}

	// 测试锚点
	img.SetPivot(0.5, 0.5)
	px, py := img.Pivot()
	if px != 0.5 || py != 0.5 {
		t.Errorf("Pivot 不正确: got (%.1f, %.1f), want (0.5, 0.5)", px, py)
	}
}

func TestAssertImage(t *testing.T) {
	img := NewImage()

	// 测试 AssertImage
	result, ok := AssertImage(img)
	if !ok {
		t.Error("AssertImage 应该成功")
	}
	if result != img {
		t.Error("AssertImage 返回的对象不正确")
	}

	// 测试 IsImage
	if !IsImage(img) {
		t.Error("IsImage 应该返回 true")
	}

	// 测试不是 Image 的情况
	obj := NewObject()
	_, ok = AssertImage(obj)
	if ok {
		t.Error("AssertImage 对非 Image 对象应该失败")
	}

	if IsImage(obj) {
		t.Error("IsImage 对非 Image 对象应该返回 false")
	}
}

func TestImage_SetPackageItem_Change(t *testing.T) {
	img := NewImage()

	item1 := &mockPackageItem{
		id:     "item1",
		name:   "image1",
		width:  100,
		height: 100,
	}

	item2 := &mockPackageItem{
		id:     "item2",
		name:   "image2",
		width:  200,
		height: 200,
	}

	// 设置第一个 item
	img.SetPackageItem(item1)
	w1, h1 := img.Size()
	if w1 != 100 || h1 != 100 {
		t.Errorf("第一次设置后的尺寸不正确: got (%.1f, %.1f)", w1, h1)
	}

	// 设置为相同的 item（不应该改变）
	img.SetPackageItem(item1)
	w1b, h1b := img.Size()
	if w1b != 100 || h1b != 100 {
		t.Errorf("重复设置相同 item 后尺寸改变: got (%.1f, %.1f)", w1b, h1b)
	}

	// 设置为不同的 item
	img.SetPackageItem(item2)
	w2, h2 := img.Size()
	if w2 != 200 || h2 != 200 {
		t.Errorf("第二次设置后的尺寸不正确: got (%.1f, %.1f)", w2, h2)
	}
}

func TestImage_CustomDraw(t *testing.T) {
	img := NewImage()

	// 设置自定义绘制
	drawn := false
	img.SetCustomDraw(func(screen *ebiten.Image) {
		drawn = true
	})

	// 创建离屏图像并绘制
	screen := ebiten.NewImage(800, 600)
	img.Draw(screen)

	if !drawn {
		t.Error("自定义绘制回调没有被调用")
	}
}
