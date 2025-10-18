# FairyGUI Ebiten Port Architecture

## Project Context
- Goal: reimplement the LayaAir-based TypeScript FairyGUI runtime (`laya_src/fairygui`) in Go on top of Ebiten while preserving the public `fgui` API.
- Constraints: Ebiten provides a frame-driven game loop, software rendering primitives, and Go concurrency, whereas LayaAir offers a retained UI tree, asset pipeline, and utility classes that the current code depends on.
- Approach: introduce a compatibility layer that mimics the subset of LayaAir services required by FairyGUI, translate the core UI modules to Go, and supply exhaustive unit tests to guard the behaviour.

## Layered Architecture
- **Application** (`cmd/*`, samples, games): owns `ebiten.Game`, drives update/draw, integrates packages produced here.
- **FGUI Runtime** (`pkg/fgui/...`): public Go API equivalent of the TypeScript classes. Packages mirror the original folders (`core`, `display`, `gears`, `tween`, `utils`, `assets`, `controllers`).
- **Compatibility Layer** (`internal/compat/laya`): shims that emulate Laya types (sprite hierarchies, events, timers, loaders, math structs) backed by Ebiten and standard Go libraries.
- **Infrastructure** (`internal/assets`, `internal/render`, `internal/text`): helpers for resource loading, font management, batching, texture atlases.
- **Tests** (`pkg/fgui/.../*_test.go`, `internal/.../*_test.go`): verify behavioural parity, asset parsing, layout math, tweening timelines, and the compatibility layer itself.

## Module Migration Plan

| TS Namespace / File                | Responsibility                                            | Go Package                         | Notes |
|------------------------------------|-----------------------------------------------------------|------------------------------------|-------|
| `fgui.GObject`, `GComponent`       | Base node management, layout, event hooks                | `pkg/fgui/core`                    | Depends on compat sprite, event dispatcher, relations system. |
| `fgui.GRoot`, `GTree`, `Window`    | Root stage, popups, windowing                             | `pkg/fgui/core`                    | Requires stage abstraction and input routing. |
| `fgui.display.*`                   | Renderable surfaces (`Image`, `MovieClip`)                | `pkg/fgui/display`                 | Wrap Ebiten draw operations behind compat sprites. |
| `fgui.gears.*`                     | Stateful UI gears (size, position, animation)             | `pkg/fgui/gears`                   | Use Go interfaces to decouple from specific components. |
| `fgui.tween.*`                     | Tweening engine                                           | `pkg/fgui/tween`                   | Tied to global scheduler in compat timer. |
| `fgui.utils.*`                     | Byte buffers, hit testing, colour math                    | `pkg/fgui/utils`                   | Many can port almost verbatim using Go equivalents. |
| `fgui.UIPackage`, `AssetProxy`     | Package loading, asset lookup                            | `pkg/fgui/assets` with `internal/assets` | Requires new loader abstraction for Ebiten-friendly IO. |
| `fgui.Controller`, `Transition`    | State machines and animation sequences                    | `pkg/fgui/controller`              | Tests should cover timeline correctness. |
| Global config (`UIConfig`, etc.)   | Defaults and feature flags                                | `pkg/fgui/config`                  | Keep static configuration with Go init patterns. |

## Compatibility Layer Blueprint

- **Display Tree (`DisplayObject`, `Sprite`)**
  - Wrap `*ebiten.Image` and metadata (transform, alpha, hit area) in Go structs implementing a retained hierarchy similar to Laya's `Sprite`.
  - Provide methods used by FairyGUI (`AddChild`, `RemoveChild`, bounds transforms, local/global matrix conversion).
  - Integrate with Ebiten via a traversal invoked from `GRoot.Draw`.

- **Math Types**
  - Implement lightweight `Point`, `Rect`, `Matrix` structs with methods matching the TypeScript signatures. Keep conversions to `image.Point` when calling Ebiten.

- **Event System**
  - Introduce `EventDispatcher` interface with `On`, `Off`, `Emit`, `Bubble`. Backed by Go maps of listener IDs. Provide common event constants mirroring `Laya.Event`.
  - Translate Ebiten input events (mouse, touch, keyboard) into compat events via an input router on each update tick.

- **Timer & Scheduler**
  - Implement a `Timer` singleton that tracks elapsed time from `Game.Update`. Support `CallLater`, frame loops, and delayed callbacks used by gears and tweens.

- **Loader & Assets**
  - Build an async loader service that reads from Go filesystem or embedded resources, returning `[]byte`/`*ebiten.Image`. Support batched loading analogous to `AssetProxy`.
  - For `.fui` / `.bin` packages, reuse `ByteBuffer` port to parse descriptors.

- **Text & Fonts**
  - Introduce a text subsystem translating Laya text metrics into Go text rendering using `golang.org/x/image/font`. Cache fonts, manage rich text fallback strategy.

- **Sound**
  - Map `SoundManager` calls to an abstraction that can be plugged into `ebiten/audio`.

- **Threading**
  - Replace `Laya.Handler` with Go function types / channels. For operations requiring async completion, return `Future`-like struct or use `context.Context`.

## Rendering & Game Loop Integration
- Provide a `fgui.GameHost` helper that wraps an `ebiten.Game`, wiring `Update`, `Draw`, `Layout` so that clients only register UI roots and respond to high-level callbacks.
- `GRoot` holds the compat stage, processes timer ticks, tweens, and input every update before nodes render in depth order.

## Testing Strategy
- Unit-test ports of deterministic logic: `ByteBuffer`, `Relations`, `GearSize`, `TweenManager`, `UIPackage` parsing.
- Snapshot-style layout tests: compute expected bounds/positions for sample package data imported from `laya_src`.
- Compatibility layer tests: event bubbling, timer scheduling accuracy, loader error handling.
- Use Go benchmarks where layout/tween performance is critical.
- Plan integration tests that run a headless `ebiten.Image` draw pass and assert pixel deltas for simple components (using `ebiten` off-screen images).

## Migration Phases
1. **Bootstrap**: establish compat layer skeleton (math, sprite, timer, events) with tests.
2. **Core Port**: translate `GObject`, `GComponent`, relations, controllers relying on the compat primitives.
3. **Rendering Components**: port display objects, text, loaders; verify atlas handling.
4. **Advanced Features**: gears, transitions, tweens, drag-drop.
5. **Package & Asset Flow**: hook `UIPackage` to new loader, support fonts/sounds.
6. **Validation**: run TypeScript sample data through Go runtime, compare outputs.
7. **Optimization**: profile, introduce batching or caching as needed.

## Deliverables Checklist
- Architecture and migration documentation (this file, kept up to date).
- Progress log (`docs/refactor-progress.md`) updated per milestone.
- Go packages with idiomatic APIs and lint-clean code.
- Comprehensive unit and integration tests to ensure behavioural parity.

