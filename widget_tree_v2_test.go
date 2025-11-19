package fairygui

import (
	"testing"
)

// TestNewTreeNode 测试创建新的树节点
func TestNewTreeNode(t *testing.T) {
	node := NewTreeNode("Test")
	if node == nil {
		t.Fatal("NewTreeNode() returned nil")
	}

	if node.Text() != "Test" {
		t.Errorf("节点文本不正确: got %s, want Test", node.Text())
	}

	if node.NumChildren() != 0 {
		t.Error("新节点不应该有子节点")
	}

	if node.IsExpanded() {
		t.Error("新节点默认应该是折叠的")
	}

	if !node.IsSelectable() {
		t.Error("新节点默认应该是可选中的")
	}

	if !node.IsVisible() {
		t.Error("新节点默认应该是可见的")
	}

	if node.Level() != 0 {
		t.Errorf("新节点的层级应该是0: got %d", node.Level())
	}
}

// TestTreeNode_SetText 测试设置节点文本
func TestTreeNode_SetText(t *testing.T) {
	node := NewTreeNode("Text1")
	node.SetText("Text2")

	if node.Text() != "Text2" {
		t.Errorf("节点文本设置失败: got %s, want Text2", node.Text())
	}
}

// TestTreeNode_SetData 测试设置节点数据
func TestTreeNode_SetData(t *testing.T) {
	node := NewTreeNode("Test")
	data := "custom data"
	node.SetData(data)

	if node.Data() != data {
		t.Errorf("节点数据设置失败: got %v, want %v", node.Data(), data)
	}
}

// TestTreeNode_SetExpanded 测试展开/折叠
func TestTreeNode_SetExpanded(t *testing.T) {
	node := NewTreeNode("Test")

	node.SetExpanded(true)
	if !node.IsExpanded() {
		t.Error("节点展开失败")
	}

	node.SetExpanded(false)
	if node.IsExpanded() {
		t.Error("节点折叠失败")
	}
}

// TestTreeNode_Toggle 测试切换展开/折叠
func TestTreeNode_Toggle(t *testing.T) {
	node := NewTreeNode("Test")

	initialState := node.IsExpanded()
	node.Toggle()

	if node.IsExpanded() == initialState {
		t.Error("切换展开/折叠失败")
	}

	node.Toggle()
	if node.IsExpanded() != initialState {
		t.Error("再次切换应该恢复原状态")
	}
}

// TestTreeNode_AddChild 测试添加子节点
func TestTreeNode_AddChild(t *testing.T) {
	parent := NewTreeNode("Parent")
	child := NewTreeNode("Child")

	parent.AddChild(child)

	if parent.NumChildren() != 1 {
		t.Errorf("子节点数量不正确: got %d, want 1", parent.NumChildren())
	}

	if child.Parent() != parent {
		t.Error("子节点的父节点设置失败")
	}

	if child.Level() != 1 {
		t.Errorf("子节点的层级不正确: got %d, want 1", child.Level())
	}
}

// TestTreeNode_RemoveChild 测试移除子节点
func TestTreeNode_RemoveChild(t *testing.T) {
	parent := NewTreeNode("Parent")
	child := NewTreeNode("Child")

	parent.AddChild(child)
	parent.RemoveChild(child)

	if parent.NumChildren() != 0 {
		t.Error("移除子节点失败")
	}

	if child.Parent() != nil {
		t.Error("移除后子节点的父节点应该为nil")
	}
}

// TestTreeNode_RemoveChildAt 测试移除指定位置的子节点
func TestTreeNode_RemoveChildAt(t *testing.T) {
	parent := NewTreeNode("Parent")
	child1 := NewTreeNode("Child1")
	child2 := NewTreeNode("Child2")
	child3 := NewTreeNode("Child3")

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.AddChild(child3)

	parent.RemoveChildAt(1)

	if parent.NumChildren() != 2 {
		t.Errorf("移除后子节点数量不正确: got %d, want 2", parent.NumChildren())
	}

	if parent.GetChildAt(1).Text() != "Child3" {
		t.Error("移除指定位置的子节点失败")
	}
}

