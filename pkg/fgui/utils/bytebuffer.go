package utils

import (
	"encoding/binary"
	"fmt"
	"math"
)

const (
	indexNull  = 0xFFFE
	indexEmpty = 0xFFFD
)

// ByteBuffer mimics the behaviour of fgui.ByteBuffer backed by big-endian data.
type ByteBuffer struct {
	data []byte
	pos  int
	// StringTable stores shared strings referenced by ReadS.
	StringTable []string
	// Version carries the package version the buffer belongs to.
	Version int
}

// NewByteBuffer creates a buffer that wraps the whole data slice.
func NewByteBuffer(data []byte) *ByteBuffer {
	return &ByteBuffer{data: data}
}

// NewByteBufferRange wraps a subsection of data starting at offset for length bytes.
func NewByteBufferRange(data []byte, offset, length int) (*ByteBuffer, error) {
	if offset < 0 || length < 0 || offset+length > len(data) {
		return nil, fmt.Errorf("bytebuffer: invalid range offset=%d length=%d len(data)=%d", offset, length, len(data))
	}
	sub := data[offset : offset+length]
	return &ByteBuffer{data: sub}, nil
}

// Len returns the total length of the buffer.
func (b *ByteBuffer) Len() int {
	return len(b.data)
}

// Pos returns the current read cursor.
func (b *ByteBuffer) Pos() int {
	return b.pos
}

// SetPos moves the read cursor to the specified position.
func (b *ByteBuffer) SetPos(pos int) error {
	if pos < 0 || pos > len(b.data) {
		return fmt.Errorf("bytebuffer: position %d out of range", pos)
	}
	b.pos = pos
	return nil
}

// Skip advances the cursor by count bytes.
func (b *ByteBuffer) Skip(count int) error {
	return b.SetPos(b.pos + count)
}

func (b *ByteBuffer) ensure(count int) {
	if b.pos+count > len(b.data) {
		panic(fmt.Sprintf("bytebuffer: attempted read beyond buffer: pos=%d count=%d len=%d", b.pos, count, len(b.data)))
	}
}

// ReadBool returns true when the next byte equals 1.
func (b *ByteBuffer) ReadBool() bool {
	return b.ReadUint8() == 1
}

// ReadUint8 reads an unsigned byte.
func (b *ByteBuffer) ReadUint8() uint8 {
	b.ensure(1)
	value := b.data[b.pos]
	b.pos++
	return value
}

// ReadUint16 reads a big-endian uint16.
func (b *ByteBuffer) ReadUint16() uint16 {
	b.ensure(2)
	value := binary.BigEndian.Uint16(b.data[b.pos:])
	b.pos += 2
	return value
}

// ReadInt16 reads a big-endian int16.
func (b *ByteBuffer) ReadInt16() int16 {
	return int16(b.ReadUint16())
}

// ReadUint32 reads a big-endian uint32.
func (b *ByteBuffer) ReadUint32() uint32 {
	b.ensure(4)
	value := binary.BigEndian.Uint32(b.data[b.pos:])
	b.pos += 4
	return value
}

// ReadInt32 reads a big-endian int32.
func (b *ByteBuffer) ReadInt32() int32 {
	return int32(b.ReadUint32())
}

// ReadFloat32 reads a big-endian float32 value.
func (b *ByteBuffer) ReadFloat32() float32 {
	return math.Float32frombits(b.ReadUint32())
}

// ReadByte reads a signed byte.
func (b *ByteBuffer) ReadByte() int8 {
	return int8(b.ReadUint8())
}

// ReadS returns a pointer to the resolved string or nil when the buffer encoded null.
func (b *ByteBuffer) ReadS() *string {
	index := b.ReadUint16()
	switch index {
	case 0xFFFF:
		return nil
	case indexNull:
		return nil
	case indexEmpty:
		empty := ""
		return &empty
	default:
		if int(index) >= len(b.StringTable) {
			empty := ""
			return &empty
		}
		value := b.StringTable[index]
		// Allocate copy so callers can mutate without affecting the table.
		copyValue := value
		return &copyValue
	}
}

// ReadSArray returns an array of string pointers for the given count.
func (b *ByteBuffer) ReadSArray(cnt int) []*string {
	result := make([]*string, cnt)
	for i := 0; i < cnt; i++ {
		result[i] = b.ReadS()
	}
	return result
}

