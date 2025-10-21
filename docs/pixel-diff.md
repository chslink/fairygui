# Pixel Comparison Guide

This utility helps you compare the Go implementation with a reference render exported from the original FairyGUI runtime.

## 1. Build an Actual Render

```bash
# Requires go 1.24+ and the ebiten build tag
GOCACHE=$(pwd)/.gocache go run -tags ebiten ./cmd/pixeldiff \
  -assets demo/assets \
  -package MainMenu \
  -component Main \
  -out snapshots/mainmenu_go.png
```

The command loads `demo/assets/MainMenu.fui`, builds the `Main` component, renders it off-screen using the Ebiten atlas manager, and writes the PNG to `snapshots/mainmenu_go.png`.

If the component depends on other packages the tool will resolve them automatically as long as the assets live in the same directory tree.

## 2. Capture A Reference Image

Grab a baseline export from the official FairyGUI runtime (Unity, Laya, or another trusted client). Save the screenshot as a PNG with the same resolution as the Go render. For consistent results, disable post-processing and ensure the UI is rendered on a solid background.

## 3. Compare Against The Baseline

```bash
GOCACHE=$(pwd)/.gocache go run -tags ebiten ./cmd/pixeldiff \
  -assets demo/assets \
  -package MainMenu \
  -component Main \
  -baseline snapshots/mainmenu_fairygui.png \
  -out snapshots/mainmenu_go.png \
  -diff snapshots/mainmenu_diff.png
```

The tool reports the number of differing pixels, the maximum channel delta, and the average per-channel delta. When `-diff` is provided, a heatmap is produced where the RGB intensity reflects the absolute difference per channel.

### Exit Codes

The command always exits with status 0; use the printed diff metrics in CI to decide whether to fail the build.

## Known Limitations

- Text rendering uses the default bitmap font and may not match the original. Supply matching fonts or ignore text layers when interpreting diffs.
- Current rendering honours position, scale, rotation, and skew. Advanced effects (shaders, filters) are skipped.
- Baseline and rendered images **must** share identical dimensions; resize them in advance if necessary.

Feel free to extend the tool with per-layer masks or tolerance thresholds in your automation. Contributions welcome!

## Preview A Single Image Asset

When you only need to inspect a solitary atlas sprite, run the lightweight helper:

```bash
GOCACHE=$(pwd)/.gocache go run -tags ebiten ./cmd/gimage-demo \
  -assets demo/assets \
  -package Basics \
  -image pic_icon \
  -out snapshots/pic_icon.png
```

The command loads the chosen `.fui`, resolves the sprite (including pixel hit-tests), and renders it directly without wrapping it in a full component hierarchy. Width/height defaults to the sprite's size—pass `-width` or `-height` to override when experimenting with layout padding.