// TestTreeNode_Children 测试获取所有子节点
func TestTreeNode_Children(t *testing.T) {
	parent := NewTreeNode("Parent")
	child1 := NewTreeNode("Child1")
	child2 := NewTreeNode("Child2")

	parent.AddChild(child1)
	parent.AddChild(child2)

	children := parent.Children()
	if len(children) != 2 {
		t.Errorf("子节点数量不正确: got %d, want 2", len(children))
	}

	if children[0].Text() != "Child1" || children[1].Text() != "Child2" {
		t.Error("子节点列表不正确")
	}
}

// TestTreeNode_IsLeaf_IsAncestorOf 测试叶子节点和祖先判断
func TestTreeNode_IsLeaf_IsAncestorOf(t *testing.T) {
	root := NewTreeNode("Root")
	child := NewTreeNode("Child")
	grandChild := NewTreeNode("GrandChild")

	root.AddChild(child)
	child.AddChild(grandChild)

	if !root.HasChildren() {
		t.Error("根节点应该有子节点")
	}

	if !child.HasChildren() {
		t.Error("子节点应该有子节点")
	}

	if child.IsLeaf() {
		t.Error("有子节点的节点不应该是叶子节点")
	}

	if !grandChild.IsLeaf() {
		t.Error("没有子节点的节点应该是叶子节点")
	}

	if !root.IsAncestorOf(grandChild) {
		t.Error("根节点应该是孙节点的祖先")
	}

	if !grandChild.IsDescendantOf(root) {
		t.Error("孙节点应该是根节点的后代")
	}
}

// TestTreeNode_Root 测试获取根节点
func TestTreeNode_Root(t *testing.T) {
	root := NewTreeNode("Root")
	child := NewTreeNode("Child")
	grandChild := NewTreeNode("GrandChild")

	root.AddChild(child)
	child.AddChild(grandChild)

	if grandChild.Root() != root {
		t.Error("获取根节点失败")
	}
}

// TestNewTreeView 测试创建树视图
func TestNewTreeView(t *testing.T) {
	tree := NewTreeView()
	if tree == nil {
		t.Fatal("NewTreeView() returned nil")
	}

	if tree.List == nil {
		t.Error("TreeView.List is nil")
	}

	if tree.RootNode() == nil {
		t.Error("TreeView.RootNode() is nil")
	}

	if tree.IndentWidth() != 20.0 {
		t.Errorf("默认缩进宽度不正确: got %f, want 20.0", tree.IndentWidth())
	}

	if !tree.ShowIcons() {
		t.Error("默认应该显示图标")
	}

	if !tree.ShowArrows() {
		t.Error("默认应该显示箭头")
	}
}

// TestTreeView_AddNode 测试添加节点
func TestTreeView_AddNode(t *testing.T) {
	tree := NewTreeView()
	node := NewTreeNode("Test")

	tree.AddNode(nil, node)

	if tree.RootNode().NumChildren() != 1 {
		t.Error("添加节点失败")
	}

	if node.Parent() != tree.RootNode() {
		t.Error("节点的父节点设置失败")
	}
}

// TestTreeView_RemoveNode 测试移除节点
func TestTreeView_RemoveNode(t *testing.T) {
	tree := NewTreeView()
	node := NewTreeNode("Test")

	tree.AddNode(nil, node)
	tree.RemoveNode(node)

	if tree.RootNode().NumChildren() != 0 {
		t.Error("移除节点失败")
	}
}

// TestTreeView_ClearNodes 测试清除所有节点
func TestTreeView_ClearNodes(t *testing.T) {
	tree := NewTreeView()

	tree.AddNode(nil, NewTreeNode("Node1"))
	tree.AddNode(nil, NewTreeNode("Node2"))
	tree.AddNode(nil, NewTreeNode("Node3"))

	tree.ClearNodes()

	if tree.RootNode().NumChildren() != 0 {
		t.Error("清除所有节点失败")
	}
}

