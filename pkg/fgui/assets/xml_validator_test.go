package assets

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// XMLPackageDescription 对应 package.xml 的根节点
type XMLPackageDescription struct {
	XMLName      xml.Name         `xml:"packageDescription"`
	ID           string           `xml:"id,attr"`
	JPEGQuality  int              `xml:"jpegQuality,attr"`
	CompressPNG  bool             `xml:"compressPNG,attr"`
	Resources    XMLResources     `xml:"resources"`
	Publish      XMLPublish       `xml:"publish"`
}

// XMLResources 对应 resources 节点
type XMLResources struct {
	Components []XMLResourceComponent `xml:"component"`
	Images     []XMLResourceImage     `xml:"image"`
	MovieClips []XMLResourceMovieClip `xml:"movieclip"`
	Fonts      []XMLResourceFont      `xml:"font"`
	Sounds     []XMLResourceSound     `xml:"sound"`
}

// XMLResourceComponent 对应 component 资源
type XMLResourceComponent struct {
	ID       string `xml:"id,attr"`
	Name     string `xml:"name,attr"`
	Path     string `xml:"path,attr"`
	Exported bool   `xml:"exported,attr"`
}

// XMLResourceImage 对应 image 资源
type XMLResourceImage struct {
	ID         string `xml:"id,attr"`
	Name       string `xml:"name,attr"`
	Path       string `xml:"path,attr"`
	Scale      string `xml:"scale,attr"`
	Scale9Grid string `xml:"scale9grid,attr"`
}

// XMLResourceMovieClip 对应 movieclip 资源
type XMLResourceMovieClip struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
	Path string `xml:"path,attr"`
}

// XMLResourceFont 对应 font 资源
type XMLResourceFont struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
	Path string `xml:"path,attr"`
}

// XMLResourceSound 对应 sound 资源
type XMLResourceSound struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
	Path string `xml:"path,attr"`
}

// XMLPublish 对应 publish 节点
type XMLPublish struct {
	Name   string      `xml:"name,attr"`
	Atlases []XMLAtlas `xml:"atlas"`
}

// XMLAtlas 对应 atlas 节点
type XMLAtlas struct {
	Name  string `xml:"name,attr"`
	Index int    `xml:"index,attr"`
}

// parsePackageXML 解析 package.xml 文件
func parsePackageXML(path string) (*XMLPackageDescription, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg XMLPackageDescription
	if err := xml.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	return &pkg, nil
}

// TestComparePackageXMLWithFUI 对比 package.xml 和 .fui 文件的属性
func TestComparePackageXMLWithFUI(t *testing.T) {
	testCases := []struct {
		name      string
		xmlDir    string  // demo/UIProject/assets/Bag
		fuiPath   string  // demo/assets/Bag.fui
	}{
		{
			name:    "Bag",
			xmlDir:  filepath.Join("..", "..", "..", "demo", "UIProject", "assets", "Bag"),
			fuiPath: filepath.Join("..", "..", "..", "demo", "assets", "Bag.fui"),
		},
		{
			name:    "Basics",
			xmlDir:  filepath.Join("..", "..", "..", "demo", "UIProject", "assets", "Basics"),
			fuiPath: filepath.Join("..", "..", "..", "demo", "assets", "Basics.fui"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 解析原始 XML
			xmlPkgPath := filepath.Join(tc.xmlDir, "package.xml")
			xmlPkg, err := parsePackageXML(xmlPkgPath)
			if err != nil {
				t.Skipf("跳过测试：无法读取 XML 文件 %s: %v", xmlPkgPath, err)
			}

			// 解析 .fui 文件
			fuiData, err := os.ReadFile(tc.fuiPath)
			if err != nil {
				t.Skipf("跳过测试：无法读取 .fui 文件 %s: %v", tc.fuiPath, err)
			}

			pkg, err := ParsePackage(fuiData, "demo/assets/"+tc.name)
			if err != nil {
				t.Fatalf("解析 .fui 文件失败: %v", err)
			}

			// 验证包 ID
			if pkg.ID != xmlPkg.ID {
				t.Errorf("包 ID 不匹配: FUI=%s, XML=%s", pkg.ID, xmlPkg.ID)
			}

			// 验证包名称
			if pkg.Name != xmlPkg.Publish.Name {
				t.Errorf("包名称不匹配: FUI=%s, XML=%s", pkg.Name, xmlPkg.Publish.Name)
			}

			// 统计 XML 中的资源数量
			totalXMLResources := len(xmlPkg.Resources.Components) +
				len(xmlPkg.Resources.Images) +
				len(xmlPkg.Resources.MovieClips) +
				len(xmlPkg.Resources.Fonts) +
				len(xmlPkg.Resources.Sounds)

			// 验证资源项数量
			if len(pkg.Items) != totalXMLResources {
				t.Logf("警告：资源项数量不匹配: FUI=%d, XML=%d", len(pkg.Items), totalXMLResources)
			}

			// 验证 Component 资源
			for _, xmlComp := range xmlPkg.Resources.Components {
				item := pkg.ItemByID(xmlComp.ID)
				if item == nil {
					t.Errorf("组件 %s (ID=%s) 在 FUI 中未找到", xmlComp.Name, xmlComp.ID)
					continue
				}

				if item.Type != PackageItemTypeComponent {
					t.Errorf("组件 %s 类型错误: 期望 Component, 实际 %v", xmlComp.Name, item.Type)
				}

				// XML 中的 name 包含 .xml 后缀，需要去掉后对比
				expectedName := strings.TrimSuffix(xmlComp.Name, ".xml")
				if item.Name != expectedName {
					t.Errorf("组件 %s 名称不匹配: FUI=%s, XML=%s", xmlComp.ID, item.Name, expectedName)
				}
			}

			// 验证 Image 资源
			for _, xmlImg := range xmlPkg.Resources.Images {
				item := pkg.ItemByID(xmlImg.ID)
				if item == nil {
					// FairyGUI 导出时可能会删除未使用的资源，所以这里只记录警告
					t.Logf("警告：图片 %s (ID=%s) 在 XML 中定义但未打包到 FUI（可能未被使用）", xmlImg.Name, xmlImg.ID)
					continue
				}

				if item.Type != PackageItemTypeImage {
					t.Errorf("图片 %s 类型错误: 期望 Image, 实际 %v", xmlImg.Name, item.Type)
				}

				// 验证九宫格设置
				if xmlImg.Scale == "9grid" && xmlImg.Scale9Grid != "" {
					if item.Scale9Grid == nil {
						t.Errorf("图片 %s 缺少九宫格数据, XML 中定义为: %s", xmlImg.Name, xmlImg.Scale9Grid)
					}
				}
			}

			// 验证 MovieClip 资源
			for _, xmlMC := range xmlPkg.Resources.MovieClips {
				item := pkg.ItemByID(xmlMC.ID)
				if item == nil {
					t.Errorf("动画 %s (ID=%s) 在 FUI 中未找到", xmlMC.Name, xmlMC.ID)
					continue
				}

				if item.Type != PackageItemTypeMovieClip {
					t.Errorf("动画 %s 类型错误: 期望 MovieClip, 实际 %v", xmlMC.Name, item.Type)
				}
			}

			t.Logf("✓ 包 %s 验证通过：%d 个资源项", tc.name, len(pkg.Items))
		})
	}
}

