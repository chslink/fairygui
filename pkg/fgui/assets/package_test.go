package assets

import (
	"bytes"
	"compress/flate"
	"encoding/binary"
	"os"
	"path/filepath"
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
	data := buildPackageBytes(t, true)
	pkg, err := ParsePackage(data, "ui/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkg.Items) == 0 {
		t.Fatalf("expected items after decompression")
	}
}

func TestParsePackageInvalidSignature(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02}
	if _, err := ParsePackage(data, "ui/test"); err != ErrInvalidPackage {
		t.Fatalf("expected invalid package error, got %v", err)
	}
}

func TestParsePackageWithItems(t *testing.T) {
	data := buildPackageBytes(t, false)
	pkg, err := ParsePackage(data, "ui/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkg.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(pkg.Items))
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
	if item.File != "" {
		t.Fatalf("expected image file blank, got %q", item.File)
	}

	if len(pkg.Dependencies) != 1 || pkg.Dependencies[0].ID != "dep-id" {
		t.Fatalf("unexpected dependencies: %+v", pkg.Dependencies)
	}

	atlas := pkg.ItemByID("atlas-item")
	if atlas == nil || atlas.File != "ui/test_atlas.png" {
		t.Fatalf("expected atlas file with prefix, got %+v", atlas)
	}
	if item.Atlas != atlas {
		t.Fatalf("expected image to reference atlas")
	}
	sprite := pkg.Sprites["item-001"]
	if sprite == nil || sprite.Rect.Width != 100 || sprite.Rect.Height != 200 {
		t.Fatalf("expected sprite metadata, got %+v", sprite)
	}
	if item.Sprite == nil || item.Sprite != sprite {
		t.Fatalf("expected item to link sprite")
	}
	if item.PixelHitTest == nil || item.PixelHitTest.Width != 10 || len(item.PixelHitTest.Data) != 2 {
		t.Fatalf("unexpected pixel hit test data: %+v", item.PixelHitTest)
	}
	if item.PixelHitTest.Scale <= 0 || item.PixelHitTest.Scale >= 1 {
		t.Fatalf("unexpected pixel hit scale: %v", item.PixelHitTest.Scale)
	}
}

func TestParseRealFUI(t *testing.T) {
	path := filepath.Join("..", "..", "..", "demo", "assets", "Bag.fui")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("real asset not available: %v", err)
	}
	pkg, err := ParsePackage(data, "demo/assets/Bag")
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	if len(pkg.Items) == 0 {
		t.Fatalf("expected real package to contain items")
	}

	var component *PackageItem
	for _, item := range pkg.Items {
		if item.Type == PackageItemTypeComponent && item.Component != nil {
			component = item
			break
		}
	}
	if component == nil {
		t.Fatalf("expected component item with parsed data")
	}
	if len(component.Component.Children) == 0 {
		t.Fatalf("expected component to contain children metadata")
	}
}

