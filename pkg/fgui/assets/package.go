package assets

import (
	"errors"
	"strings"

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
	Branches     []string
	BranchIndex  int
	StringTable  []string

	Items       []*PackageItem
	itemsByID   map[string]*PackageItem
	itemsByName map[string]*PackageItem

	rawData       []byte
	indexTablePos int
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
		ResKey:      resKey,
		ID:          id,
		Name:        name,
		Version:     version,
		BranchIndex: -1,
		rawData:     data,
		itemsByID:   make(map[string]*PackageItem),
		itemsByName: make(map[string]*PackageItem),
	}

	buf.Version = version

	indexTablePos := buf.Pos()

	if buf.Seek(indexTablePos, 4) {
		count := int(buf.ReadInt32())
		strings := make([]string, count)
		for i := 0; i < count; i++ {
			strings[i] = buf.ReadUTFString()
		}
		buf.StringTable = strings
		pkg.StringTable = strings
	}

	if buf.Seek(indexTablePos, 5) {
		count := int(buf.ReadInt32())
		for i := 0; i < count; i++ {
			index := int(buf.ReadUint16())
			length := int(buf.ReadInt32())
			value := string(buf.ReadBytes(length))
			if index >= 0 && index < len(buf.StringTable) {
				buf.StringTable[index] = value
			}
		}
		pkg.StringTable = buf.StringTable
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

		if version >= 2 {
			branchCount := int(buf.ReadInt16())
			if branchCount > 0 {
				branches := make([]string, branchCount)
				for i := 0; i < branchCount; i++ {
					branches[i] = stringValue(buf.ReadS())
				}
				pkg.Branches = branches
				pkg.BranchIndex = -1
			}
		}
	}

	_ = buf.SetPos(indexTablePos)
	pkg.indexTablePos = indexTablePos

	if err := parseItems(buf, pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func parseItems(buf *utils.ByteBuffer, pkg *Package) error {
	if !buf.Seek(pkg.indexTablePos, 1) {
		return nil
	}

	count := int(buf.ReadUint16())
	if count <= 0 {
		return nil
	}

	pkg.Items = make([]*PackageItem, 0, count)

	path := pkg.ResKey
	shortPath := ""
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		shortPath = path[:idx+1]
	}
	path = path + "_"

	ver2 := pkg.Version >= 2
	branchIncluded := len(pkg.Branches) > 0

	for i := 0; i < count; i++ {
		nextPos := int(buf.ReadInt32())
		nextPos += buf.Pos()

		item := &PackageItem{Owner: pkg}
		item.Type = PackageItemType(uint8(buf.ReadByte()))
		item.ID = stringValue(buf.ReadS())
		item.Name = stringValue(buf.ReadS())
		_ = buf.ReadS() // path (unused)
		item.File = stringValue(buf.ReadS())
		_ = buf.ReadBool() // exported
		item.Width = int(buf.ReadInt32())
		item.Height = int(buf.ReadInt32())

		switch item.Type {
		case PackageItemTypeImage:
			item.ObjectType = ObjectTypeImage
			scaleOption := buf.ReadByte()
			if scaleOption == 1 {
				item.Scale9Grid = &Rect{
					X:      int(buf.ReadInt32()),
					Y:      int(buf.ReadInt32()),
					Width:  int(buf.ReadInt32()),
					Height: int(buf.ReadInt32()),
				}
				item.TileGridIndice = int(buf.ReadInt32())
			} else if scaleOption == 2 {
				item.ScaleByTile = true
			}
			item.Smoothing = buf.ReadBool()
		case PackageItemTypeMovieClip:
			item.Smoothing = buf.ReadBool()
			item.ObjectType = ObjectTypeMovieClip
			item.RawData = buf.ReadBuffer()
		case PackageItemTypeFont:
			item.RawData = buf.ReadBuffer()
		case PackageItemTypeComponent:
			ext := buf.ReadByte()
			if ext > 0 {
				item.ObjectType = ObjectType(uint8(ext))
			} else {
				item.ObjectType = ObjectTypeComponent
			}
			item.RawData = buf.ReadBuffer()
		case PackageItemTypeAtlas, PackageItemTypeSound, PackageItemTypeMisc:
			if item.File != "" {
				item.File = path + item.File
			} else if item.Type == PackageItemTypeAtlas {
				item.File = path + item.ID
			}
		case PackageItemTypeSpine, PackageItemTypeDragonBones:
			if item.File != "" {
				item.File = shortPath + item.File
			}
			item.SkeletonAnchor = &Point{X: buf.ReadFloat32(), Y: buf.ReadFloat32()}
		default:
			// No additional payload to read.
		}

		if ver2 {
			branchPath := stringValue(buf.ReadS())
			if branchPath != "" {
				if item.Name != "" {
					item.Name = branchPath + "/" + item.Name
				} else {
					item.Name = branchPath
				}
			}
			branchCnt := int(buf.ReadUint8())
			if branchCnt > 0 {
				if branchIncluded {
					item.Branches = readSArrayValues(buf, branchCnt)
				} else {
					alias := stringValue(buf.ReadS())
					if alias != "" {
						pkg.itemsByID[alias] = item
					}
					for j := 1; j < branchCnt; j++ {
						_ = buf.ReadS()
					}
				}
			}
			highResCnt := int(buf.ReadUint8())
			if highResCnt > 0 {
				item.HighResolution = readSArrayValues(buf, highResCnt)
			}
		}

		pkg.Items = append(pkg.Items, item)
		if item.ID != "" {
			pkg.itemsByID[item.ID] = item
		}
		if item.Name != "" {
			pkg.itemsByName[item.Name] = item
		}

		_ = buf.SetPos(nextPos)
	}
	return nil
}

func readSArrayValues(buf *utils.ByteBuffer, cnt int) []string {
	raw := buf.ReadSArray(cnt)
	out := make([]string, cnt)
	for i, s := range raw {
		out[i] = stringValue(s)
	}
	return out
}

// ItemByID returns the package item with the given id.
func (p *Package) ItemByID(id string) *PackageItem {
	if p == nil {
		return nil
	}
	return p.itemsByID[id]
}

// ItemByName returns the package item with the given name.
func (p *Package) ItemByName(name string) *PackageItem {
	if p == nil {
		return nil
	}
	return p.itemsByName[name]
}