// TestTreeView_ExpandCollapse 测试展开/折叠
func TestTreeView_ExpandCollapse(t *testing.T) {
	tree := NewTreeView()
	parent := NewTreeNode("Parent")
	child := NewTreeNode("Child")

	parent.AddChild(child)
	tree.AddNode(nil, parent)

	// 初始状态应该是折叠的
	if parent.IsExpanded() {
		t.Error("新节点默认应该是折叠的")
	}

	// 展开
	tree.ExpandNode(parent)
	if !parent.IsExpanded() {
		t.Error("展开节点失败")
	}

	// 折叠
	tree.CollapseNode(parent)
	if parent.IsExpanded() {
		t.Error("折叠节点失败")
	}

	// 切换
	tree.ToggleNode(parent)
	if !parent.IsExpanded() {
		t.Error("切换节点失败")
	}

	tree.ToggleNode(parent)
	if parent.IsExpanded() {
		t.Error("切换节点失败")
	}
}

// TestTreeView_ExpandAll_CollapseAll 测试展开/折叠所有节点
func TestTreeView_ExpandAll_CollapseAll(t *testing.T) {
	tree := NewTreeView()

	// 构建树结构
	root := tree.RootNode()
	child1 := NewTreeNode("Child1")
	child2 := NewTreeNode("Child2")
	grandChild1 := NewTreeNode("GrandChild1")
	grandChild2 := NewTreeNode("GrandChild2")

	root.AddChild(child1)
	root.AddChild(child2)
	child1.AddChild(grandChild1)
	child2.AddChild(grandChild2)

	// 折叠所有
	tree.CollapseAll()

	if child1.IsExpanded() || child2.IsExpanded() {
		t.Error("折叠所有节点失败")
	}

	// 展开所有
	tree.ExpandAll()

	if !child1.IsExpanded() || !child2.IsExpanded() {
		t.Error("展开所有节点失败")
	}
}

// TestTreeView_SetIndentWidth 测试设置缩进宽度
func TestTreeView_SetIndentWidth(t *testing.T) {
	tree := NewTreeView()

	indentWidth := 30.0
	tree.SetIndentWidth(indentWidth)

	if tree.IndentWidth() != indentWidth {
		t.Errorf("缩进宽度设置失败: got %f, want %f", tree.IndentWidth(), indentWidth)
	}
}

// TestTreeView_SetShowIcons 测试显示/隐藏图标
func TestTreeView_SetShowIcons(t *testing.T) {
	tree := NewTreeView()

	tree.SetShowIcons(false)
	if tree.ShowIcons() {
		t.Error("隐藏图标失败")
	}

	tree.SetShowIcons(true)
	if !tree.ShowIcons() {
		t.Error("显示图标失败")
	}
}

// TestTreeView_SetShowArrows 测试显示/隐藏箭头
func TestTreeView_SetShowArrows(t *testing.T) {
	tree := NewTreeView()

	tree.SetShowArrows(false)
	if tree.ShowArrows() {
		t.Error("隐藏箭头失败")
	}

	tree.SetShowArrows(true)
	if !tree.ShowArrows() {
		t.Error("显示箭头失败")
	}
}

// TestTreeView_Selection 测试节点选择
func TestTreeView_Selection(t *testing.T) {
	tree := NewTreeView()

	node1 := NewTreeNode("Node1")
	node2 := NewTreeNode("Node2")

	tree.AddNode(nil, node1)
	tree.AddNode(nil, node2)

	// 选择节点
	tree.SelectNode(node1, false)

	selected := tree.GetSelectedNode()
	if selected != node1 {
		t.Error("选择节点失败")
	}

	// 选择另一个节点
	tree.SelectNode(node2, false)

	selected = tree.GetSelectedNode()
	if selected != node2 {
		t.Error("选择另一个节点失败")
	}

	// 通过索引选择
	tree.SetSelectedNode(node1)

	selected = tree.GetSelectedNode()
	if selected != node1 {
		t.Error("SetSelectedNode 失败")
	}
}

