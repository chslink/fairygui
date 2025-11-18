package scenes

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// LoopListDemo 循环列表演示场景
// 参考 TypeScript 版本: laya_src/demo/LoopListDemo.ts
type LoopListDemo struct {
	view *fgui.GComponent
	list *widgets.GList
}

// NewLoopListDemo 创建循环列表演示场景
func NewLoopListDemo() Scene {
	return &LoopListDemo{}
}

func (d *LoopListDemo) Name() string {
	return "LoopListDemo"
}

// Load 加载场景
func (d *LoopListDemo) Load(ctx context.Context, mgr *Manager) (*fgui.GComponent, error) {
	log.Println("📦 加载循环列表 demo...")

	env := mgr.Environment()

	// 加载LoopList资源包
	pkg, err := env.Package(ctx, "LoopList")
	if err != nil {
		return nil, err
	}

	// 加载Main组件
	item := chooseComponent(pkg, "Main")
	if item == nil {
		return nil, newMissingComponentError("LoopList", "Main")
	}

	view, err := env.Factory.BuildComponent(ctx, pkg, item)
	if err != nil {
		return nil, err
	}

	d.view = view

	// 查找list组件
	listObj := view.ChildByName("list")
	if listObj == nil {
		return nil, fmt.Errorf("找不到 list 组件")
	}

	// 转换为GList
	if data := listObj.Data(); data != nil {
		if list, ok := data.(*widgets.GList); ok {
			d.list = list

			// 调试：检查包加载情况
			defaultItemURL := list.DefaultItem()
			log.Printf("🔍 List defaultItem: %s", defaultItemURL)

			// 测试URL解析
			if defaultItemURL != "" {
				if item := fgui.GetItemByURL(defaultItemURL); item != nil {
					log.Printf("✅ 成功解析 defaultItem: 类型=%d, ID=%s, Name=%s",
						item.Type, item.ID, item.Name)
				} else {
					log.Printf("❌ 无法解析 defaultItem: %s", defaultItemURL)
				}
			}

			// 参考TypeScript版本：直接调用 SetVirtualAndLoop()
			list.SetVirtual(true)
			list.SetLoop(true)

			// 调试：检查滚动类型和列间距
			sp := list.GComponent.ScrollPane()
			if sp != nil {
				log.Printf("🔍 ScrollPane状态: 类型=%v, viewSize=%.0fx%.0f",
					sp.ScrollType(), sp.ViewWidth(), sp.ViewHeight())
			} else {
				log.Printf("⚠️  ScrollPane为nil")
			}
			log.Printf("🔍 List配置: columnGap=%d, lineGap=%d, layout=%d, autoResizeItem=%v",
				list.ColumnGap(), list.LineGap(), list.Layout(), list.AutoResizeItem())

			// 设置项目渲染器
			list.SetItemRenderer(d.renderListItem)

			// 设置项目数量
			list.SetNumItems(5)

			// 添加滚动事件
			list.On(laya.EventScroll, func(evt *laya.Event) {
				d.doSpecialEffect()
			})

			// 初始执行特效
			d.doSpecialEffect()

			log.Printf("✅ 循环列表配置完成: NumItems=%d, IsLoop=%v",
				list.NumItems(), list.IsLoop())
		} else {
			log.Printf("⚠️  list 不是 GList 类型: %T", data)
		}
	}

	log.Println("✅ 循环列表 demo 加载完成")
	return view, nil
}

// doSpecialEffect 执行特殊效果
// 根据与中间位置的距离改变缩放
// 对应 TypeScript: private doSpecialEffect(): void
func (d *LoopListDemo) doSpecialEffect() {
	if d.list == nil || d.view == nil {
		return
	}

	// 获取中间位置
	sp := d.list.GComponent.ScrollPane()
	if sp == nil {
		return
	}

	midX := sp.PosX() + d.list.GComponent.Width()/2

	// 遍历所有子项，根据距离中间位置的远近调整缩放
	cnt := d.list.NumChildren()
	for i := 0; i < cnt; i++ {
		obj := d.list.ChildAt(i)
		if obj == nil {
			continue
		}

		// 计算距离中间位置的距离
		dist := math.Abs(midX - (obj.X() + obj.Width()/2))

		if dist > obj.Width() { // 无交集
			obj.SetScale(1.0, 1.0)
		} else {
			// 根据距离调整缩放比例
			ss := 1.0 + (1.0-dist/obj.Width())*0.24
			obj.SetScale(ss, ss)
		}
	}

	// 更新文本显示，使用GetFirstChildInView方法
	// 修复：计算循环索引，对应TypeScript版本的逻辑
	// (getFirstChildInView() + 1) % numItems
	if textObj := d.view.ChildByName("n3"); textObj != nil {
		if textData := textObj.Data(); textData != nil {
			if textField, ok := textData.(*widgets.GTextField); ok {
				firstVisibleIndex := d.list.GetFirstChildInView()
				if firstVisibleIndex >= 0 {
					// 计算循环索引：对5取模得到0-4的范围
					cycledIndex := (firstVisibleIndex + 1) % d.list.NumItems()
					textField.SetText(fmt.Sprintf("%d", cycledIndex))
					log.Printf("🔄 循环索引: firstVisible=%d, cycled=%d, numItems=%d",
						firstVisibleIndex, cycledIndex, d.list.NumItems())
				} else {
					textField.SetText("No visible items")
				}
			}
		}
	}
}

// renderListItem 渲染列表项
// 对应 TypeScript: private renderListItem(index: number, obj: fgui.GObject): void
func (d *LoopListDemo) renderListItem(index int, obj *fgui.GObject) {
	if obj == nil {
		log.Printf("❌ obj is nil")
		return
	}

	// 设置中心点
	obj.SetPivot(0.5, 0.5)

	// 设置图标
	if button, ok := obj.Data().(*widgets.GButton); ok {
		// 构建图标URL
		iconURL := fmt.Sprintf("ui://LoopList/n%d", index+1)
		button.SetIcon(iconURL)
	} else {
		// 如果不是按钮，尝试通过其他方式设置图标
		if comp := fgui.ComponentFrom(obj); comp != nil {
			if iconObj := comp.ChildByName("icon"); iconObj != nil {
				if loader, ok := iconObj.Data().(*widgets.GLoader); ok {
					iconURL := fmt.Sprintf("ui://LoopList/n%d", index+1)
					loader.SetURL(iconURL)
				}
			}
		}
	}
}

// Dispose 销毁场景
func (d *LoopListDemo) Dispose() {
	d.view = nil
	d.list = nil
}
