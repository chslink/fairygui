package utils

import (
	"encoding/binary"
	"testing"
)

func TestByteBufferReadSVariants(t *testing.T) {
	data := []byte{
		0x00, 0x00, // index 0 -> "hello"
		0xFF, 0xFD, // empty string
		0xFF, 0xFE, // null
	}
	buf := NewByteBuffer(data)
	buf.StringTable = []string{"hello"}

	str := buf.ReadS()
	if str == nil || *str != "hello" {
		t.Fatalf("expected hello, got %v", str)
	}
	empty := buf.ReadS()
	if empty == nil || *empty != "" {
		t.Fatalf("expected empty string, got %v", empty)
	}
	nullStr := buf.ReadS()
	if nullStr != nil {
		t.Fatalf("expected nil for null string, got %v", *nullStr)
	}
}

func TestByteBufferWriteSUpdatesStringTable(t *testing.T) {
	data := []byte{0x00, 0x01}
	buf := NewByteBuffer(data)
	buf.StringTable = []string{"a", "b"}
	buf.WriteS("updated")

	if buf.StringTable[1] != "updated" {
		t.Fatalf("expected string table index 1 to be updated, got %q", buf.StringTable[1])
	}
	if buf.Pos() != 2 {
		t.Fatalf("expected position 2 after write, got %d", buf.Pos())
	}
}

func TestByteBufferReadColor(t *testing.T) {
	data := []byte{0x10, 0x20, 0x30, 0x40}
	buf := NewByteBuffer(data)

	withAlpha := buf.ReadColor(true)
	if withAlpha != 0x40102030 {
		t.Fatalf("expected 0x40102030, got %#x", withAlpha)
	}
	buf = NewByteBuffer(data)
	withoutAlpha := buf.ReadColor(false)
	if withoutAlpha != 0x00102030 {
		t.Fatalf("expected 0x00102030, got %#x", withoutAlpha)
	}

	buf = NewByteBuffer([]byte{0xFF, 0x00, 0x7F, 0x7F})
	css := buf.ReadColorString(true)
	if css != "rgba(255,0,127,0.498)" {
		t.Fatalf("unexpected css color string %q", css)
	}
	buf = NewByteBuffer([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	css = buf.ReadColorString(false)
	if css != "#ffffff" {
		t.Fatalf("unexpected css color string %q", css)
	}
}

func TestByteBufferReadBuffer(t *testing.T) {
	data := []byte{
		0x00, 0x00, 0x00, 0x04, // length 4
		0xAA, 0xBB, 0xCC, 0xDD,
	}
	parent := NewByteBuffer(data)
	parent.StringTable = []string{"parent"}
	parent.Version = 3

	child := parent.ReadBuffer()
	if child.Len() != 4 {
		t.Fatalf("expected child length 4, got %d", child.Len())
	}
	if child.StringTable[0] != "parent" || child.Version != 3 {
		t.Fatalf("child should inherit string table/version")
	}
	if child.ReadUint32() != 0xAABBCCDD {
		t.Fatalf("unexpected child data")
	}
	if parent.Pos() != 8 {
		t.Fatalf("parent position not advanced, got %d", parent.Pos())
	}
}

func TestByteBufferSeek(t *testing.T) {
	data := make([]byte, 32)
	data[0] = 2                              // segCount
	data[1] = 1                              // useShort
	binary.BigEndian.PutUint16(data[2:], 6)  // block 0 offset
	binary.BigEndian.PutUint16(data[4:], 10) // block 1 offset

	buf := NewByteBuffer(data)

	if !buf.Seek(0, 1) {
		t.Fatalf("expected seek success")
	}
	if buf.Pos() != 10 {
		t.Fatalf("expected position 10, got %d", buf.Pos())
	}

	if buf.Seek(0, 5) {
		t.Fatalf("expected seek failure for invalid block index")
	}

	// set block offset to zero to test fallback
	data[4] = 0
	data[5] = 0
	buf = NewByteBuffer(data)
	if buf.Seek(0, 1) {
		t.Fatalf("expected seek failure for zero offset")
	}
}

func TestByteBufferSkipBounds(t *testing.T) {
	buf := NewByteBuffer(make([]byte, 2))
	if err := buf.Skip(2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := buf.Skip(1); err == nil {
		t.Fatalf("expected error when skipping beyond buffer")
	}
}
