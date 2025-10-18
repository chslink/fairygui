package assets

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func writeUTF(buf *bytes.Buffer, value string) {
	_ = binary.Write(buf, binary.BigEndian, uint16(len(value)))
	buf.WriteString(value)
}

func TestParsePackageHeader(t *testing.T) {
	var data bytes.Buffer
	_ = binary.Write(&data, binary.BigEndian, packageSignature)
	_ = binary.Write(&data, binary.BigEndian, int32(3))
	data.WriteByte(0) // compressed flag
	writeUTF(&data, "pkg12345")
	writeUTF(&data, "TestPackage")
	data.Write(make([]byte, 20))
	data.WriteByte(0) // segCount => no index segments

	pkg, err := ParsePackage(data.Bytes(), "ui/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.ID != "pkg12345" || pkg.Name != "TestPackage" {
		t.Fatalf("unexpected package metadata: %+v", pkg)
	}
	if pkg.Version != 3 {
		t.Fatalf("expected version 3, got %d", pkg.Version)
	}
	if pkg.ResKey != "ui/test" {
		t.Fatalf("expected resKey 'ui/test', got %s", pkg.ResKey)
	}
	if len(pkg.Dependencies) != 0 {
		t.Fatalf("expected no dependencies, got %v", pkg.Dependencies)
	}
}

func TestParsePackageCompressed(t *testing.T) {
	var data bytes.Buffer
	_ = binary.Write(&data, binary.BigEndian, packageSignature)
	_ = binary.Write(&data, binary.BigEndian, int32(1))
	data.WriteByte(1) // compressed
	writeUTF(&data, "pkg")
	writeUTF(&data, "name")
	data.Write(make([]byte, 20))

	if _, err := ParsePackage(data.Bytes(), "ui/test"); err != ErrCompressedPackage {
		t.Fatalf("expected compressed package error, got %v", err)
	}
}

func TestParsePackageInvalidSignature(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02}
	if _, err := ParsePackage(data, "ui/test"); err != ErrInvalidPackage {
		t.Fatalf("expected invalid package error, got %v", err)
	}
}

func TestParsePackageWithItems(t *testing.T) {
	data := buildPackageWithItemsBuffer(t)
	pkg, err := ParsePackage(data, "ui/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkg.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(pkg.Items))
	}
	item := pkg.ItemByID("item-001")
	if item == nil {
		t.Fatalf("expected item lookup by id")
	}
	if item.Type != PackageItemTypeImage {
		t.Fatalf("expected image type, got %v", item.Type)
	}
	if item.Width != 100 || item.Height != 200 {
		t.Fatalf("unexpected size: %dx%d", item.Width, item.Height)
	}
	if !item.Smoothing {
		t.Fatalf("expected smoothing flag to be true")
	}
	if item.File != "atlas.png" {
		t.Fatalf("expected file 'atlas.png', got %q", item.File)
	}

	if len(pkg.Dependencies) != 1 || pkg.Dependencies[0].ID != "dep-id" {
		t.Fatalf("unexpected dependencies: %+v", pkg.Dependencies)
	}
}

func buildPackageWithItemsBuffer(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, packageSignature)
	_ = binary.Write(&buf, binary.BigEndian, int32(1))
	buf.WriteByte(0) // compressed flag
	writeUTF(&buf, "pkg")
	writeUTF(&buf, "name")
	buf.Write(make([]byte, 20))

	indexTablePos := buf.Len()
	buf.WriteByte(6) // segCount
	buf.WriteByte(0) // use uint32 offsets
	offsetsPos := buf.Len()
	for i := 0; i < 6; i++ {
		_ = binary.Write(&buf, binary.BigEndian, uint32(0))
	}

	depPos := buf.Len()
	_ = binary.Write(&buf, binary.BigEndian, uint16(1))
	_ = binary.Write(&buf, binary.BigEndian, uint16(0))
	_ = binary.Write(&buf, binary.BigEndian, uint16(1))

	itemPos := buf.Len()
	var items bytes.Buffer
	_ = binary.Write(&items, binary.BigEndian, uint16(1))

	var itemChunk bytes.Buffer
	itemChunk.WriteByte(uint8(PackageItemTypeImage))
	_ = binary.Write(&itemChunk, binary.BigEndian, uint16(2))
	_ = binary.Write(&itemChunk, binary.BigEndian, uint16(3))
	_ = binary.Write(&itemChunk, binary.BigEndian, uint16(0xFFFD))
	_ = binary.Write(&itemChunk, binary.BigEndian, uint16(4))
	itemChunk.WriteByte(1) // exported flag
	_ = binary.Write(&itemChunk, binary.BigEndian, int32(100))
	_ = binary.Write(&itemChunk, binary.BigEndian, int32(200))
	itemChunk.WriteByte(0) // scale option
	itemChunk.WriteByte(1) // smoothing

	blockLen := itemChunk.Len()
	_ = binary.Write(&items, binary.BigEndian, int32(blockLen))
	items.Write(itemChunk.Bytes())

	buf.Write(items.Bytes())

	stringPos := buf.Len()
	strings := []string{"dep-id", "dep-name", "item-001", "ItemName", "atlas.png"}
	_ = binary.Write(&buf, binary.BigEndian, int32(len(strings)))
	for _, s := range strings {
		writeUTF(&buf, s)
	}

	data := buf.Bytes()
	offsets := []uint32{
		uint32(depPos - indexTablePos),
		uint32(itemPos - indexTablePos),
		0,
		0,
		uint32(stringPos - indexTablePos),
		0,
	}
	for i, off := range offsets {
		binary.BigEndian.PutUint32(data[offsetsPos+i*4:], off)
	}

	return data
}
