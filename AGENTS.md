# Repository Guidelines


##  Guidelines
- 全程使用中文进行问答
- 代码架构 docs/architecture.md
- 代码更新进度 docs/refactor-progress.md
- 因为开发环境处于沙盒环境,项目又是GUI相关,所以无法运行大部分需要实际渲染的单元测试，demo。请输出需要运行的命令，这边会在GUI环境运行之后反馈给你。

## Project Structure & Module Organization
`pkg/fgui` contains the public FairyGUI runtime; its subpackages (`core`, `display`, `gears`, `tween`, `utils`, `assets`) mirror the original TypeScript layout. LayaAir compatibility shims live in `internal/compat/laya`, with rendering, text, and loading helpers in `internal/render`, `internal/text`, and `internal/assets`. Tests sit beside their packages as `*_test.go`. The playable sample in `demo/main.go` draws assets from `demo/assets` and `demo/UIProject`; use it to confirm behaviour. `laya_src/fairygui` is the upstream reference, while `docs/` tracks architectural decisions—update those when you move or add modules.

## Build, Test, and Development Commands
Run `go build ./...` to guard against compilation regressions on Go 1.24. Execute `go test ./...` for the full suite; narrow the focus with `go test ./pkg/fgui/core -run TestGComponent` when iterating. Launch the sample UI with `go run ./demo` to validate rendering, input, and asset loading together. Benchmarks for layout or tween hot paths belong under `go test -bench . ./pkg/fgui/...`.

## Coding Style & Naming Conventions
Format every change with `gofmt` (tabs, blank lines as per Go style) and tidy imports using `goimports`. Exported identifiers stay in `CamelCase`; internal helpers remain lowercase. Keep package names short, lowercase, and aligned with FairyGUI concepts. Prefer explicit constructors or factory functions to hidden globals, and reserve comments for non-obvious behaviour or porting caveats.

## Testing Guidelines
Write table-driven tests alongside implementations, covering layout math, asset parsing, timers, and event propagation—the areas most likely to regress when porting. Deterministic timer tests should inject fake clocks instead of sleeping. Maintain strong coverage in `pkg/fgui` and `internal/compat`, since they anchor the public API. After notable changes, run `go run ./demo` and confirm assets load without warnings.

## Commit & Pull Request Guidelines
Follow the conventional commit format used to date (`feat(assets): …`, `chore: …`). Scopes should match package directories or concise domain names; localized scopes such as `组件系统` are acceptable when they clarify intent. Pull requests must spell out purpose, notable code paths, and test evidence (command output or screenshots for UI tweaks). Link related issues and call out migrations or breaking API changes so downstream consumers can plan.

## Asset & Configuration Notes
Store runtime assets under `demo/assets`; place fixture binaries for tests in `internal/assets/testdata` to avoid polluting the main sample. Document new large files or asset workflows in `docs/refactor-progress.md`. When adding Ebiten configuration toggles, wire them through `demo/main.go` and describe expected values to keep the exported `pkg/fgui` surface stable.
