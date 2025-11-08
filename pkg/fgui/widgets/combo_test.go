package widgets

import (
	"encoding/binary"
	"testing"

	"github.com/chslink/fairygui/pkg/fgui/assets"
	"github.com/chslink/fairygui/pkg/fgui/core"
	"github.com/chslink/fairygui/pkg/fgui/utils"
)

func TestGComboBoxSetupAfterAddParsesItems(t *testing.T) {
	bufData := make([]byte, 128)
	bufData[0] = 7 // segCount
	bufData[1] = 1 // useShort
	// block6 starts at offset 30
	binary.BigEndian.PutUint16(bufData[14:], 0x001E)

	// block payload at offset 30
	pos := 30
	bufData[pos] = byte(assets.ObjectTypeComboBox)
	pos++
	binary.BigEndian.PutUint16(bufData[pos:], uint16(2))
	pos += 2

	// item 0
	binary.BigEndian.PutUint16(bufData[pos:], uint16(6))
	pos += 2
	binary.BigEndian.PutUint16(bufData[pos:], uint16(1)) // item text
	pos += 2
	binary.BigEndian.PutUint16(bufData[pos:], uint16(2)) // value
	pos += 2
	binary.BigEndian.PutUint16(bufData[pos:], uint16(3)) // icon
	pos += 2

	// item 1
	binary.BigEndian.PutUint16(bufData[pos:], uint16(6))
	pos += 2
	binary.BigEndian.PutUint16(bufData[pos:], uint16(4))
	pos += 2
	binary.BigEndian.PutUint16(bufData[pos:], uint16(5))
	pos += 2
	binary.BigEndian.PutUint16(bufData[pos:], uint16(6))
	pos += 2

	// default text uses item index 1
	binary.BigEndian.PutUint16(bufData[pos:], uint16(4))
	pos += 2
	// icon string
	binary.BigEndian.PutUint16(bufData[pos:], uint16(3))
	pos += 2
	// titleColor flag + color
	bufData[pos] = 1
	pos++
	bufData[pos+0] = 0x00
	bufData[pos+1] = 0x80
	bufData[pos+2] = 0x00
	bufData[pos+3] = 0xFF
	pos += 4
	// visible item count
	binary.BigEndian.PutUint32(bufData[pos:], uint32(5))
	pos += 4
	// popup direction
	bufData[pos] = byte(PopupDirectionUp)
	pos++
	// selection controller index
	binary.BigEndian.PutUint16(bufData[pos:], uint16(0))

	buf := utils.NewByteBuffer(bufData)
	buf.StringTable = []string{
		"",
		"Option A",
		"ValueA",
		"icon-a",
		"Option B",
		"ValueB",
		"",
		"Unused",
	}

	parent := core.NewGComponent()
	ctrl := core.NewController("state")
	ctrl.SetPages([]string{"0"}, []string{"page0"})
	parent.AddController(ctrl)

	child := &assets.ComponentChild{Type: assets.ObjectTypeComboBox}
	ctx := &SetupContext{
		Child:  child,
		Parent: parent,
	}

	combo := NewComboBox()
	combo.SetupAfterAdd(ctx, buf)

	if got := combo.Items(); len(got) != 2 || got[0] != "Option A" || got[1] != "Option B" {
		t.Fatalf("unexpected items %+v", got)
	}
	if got := combo.Values(); len(got) != 2 || got[0] != "ValueA" || got[1] != "ValueB" {
		t.Fatalf("unexpected values %+v", got)
	}
	if got := combo.Icons(); len(got) != 2 || got[0] != "icon-a" || got[1] != "" {
		t.Fatalf("unexpected icons %+v", got)
	}
	if combo.SelectedIndex() != 1 {
		t.Fatalf("expected selected index 1, got %d", combo.SelectedIndex())
	}
	if combo.Text() != "Option B" {
		t.Fatalf("expected text Option B, got %q", combo.Text())
	}
	if combo.Icon() != "" {
		t.Fatalf("expected icon to be empty, got %q", combo.Icon())
	}
	if combo.VisibleItemCount() != 5 {
		t.Fatalf("expected visible count 5, got %d", combo.VisibleItemCount())
	}
	if combo.PopupDirection() != PopupDirectionUp {
		t.Fatalf("expected popup direction Up, got %d", combo.PopupDirection())
	}
	if combo.TitleColor() != "#008000" {
		t.Fatalf("expected title color #008000, got %q", combo.TitleColor())
	}
	if combo.SelectionController() != ctrl {
		t.Fatalf("expected selection controller to be assigned")
	}
}

func TestGTextInputSetupBeforeAdd(t *testing.T) {
	data := make([]byte, 128)
	data[0] = 6 // segCount (need blocks 0..5)
	data[1] = 1 // useShort
	// block4 offset (indices start after header)
	binary.BigEndian.PutUint16(data[10:], uint16(20))

	pos := 20
	// prompt string index
	binary.BigEndian.PutUint16(data[pos:], uint16(1))
	pos += 2
	// restrict string index
	binary.BigEndian.PutUint16(data[pos:], uint16(2))
	pos += 2
	// max length
	binary.BigEndian.PutUint32(data[pos:], uint32(16))
	pos += 4
	// keyboard type code (4 => number)
	binary.BigEndian.PutUint32(data[pos:], uint32(4))
	pos += 4
	// password flag
	data[pos] = 1

	buf := utils.NewByteBuffer(data)
	buf.StringTable = []string{
		"",
		"${prompt}",
		"A-Z",
	}

	input := NewTextInput()
	input.SetupBeforeAdd(buf, 0)

	if input.PromptText() != "${prompt}" {
		t.Fatalf("expected prompt ${prompt}, got %q", input.PromptText())
	}
	if input.Restrict() != "A-Z" {
		t.Fatalf("expected restrict A-Z, got %q", input.Restrict())
	}
	if input.MaxLength() != 16 {
		t.Fatalf("expected max length 16, got %d", input.MaxLength())
	}
	if input.KeyboardType() != KeyboardTypeNumber {
		t.Fatalf("expected keyboard type number, got %s", input.KeyboardType())
	}
	if !input.Password() {
		t.Fatalf("expected password flag true")
	}
	if !input.Editable() {
		t.Fatalf("expected editable default true")
	}
}
