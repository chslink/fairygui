package widgets

import (
	"testing"
	"time"

	"github.com/chslink/fairygui/internal/compat/laya"
	"github.com/chslink/fairygui/internal/compat/laya/testutil"
	"github.com/chslink/fairygui/pkg/fgui/core"
)

func TestListAccessors(t *testing.T) {
	list := NewList()
	list.SetDefaultItem("ui://pkg/item")
	list.SetResource("resourceID")
	if list.DefaultItem() != "ui://pkg/item" {
		t.Fatalf("unexpected default item: %s", list.DefaultItem())
	}
	if list.Resource() != "resourceID" {
		t.Fatalf("unexpected resource: %s", list.Resource())
	}
}

func TestListSelectionOnClick(t *testing.T) {
	list := NewList()
	list.GComponent.GObject.SetSize(200, 60)

	itemA := core.NewGObject()
	itemA.SetSize(60, 40)
	itemA.SetPosition(10, 10)
	list.AddItem(itemA)

	itemB := core.NewGObject()
	itemB.SetSize(60, 40)
	itemB.SetPosition(90, 10)
	list.AddItem(itemB)

	env := testutil.NewStageEnv(t, 240, 120)
	env.Stage.AddChild(list.GComponent.GObject.DisplayObject())

	env.Advance(16*time.Millisecond, laya.MouseState{X: 100, Y: 30, Primary: true})
	env.Advance(16*time.Millisecond, laya.MouseState{X: 100, Y: 30, Primary: false})

	if idx := list.SelectedIndex(); idx != 1 {
		t.Fatalf("expected selected index 1, got %d", idx)
	}
	if list.SelectedItem() != itemB {
		t.Fatalf("expected selected item to be itemB")
	}
	indices := list.SelectedIndices()
	if len(indices) != 1 || indices[0] != 1 {
		t.Fatalf("expected indices [1], got %v", indices)
	}
}

func TestListSelectionControllerSync(t *testing.T) {
	list := NewList()
	itemA := core.NewGObject()
	itemB := core.NewGObject()
	list.AddItem(itemA)
	list.AddItem(itemB)

	ctrl := core.NewController("selection")
	ctrl.SetPages([]string{"id-0", "id-1"}, []string{"page-0", "page-1"})

	list.SetSelectionController(ctrl)

	if ctrl.SelectedIndex() != 0 {
		t.Fatalf("expected controller to default to first page, got %d", ctrl.SelectedIndex())
	}
	if list.SelectedIndex() != 0 {
		t.Fatalf("expected list to sync with controller default selection, got %d", list.SelectedIndex())
	}

	list.SetSelectedIndex(1)
	if ctrl.SelectedIndex() != 1 {
		t.Fatalf("expected controller to follow list selection, got %d", ctrl.SelectedIndex())
	}
	if idx := list.SelectedIndex(); idx != 1 {
		t.Fatalf("expected list selected index 1, got %d", idx)
	}
	if indices := list.SelectedIndices(); len(indices) != 1 || indices[0] != 1 {
		t.Fatalf("expected list indices [1], got %v", indices)
	}

	ctrl.SetSelectedIndex(0)
	if list.SelectedIndex() != 0 {
		t.Fatalf("expected list to follow controller selection, got %d", list.SelectedIndex())
	}

	ctrl.SetSelectedIndex(-1)
	if list.SelectedIndex() != 0 {
		t.Fatalf("expected list to follow controller clamped selection, got %d", list.SelectedIndex())
	}
	// controller clamps negative when pages exist, so we expect it to stay on 0.
	if ctrl.SelectedIndex() != 0 {
		t.Fatalf("expected controller to clamp negative selection to 0, got %d", ctrl.SelectedIndex())
	}
}

func TestListSelectionControllerDetach(t *testing.T) {
	list := NewList()
	itemA := core.NewGObject()
	itemB := core.NewGObject()
	list.AddItem(itemA)
	list.AddItem(itemB)

	ctrl := core.NewController("selection")
	ctrl.SetPages([]string{"id-0", "id-1"}, []string{"page-0", "page-1"})

	list.SetSelectionController(ctrl)
	if list.SelectedIndex() != 0 {
		t.Fatalf("expected list to adopt controller default selection, got %d", list.SelectedIndex())
	}
	list.SetSelectedIndex(1)
	if ctrl.SelectedIndex() != 1 {
		t.Fatalf("expected controller to sync before detach")
	}

	list.SetSelectionController(nil)

	ctrl.SetSelectedIndex(0)
	if list.SelectedIndex() != 1 {
		t.Fatalf("expected list selection to remain unchanged after detach, got %d", list.SelectedIndex())
	}
	if indices := list.SelectedIndices(); len(indices) != 1 || indices[0] != 1 {
		t.Fatalf("expected list indices [1] after detached controller change, got %v", indices)
	}

	list.SetSelectedIndex(0)
	if list.SelectedIndex() != 0 {
		t.Fatalf("expected list to accept manual selection after detach, got %d", list.SelectedIndex())
	}
	if ctrl.SelectedIndex() != 0 {
		t.Fatalf("expected controller to remain unchanged after detach, got %d", ctrl.SelectedIndex())
	}
	if indices := list.SelectedIndices(); len(indices) != 1 || indices[0] != 0 {
		t.Fatalf("expected list indices [0] after manual selection, got %v", indices)
	}
}

