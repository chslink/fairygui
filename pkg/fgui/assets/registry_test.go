package assets

import (
	"testing"
)

func TestPackageRegistry(t *testing.T) {
	// 清理注册表
	packageRegistry.Lock()
	packageRegistry.byID = make(map[string]*Package)
	packageRegistry.byName = make(map[string]*Package)
	packageRegistry.Unlock()

	// 创建测试包
	pkg1 := &Package{
		ID:   "9leh0eyf",
		Name: "Basics",
	}
	pkg1.Items = []*PackageItem{
		{ID: "rpmb6", Name: "button", Type: PackageItemTypeImage, Width: 100, Height: 50},
		{ID: "abc12", Name: "icon", Type: PackageItemTypeImage, Width: 32, Height: 32},
	}

	pkg2 := &Package{
		ID:   "test1234",
		Name: "TestPkg",
	}
	pkg2.Items = []*PackageItem{
		{ID: "item001", Name: "testitem", Type: PackageItemTypeImage, Width: 64, Height: 64},
	}

	// 测试注册
	RegisterPackage(pkg1)
	RegisterPackage(pkg2)

	// 测试通过 ID 获取包
	t.Run("GetPackageByID", func(t *testing.T) {
		got := GetPackageByID("9leh0eyf")
		if got != pkg1 {
			t.Errorf("GetPackageByID failed: expected pkg1, got %v", got)
		}

		got = GetPackageByID("test1234")
		if got != pkg2 {
			t.Errorf("GetPackageByID failed: expected pkg2, got %v", got)
		}

		got = GetPackageByID("notexist")
		if got != nil {
			t.Errorf("GetPackageByID should return nil for non-existent ID")
		}
	})

	// 测试通过名称获取包
	t.Run("GetPackageByName", func(t *testing.T) {
		got := GetPackageByName("Basics")
		if got != pkg1 {
			t.Errorf("GetPackageByName failed: expected pkg1, got %v", got)
		}

		got = GetPackageByName("basics") // 应该不区分大小写
		if got != pkg1 {
			t.Errorf("GetPackageByName should be case-insensitive")
		}

		got = GetPackageByName("TestPkg")
		if got != pkg2 {
			t.Errorf("GetPackageByName failed: expected pkg2, got %v", got)
		}

		got = GetPackageByName("NotExist")
		if got != nil {
			t.Errorf("GetPackageByName should return nil for non-existent name")
		}
	})

	// 构建包的 itemsByID 和 itemsByName 索引
	pkg1.itemsByID = make(map[string]*PackageItem)
	pkg1.itemsByName = make(map[string]*PackageItem)
	for _, item := range pkg1.Items {
		item.Owner = pkg1
		pkg1.itemsByID[item.ID] = item
		pkg1.itemsByName[item.Name] = item
	}

	pkg2.itemsByID = make(map[string]*PackageItem)
	pkg2.itemsByName = make(map[string]*PackageItem)
	for _, item := range pkg2.Items {
		item.Owner = pkg2
		pkg2.itemsByID[item.ID] = item
		pkg2.itemsByName[item.Name] = item
	}

	// 测试通过 URL 获取资源
	t.Run("GetItemByURL_PackageIDFormat", func(t *testing.T) {
		// 格式: ui://packageId+itemId
		item := GetItemByURL("ui://9leh0eyfrpmb6")
		if item == nil {
			t.Fatal("GetItemByURL returned nil")
		}
		if item.ID != "rpmb6" {
			t.Errorf("Expected item ID 'rpmb6', got '%s'", item.ID)
		}
		if item.Name != "button" {
			t.Errorf("Expected item name 'button', got '%s'", item.Name)
		}
	})

	t.Run("GetItemByURL_PackageNameFormat", func(t *testing.T) {
		// 格式: ui://packageName/itemName
		item := GetItemByURL("ui://Basics/button")
		if item == nil {
			t.Fatal("GetItemByURL returned nil")
		}
		if item.ID != "rpmb6" {
			t.Errorf("Expected item ID 'rpmb6', got '%s'", item.ID)
		}

		item = GetItemByURL("ui://basics/icon") // 包名应该不区分大小写
		if item == nil {
			t.Fatal("GetItemByURL should be case-insensitive for package name")
		}
		if item.ID != "abc12" {
			t.Errorf("Expected item ID 'abc12', got '%s'", item.ID)
		}

		item = GetItemByURL("ui://TestPkg/testitem")
		if item == nil {
			t.Fatal("GetItemByURL returned nil for TestPkg/testitem")
		}
		if item.ID != "item001" {
			t.Errorf("Expected item ID 'item001', got '%s'", item.ID)
		}
	})

	t.Run("GetItemByURL_Invalid", func(t *testing.T) {
		tests := []string{
			"",
			"invalid",
			"ui://",
			"ui://short",
			"ui://notexist/item",
			"ui://9leh0eyfnotexist",
		}

		for _, url := range tests {
			item := GetItemByURL(url)
			if item != nil {
				t.Errorf("GetItemByURL(%q) should return nil, got %v", url, item)
			}
		}
	})

	// 测试生成 URL
	t.Run("GetItemURL", func(t *testing.T) {
		url := GetItemURL("Basics", "button")
		expected := "ui://9leh0eyfrpmb6"
		if url != expected {
			t.Errorf("GetItemURL failed: expected '%s', got '%s'", expected, url)
		}

		url = GetItemURL("basics", "icon") // 包名应该不区分大小写
		expected = "ui://9leh0eyfabc12"
		if url != expected {
			t.Errorf("GetItemURL should be case-insensitive: expected '%s', got '%s'", expected, url)
		}

		url = GetItemURL("NotExist", "item")
		if url != "" {
			t.Errorf("GetItemURL should return empty string for non-existent package")
		}

		url = GetItemURL("Basics", "notexist")
		if url != "" {
			t.Errorf("GetItemURL should return empty string for non-existent item")
		}
	})

	// 测试注销包
	t.Run("UnregisterPackage", func(t *testing.T) {
		UnregisterPackage(pkg2)

		got := GetPackageByID("test1234")
		if got != nil {
			t.Errorf("UnregisterPackage failed: package should not be found by ID")
		}

		got = GetPackageByName("TestPkg")
		if got != nil {
			t.Errorf("UnregisterPackage failed: package should not be found by name")
		}

		item := GetItemByURL("ui://TestPkg/testitem")
		if item != nil {
			t.Errorf("UnregisterPackage failed: item should not be found after unregister")
		}
	})
}
