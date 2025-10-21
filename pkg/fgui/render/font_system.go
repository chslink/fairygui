package render

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const systemFontEnv = "FGUI_FONT_PATH"

var (
	systemFontData         []byte
	systemFontIsCollection bool
	systemFontIndex        int
	systemFontCache        = make(map[int]font.Face)
	systemFontMu           sync.RWMutex
)

// LoadSystemFont 尝试按操作系统默认位置加载本地字体。
// size 以逻辑像素为单位，若提供非正值则回退到 16。
// 返回值包括字体句柄和实际使用的路径。
func LoadSystemFont(size float64) (font.Face, string, error) {
	if size <= 0 {
		size = 16
	}
	requestedSize := int(math.Round(size))

	for _, candidate := range enumerateFontCandidates() {
		face, src, err := openFontFace(candidate, size)
		if err == nil {
			systemFontMu.Lock()
			systemFontData = src.data
			systemFontIsCollection = src.isCollection
			systemFontIndex = src.index
			systemFontCache = make(map[int]font.Face)
			systemFontCache[requestedSize] = face
			systemFontMu.Unlock()
			return face, candidate, nil
		}
	}

	return nil, "", errors.New("render: 未找到可用的系统字体")
}

func enumerateFontCandidates() []string {
	var paths []string

	if env := os.Getenv(systemFontEnv); env != "" {
		for _, p := range strings.Split(env, string(os.PathListSeparator)) {
			if p = strings.TrimSpace(p); p != "" {
				paths = append(paths, p)
			}
		}
	}

	switch runtime.GOOS {
	case "windows":
		base := os.Getenv("WINDIR")
		if base == "" {
			base = `C:\Windows`
		}
		fontDir := filepath.Join(base, "Fonts")
		names := []string{
			"simsun.ttf",
			"simsun.ttc",
			"msyh.ttc",
			"msyh.ttf",
			"simsunb.ttf",
			"Microsoft YaHei.ttf",
			"simhei.ttf",
			"Deng.ttf",
		}
		for _, name := range names {
			paths = append(paths, filepath.Join(fontDir, name))
		}
	case "darwin":
		paths = append(paths,
			"/System/Library/Fonts/PingFang.ttc",
			"/System/Library/Fonts/Songti.ttc",
			"/System/Library/Fonts/STHeiti Light.ttc",
			"/Library/Fonts/Arial Unicode.ttf",
		)
	default: // Linux / others
		paths = append(paths,
			"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc",
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		)
		if home, err := os.UserHomeDir(); err == nil {
			paths = append(paths,
				filepath.Join(home, ".local", "share", "fonts", "NotoSansCJK-Regular.ttc"),
				filepath.Join(home, "Library", "Fonts", "PingFang.ttc"),
			)
		}
	}

	return paths
}

type fontSource struct {
	data         []byte
	isCollection bool
	index        int
}

func openFontFace(path string, size float64) (font.Face, fontSource, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fontSource{}, err
	}
	ext := strings.ToLower(filepath.Ext(path))
	opts := &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	}

	switch ext {
	case ".ttc", ".otc":
		col, err := opentype.ParseCollection(data)
		if err != nil {
			return nil, fontSource{}, err
		}
		count := col.NumFonts()
		for i := 0; i < count; i++ {
			fnt, err := col.Font(i)
			if err != nil {
				continue
			}
			face, err := opentype.NewFace(fnt, opts)
			if err == nil {
				return face, fontSource{data: data, isCollection: true, index: i}, nil
			}
		}
		return nil, fontSource{}, fmt.Errorf("render: 无法从 TTC 解析字体 %s", path)
	default:
		fnt, err := opentype.Parse(data)
		if err != nil {
			return nil, fontSource{}, err
		}
		face, err := opentype.NewFace(fnt, opts)
		if err != nil {
			return nil, fontSource{}, err
		}
		return face, fontSource{data: data}, nil
	}
}

func getFontFace(size int) (font.Face, error) {
	systemFontMu.RLock()
	if face, ok := systemFontCache[size]; ok {
		systemFontMu.RUnlock()
		return face, nil
	}
	srcData := systemFontData
	isCollection := systemFontIsCollection
	index := systemFontIndex
	systemFontMu.RUnlock()

	if len(srcData) == 0 {
		return nil, errors.New("render: system font not loaded")
	}

	opts := &opentype.FaceOptions{
		Size:    float64(size),
		DPI:     72,
		Hinting: font.HintingFull,
	}
	var face font.Face
	var err error
	if isCollection {
		col, parseErr := opentype.ParseCollection(srcData)
		if parseErr != nil {
			return nil, parseErr
		}
		if index < 0 || index >= col.NumFonts() {
			index = 0
		}
		fnt, fontErr := col.Font(index)
		if fontErr != nil {
			return nil, fontErr
		}
		face, err = opentype.NewFace(fnt, opts)
	} else {
		fnt, parseErr := opentype.Parse(srcData)
		if parseErr != nil {
			return nil, parseErr
		}
		face, err = opentype.NewFace(fnt, opts)
	}
	if err != nil {
		return nil, err
	}

	systemFontMu.Lock()
	systemFontCache[size] = face
	systemFontMu.Unlock()
	return face, nil
}
