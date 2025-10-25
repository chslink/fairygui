package assets

import (
	"strings"
	"sync"
)

// packageRegistry 是全局包注册表,用于通过 URL 查找资源
var packageRegistry = struct {
	sync.RWMutex
	byID   map[string]*Package // key: package ID (8 chars)
	byName map[string]*Package // key: package name
}{
	byID:   make(map[string]*Package),
	byName: make(map[string]*Package),
}

// RegisterPackage 注册一个包到全局注册表
// 注册后可以通过 GetItemByURL 查找包中的资源
func RegisterPackage(pkg *Package) {
	if pkg == nil {
		return
	}
	packageRegistry.Lock()
	defer packageRegistry.Unlock()

	if pkg.ID != "" {
		packageRegistry.byID[pkg.ID] = pkg
	}
	if pkg.Name != "" {
		packageRegistry.byName[strings.ToLower(pkg.Name)] = pkg
	}
}

// UnregisterPackage 从全局注册表移除一个包
func UnregisterPackage(pkg *Package) {
	if pkg == nil {
		return
	}
	packageRegistry.Lock()
	defer packageRegistry.Unlock()

	if pkg.ID != "" {
		delete(packageRegistry.byID, pkg.ID)
	}
	if pkg.Name != "" {
		delete(packageRegistry.byName, strings.ToLower(pkg.Name))
	}
}

// GetPackageByID 通过包 ID 获取包
func GetPackageByID(id string) *Package {
	if id == "" {
		return nil
	}
	packageRegistry.RLock()
	defer packageRegistry.RUnlock()
	return packageRegistry.byID[id]
}

// GetPackageByName 通过包名获取包
func GetPackageByName(name string) *Package {
	if name == "" {
		return nil
	}
	packageRegistry.RLock()
	defer packageRegistry.RUnlock()
	return packageRegistry.byName[strings.ToLower(name)]
}

// GetItemByURL 通过 URL 获取资源项
// 支持两种格式:
//   - ui://packageId+itemId (例如: ui://9leh0eyf6pmb6, 8+8字符)
//   - ui://packageName/itemName (例如: ui://Basics/button)
//
// 参考 LayaAir UIPackage.ts:210-237
func GetItemByURL(url string) *PackageItem {
	if url == "" {
		return nil
	}

	// 查找 "//"
	pos1 := strings.Index(url, "//")
	if pos1 == -1 {
		return nil
	}

	// 查找第三个 "/"
	pos2 := strings.Index(url[pos1+2:], "/")
	if pos2 == -1 {
		// 格式: ui://packageId+itemId
		// 最小长度: "ui://" (5) + packageId (8) = 13
		if len(url) > 13 {
			pkgID := url[5:13]              // packageId (8 chars)
			itemID := url[13:]              // itemId (remaining)

			pkg := GetPackageByID(pkgID)
			if pkg != nil {
				return pkg.ItemByID(itemID)
			}
		}
	} else {
		// 格式: ui://packageName/itemName
		pos2 += pos1 + 2 // 转换为相对于 url 的位置
		pkgName := url[pos1+2 : pos2]
		itemName := url[pos2+1:]

		pkg := GetPackageByName(pkgName)
		if pkg != nil {
			return pkg.ItemByName(itemName)
		}
	}

	return nil
}

// GetItemURL 根据包名和资源名生成 URL
// 参考 LayaAir UIPackage.ts:198-208
func GetItemURL(pkgName, resName string) string {
	pkg := GetPackageByName(pkgName)
	if pkg == nil {
		return ""
	}

	item := pkg.ItemByName(resName)
	if item == nil {
		return ""
	}

	return "ui://" + pkg.ID + item.ID
}
