package fairygui

import (
	"testing"
)

// TestNewComboBox 测试创建新的组合框
func TestNewComboBox(t *testing.T) {
	cb := NewComboBox()
	if cb == nil {
		t.Fatal("NewComboBox() returned nil")
	}

	// 检查默认属性
	if cb.ComponentImpl == nil {
		t.Error("ComboBox.ComponentImpl is nil")
	}

	if cb.SelectedIndex() != -1 {
		t.Errorf("默认选中索引不正确: got %d, want -1", cb.SelectedIndex())
	}

	if cb.VisibleItemCount() != 10 {
		t.Errorf("默认可见项数量不正确: got %d, want 10", cb.VisibleItemCount())
	}

	if cb.IsDisabled() {
		t.Error("新创建的 ComboBox 不应该禁用")
	}

	if cb.NumItems() != 0 {
		t.Error("新创建的 ComboBox 应该没有项")
	}
}

// TestComboBox_AddItem 测试添加项
func TestComboBox_AddItem(t *testing.T) {
	cb := NewComboBox()

	cb.AddItem("Item1")
	cb.AddItem("Item2")
	cb.AddItem("Item3")

	if cb.NumItems() != 3 {
		t.Errorf("项数量不正确: got %d, want 3", cb.NumItems())
	}
}

// TestComboBox_AddItems 测试批量添加项
func TestComboBox_AddItems(t *testing.T) {
	cb := NewComboBox()

	items := []string{"Item1", "Item2", "Item3", "Item4"}
	cb.AddItems(items)

	if cb.NumItems() != 4 {
		t.Errorf("项数量不正确: got %d, want 4", cb.NumItems())
	}

	// 验证内容
	savedItems := cb.Items()
	for i, item := range items {
		if savedItems[i] != item {
			t.Errorf("项内容不匹配: got %s, want %s", savedItems[i], item)
		}
	}
}

// TestComboBox_SetItems 测试设置项
func TestComboBox_SetItems(t *testing.T) {
	cb := NewComboBox()

	// 先添加一些项
	cb.AddItems([]string{"Old1", "Old2"})

	// 重新设置
	newItems := []string{"New1", "New2", "New3"}
	cb.SetItems(newItems)

	if cb.NumItems() != 3 {
		t.Errorf("项数量不正确: got %d, want 3", cb.NumItems())
	}

	// 验证内容
	savedItems := cb.Items()
	for i, item := range newItems {
		if savedItems[i] != item {
			t.Errorf("项内容不匹配: got %s, want %s", savedItems[i], item)
		}
	}
}

// TestComboBox_Clear 测试清空
func TestComboBox_Clear(t *testing.T) {
	cb := NewComboBox()

	// 添加项
	cb.AddItems([]string{"Item1", "Item2", "Item3"})
	cb.SetSelectedIndex(1)

	// 清空
	cb.Clear()

	if cb.NumItems() != 0 {
		t.Errorf("清空后项数量应该为0: got %d", cb.NumItems())
	}

	if cb.SelectedIndex() != -1 {
		t.Errorf("清空后选中索引应该为-1: got %d", cb.SelectedIndex())
	}
}

// TestComboBox_SetSelectedIndex 测试设置选中索引
func TestComboBox_SetSelectedIndex(t *testing.T) {
	cb := NewComboBox()
	cb.AddItems([]string{"Item1", "Item2", "Item3", "Item4", "Item5"})

	// 设置有效索引
	cb.SetSelectedIndex(2)
	if cb.SelectedIndex() != 2 {
		t.Errorf("选中索引设置失败: got %d, want 2", cb.SelectedIndex())
	}

	if cb.GetSelection() != "Item3" {
		t.Errorf("选中项不正确: got %s, want Item3", cb.GetSelection())
	}

	// 设置为 -1（不选中）
	cb.SetSelectedIndex(-1)
	if cb.SelectedIndex() != -1 {
		t.Errorf("设置-1失败: got %d, want -1", cb.SelectedIndex())
	}

	// 设置为超出范围的索引
	cb.SetSelectedIndex(10)
	if cb.SelectedIndex() != -1 {
		t.Errorf("超出范围的索引应该设为-1: got %d", cb.SelectedIndex())
	}
}

