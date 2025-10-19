# FairyGUI Ebiten Port Progress

## 2025-10-18
- [x] Audited `laya_src/fairygui` TypeScript modules and catalogued LayaAir dependencies.
- [x] Authored `docs/architecture.md` describing Go/Ebiten layering, package mapping, and migration phases.
- [x] Bootstrapped Go scaffolding (`internal/compat/laya`, `pkg/fgui/core`) with geometry, events, scheduler, and base `GObject`/`GComponent` containers.
- [x] Expanded the compat sprite with transform state, affine matrix math, global bounds, and introduced a stage abstraction with scheduler/input integration.
- [x] Added foundational unit tests covering sprite coordinate transforms, stage mouse routing, and `GObject` size/position propagation.
- [x] Ported `fgui.utils.ByteBuffer` to Go with full string-table, colour, sub-buffer, and seek behaviours plus unit coverage.
- [x] Enhanced the stage/input layer with hit testing, pointer bubbling, click synthesis, and regression tests.
- [x] Introduced shared test utilities (stage env, event logs) and expanded coverage for scheduler and `GComponent` behaviours.
- [x] Bootstrapped Go asset pipeline scaffolding: resource loader abstraction, package header parsing, ByteBuffer enhancements, and parsing for package items, atlas sprites, and pixel hit-test metadata with unit tests.
- [x] Added raw DEFLATE decompression support, filesystem loader, verified parsing against demo `.fui` packages, and introduced Ebiten-tagged atlas manager plus pixel hit-test integration hooks.
- [x] Began component instantiation path: parsed component metadata now exposes structured child descriptors, builder scaffolding creates GObject trees from real `.fui` data, widgets bind sprite/text content, and component controllers are parsed and attached for runtime use.

## 2025-10-19
- [x] Enriched `core.GObject` with scale, rotation, and pivot state mirrored to compat sprites, enabling downstream systems to track transforms without poking display objects.
- [x] Applied component metadata transforms (scale, rotation, pivot, alpha) during factory builds so instantiated hierarchies better reflect original FairyGUI layouts.
- [x] Added focused unit tests covering the new geometry plumbing (`pkg/fgui/core`, `pkg/fgui/builder`) using `GOCACHE=$(pwd)/.gocache go test ./pkg/fgui/core ./pkg/fgui/builder`.
- [x] Introduced skew handling, pivot-anchor positioning, and cross-package asset resolution in the builder, alongside regression tests that exercise demo `.fui` dependencies.
- [x] Added compat sprite regression tests validating pivot/倾斜矩阵运算与锚点偏移，涵盖缩放、旋转、移动、尺寸变更场景。
- [x] Builder 现会将按钮、标签、列表等高级控件解析成对应 widget（携带包引用、默认项、图标资源），并在渲染阶段绘制文本及按钮图标。

### Upcoming Focus
- Wire parsed atlas sprites into texture loaders and expose hit-test data to rendering/input layers.
- Expand widget factories beyond image/text/button/loader and honor controller/gear transitions during instantiation.
- Wire atlas sprites and pixel masks into real rendering passes under Ebiten.
- Implement concrete loaders (filesystem/embedded) and integrate with Ebiten-friendly texture creation.
- Connect pointer events to higher-level UI abstractions (GRoot, drag/drop) leveraging the new compat stage.
- Profile pivot-aware transforms against upstream FairyGUI scenes and tune any drift discovered during animation playback.
