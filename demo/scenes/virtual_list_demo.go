package scenes

import (
	"context"
	"fmt"
	"log"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/pkg/fgui"
	"github.com/chslink/fairygui/pkg/fgui/widgets"
)

// VirtualListDemo 虚拟列表演示场景
// 参考 TypeScript 版本: laya_src/demo/VirtualListDemo.ts
type VirtualListDemo struct {
	view *fgui.GComponent
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
func (d *VirtualListDemo) Load(ctx context.Context, mgr *Manager) (*fgui.GComponent, error) {
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
func (d *VirtualListDemo) renderMailItem(index int, obj *fgui.GObject) {
	if obj == nil {
		log.Printf("❌ obj is nil")
		return
	}

	// 获取子组件
	// 关键修复: 使用 ComponentFrom 而不是 AsComponent
	// AsComponent() 只在 data 是 *GComponent 时返回非 nil,对 widget 类型返回 nil
	// ComponentFrom() 通过 ComponentAccessor 接口正确处理 widget 类型
	comp := fgui.ComponentFrom(obj)
	if comp == nil {
		log.Printf("❌ ComponentFrom() returned nil for index=%d", index)
		return
	}

	// 设置邮件信息（模拟数据）
	// fetched状态（每3个设置一次）
	// 对应 MailItem.ts:28 - setFetched() 方法使用 "c1" controller
	if fetchedCtrl := comp.ControllerByName("c1"); fetchedCtrl != nil {
		if index%3 == 0 {
			fetchedCtrl.SetSelectedIndex(1) // 已获取
		} else {
			fetchedCtrl.SetSelectedIndex(0) // 未获取
		}
	}

	// read状态（每2个设置一次）
	// 对应 MailItem.ts:24 - setRead() 方法使用 "IsRead" controller（注意大小写）
	if readCtrl := comp.ControllerByName("IsRead"); readCtrl != nil {
		if index%2 == 0 {
			readCtrl.SetSelectedIndex(1) // 已读
		} else {
			readCtrl.SetSelectedIndex(0) // 未读
		}
	}

	// 设置标题
	// 关键修复：如果 mailItem 是 GButton，使用 SetTitle() 而不是直接设置文本
	// 这样当点击触发 SetSelected → applyTitleState 时，标题不会被清空
	titleText := fmt.Sprintf("%d Mail title here", index)
	if button, ok := obj.Data().(*widgets.GButton); ok {
		button.SetTitle(titleText)
	} else {
		// 不是 GButton，直接设置 titleObject 的文本
		if titleChild := comp.ChildByName("title"); titleChild != nil {
			if textData := titleChild.Data(); textData != nil {
				if textField, ok := textData.(*widgets.GTextField); ok {
					textField.SetText(titleText)
				}
			}
		}
	}

	// 设置时间
	// 对应 MailItem.ts:20 - setTime() 方法使用 "timeText" 子对象
	if timeText := comp.ChildByName("timeText"); timeText != nil {
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
// 参考 TypeScript: VirtualListDemo.ts
func (d *VirtualListDemo) bindButtons(view *fgui.GComponent) {
	// n6: 添加选择按钮
	// 对应 TypeScript 版本: this._view.getChild("n6").onClick(this, () => { this._list.addSelection(500, true); });
	if btn := view.ChildByName("n6"); btn != nil {
		btn.On(laya.EventClick, func(evt *laya.Event) {
			if d.list != nil {
				// 第二个参数 true 表示自动滚动到该项
				d.list.AddSelection(500, true)
			}
		})
	}

	// n7: 滚动到顶部按钮
	if btn := view.ChildByName("n7"); btn != nil {
		btn.On(laya.EventClick, func(evt *laya.Event) {
			if d.list != nil && d.list.GComponent.ScrollPane() != nil {
				d.list.GComponent.ScrollPane().ScrollTop(false)
			}
		})
	}

	// n8: 滚动到底部按钮
	if btn := view.ChildByName("n8"); btn != nil {
		btn.On(laya.EventClick, func(evt *laya.Event) {
			if d.list != nil && d.list.GComponent.ScrollPane() != nil {
				d.list.GComponent.ScrollPane().ScrollBottom(false)
			}
		})
	}
}

// Dispose 销毁场景
func (d *VirtualListDemo) Dispose() {
	d.view = nil
	d.list = nil
}
