package core

import (
	"testing"
	"time"
)

func BenchmarkObjectCreation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		obj := NewGObject()
		obj.SetPosition(10, 20)
		obj.SetSize(100, 50)
		obj.SetAlpha(0.8)
		obj.SetVisible(true)
	}
}

func BenchmarkComponentCreation(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		comp := NewGComponent()
		comp.SetPosition(10, 20)
		comp.SetSize(400, 300)
	}
}

func BenchmarkComponentAddChild(b *testing.B) {
	b.ReportAllocs()
	comp := NewGComponent()
	children := make([]*GObject, 100)
	for i := 0; i < 100; i++ {
		children[i] = NewGObject()
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp.AddChild(children[i%100])
	}
}

func BenchmarkGearApply(b *testing.B) {
	b.ReportAllocs()
	comp := NewGComponent()
	obj := NewGObject()
	comp.AddChild(obj)
	ctrl := NewController("c1")
	ctrl.SetPages([]string{"p1", "p2"}, []string{"Page1", "Page2"})
	comp.AddController(ctrl)

	gear := obj.GetGear(1)
	gear.SetController(ctrl)
	ctrl.SetSelectedIndex(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			ctrl.SetSelectedIndex(0)
		} else {
			ctrl.SetSelectedIndex(1)
		}
	}
}

func BenchmarkTickAll(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < 10; i++ {
		RegisterTicker(func(delta time.Duration) {})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tickAll(16 * time.Millisecond)
	}
}

