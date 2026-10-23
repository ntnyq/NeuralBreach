# Neural Breach vector icon

Source: `../../../resources/icon-macos.svg`.

Preview: `../../../resources/icon-macos-preview.png` (rendered on white).

The SVG reconstructs the approved solid graphite base and continuous yellow path as editable geometry, with subtle gradients and relief. It contains no raster images, fonts, scripts, or external dependencies. The base is opaque, including beneath the five recesses; the canvas outside it is transparent.

## Matrix geometry

- Canvas: 1024 × 1024.
- Shared aperture: 80 × 80, corner radius 8.
- Horizontal and vertical center spacing: 176.
- Edge-to-edge gap along each grid axis: 96.
- Column centers: 352, 528, 704.
- Row centers: 304, 480, 656.
- Occupied cells: top-left; middle-left, middle-center, middle-right; bottom-right.

All five apertures reference one `cell` definition through `matrix-cells`. The top aperture aligns exactly with the middle-left aperture, and the bottom aperture aligns exactly with the middle-right aperture. Edit that shared definition to change every aperture consistently.

## Generation and packaging

The SVG is the single source for the application and documentation icons. Install [resvg 0.48.1](https://github.com/linebender/resvg/releases/tag/v0.48.1) on `PATH`, then run from the repository root:

```sh
make icon
make icon-check
# Alternatively, select a specific renderer executable:
go run ./tools/genicon -renderer /path/to/resvg
```

The generator produces `resources/icon.png` at 1024px, copies the SVG and PNG to `docs/public/icon.svg` and `docs/public/icon.png`, and renders 16/32/48/64px frames directly from the SVG into `docs/public/favicon.ico`. It preserves transparent outer margins and the opaque base beneath all five recesses. The renderer version is pinned to keep output consistent.

Commit the SVG, generator, generated assets, and `tools/genicon/assets.json` together. The manifest records SHA-256 hashes of the source, generator, and outputs. `make icon-check`, `make check`, `make build`, documentation builds, and CI/release packaging reject stale or missing assets. Checks and ordinary app builds use committed assets and do not require resvg or Node.js; resvg is needed only when regenerating icons.

`mygo.json` explicitly selects `resources/icon.png`. MyGo packages it as `AppIcon.icns` on macOS, embeds Windows icon resources, and installs the Linux icon. When invoking `go tool mygo build` directly, run `make icon-check` first. Do not configure MyGo's frontend `buildCommand` just for this check: it would classify the native application as a Web frontend. The documentation navbar and home page use the generated PNG to avoid SVG mask repaint artifacts during theme transitions; pages declare both SVG and ICO favicons with the site's base path.