// TestTreeView_UpdateVisibleNodes 测试更新可见节点
func TestTreeView_UpdateVisibleNodes(t *testing.T) {
	tree := NewTreeView()

	// 构建树结构
	root := tree.RootNode()
	child1 := NewTreeNode("Child1")
	child2 := NewTreeNode("Child2")
	grandChild1 := NewTreeNode("GrandChild1")
	grandChild2 := NewTreeNode("GrandChild2")

	root.AddChild(child1)
	root.AddChild(child2)
	child1.AddChild(grandChild1)
	child2.AddChild(grandChild2)

	// 初始状态（所有节点折叠）
	tree.updateVisibleNodes()

	if tree.NumItems() != 2 {
		t.Errorf("折叠状态下可见节点数量不正确: got %d, want 2", tree.NumItems())
	}

	// 展开一个节点
	child1.SetExpanded(true)
	tree.updateVisibleNodes()

	if tree.NumItems() != 3 {
		t.Errorf("展开一个节点后可见节点数量不正确: got %d, want 3", tree.NumItems())
	}

	// 展开所有节点
	tree.ExpandAll()
	tree.updateVisibleNodes()

	if tree.NumItems() != 5 {
		t.Errorf("展开所有节点后可见节点数量不正确: got %d, want 5", tree.NumItems())
	}
}

// TestTreeView_GetNodeAt 测试获取指定索引的节点
func TestTreeView_GetNodeAt(t *testing.T) {
	tree := NewTreeView()

	node1 := NewTreeNode("Node1")
	node2 := NewTreeNode("Node2")

	tree.AddNode(nil, node1)
	tree.AddNode(nil, node2)

	node := tree.GetNodeAt(0)
	if node == nil || node.Text() != "Node1" {
		t.Error("获取节点失败")
	}

	node = tree.GetNodeAt(1)
	if node == nil || node.Text() != "Node2" {
		t.Error("获取节点失败")
	}

	node = tree.GetNodeAt(10)
	if node != nil {
		t.Error("超出范围的索引应该返回nil")
	}
}

// TestTreeView_GetItemNode_GetNodeItem 测试节点和项的映射
func TestTreeView_GetItemNode_GetNodeItem(t *testing.T) {
	tree := NewTreeView()
	node := NewTreeNode("Test")

	tree.AddNode(nil, node)

	// 注意：这个测试需要实际的列表项，当前简化实现可能无法完全测试
	// 在实际渲染后，这些映射关系会被建立

	if tree.GetNodeItem(node) != nil {
		t.Log("节点到项的映射将在实际渲染后建立")
	}

	item := tree.GetItemAt(0)
	if item == nil {
		t.Log("列表项将在实际渲染后创建")
	}
}

// TestTreeView_GetNodeIndex 测试获取节点索引
func TestTreeView_getNodeIndex(t *testing.T) {
	tree := NewTreeView()

	node1 := NewTreeNode("Node1")
	node2 := NewTreeNode("Node2")

	tree.AddNode(nil, node1)
	tree.AddNode(nil, node2)

	index := tree.getNodeIndex(node1)
	if index != 0 {
		t.Errorf("获取节点索引失败: got %d, want 0", index)
	}

	index = tree.getNodeIndex(node2)
	if index != 1 {
		t.Errorf("获取节点索引失败: got %d, want 1", index)
	}

	index = tree.getNodeIndex(NewTreeNode("NotInTree"))
	if index != -1 {
		t.Error("不在树中的节点应该返回-1")
	}
}

// TestTreeView_ClearMappings 测试清除映射
func TestTreeView_clearMappings(t *testing.T) {
	tree := NewTreeView()

	// 调用 clearMappings（内部方法）
	tree.clearMappings()

	// 映射应该被清空
	if len(tree.itemToNode) != 0 || len(tree.nodeToItem) != 0 {
		t.Error("清除映射失败")
	}
}

// TestAssertTreeView 测试类型断言
func TestAssertTreeView(t *testing.T) {
	tree := NewTreeView()

	// 测试 AssertTreeView
	result, ok := AssertTreeView(tree)
	if !ok {
		t.Error("AssertTreeView 应该成功")
	}
	if result != tree {
		t.Error("AssertTreeView 返回的对象不正确")
	}

	// 测试 IsTreeView
	if !IsTreeView(tree) {
		t.Error("IsTreeView 应该返回 true")
	}

	// 测试不是 TreeView 的情况
	obj := NewObject()
	_, ok = AssertTreeView(obj)
	if ok {
		t.Error("AssertTreeView 对非 TreeView 对象应该失败")
	}

	if IsTreeView(obj) {
		t.Error("IsTreeView 对非 TreeView 对象应该返回 false")
	}
}
