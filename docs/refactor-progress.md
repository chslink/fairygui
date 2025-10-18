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
- [x] Bootstrapped Go asset pipeline scaffolding: resource loader abstraction, package header parsing, ByteBuffer enhancements, and initial package item metadata parsing/tests.

### Upcoming Focus
- Expand `ParsePackage` to read remaining package tables (sprites, pixel tests) and wire atlas metadata into texture loaders.
- Implement concrete loaders (filesystem/embedded) and integrate with Ebiten-friendly texture creation.
- Connect pointer events to higher-level UI abstractions (GRoot, drag/drop) leveraging the new compat stage.
