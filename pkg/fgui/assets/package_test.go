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