// TestComboBox_SetSelectedItem 测试根据项文本选中
func TestComboBox_SetSelectedItem(t *testing.T) {
	cb := NewComboBox()
	cb.AddItems([]string{"Apple", "Banana", "Cherry", "Date"})

	// 选择现有项
	success := cb.SetSelectedItem("Cherry")
	if !success {
		t.Error("选择现有项应该成功")
	}

	if cb.SelectedIndex() != 2 {
		t.Errorf("选中索引不正确: got %d, want 2", cb.SelectedIndex())
	}

	// 选择不存在的项
	success = cb.SetSelectedItem("Grape")
	if success {
		t.Error("选择不存在的项应该失败")
	}

	if cb.SelectedIndex() != 2 {
		t.Error("选择失败后索引应该保持不变")
	}
}

// TestComboBox_SetSelectedValue 测试根据值选中
func TestComboBox_SetSelectedValue(t *testing.T) {
	cb := NewComboBox()
	cb.AddItems([]string{"Item1", "Item2", "Item3"})
	cb.SetValues([]string{"val1", "val2", "val3"})

	// 选择现有值
	success := cb.SetSelectedValue("val2")
	if !success {
		t.Error("选择现有值应该成功")
	}

	if cb.SelectedIndex() != 1 {
		t.Errorf("选中索引不正确: got %d, want 1", cb.SelectedIndex())
	}

	// 选择不存在的值
	success = cb.SetSelectedValue("val99")
	if success {
		t.Error("选择不存在的值应该失败")
	}
}

// TestComboBox_GetValue 测试返回值
func TestComboBox_GetValue(t *testing.T) {
	cb := NewComboBox()
	cb.AddItems([]string{"Item1", "Item2", "Item3"})

	// 没有设置 values，应该返回 item
	cb.SetSelectedIndex(1)
	if cb.GetValue() != "Item2" {
		t.Errorf("GetValue 返回值不正确: got %s, want Item2", cb.GetValue())
	}

	// 设置 values
	cb.SetValues([]string{"val1", "val2", "val3"})
	if cb.GetValue() != "val2" {
		t.Errorf("GetValue 返回值不正确: got %s, want val2", cb.GetValue())
	}
}

// TestComboBox_SetValues 测试设置值
func TestComboBox_SetValues(t *testing.T) {
	cb := NewComboBox()
	cb.AddItems([]string{"Item1", "Item2", "Item3"})

	values := []string{"Value1", "Value2", "Value3"}
	cb.SetValues(values)

	savedValues := cb.Values()
	for i, val := range values {
		if savedValues[i] != val {
			t.Errorf("值设置失败: got %s, want %s", savedValues[i], val)
		}
	}
}

// TestComboBox_SetIcons 测试设置图标
func TestComboBox_SetIcons(t *testing.T) {
	cb := NewComboBox()

	icons := []string{"icon1.png", "icon2.png", "icon3.png"}
	cb.SetIcons(icons)

	savedIcons := cb.Icons()
	for i, icon := range icons {
		if savedIcons[i] != icon {
			t.Errorf("图标设置失败: got %s, want %s", savedIcons[i], icon)
		}
	}
}

// TestComboBox_GetItemAt 测试获取指定项
func TestComboBox_GetItemAt(t *testing.T) {
	cb := NewComboBox()
	cb.AddItems([]string{"Item1", "Item2", "Item3"})

	if cb.GetItemAt(0) != "Item1" {
		t.Error("获取项失败")
	}

	if cb.GetItemAt(2) != "Item3" {
		t.Error("获取项失败")
	}

	if cb.GetItemAt(10) != "" {
		t.Error("超出范围的索引应该返回空字符串")
	}
}

// TestComboBox_GetValueAt 测试获取指定值
func TestComboBox_GetValueAt(t *testing.T) {
	cb := NewComboBox()
	cb.AddItems([]string{"Item1", "Item2", "Item3"})
	cb.SetValues([]string{"val1", "val2", "val3"})

	if cb.GetValueAt(1) != "val2" {
		t.Error("获取值失败")
	}

	if cb.GetValueAt(10) != "" {
		t.Error("超出范围的索引应该返回空字符串")
	}
}

// TestComboBox_GetIconAt 测试获取指定图标
func TestComboBox_GetIconAt(t *testing.T) {
	cb := NewComboBox()
	cb.SetIcons([]string{"icon1.png", "icon2.png", "icon3.png"})

	if cb.GetIconAt(1) != "icon2.png" {
		t.Error("获取图标失败")
	}

	if cb.GetIconAt(10) != "" {
		t.Error("超出范围的索引应该返回空字符串")
	}
}

// TestComboBox_SetDisabled 测试禁用状态
func TestComboBox_SetDisabled(t *testing.T) {
	cb := NewComboBox()

	cb.SetDisabled(true)
	if !cb.IsDisabled() {
		t.Error("设置禁用失败")
	}

	cb.SetDisabled(false)
	if cb.IsDisabled() {
		t.Error("取消禁用失败")
	}
}