func buildPackageFixture(t *testing.T) ([]byte, int) {
	t.Helper()

	strings := []string{
		"dep-id",     // 0
		"dep-name",   // 1
		"atlas-item", // 2
		"AtlasItem",  // 3
		"atlas.png",  // 4
		"item-001",   // 5
		"ItemName",   // 6
	}
	stringIndex := map[string]uint16{}
	for i, s := range strings {
		stringIndex[s] = uint16(i)
	}

	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, packageSignature)
	_ = binary.Write(&buf, binary.BigEndian, int32(1))
	buf.WriteByte(0)
	writeUTF(&buf, "pkg")
	writeUTF(&buf, "name")
	buf.Write(make([]byte, 20))

	indexTablePos := buf.Len()
	buf.WriteByte(6)
	buf.WriteByte(0)
	offsetsPos := buf.Len()
	buf.Write(make([]byte, 6*4))

	segments := make([][]byte, 6)

	// Dependencies segment (0)
	dep := bytes.Buffer{}
	_ = binary.Write(&dep, binary.BigEndian, uint16(1))
	_ = binary.Write(&dep, binary.BigEndian, stringIndex["dep-id"])
	_ = binary.Write(&dep, binary.BigEndian, stringIndex["dep-name"])
	segments[0] = dep.Bytes()

	// Items segment (1)
	items := bytes.Buffer{}
	_ = binary.Write(&items, binary.BigEndian, uint16(2))
	// Atlas item chunk
	atlasChunk := bytes.Buffer{}
	atlasChunk.WriteByte(uint8(PackageItemTypeAtlas))
	_ = binary.Write(&atlasChunk, binary.BigEndian, stringIndex["atlas-item"])
	_ = binary.Write(&atlasChunk, binary.BigEndian, stringIndex["AtlasItem"])
	_ = binary.Write(&atlasChunk, binary.BigEndian, uint16(0xFFFD))
	_ = binary.Write(&atlasChunk, binary.BigEndian, stringIndex["atlas.png"])
	atlasChunk.WriteByte(0)
	_ = binary.Write(&atlasChunk, binary.BigEndian, int32(512))
	_ = binary.Write(&atlasChunk, binary.BigEndian, int32(512))
	_ = binary.Write(&items, binary.BigEndian, int32(atlasChunk.Len()))
	items.Write(atlasChunk.Bytes())

	// Image item chunk
	imageChunk := bytes.Buffer{}
	imageChunk.WriteByte(uint8(PackageItemTypeImage))
	_ = binary.Write(&imageChunk, binary.BigEndian, stringIndex["item-001"])
	_ = binary.Write(&imageChunk, binary.BigEndian, stringIndex["ItemName"])
	_ = binary.Write(&imageChunk, binary.BigEndian, uint16(0xFFFD))
	_ = binary.Write(&imageChunk, binary.BigEndian, uint16(0xFFFE))
	imageChunk.WriteByte(1)
	_ = binary.Write(&imageChunk, binary.BigEndian, int32(100))
	_ = binary.Write(&imageChunk, binary.BigEndian, int32(200))
	imageChunk.WriteByte(0)
	imageChunk.WriteByte(1)
	_ = binary.Write(&items, binary.BigEndian, int32(imageChunk.Len()))
	items.Write(imageChunk.Bytes())
	segments[1] = items.Bytes()

	// Atlas sprites segment (2)
	sprites := bytes.Buffer{}
	_ = binary.Write(&sprites, binary.BigEndian, uint16(1))
	spriteChunk := bytes.Buffer{}
	_ = binary.Write(&spriteChunk, binary.BigEndian, stringIndex["item-001"])
	_ = binary.Write(&spriteChunk, binary.BigEndian, stringIndex["atlas-item"])
	_ = binary.Write(&spriteChunk, binary.BigEndian, int32(0))
	_ = binary.Write(&spriteChunk, binary.BigEndian, int32(0))
	_ = binary.Write(&spriteChunk, binary.BigEndian, int32(100))
	_ = binary.Write(&spriteChunk, binary.BigEndian, int32(200))
	spriteChunk.WriteByte(0)
	spriteLen := spriteChunk.Len()
	_ = binary.Write(&sprites, binary.BigEndian, uint16(spriteLen))
	sprites.Write(spriteChunk.Bytes())
	segments[2] = sprites.Bytes()

	// Pixel hit test segment (3)
	pixels := bytes.Buffer{}
	_ = binary.Write(&pixels, binary.BigEndian, uint16(1))
	pxChunk := bytes.Buffer{}
	_ = binary.Write(&pxChunk, binary.BigEndian, stringIndex["item-001"])
	_ = binary.Write(&pxChunk, binary.BigEndian, int32(0))
	_ = binary.Write(&pxChunk, binary.BigEndian, int32(10))
	pxChunk.WriteByte(2)
	_ = binary.Write(&pxChunk, binary.BigEndian, int32(2))
	pxChunk.Write([]byte{0xFF, 0x0F})
	pxLen := pxChunk.Len()
	_ = binary.Write(&pixels, binary.BigEndian, int32(pxLen))
	pixels.Write(pxChunk.Bytes())
	segments[3] = pixels.Bytes()

	// String table segment (4)
	strSeg := bytes.Buffer{}
	_ = binary.Write(&strSeg, binary.BigEndian, int32(len(strings)))
	for _, s := range strings {
		writeUTF(&strSeg, s)
	}
	segments[4] = strSeg.Bytes()

	data := buf.Bytes()
	for i, seg := range segments {
		if len(seg) == 0 {
			continue
		}
		start := len(data)
		data = append(data, seg...)
		binary.BigEndian.PutUint32(data[offsetsPos+i*4:], uint32(start-indexTablePos))
	}

	return data, indexTablePos
}

func buildPackageBytes(t *testing.T, compressed bool) []byte {
	data, index := buildPackageFixture(t)
	if !compressed {
		return data
	}
	body := data[index:]
	var comp bytes.Buffer
	w, err := flate.NewWriter(&comp, flate.BestSpeed)
	if err != nil {
		t.Fatalf("flate writer: %v", err)
	}
	if _, err := w.Write(body); err != nil {
		t.Fatalf("flate write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("flate close: %v", err)
	}
	header := append([]byte(nil), data[:index]...)
	if len(header) > 8 {
		header[8] = 1
	}
	return append(header, comp.Bytes()...)
}
