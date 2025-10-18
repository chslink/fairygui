# FairyGUI Ebiten Port Progress

## 2025-10-18
- [x] Audited `laya_src/fairygui` TypeScript modules and catalogued LayaAir dependencies.
- [x] Authored `docs/architecture.md` describing Go/Ebiten layering, package mapping, and migration phases.
- [x] Bootstrapped Go scaffolding (`internal/compat/laya`, `pkg/fgui/core`) with geometry, events, scheduler, and base `GObject`/`GComponent` containers.
- [x] Expanded the compat sprite with transform state, affine matrix math, global bounds, and introduced a stage abstraction with scheduler/input integration.
- [x] Added foundational unit tests covering sprite coordinate transforms, stage mouse routing, and `GObject` size/position propagation.
- [x] Ported `fgui.utils.ByteBuffer` to Go with full string-table, colour, sub-buffer, and seek behaviours plus unit coverage.

### Upcoming Focus
- Extend the stage/input layer to support hit testing and event bubbling.
- Build reusable test utilities for compat timers/events and component hierarchies.
- Begin wiring asset/package parsing using the new byte buffer as foundation.