// WriteS updates the string table entry referenced by the next uint16.
func (b *ByteBuffer) WriteS(value string) {
	index := b.ReadUint16()
	if index == indexNull || index == indexEmpty {
		return
	}
	if int(index) >= len(b.StringTable) {
		panic(fmt.Sprintf("bytebuffer: write string index %d out of range (%d entries)", index, len(b.StringTable)))
	}
	b.StringTable[index] = value
}

// ReadColor returns an RGBA value in 0xAARRGGBB format when hasAlpha is true, otherwise 0x00RRGGBB.
func (b *ByteBuffer) ReadColor(hasAlpha bool) uint32 {
	r := uint32(b.ReadUint8())
	g := uint32(b.ReadUint8())
	bl := uint32(b.ReadUint8())
	a := uint32(b.ReadUint8())
	if hasAlpha {
		return (a << 24) | (r << 16) | (g << 8) | bl
	}
	return (r << 16) | (g << 8) | bl
}

// ReadColorString returns a CSS-style colour string.
func (b *ByteBuffer) ReadColorString(hasAlpha bool) string {
	r := b.ReadUint8()
	g := b.ReadUint8()
	bl := b.ReadUint8()
	a := b.ReadUint8()

	if hasAlpha && a != 255 {
		return fmt.Sprintf("rgba(%d,%d,%d,%.3g)", r, g, bl, float64(a)/255.0)
	}
	return fmt.Sprintf("#%02x%02x%02x", r, g, bl)
}

// ReadChar reads a UTF-16 code unit and returns it as a rune.
func (b *ByteBuffer) ReadChar() rune {
	return rune(b.ReadUint16())
}

// ReadBuffer slices out a sub-buffer of the given size.
func (b *ByteBuffer) ReadBuffer() *ByteBuffer {
	count := int(b.ReadUint32())
	b.ensure(count)
	sub := b.data[b.pos : b.pos+count]
	b.pos += count
	child := NewByteBuffer(sub)
	child.StringTable = b.StringTable
	child.Version = b.Version
	return child
}

// Seek moves the cursor to a block entry referenced by an index table.
func (b *ByteBuffer) Seek(indexTablePos int, blockIndex int) bool {
	tmp := b.pos
	if err := b.SetPos(indexTablePos); err != nil {
		return false
	}
	segCount := int(b.ReadUint8())
	if blockIndex >= segCount {
		b.pos = tmp
		return false
	}
	useShort := b.ReadUint8() == 1
	var newPos int
	if useShort {
		if err := b.Skip(blockIndex * 2); err != nil {
			b.pos = tmp
			return false
		}
		newPos = int(b.ReadUint16())
	} else {
		if err := b.Skip(blockIndex * 4); err != nil {
			b.pos = tmp
			return false
		}
		newPos = int(b.ReadUint32())
	}

	if newPos > 0 {
		_ = b.SetPos(indexTablePos + newPos)
		return true
	}

	b.pos = tmp
	return false
}

// ReadUTFString reads a length-prefixed UTF-8 string.
func (b *ByteBuffer) ReadUTFString() string {
	length := int(b.ReadUint16())
	if length == 0 {
		return ""
	}
	b.ensure(length)
	value := string(b.data[b.pos : b.pos+length])
	b.pos += length
	return value
}

// ReadBytes returns a copy of the next count bytes.
func (b *ByteBuffer) ReadBytes(count int) []byte {
	b.ensure(count)
	out := make([]byte, count)
	copy(out, b.data[b.pos:b.pos+count])
	b.pos += count
	return out
}

// Remaining returns the unread byte count.
func (b *ByteBuffer) Remaining() int {
	return len(b.data) - b.pos
}

// SubBuffer returns a view that starts at offset and spans length bytes.
func (b *ByteBuffer) SubBuffer(offset, length int) (*ByteBuffer, error) {
	if offset < 0 || length < 0 || offset+length > len(b.data) {
		return nil, fmt.Errorf("bytebuffer: invalid sub buffer range offset=%d length=%d len=%d", offset, length, len(b.data))
	}
	sub := b.data[offset : offset+length]
	child := NewByteBuffer(sub)
	child.StringTable = b.StringTable
	child.Version = b.Version
	return child, nil
}
