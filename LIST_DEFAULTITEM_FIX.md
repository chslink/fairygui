# List defaultItem 渲染问题修复报告

## 问题描述

在FairyGUI Go移植版本中，List组件的defaultItem（道具格子）存在渲染问题：
- 道具格子是继承GButton的模板组件（BagGridSub）
- 模板中定义了title（GTextField）和icon（GLoader）子组件
- List的item通过设置Button的title和icon属性来覆盖模板中的属性
- **实际表现**：只显示道具格子的背景，道具的title和icon没有显示

## 问题分析

经过详细调查发现，问题的根本原因在于**GButton的titleObject和iconObject字段未正确设置**：

1. GButton的`SetTitle()`和`SetIcon()`方法会调用`applyTitleState()`和`applyIconState()`
2. `applyTitleState()`和`applyIconState()`依赖于`titleObject`和`iconObject`字段
3. 如果`titleObject`或`iconObject`为nil，这些方法会直接返回，不会设置文本和图标
4. **关键问题**：在对象池复用Button时，`titleObject`和`iconObject`可能为nil

## 解决方案

修改`GButton`的`applyTitleState()`和`applyIconState()`方法，添加自动查找逻辑：

```go
func (b *GButton) applyTitleState() {
    text := b.title
    if b.selected && b.selectedTitle != "" {
        text = b.selectedTitle
    }
    if b.titleObject == nil {
        // 修复：在没有titleObject时自动查找
        // 参考 TypeScript 版本中 titleObject 是通过 getChild("title") 查找的
        if titleChild := b.GComponent.ChildByName("title"); titleChild != nil {
            b.SetTitleObject(titleChild)
        } else if b.template != nil {
            // 如果Button有template，从template中查找
            if titleChild := b.template.ChildByName("title"); titleChild != nil {
                b.SetTitleObject(titleChild)
            }
        }
        // 如果仍然没有找到titleObject，直接返回
        if b.titleObject == nil {
            return
        }
    }
    // ... 后续设置文本逻辑
}
```

同样修改`applyIconState()`方法。

## 修复文件

- `pkg/fgui/widgets/button.go`：
  - 修改`applyTitleState()`方法，添加自动查找titleObject逻辑
  - 修改`applyIconState()`方法，添加自动查找iconObject逻辑

- `pkg/fgui/builder/component.go`：
  - 修改`applyButtonTemplate()`方法，确保template和Button子组件都被检查
  - 添加fallback逻辑，当template中没有找到title/icon时，从Button直接子组件中查找

## 测试验证

创建了两个单元测试来验证修复：

1. `TestListDefaultItemWithItemAttributes`：验证List的defaultItem能正确解析title和icon属性
2. `TestButtonTitleAndIconObject`：验证Button的titleObject和iconObject被正确设置

测试结果显示：
- ✅ Button的titleObject和iconObject现在被正确设置
- ✅ List的item能正确显示title和icon
- ✅ 属性值被正确传递到模板组件的子对象

## 兼容性说明

此修复与TypeScript版本保持一致：
- TypeScript版本中，GButton在`constructExtension()`中设置`titleObject`和`iconObject`
- Go版本中，我们在`applyTitleState()`和`applyIconState()`中按需查找
- 这种方式更灵活，能够处理对象池复用场景

## 总结

这个修复解决了List组件中道具格子只显示背景、不显示道具信息的bug。现在道具的标题和图标能够正确渲染和显示。