func TestListSelectionModeNone(t *testing.T) {
	list := NewList()
	list.AddItem(core.NewGObject())
	list.AddItem(core.NewGObject())

	list.SetSelectionMode(ListSelectionModeNone)
	list.handleItemClick(0)

	if list.SelectedIndex() != -1 {
		t.Fatalf("expected no selection when mode is none, got %d", list.SelectedIndex())
	}
	if indices := list.SelectedIndices(); len(indices) != 0 {
		t.Fatalf("expected empty selection indices, got %v", indices)
	}
}

func TestListSelectionMultipleSingleClick(t *testing.T) {
	list := NewList()
	list.AddItem(core.NewGObject())
	list.AddItem(core.NewGObject())
	list.AddItem(core.NewGObject())

	list.SetSelectionMode(ListSelectionModeMultipleSingleClick)

	list.handleItemClick(0)
	list.handleItemClick(1)

	indices := list.SelectedIndices()
	if len(indices) != 2 || indices[0] != 0 || indices[1] != 1 {
		t.Fatalf("expected indices [0 1], got %v", indices)
	}
	if list.SelectedIndex() != 1 {
		t.Fatalf("expected primary selected index 1, got %d", list.SelectedIndex())
	}

	list.handleItemClick(0)
	indices = list.SelectedIndices()
	if len(indices) != 1 || indices[0] != 1 {
		t.Fatalf("expected indices [1] after toggling, got %v", indices)
	}

	list.handleItemClick(1)
	if indices = list.SelectedIndices(); len(indices) != 0 {
		t.Fatalf("expected no selection after toggling off, got %v", indices)
	}
	if list.SelectedIndex() != -1 {
		t.Fatalf("expected primary index -1 after clearing, got %d", list.SelectedIndex())
	}
}

func TestListSetSelectedIndices(t *testing.T) {
	list := NewList()
	for i := 0; i < 4; i++ {
		list.AddItem(core.NewGObject())
	}
	list.SetSelectionMode(ListSelectionModeMultiple)

	list.SetSelectedIndices([]int{2, 1, 1, 3, -1})

	indices := list.SelectedIndices()
	if len(indices) != 3 || indices[0] != 1 || indices[1] != 2 || indices[2] != 3 {
		t.Fatalf("expected indices [1 2 3], got %v", indices)
	}
	if list.SelectedIndex() != 1 {
		t.Fatalf("expected primary index 1, got %d", list.SelectedIndex())
	}

	list.RemoveSelection(2)
	indices = list.SelectedIndices()
	if len(indices) != 2 || indices[0] != 1 || indices[1] != 3 {
		t.Fatalf("expected indices [1 3], got %v", indices)
	}

	list.ClearSelection()
	if list.SelectedIndex() != -1 {
		t.Fatalf("expected selection cleared, got %d", list.SelectedIndex())
	}
	if len(list.SelectedIndices()) != 0 {
		t.Fatalf("expected empty indices after clear, got %v", list.SelectedIndices())
	}
}

func TestListSelectionSyncsButtonState(t *testing.T) {
	list := NewList()
	btn := NewButton()
	btn.SetMode(ButtonModeCheck)
	list.AddItem(btn.GComponent.GObject)

	if btn.Selected() {
		t.Fatalf("expected button to start unselected")
	}

	list.SetSelectedIndex(0)
	if !btn.Selected() {
		t.Fatalf("expected button to become selected with list")
	}

	list.ClearSelection()
	if btn.Selected() {
		t.Fatalf("expected button to clear selection with list")
	}
}

func TestListControllerRetainsPageAfterClear(t *testing.T) {
	list := NewList()
	list.AddItem(core.NewGObject())
	list.AddItem(core.NewGObject())

	ctrl := core.NewController("selection")
	ctrl.SetPages([]string{"p0", "p1"}, []string{"page0", "page1"})
	list.SetSelectionController(ctrl)

	list.SetSelectedIndex(1)
	if ctrl.SelectedIndex() != 1 {
		t.Fatalf("expected controller to sync with list selection, got %d", ctrl.SelectedIndex())
	}

	list.ClearSelection()
	if ctrl.SelectedIndex() != 1 {
		t.Fatalf("expected controller to retain last page after clear, got %d", ctrl.SelectedIndex())
	}
	if list.SelectedIndex() != -1 {
		t.Fatalf("expected list selection cleared, got %d", list.SelectedIndex())
	}
}
