package scenes

import (
	"context"
	"fmt"
	"log"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// VirtualListDemo 虚拟列表演示场景
// 参考 TypeScript 版本: laya_src/demo/VirtualListDemo.ts
type VirtualListDemo struct {
	view *core.GComponent
	list *widgets.GList
}

// NewVirtualListDemo 创建虚拟列表演示场景
func NewVirtualListDemo() Scene {
	return &VirtualListDemo{}
}

func (d *VirtualListDemo) Name() string {
	return "VirtualListDemo"
}

// Load 加载场景
func (d *VirtualListDemo) Load(ctx context.Context, mgr *Manager) (*core.GComponent, error) {
	log.Println("📦 加载虚拟列表 demo...")

	env := mgr.Environment()

	// 加载VirtualList资源包
	pkg, err := env.Package(ctx, "VirtualList")
	if err != nil {
		return nil, err
	}

	// 加载Main组件
	item := chooseComponent(pkg, "Main")
	if item == nil {
		return nil, newMissingComponentError("VirtualList", "Main")
	}

	view, err := env.Factory.BuildComponent(ctx, pkg, item)
	if err != nil {
		return nil, err
	}

	d.view = view

	// 查找mailList组件
	mailListObj := view.ChildByName("mailList")
	if mailListObj == nil {
		return nil, fmt.Errorf("找不到 mailList 组件")
	}

	// 转换为GList
	if data := mailListObj.Data(); data != nil {
		if list, ok := data.(*widgets.GList); ok {
			d.list = list

			// 启用虚拟化
			list.SetVirtual(true)

			// 设置项目渲染器
			list.SetItemRenderer(d.renderMailItem)

			// 设置项目数量（模拟1000封邮件）
			list.SetNumItems(1000)

			log.Printf("✅ 虚拟列表配置完成: NumItems=%d", list.NumItems())
		} else {
			log.Printf("⚠️  mailList 不是 GList 类型: %T", data)
		}
	}

	// 绑定按钮事件
	d.bindButtons(view)

	log.Println("✅ 虚拟列表 demo 加载完成")
	return view, nil
}

// renderMailItem 渲染邮件项目
// 对应 TypeScript: private renderListItem(index: number, obj: fgui.GObject)
func (d *VirtualListDemo) renderMailItem(index int, obj *core.GObject) {
	if obj == nil {
		return
	}

	// 获取子组件
	comp, ok := obj.Data().(*core.GComponent)
	if !ok {
		return
	}

	// 设置邮件信息（模拟数据）
	// fetched状态（每3个设置一次）
	if fetchedCtrl := comp.ControllerByName("fetched"); fetchedCtrl != nil {
		if index%3 == 0 {
			fetchedCtrl.SetSelectedIndex(1) // 已获取
		} else {
			fetchedCtrl.SetSelectedIndex(0) // 未获取
		}
	}

	// read状态（每2个设置一次）
	if readCtrl := comp.ControllerByName("isRead"); readCtrl != nil {
		if index%2 == 0 {
			readCtrl.SetSelectedIndex(1) // 已读
		} else {
			readCtrl.SetSelectedIndex(0) // 未读
		}
	}

	// 设置标题
	if nameText := comp.ChildByName("name"); nameText != nil {
		if textData := nameText.Data(); textData != nil {
			if textField, ok := textData.(*widgets.GTextField); ok {
				textField.SetText(fmt.Sprintf("%d Mail title here", index))
			}
		}
	}

	// 设置时间
	if timeText := comp.ChildByName("time"); timeText != nil {
		if textData := timeText.Data(); textData != nil {
			if textField, ok := textData.(*widgets.GTextField); ok {
				textField.SetText("5 Nov 2015 16:24:33")
			}
		}
	}

	// 设置调试名称
	obj.SetName(fmt.Sprintf("mail_%d", index))
}

// bindButtons 绑定按钮事件
func (d *VirtualListDemo) bindButtons(view *core.GComponent) {
	// 添加选择按钮
	if btn := view.ChildByName("btnAddSelect"); btn != nil {
		btn.On(laya.EventClick, func(evt laya.Event) {
			if d.list != nil {
				d.list.AddSelection(500)
				log.Printf("🎯 添加选择: index=500")
			}
		})
	}

	// 滚动到顶部按钮
	if btn := view.ChildByName("btnScrollToTop"); btn != nil {
		btn.On(laya.EventClick, func(evt laya.Event) {
			if d.list != nil && d.list.GComponent.ScrollPane() != nil {
				d.list.GComponent.ScrollPane().SetPos(0, 0, false)
				log.Printf("⬆️  滚动到顶部")
			}
		})
	}

	// 滚动到底部按钮
	if btn := view.ChildByName("btnScrollToBottom"); btn != nil {
		btn.On(laya.EventClick, func(evt laya.Event) {
			if d.list != nil && d.list.GComponent.ScrollPane() != nil {
				scrollPane := d.list.GComponent.ScrollPane()
				// 滚动到最大Y位置
				maxY := d.list.GComponent.Height() - scrollPane.ViewHeight()
				if maxY < 0 {
					maxY = 0
				}
				scrollPane.SetPos(0, maxY, false)
				log.Printf("⬇️  滚动到底部")
			}
		})
	}

	// 刷新列表按钮
	if btn := view.ChildByName("btnRefresh"); btn != nil {
		btn.On(laya.EventClick, func(evt laya.Event) {
			if d.list != nil {
				d.list.RefreshVirtualList()
				log.Printf("🔄 刷新虚拟列表")
			}
		})
	}
}

// Dispose 销毁场景
func (d *VirtualListDemo) Dispose() {
	log.Println("🗑️  虚拟列表 demo 已销毁")
	d.view = nil
	d.list = nil
}
