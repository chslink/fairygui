package widgets

import (
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/utils"
)

// TestComboBox_readItemsFromBuffer_Empty verifies no panic with empty buffer.
func TestComboBox_readItemsFromBuffer_Empty(t *testing.T) {
	c := &GComboBox{}
	buf := utils.NewByteBuffer([]byte{})
	c.readItemsFromBuffer(buf)
	if len(c.items) != 0 {
		t.Errorf("expected 0 items, got %d", len(c.items))
	}
}

// TestComboBox_readItemsFromBuffer_Short verifies no panic with buffer too short.
func TestComboBox_readItemsFromBuffer_Short(t *testing.T) {
	c := &GComboBox{}
	buf := utils.NewByteBuffer([]byte{0x00}) // 1 byte, not enough
	c.readItemsFromBuffer(buf)
	if len(c.items) != 0 {
		t.Errorf("expected 0 items, got %d", len(c.items))
	}
}

// TestComboBox_readItemsFromBuffer_ZeroItems verifies reading zero items.
func TestComboBox_readItemsFromBuffer_ZeroItems(t *testing.T) {
	c := &GComboBox{}
	// 1 byte object type + 2 bytes int16(0) = 3 bytes
	buf := utils.NewByteBuffer([]byte{0x01, 0x00, 0x00})
	c.readItemsFromBuffer(buf)
	if len(c.items) != 0 {
		t.Errorf("expected 0 items, got %d", len(c.items))
	}
}

// TestComboBox_readItemsFromBuffer_NegativeCount verifies negative count returns early.
func TestComboBox_readItemsFromBuffer_NegativeCount(t *testing.T) {
	c := &GComboBox{}
	// 1 byte object type + 2 bytes int16(-1) = 0xFFFF
	buf := utils.NewByteBuffer([]byte{0x01, 0xFF, 0xFF})
	c.readItemsFromBuffer(buf)
	if len(c.items) != 0 {
		t.Errorf("expected 0 items, got %d", len(c.items))
	}
}

// TestComboBox_readItemsFromBuffer_LargeCount verifies large count is rejected.
func TestComboBox_readItemsFromBuffer_LargeCount(t *testing.T) {
	c := &GComboBox{}
	// 1 byte object type + 2 bytes int16(2000) — exceeds 1000 limit
	buf := utils.NewByteBuffer([]byte{0x01, 0xD0, 0x07})
	c.readItemsFromBuffer(buf)
	if len(c.items) != 0 {
		t.Errorf("expected 0 items, got %d", len(c.items))
	}
}

// TestComboBox_ConstructExtension_NoCrash verifies constructExtension doesn't panic.
func TestComboBox_ConstructExtension_NoCrash(t *testing.T) {
	c := NewComboBox()
	// Empty buffer — should not panic
	err := c.ConstructExtension(nil)
	if err != nil {
		t.Logf("ConstructExtension with nil: %v", err)
	}

	buf := utils.NewByteBuffer([]byte{})
	err = c.ConstructExtension(buf)
	if err != nil {
		t.Logf("ConstructExtension with empty: %v", err)
	}

	// Buffer with minimum data for Seek(0,6) to fail
	buf = utils.NewByteBuffer([]byte{0x01, 0x01, 0x00, 0x00, 0x00, 0x00})
	err = c.ConstructExtension(buf)
	if err != nil {
		t.Logf("ConstructExtension with minimal: %v", err)
	}
}

// TestComboBox_ItemsDefault verifies itemsUpdated starts true and items is initialized.
func TestComboBox_ItemsDefault(t *testing.T) {
	c := NewComboBox()
	if len(c.items) != 0 {
		t.Errorf("expected 0 items initially, got %d", len(c.items))
	}
	if !c.itemsUpdated {
		t.Error("expected itemsUpdated to be true initially")
	}
}

// TestComboBox_SetItems marks items as updated.
func TestComboBox_SetItems(t *testing.T) {
	c := NewComboBox()
	c.SetItems([]string{"A", "B", "C"}, nil, nil)
	if len(c.items) != 3 {
		t.Errorf("expected 3 items, got %d", len(c.items))
	}
	if !c.itemsUpdated {
		t.Error("SetItems should set itemsUpdated = true")
	}
}
