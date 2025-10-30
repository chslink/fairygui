# FUI 资源解析测试说明

本目录包含了完整的 FairyGUI 资源包解析验证测试套件，通过对比原始 XML 文件和打包后的 .fui 文件来确保解析的正确性。

## 测试文件

### xml_validator_test.go
包级别和组件级别的基础验证测试。

**测试内容：**
- `TestComparePackageXMLWithFUI` - 验证包元数据（ID、名称、资源列表）
- `TestCompareComponentXMLWithFUI` - 验证组件基本属性（尺寸、子元素数量、控制器）

**覆盖的包：**
- Bag.fui
- Basics.fui

### component_validation_test.go
针对各种组件类型的详细验证测试。

**测试内容：**

1. **TestBasicsComponents** - 验证常用组件的特定属性
   - Button - 控制器、按钮状态
   - ProgressBar - bar 元素存在性
   - Slider - grip 和 bar 元素
   - ComboBox - 子元素结构
   - Checkbox - 控制器配置
   - ComboBoxPopup - list 元素、滚动条配置、默认列表项
   - GridItem - graph 元素、齿轮系统、控制器配置
   - WindowFrame - Label 扩展、窗口元素（closeButton、dragArea、contentArea、title）
   - Dropdown2 - list 元素、下拉列表配置

2. **TestDemoScenes** - 验证各种演示场景
   - Demo_Button, Demo_Text, Demo_Image
   - Demo_Loader, Demo_ProgressBar, Demo_Slider
   - Demo_ComboBox, Demo_List
   - Demo_Controller, Demo_Relation, Demo_MovieClip

3. **TestComponentRelations** - 验证关系系统
   - 检测元素间的关系定义
   - 验证 relation target 和 sidePair

4. **TestComponentGears** - 验证齿轮系统（简单版本）
   - 检测 gearDisplay 配置
   - 验证控制器关联

**注意**：TestComponentGears 是简单版本的齿轮测试，主要用于向后兼容。更详细的控制器和齿轮测试请参见 TestControllerSystem 和 TestGearSystem。

5. **TestGraphComponents** - 验证 Graph 元素的详细属性
   - Graph 类型分布统计（rect, eclipse, polygon, regular_polygon）
   - 矩形 Graph - lineSize, fillColor, corner 等属性
   - 椭圆 Graph - lineSize, fillColor 等属性
   - 多边形 Graph - points 自定义点坐标
   - 正多边形 Graph - sides 边数、startAngle 起始角度、distances 距离

6. **TestListComponents** - 验证 List 元素的详细属性
   - List 布局模式（column, row, flow_hz, flow_vt）
   - overflow 和 scroll 配置
   - lineGap 和 colGap 间距配置
   - defaultItem 默认列表项
   - clipSoftness 裁剪柔和度
   - 列表项（item）数据验证

7. **TestLoaderComponents** - 验证 Loader 元素的详细属性
   - URL 资源配置验证
   - 填充模式（scale, scaleMatchHeight, scaleMatchWidth）
   - 缩放比例（scale）配置
   - 对齐方式（align, vAlign）配置
   - 统计各类属性的使用分布

8. **TestLabelExtension** - 验证 Label 扩展组件
   - WindowFrame 作为 Label 扩展的验证
   - 窗口必需元素检查（closeButton、dragArea、contentArea、title）
   - 统计 Basics 包中所有使用 Label 扩展的组件

9. **TestControllerSystem** - 验证控制器系统
   - 测试 Demo_Controller 场景（包含多个控制器）
   - 验证控制器的页面定义和默认选中
   - 检查 XML 与 FUI 中控制器的一致性
   - 支持多控制器复杂状态管理

10. **TestGearSystem** - 验证齿轮系统的各种类型
    - 统计所有 gear 类型的使用情况
    - **gearDisplay** - 显示/隐藏控制
    - **gearDisplay2** - 带条件的显示控制
    - **gearXY** - 位置控制（支持 tween）
    - **gearSize** - 尺寸控制（支持 tween）
    - **gearLook** - 外观控制（alpha、rotation，支持 tween）
    - **gearColor** - 颜色控制
    - **gearAni** - 动画控制（MovieClip）
    - **gearFontSize** - 字体大小控制
    - 详细测试 gearXY、gearColor、gearLook 的属性
    - 验证 tween 动画支持

11. **TestMainMenuScene** - 验证 MainMenu 场景的完整性
    - 测试真实场景的组件属性
    - 验证场景基本信息（尺寸：1136x640，15个按钮组件）
    - **ElementTypes** - 元素类型统计（1个 graph 背景，15个按钮）
    - **Background** - 背景属性验证（矩形、颜色、relation 关系）
    - **Buttons** - 按钮布局验证（3列分布：6+6+3）
    - **ButtonComponent** - Button 组件详细验证
      - 控制器：button（4个页面：up, down, over, selectedOver）
      - 3个图片使用 gearDisplay 控制不同状态显示
      - 标题文本居中对齐
      - 所有元素配置 relation 关系
    - **PackageIntegrity** - 包完整性验证（3个组件，5个图片，1个图集）

