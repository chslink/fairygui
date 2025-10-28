package widgets

import (
	"testing"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

// MockObjectCreator 模拟对象创建器
type MockObjectCreator struct {
	objects map[string]*core.GObject
}

func NewMockObjectCreator() *MockObjectCreator {
	return &MockObjectCreator{
		objects: make(map[string]*core.GObject),
	}
}

func (m *MockObjectCreator) CreateObject(url string) *core.GObject {
	if obj, exists := m.objects[url]; exists {
		return obj
	}

	// 创建模拟对象
	obj := core.NewGObject()
	obj.SetName(url)
	obj.SetSize(100, 30) // 模拟项目尺寸
	m.objects[url] = obj
	return obj
}

func TestGListVirtualBasic(t *testing.T) {
	list := NewList()

	// 设置列表尺寸（重要：否则视图高度为0，不会渲染任何项目）
	list.SetSize(200, 300)

	// 启用虚拟化
	list.SetVirtual(true)
	if !list.IsVirtual() {
		t.Fatal("虚拟化应该被启用")
	}

	// 设置项目数量
	list.SetNumItems(100)
	if list.NumItems() != 100 {
		t.Fatalf("期望项目数量100，实际%d", list.NumItems())
	}

	// 设置默认项目（重要：否则无法创建对象）
	list.SetDefaultItem("ui://test/item")

	// 设置项目尺寸
	itemSize := &laya.Point{X: 100, Y: 30}
	list.SetVirtualItemSize(itemSize)

	// 设置对象创建器
	creator := NewMockObjectCreator()
	list.SetObjectCreator(creator)

	// 设置项目渲染器
	renderCount := 0
	list.SetItemRenderer(func(index int, item *core.GObject) {
		renderCount++
		t.Logf("渲染项目 %d", index)
	})

	// 检查列表状态
	t.Logf("列表尺寸: %.0fx%.0f", list.Width(), list.Height())
	t.Logf("视图高度: %d", list.getViewHeight())
	t.Logf("项目尺寸: %.0fx%.0f", list.itemSize.X, list.itemSize.Y)
	t.Logf("项目数量: %d", list.numItems)
	t.Logf("实际项目数量: %d", list.realNumItems)
	t.Logf("当前起始索引: %d", list.firstIndex)

	// 刷新虚拟列表
	list.RefreshVirtualList()

	// 验证渲染被调用
	t.Logf("渲染调用次数: %d", renderCount)
	if renderCount == 0 {
		t.Fatal("项目渲染器应该被调用")
	}
}

func TestGListVirtualItemProvider(t *testing.T) {
	list := NewList()
	list.SetSize(200, 300) // 设置列表尺寸
	list.SetVirtual(true)
	list.SetNumItems(10)

	// 设置项目提供者
	providerCalls := 0
	list.SetItemProvider(func(index int) string {
		providerCalls++
		return "ui://test/item"
	})

	// 设置对象创建器
	creator := NewMockObjectCreator()
	list.SetObjectCreator(creator)

	// 设置项目渲染器
	rendererCalls := 0
	list.SetItemRenderer(func(index int, item *core.GObject) {
		rendererCalls++
	})

	// 刷新虚拟列表
	list.RefreshVirtualList()

	if providerCalls == 0 {
		t.Fatal("项目提供者应该被调用")
	}
	if rendererCalls == 0 {
		t.Fatal("项目渲染器应该被调用")
	}
}

func TestGListVirtualLoop(t *testing.T) {
	list := NewList()
	list.SetVirtual(true)
	list.SetLoop(true)

	if !list.IsLoop() {
		t.Fatal("循环模式应该被启用")
	}

	list.SetNumItems(10)

	// 循环模式下应该有6倍的项目
	if list.realNumItems != 60 {
		t.Fatalf("循环模式下期望60个项目，实际%d", list.realNumItems)
	}
}

func TestGListVirtualSelection(t *testing.T) {
	list := NewList()
	list.SetSize(200, 300) // 设置列表尺寸
	list.SetVirtual(true)
	list.SetNumItems(10)

	// 设置对象创建器
	creator := NewMockObjectCreator()
	list.SetObjectCreator(creator)

	// 设置项目渲染器
	list.SetItemRenderer(func(index int, item *core.GObject) {
		// 模拟选择状态
		if index == 5 {
			if button, ok := item.Data().(*GButton); ok {
				button.SetSelected(true)
			}
		}
	})

	// 刷新虚拟列表
	list.RefreshVirtualList()

	// 注意：虚拟列表的选择机制与非虚拟列表不同
	// 这里我们主要测试渲染器是否正确调用
	// 实际的选择状态应该在itemRenderer中处理
	renderedIndex5 := false
	list.SetItemRenderer(func(index int, item *core.GObject) {
		if index == 5 {
			renderedIndex5 = true
			if button, ok := item.Data().(*GButton); ok {
				button.SetSelected(true)
			}
		}
	})

	// 重新刷新以测试新的渲染器
	list.RefreshVirtualList()

	if !renderedIndex5 {
		t.Fatal("索引5的项目应该被渲染")
	}
}

func TestGListVirtualToNonVirtual(t *testing.T) {
	list := NewList()

	// 先启用虚拟化
	list.SetVirtual(true)
	list.SetNumItems(10)

	// 添加一些实际项目
	for i := 0; i < 3; i++ {
		item := core.NewGObject()
		item.SetName("item")
		list.AddItem(item)
	}

	// 禁用虚拟化
	list.SetVirtual(false)

	if list.IsVirtual() {
		t.Fatal("虚拟化应该被禁用")
	}

	// 验证实际项目仍然存在
	if len(list.Items()) != 3 {
		t.Fatalf("期望3个实际项目，实际%d", len(list.Items()))
	}
}

func TestGObjectPool(t *testing.T) {
	pool := NewGObjectPool()

	// 创建模拟对象创建器
	creator := NewMockObjectCreator()
	pool.creator = creator

	// 第一次获取对象，应该创建新对象
	obj1 := pool.GetObject("ui://test/item1")
	if obj1 == nil {
		t.Fatal("应该能获取到对象")
	}

	// 返回对象到池中
	pool.ReturnObject(obj1)

	// 再次获取，应该从池中获取
	obj2 := pool.GetObject("ui://test/item1")
	if obj2 != obj1 {
		t.Fatal("应该从池中获取相同的对象")
	}

	// 验证池的计数
	if pool.Count() != 0 {
		t.Fatalf("池计数应该为0，实际%d", pool.Count())
	}
}

func TestGListLayoutCalculation(t *testing.T) {
	list := NewList()
	list.SetVirtual(true)
	list.SetNumItems(20)

	// 设置单列布局
	list.layout = ListLayoutTypeSingleColumn
	list.SetVirtualItemSize(&laya.Point{X: 100, Y: 30})

	// 设置对象创建器
	creator := NewMockObjectCreator()
	list.SetObjectCreator(creator)

	// 刷新虚拟列表
	list.RefreshVirtualList()

	// 验证每行项目数
	if list.curLineItemCount != 1 {
		t.Fatalf("单列布局期望每行1个项目，实际%d", list.curLineItemCount)
	}
}