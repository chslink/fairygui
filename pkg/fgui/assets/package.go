package assets

import (
	"errors"

	"github.com/chslink/fairygui/pkg/fgui/utils"
)

const packageSignature uint32 = 0x46475549 // "FGUI"

var (
	// ErrInvalidPackage is returned when the data does not look like a FairyGUI package.
	ErrInvalidPackage = errors.New("assets: invalid package signature")
	// ErrCompressedPackage indicates the package uses compression that is not yet implemented.
	ErrCompressedPackage = errors.New("assets: compressed packages are not supported yet")
)

// Dependency represents a package dependency entry.
type Dependency struct {
	ID   string
	Name string
}

// Package holds the parsed metadata for a FairyGUI package.
type Package struct {
	ResKey       string
	ID           string
	Name         string
	Version      int
	Dependencies []Dependency

	rawData       []byte
	indexTablePos int
	stringTable   []string
}

// ParsePackage reads the package header and shared metadata from the provided binary data.
func ParsePackage(data []byte, resKey string) (*Package, error) {
	buf := utils.NewByteBuffer(data)
	if buf.Len() < 4 {
		return nil, ErrInvalidPackage
	}

	if buf.ReadUint32() != packageSignature {
		return nil, ErrInvalidPackage
	}

	version := int(buf.ReadInt32())
	compressed := buf.ReadBool()
	id := buf.ReadUTFString()
	name := buf.ReadUTFString()
	if err := buf.Skip(20); err != nil {
		return nil, err
	}

	if compressed {
		return nil, ErrCompressedPackage
	}

	pkg := &Package{
		ResKey:  resKey,
		ID:      id,
		Name:    name,
		Version: version,
		rawData: data,
	}

	indexTablePos := buf.Pos()

	if buf.Seek(indexTablePos, 4) {
		count := int(buf.ReadInt32())
		strings := make([]string, count)
		for i := 0; i < count; i++ {
			strings[i] = buf.ReadUTFString()
		}
		buf.StringTable = strings
		pkg.stringTable = strings
	}

	if buf.Seek(indexTablePos, 0) {
		depCount := int(buf.ReadInt16())
		if depCount > 0 {
			deps := make([]Dependency, 0, depCount)
			for i := 0; i < depCount; i++ {
				depID := stringValue(buf.ReadS())
				depName := stringValue(buf.ReadS())
				deps = append(deps, Dependency{ID: depID, Name: depName})
			}
			pkg.Dependencies = deps
		}
	}

	_ = buf.SetPos(indexTablePos)
	pkg.indexTablePos = indexTablePos

	return pkg, nil
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