## 运行测试

### 运行所有验证测试
```bash
cd pkg/fgui/assets
go test -v -run "XML|Component|Demo|Relations|Gears|Graph|List|Loader|Label|Controller|MainMenu"
```

### 运行特定测试
```bash
# 包级别验证
go test -v -run TestComparePackageXMLWithFUI

# 组件验证
go test -v -run TestBasicsComponents

# 演示场景验证
go test -v -run TestDemoScenes

# 关系和齿轮系统（简单版本）
go test -v -run "TestComponentRelations|TestComponentGears"

# 控制器系统（详细测试）
go test -v -run TestControllerSystem

# 齿轮系统（详细测试）
go test -v -run TestGearSystem

# Graph 元素验证
go test -v -run TestGraphComponents

# List 元素验证
go test -v -run TestListComponents

# Loader 元素验证
go test -v -run TestLoaderComponents

# Label 扩展验证
go test -v -run TestLabelExtension

# MainMenu 场景验证（真实场景完整性测试）
go test -v -run TestMainMenuScene
```

### 快速检查
```bash
# 无输出运行（快速）
go test -run "XML|Component"
```

## 测试数据来源

### 原始 XML 文件
- **位置**: `demo/UIProject/assets/`
- **目录结构**:
  - `Bag/` - Bag 包的组件 XML
  - `Basics/` - Basics 包的组件 XML
    - `components/` - 基础组件（Button, Slider, ProgressBar 等）
    - `Demo_*.xml` - 演示场景

### 打包 FUI 文件
- **位置**: `demo/assets/`
- **文件**: `Bag.fui`, `Basics.fui`

## 验证逻辑

### 包验证
1. 解析 `package.xml` 获取资源列表
2. 解析 `.fui` 文件获取打包数据
3. 对比：
   - 包 ID 和名称
   - 资源类型（Component, Image, MovieClip, Font, Sound）
   - 资源名称（注意：FUI 中组件名不含 .xml 后缀）
   - 图片九宫格配置

### 组件验证
1. 解析组件 XML 文件（如 `Button.xml`）
2. 从 FUI 包中查找对应组件
3. 对比：
   - 组件尺寸（SourceWidth, SourceHeight）
   - 扩展类型（extension 属性）
   - 控制器定义（名称、页面配置）
   - 子元素数量和类型
   - 关系定义（relation）
   - 齿轮配置（gearDisplay, gearColor 等）

## 已知差异

### 未使用资源优化
FairyGUI 导出时会删除未使用的资源。例如：
- `Bag/2.png` (ID=jlmga) - XML 中定义但未打包
- `Basics/k27.png` (ID=wa8u2q) - XML 中定义但未打包

测试会对此发出警告而非错误。

### 子元素数量差异
某些场景的子元素数量可能不匹配，原因包括：
- XML 中可能有隐藏或特殊元素未计入
- FairyGUI 导出时可能展开或合并某些元素
- 测试的 XML 解析可能未涵盖所有元素类型

这些差异会记录为警告，不会导致测试失败。

## 扩展测试

### 添加新组件测试
在 `TestBasicsComponents` 的 `testCases` 中添加：

```go
{
    name:          "YourComponent",
    xmlPath:       "components/YourComponent.xml",
    componentName: "YourComponent",
    extension:     "ComponentType",
    validate: func(t *testing.T, xmlComp *XMLComponent, fuiItem *PackageItem) {
        // 自定义验证逻辑
    },
},
```

### 添加新包测试
在 `TestComparePackageXMLWithFUI` 的 `testCases` 中添加：

```go
{
    name:    "YourPackage",
    xmlDir:  filepath.Join("..", "..", "..", "demo", "UIProject", "assets", "YourPackage"),
    fuiPath: filepath.Join("..", "..", "..", "demo", "assets", "YourPackage.fui"),
},
```

### 扩展 XML 结构
如需验证更多属性，扩展 `XMLComponent` 结构：

```go
type XMLComponent struct {
    // ... 现有字段
    Transitions []XMLTransition `xml:"transition"`
    CustomData  string          `xml:"customData,attr"`
}
```

## 测试结果解读

### ✓ 通过标记
```
✓ 找到控制器: button
✓ Slider 包含 grip 和 bar 元素
✓ 扩展类型匹配: Button
```

### 警告信息
```
警告：图片 2.png (ID=jlmga) 在 XML 中定义但未打包到 FUI（可能未被使用）
警告：子元素数量不匹配
```

### 错误信息
```
未找到组件: XXX
组件类型错误: 期望 Component, 实际 Image
```

## 维护建议

1. **新增组件时**：同时添加对应的验证测试用例
2. **更新 FUI 格式时**：检查所有测试是否仍然通过
3. **发现解析 bug 时**：先添加失败的测试用例，然后修复
4. **定期运行**：在 CI/CD 中集成这些测试

## 相关文档

- `docs/architecture.md` - 系统架构说明
- `docs/refactor-progress.md` - 迁移进度跟踪
- `CLAUDE.md` - 项目开发指南