// TestComboBox_SetVisibleItemCount 测试设置可见项数量
func TestComboBox_SetVisibleItemCount(t *testing.T) {
	cb := NewComboBox()

	cb.SetVisibleItemCount(5)
	if cb.VisibleItemCount() != 5 {
		t.Errorf("设置可见项数量失败: got %d, want 5", cb.VisibleItemCount())
	}
}

// TestComboBox_IsDropdownVisible 测试下拉可见性
func TestComboBox_IsDropdownVisible(t *testing.T) {
	cb := NewComboBox()

	// 初始状态
	if cb.IsDropdownVisible() {
		t.Error("初始状态下拉应该不可见")
	}

	// 显示下拉（简化测试）
	// 注意：实际显示需要 UI 环境
	cb.dropDownVisible = true
	if !cb.IsDropdownVisible() {
		t.Error("设置后下拉应该可见")
	}
}

// TestComboBox_ParseList 测试解析列表字符串
func TestComboBox_ParseList(t *testing.T) {
	cb := NewComboBox()

	listStr := "Apple,Banana,Cherry,Date"
	cb.ParseList(listStr)

	if cb.NumItems() != 4 {
		t.Errorf("解析后项数量不正确: got %d, want 4", cb.NumItems())
	}

	expectedItems := []string{"Apple", "Banana", "Cherry", "Date"}
	savedItems := cb.Items()
	for i, item := range expectedItems {
		if savedItems[i] != item {
			t.Errorf("解析项内容不正确: got %s, want %s", savedItems[i], item)
		}
	}
}

// TestComboBox_Events 测试事件
func TestComboBox_Events(t *testing.T) {
	cb := NewComboBox()
	cb.AddItems([]string{"Item1", "Item2", "Item3"})

	// 下拉事件
	dropdownCalled := false
	cb.OnDropDown(func(event Event) {
		dropdownCalled = true
	})

	cb.emitDropDown()
	if !dropdownCalled {
		t.Error("下拉事件未触发")
	}

	// 选择事件
	selectionCalled := false
	var eventData interface{}
	cb.OnSelection(func(event Event) {
		selectionCalled = true
		// 检查事件数据
		if uiEvent, ok := event.(*UIEvent); ok {
			eventData = uiEvent.Data
		}
	})

	cb.SetSelectedIndex(1)
	cb.emitSelection()

	if !selectionCalled {
		t.Error("选择事件未触发")
	}

	// 验证数据是否正确
	// event.Data 应该是 cb.selectedIndex (1)
	if eventData == nil {
		t.Error("选择事件数据为 nil")
	} else {
		if idx, ok := eventData.(int); !ok || idx != 1 {
			t.Errorf("选择事件数据不正确: got %v (%T), want 1 (int)", eventData, eventData)
		}
	}
}

// TestComboBox_Chaining 测试链式调用
func TestComboBox_Chaining(t *testing.T) {
	cb := NewComboBox()

	// 设置项
	cb.AddItem("Item1")
	cb.AddItem("Item2")
	cb.SetSelectedIndex(0)
	cb.SetVisibleItemCount(5)
	cb.SetDisabled(true)

	if cb.NumItems() != 2 {
		t.Error("链式调用失败")
	}

	if cb.SelectedIndex() != 0 {
		t.Error("链式调用设置选中索引失败")
	}

	if cb.VisibleItemCount() != 5 {
		t.Error("链式调用设置可见项数量失败")
	}

	if !cb.IsDisabled() {
		t.Error("链式调用设置禁用状态失败")
	}
}

// TestAssertComboBox 测试类型断言
func TestAssertComboBox(t *testing.T) {
	cb := NewComboBox()

	// 测试 AssertComboBox
	result, ok := AssertComboBox(cb)
	if !ok {
		t.Error("AssertComboBox 应该成功")
	}
	if result != cb {
		t.Error("AssertComboBox 返回的对象不正确")
	}

	// 测试 IsComboBox
	if !IsComboBox(cb) {
		t.Error("IsComboBox 应该返回 true")
	}

	// 测试不是 ComboBox 的情况
	obj := NewObject()
	_, ok = AssertComboBox(obj)
	if ok {
		t.Error("AssertComboBox 对非 ComboBox 对象应该失败")
	}

	if IsComboBox(obj) {
		t.Error("IsComboBox 对非 ComboBox 对象应该返回 false")
	}
}
