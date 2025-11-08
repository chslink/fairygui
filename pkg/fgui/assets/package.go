package assets

import (
	"bytes"
	"compress/flate"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/chslink/fairygui/pkg/fgui/utils"
)

const packageSignature uint32 = 0x46475549 // "FGUI"

var (
	// ErrInvalidPackage is returned when the data does not look like a FairyGUI package.
	ErrInvalidPackage = errors.New("assets: invalid package signature")
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
	Sprites     map[string]*AtlasSprite

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

	pkg := &Package{
		ResKey:      resKey,
		ID:          id,
		Name:        name,
		Version:     version,
		BranchIndex: -1,
		rawData:     data,
		itemsByID:   make(map[string]*PackageItem),
		itemsByName: make(map[string]*PackageItem),
		Sprites:     make(map[string]*AtlasSprite),
	}

	if compressed {
		decompressed, err := decompressRaw(buf.ReadBytes(buf.Remaining()))
		if err != nil {
			return nil, err
		}
		buf = utils.NewByteBuffer(decompressed)
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
			parseMovieClipData(item)
		case PackageItemTypeFont:
			item.RawData = buf.ReadBuffer()
		case PackageItemTypeComponent:
			ext := buf.ReadByte()
			// 诊断日志：打印 ext byte 值，特别关注滚动条组件
			if strings.Contains(strings.ToLower(item.Name), "scrollbar") {
				fmt.Printf("[DEBUG ScrollBar] Component %s: ext byte = %d (expected 16 for ScrollBar, got ObjectType=%d)\n",
					item.Name, ext, ObjectType(uint8(ext)))
			}
			if ext > 0 {
				item.ObjectType = ObjectType(uint8(ext))
			} else {
				item.ObjectType = ObjectTypeComponent
			}
			item.RawData = buf.ReadBuffer()
			parseComponentData(item)
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

	if err := parseAtlasSprites(buf, pkg); err != nil {
		return err
	}
	if err := parsePixelHitTests(buf, pkg); err != nil {
		return err
	}
	linkMovieClipSprites(pkg)

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

func parseAtlasSprites(buf *utils.ByteBuffer, pkg *Package) error {
	if !buf.Seek(pkg.indexTablePos, 2) {
		return nil
	}

	count := int(buf.ReadUint16())
	if count <= 0 {
		return nil
	}

	for i := 0; i < count; i++ {
		nextPos := int(buf.ReadUint16())
		nextPos += buf.Pos()

		itemID := stringValue(buf.ReadS())
		atlasID := stringValue(buf.ReadS())
		atlas := pkg.ItemByID(atlasID)

		sprite := &AtlasSprite{Atlas: atlas}
		sprite.Rect = Rect{
			X:      int(buf.ReadInt32()),
			Y:      int(buf.ReadInt32()),
			Width:  int(buf.ReadInt32()),
			Height: int(buf.ReadInt32()),
		}
		sprite.Rotated = buf.ReadBool()
		if pkg.Version >= 2 && buf.ReadBool() {
			sprite.Offset = Point{X: float32(buf.ReadInt32()), Y: float32(buf.ReadInt32())}
			sprite.OriginalSize = Point{X: float32(buf.ReadInt32()), Y: float32(buf.ReadInt32())}
		} else {
			sprite.OriginalSize = Point{X: float32(sprite.Rect.Width), Y: float32(sprite.Rect.Height)}
		}

		pkg.Sprites[itemID] = sprite
		if item := pkg.ItemByID(itemID); item != nil {
			item.Atlas = atlas
			item.Sprite = sprite
		}

		_ = buf.SetPos(nextPos)
	}
	return nil
}

func parsePixelHitTests(buf *utils.ByteBuffer, pkg *Package) error {
	if !buf.Seek(pkg.indexTablePos, 3) {
		return nil
	}

	count := int(buf.ReadUint16())
	for i := 0; i < count; i++ {
		nextPos := int(buf.ReadInt32())
		nextPos += buf.Pos()

		item := pkg.ItemByID(stringValue(buf.ReadS()))
		if item != nil && item.Type == PackageItemTypeImage {
			data := &PixelHitTestData{}
			data.Load(buf)
			item.PixelHitTest = data
		}
		_ = buf.SetPos(nextPos)
	}
	return nil
}

func decompressRaw(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	fr := flate.NewReader(reader)
	defer fr.Close()
	return io.ReadAll(fr)
}

func parseMovieClipData(item *PackageItem) {
	if item == nil || item.RawData == nil {
		return
	}
	buf := item.RawData
	saved := buf.Pos()
	defer func() { _ = buf.SetPos(saved) }()

	if !buf.Seek(0, 0) {
		return
	}
	item.Interval = int(buf.ReadInt32())
	item.Swing = buf.ReadBool()
	item.RepeatDelay = int(buf.ReadInt32())

	if !buf.Seek(0, 1) {
		return
	}
	frameCount := int(buf.ReadInt16())
	if frameCount <= 0 {
		item.Frames = nil
		return
	}

	frames := make([]*MovieClipFrame, 0, frameCount)
	for i := 0; i < frameCount; i++ {
		nextPos := int(buf.ReadInt16()) + buf.Pos()

		offsetX := int(buf.ReadInt32())
		offsetY := int(buf.ReadInt32())
		width := int(buf.ReadInt32())
		height := int(buf.ReadInt32())
		addDelay := int(buf.ReadInt32())
		spriteID := stringValue(buf.ReadS())

		frame := &MovieClipFrame{
			SpriteID: spriteID,
			AddDelay: addDelay,
			OffsetX:  offsetX,
			OffsetY:  offsetY,
			Width:    width,
			Height:   height,
		}
		if spriteID != "" && item.Owner != nil {
			if sprite := item.Owner.Sprites[spriteID]; sprite != nil {
				frame.Sprite = sprite
			}
		}
		frames = append(frames, frame)

		_ = buf.SetPos(nextPos)
	}
	item.Frames = frames
}

func linkMovieClipSprites(pkg *Package) {
	if pkg == nil {
		return
	}
	for _, item := range pkg.Items {
		if item == nil || item.Type != PackageItemTypeMovieClip || len(item.Frames) == 0 {
			continue
		}
		for _, frame := range item.Frames {
			if frame == nil || frame.Sprite != nil || frame.SpriteID == "" {
				continue
			}
			if sprite := pkg.Sprites[frame.SpriteID]; sprite != nil {
				frame.Sprite = sprite
			}
		}
	}
}

func parseComponentData(item *PackageItem) {
	if item == nil || item.RawData == nil {
		return
	}
	buf := item.RawData
	if buf == nil {
		return
	}
	saved := buf.Pos()
	if !buf.Seek(0, 0) {
		_ = buf.SetPos(saved)
		return
	}

	cd := &ComponentData{}
	cd.SourceWidth = int(buf.ReadInt32())
	cd.SourceHeight = int(buf.ReadInt32())
	cd.InitWidth = cd.SourceWidth
	cd.InitHeight = cd.SourceHeight

	if buf.ReadBool() {
		cd.MinWidth = int(buf.ReadInt32())
		cd.MaxWidth = int(buf.ReadInt32())
		cd.MinHeight = int(buf.ReadInt32())
		cd.MaxHeight = int(buf.ReadInt32())
	}

	if buf.ReadBool() {
		cd.PivotX = buf.ReadFloat32()
		cd.PivotY = buf.ReadFloat32()
		cd.PivotAnchor = buf.ReadBool()
	}

	if buf.ReadBool() {
		cd.Margin.Top = int(buf.ReadInt32())
		cd.Margin.Bottom = int(buf.ReadInt32())
		cd.Margin.Left = int(buf.ReadInt32())
		cd.Margin.Right = int(buf.ReadInt32())
	}

	cd.Overflow = OverflowType(buf.ReadUint8())

	if buf.ReadBool() {
		_ = buf.Skip(8)
	}

	if buf.Seek(0, 1) {
		controllerCount := int(buf.ReadInt16())
		if controllerCount > 0 {
			controllers := make([]ControllerData, 0, controllerCount)
			for i := 0; i < controllerCount; i++ {
				nextPos := int(buf.ReadInt16()) + buf.Pos()

				begin := buf.Pos()
				buf.Seek(begin, 0)
				ctrl := ControllerData{}
				ctrl.Name = readSValue(buf)
				ctrl.AutoRadio = buf.ReadBool()

				buf.Seek(begin, 1)
				pageCount := int(buf.ReadInt16())
				if pageCount > 0 {
					ctrl.PageIDs = make([]string, pageCount)
					ctrl.PageNames = make([]string, pageCount)
					for j := 0; j < pageCount; j++ {
						ctrl.PageIDs[j] = readSValue(buf)
						ctrl.PageNames[j] = readSValue(buf)
					}
				}

				buf.Seek(begin, 2)
				actionCount := int(buf.ReadInt16())
				for j := 0; j < actionCount; j++ {
					actionNext := int(buf.ReadInt16()) + buf.Pos()
					_ = buf.SetPos(actionNext)
				}

				controllers = append(controllers, ctrl)
				_ = buf.SetPos(nextPos)
			}
			cd.Controllers = controllers
		}
	}

	if buf.Seek(0, 2) {
		childCount := int(buf.ReadInt16())
		children := make([]ComponentChild, 0, childCount)
		for i := 0; i < childCount; i++ {
			dataLen := int(buf.ReadInt16())
			curPos := buf.Pos()
			child := parseComponentChild(buf, curPos, dataLen)
			children = append(children, child)
			_ = buf.SetPos(curPos + dataLen)
		}
		cd.Children = children
	}

	item.Component = cd
	_ = buf.SetPos(saved)
}

func parseComponentChild(buf *utils.ByteBuffer, start int, length int) ComponentChild {
	child := ComponentChild{
		Width:     -1,
		Height:    -1,
		ScaleX:    1,
		ScaleY:    1,
		Visible:   true,
		Touchable: true,
		Alpha:     1,
	}
	child.RawDataOffset = start
	child.RawDataLength = length
	limit := start + length
	remaining := func() int {
		if limit <= buf.Pos() {
			return 0
		}
		return limit - buf.Pos()
	}
	readBool := func() bool {
		if remaining() <= 0 {
			return false
		}
		return buf.ReadBool()
	}

	if buf.Seek(start, 0) {
		child.Type = ObjectType(buf.ReadByte())
		child.Src = readSValue(buf)
		child.PackageID = readSValue(buf)
		child.ID = readSValue(buf)
		child.Name = readSValue(buf)
		child.X = int(buf.ReadInt32())
		child.Y = int(buf.ReadInt32())

		if readBool() {
			child.Width = int(buf.ReadInt32())
			child.Height = int(buf.ReadInt32())
		}
		if readBool() {
			child.MinWidth = int(buf.ReadInt32())
			child.MaxWidth = int(buf.ReadInt32())
			child.MinHeight = int(buf.ReadInt32())
			child.MaxHeight = int(buf.ReadInt32())
		}
		if readBool() {
			child.ScaleX = buf.ReadFloat32()
			child.ScaleY = buf.ReadFloat32()
		}
		if readBool() {
			child.SkewX = buf.ReadFloat32()
			child.SkewY = buf.ReadFloat32()
		}
		if readBool() {
			child.PivotX = buf.ReadFloat32()
			child.PivotY = buf.ReadFloat32()
			child.PivotAnchor = buf.ReadBool()
		}
		child.Alpha = buf.ReadFloat32()
		if child.Alpha == 0 {
			child.Alpha = 1
		}
		child.Rotation = buf.ReadFloat32()
		if remaining() > 0 {
			child.Visible = buf.ReadBool()
		}
		if remaining() > 0 {
			child.Touchable = buf.ReadBool()
		}
		if remaining() > 0 {
			child.Grayed = buf.ReadBool()
		}
		if remaining() > 0 {
			child.BlendMode = int(buf.ReadByte())
		}
		if remaining() > 0 {
			switch filter := buf.ReadByte(); filter {
			case 1:
				_ = buf.ReadFloat32()
				_ = buf.ReadFloat32()
				_ = buf.ReadFloat32()
				_ = buf.ReadFloat32()
			case 2:
				_ = buf.ReadFloat32()
			}
		}
		child.Data = readSValue(buf)
		if child.Type == ObjectTypeText || child.Type == ObjectTypeRichText {
			child.Text = readSValue(buf)
		}
	}

	return child
}

func readSValue(buf *utils.ByteBuffer) string {
	idx := buf.ReadUint16()
	switch idx {
	case 0xFFFE, 0xFFFD:
		return ""
	default:
		if int(idx) >= 0 && int(idx) < len(buf.StringTable) {
			return buf.StringTable[idx]
		}
		return ""
	}
}

func stringFromIndex(idx uint16, table []string) string {
	if idx == 0xFFFE || idx == 0xFFFD {
		return ""
	}
	i := int(idx)
	if i >= 0 && i < len(table) {
		return table[i]
	}
	return ""
}
