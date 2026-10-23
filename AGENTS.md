# Repository Guidelines

## Project structure

Neural Breach is an offline native puzzle game built with Go 1.27.3 and MyGo v0.6.2. The pinned MyGo CLI is a Go tool; Node.js is only used for the VitePress documentation.

- `main.go`, `window.go`, `view.go`, and `menu.go` compose lifecycle, game navigation, native views, and commands.
- `internal/game` owns deterministic rules, puzzle generation, bounded hint search, and scoring.
- `internal/records` stores preferences, the active slot, and history in one versioned snapshot.
- `internal/sound` synthesizes original music and preloads effects. `storage` supplies atomic writes; `tasks` rejects stale background callbacks.
- `locales/` contains five catalogs and generated accessors. `resources/fonts/` embeds licensed Rajdhani fonts.
- `docs/` contains the player guide, architecture, release instructions, and the original design reference.
- `tools/genicon` generates the app icon; `tools/releaseassets` validates release packages.

## Commands

```sh
make run
make dev
make build
make fmt
make check
make race
go tool mygo vet
make generate
make snapshots
pnpm docs:dev
pnpm docs:build
```

After Go changes, run `make check`, `make race`, and `go tool mygo vet`. For UI changes, inspect headless snapshots at 1100×700 and enlarged text sizes across languages. Native focus, audio, dialogs, and packaging need platform checks; cross-compilation does not verify their runtime behavior.

## Invariants

- Keep game rules independent of MyGo. Do not change generator behavior or scoring without updating the rule version and deterministic tests.
- Inputs start in row one and alternate column/row. Used cells cannot be reused. Match continuous substrings, preserve completed targets, and resolve success before buffer exhaustion.
- Use elapsed wall time with monotonic timestamps, not frame counts. Observation and explicit pauses consume no time. Paused views must hide both matrix and targets.
- Copy inputs before background work. Dispatch UI changes through `Window.Update`; route blocking dialogs and I/O through `window.run` or the revision-checked persistence path.
- Commit completed history and clear the active slot together. Never overwrite corrupt/unknown-version data automatically. Preserve backup bytes before repair.
- Standard scores exclude practice, tutorials, and abandoned games. Replaying a seed creates practice. Replay views never change records.
- Maintain one music decoder and release it on close. Reduced motion overrides glitch and edge-pulse settings.

## Style and tests

Follow `.editorconfig`, `gofmt`, and the configured `oxfmt`. Update all five catalogs together and run `make generate` and `make i18n`; do not edit generated accessors by hand. Retained messages resolve at display time. Tests use temporary data directories and never touch real settings or audio devices. Locale-provider tests stay serial.

Use concise conventional commit subjects. PRs describe behavior, validation and tested platforms, with screenshots for UI changes. Application versions live in `main.go` and `mygo.json`; release guidance is in `docs/releasing.md`.