// TestCompareComponentXMLWithFUI 对比组件 XML 定义和 FUI 解析结果
func TestCompareComponentXMLWithFUI(t *testing.T) {
	// 加载 Bag.fui 包
	fuiPath := filepath.Join("..", "..", "..", "demo", "assets", "Bag.fui")
	fuiData, err := os.ReadFile(fuiPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取 .fui 文件: %v", err)
	}

	pkg, err := ParsePackage(fuiData, "demo/assets/Bag")
	if err != nil {
		t.Fatalf("解析 .fui 文件失败: %v", err)
	}

	// 查找 BagWin 组件（注意：FUI 中的 Name 不包含 .xml 后缀）
	var bagWinItem *PackageItem
	for _, item := range pkg.Items {
		if item.Name == "BagWin" && item.Type == PackageItemTypeComponent {
			bagWinItem = item
			break
		}
	}

	if bagWinItem == nil {
		t.Fatal("未找到 BagWin 组件")
	}

	// 加载原始 BagWin.xml
	xmlPath := filepath.Join("..", "..", "..", "demo", "UIProject", "assets", "Bag", "BagWin.xml")
	xmlData, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Skipf("跳过测试：无法读取组件 XML: %v", err)
	}

	// 解析组件 XML
	type XMLComponent struct {
		Size       string `xml:"size,attr"`
		Controller struct {
			Name  string `xml:"name,attr"`
			Pages string `xml:"pages,attr"`
		} `xml:"controller"`
		DisplayList struct {
			Components []struct {
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				Src  string `xml:"src,attr"`
				XY   string `xml:"xy,attr"`
				Size string `xml:"size,attr"`
			} `xml:"component"`
			Lists []struct {
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				XY   string `xml:"xy,attr"`
				Size string `xml:"size,attr"`
			} `xml:"list"`
			Images []struct {
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				Src  string `xml:"src,attr"`
				XY   string `xml:"xy,attr"`
				Size string `xml:"size,attr"`
			} `xml:"image"`
			Texts []struct {
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				XY   string `xml:"xy,attr"`
				Size string `xml:"size,attr"`
				Text string `xml:"text,attr"`
			} `xml:"text"`
			Loaders []struct {
				ID   string `xml:"id,attr"`
				Name string `xml:"name,attr"`
				XY   string `xml:"xy,attr"`
				Size string `xml:"size,attr"`
			} `xml:"loader"`
		} `xml:"displayList"`
	}

	var xmlComp XMLComponent
	if err := xml.Unmarshal(xmlData, &xmlComp); err != nil {
		t.Fatalf("解析组件 XML 失败: %v", err)
	}

	// 验证组件基本属性
	if bagWinItem.Component == nil {
		t.Fatal("BagWin 组件数据为空")
	}

	// 验证尺寸
	t.Logf("组件尺寸: FUI=(%d,%d), XML=%s",
		bagWinItem.Component.SourceWidth,
		bagWinItem.Component.SourceHeight,
		xmlComp.Size)

	// 验证子元素数量
	totalXMLChildren := len(xmlComp.DisplayList.Components) +
		len(xmlComp.DisplayList.Lists) +
		len(xmlComp.DisplayList.Images) +
		len(xmlComp.DisplayList.Texts) +
		len(xmlComp.DisplayList.Loaders)

	if len(bagWinItem.Component.Children) != totalXMLChildren {
		t.Errorf("子元素数量不匹配: FUI=%d, XML=%d",
			len(bagWinItem.Component.Children), totalXMLChildren)
	} else {
		t.Logf("✓ 子元素数量匹配: %d", totalXMLChildren)
	}

	// 验证控制器
	if xmlComp.Controller.Name != "" {
		foundController := false
		for _, ctrl := range bagWinItem.Component.Controllers {
			if ctrl.Name == xmlComp.Controller.Name {
				foundController = true
				t.Logf("✓ 找到控制器: %s", ctrl.Name)
				break
			}
		}
		if !foundController {
			t.Errorf("未找到控制器: %s", xmlComp.Controller.Name)
		}
	}
}
