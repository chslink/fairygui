package main

import (
	"fmt"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
)

func main() {
	env := testutil.NewTest()
	
	// 创建ScrollBar
	sb := core.NewGComponent()
	sb.SetSize(20, 100)
	
	// 创建template
	tmpl := core.NewGComponent()
	tmpl.SetSize(20, 100)
	
	bar := core.NewGObject()
	bar.SetName("bar")
	bar.SetSize(10, 90)
	tmpl.AddChild(bar)
	
	grip := core.NewGObject()
	grip.SetName("grip")
	grip.SetSize(10, 30)
	grip.SetXY(0, 0) // 设置在template中的位置
	tmpl.AddChild(grip)
	
	sb.AddChild(tmpl.GObject)
	
	// 获取container display object
	display := tmpl.GObject.DisplayObject()
	
	fmt.Printf("Template GObject: %p\n", tmpl.GObject)
	fmt.Printf("Template DisplayObject: %p\n", display)
	fmt.Printf("Bar position in template: (%.2f, %.2f)\n", bar.X(), bar.Y())
	fmt.Printf("Grip position in template: (%.2f, %.2f)\n", grip.X(), grip.Y())
	
	// 测试坐标转换
	globalPos := laya.Point{X: 10, Y: 15}
	localPos := display.GlobalToLocal(globalPos)
	fmt.Printf("Global (10, 15) -> Local (%.2f, %.2f)\n", localPos.X, localPos.Y)
}
