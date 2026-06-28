# Project Guide for Coding Agents

## Purpose

Resolume Timecode Monitor is a Go desktop application that receives clip data
from Resolume over OSC, calculates the remaining time, and serves a live web
display over HTTP and WebSocket. The desktop UI uses Fyne. Browser assets are
embedded into the Go binary.

## Repository Map

- `main.go`: process startup, OSC/HTTP/WebSocket servers, and embedded assets.
- `gui.go`: Fyne desktop UI and persisted application preferences.
- `procmessage.go`: OSC message parsing and timecode calculation.
- `distributor.go`: WebSocket client fan-out.
- `index.html`, `main.js`, `osc.min.js`: browser display.
- `dev.sh`: local build, run, browser, and OSC simulation helper.
- `.goreleaser.yml`: macOS Universal and Windows amd64 release packaging.
- `.github/workflows/workflow.yml`: GitHub Release workflow.
- `macos.sh`: creates the macOS `.app` bundle and icon.

## Development

The module targets Go 1.17 and requires CGO plus platform GUI dependencies.

```bash
go test ./...
./dev.sh build
./dev.sh run
./dev.sh send-osc
./dev.sh open
./dev.sh clean
```

`./dev.sh build` creates `resolume-timecode-dev`, which is intentionally
gitignored. Do not commit generated binaries, `dist/`, or local `.claude/`
configuration.

## Release Process

Release tags use semantic versions prefixed with `v`. GoReleaser produces:

- `resolume-timecode_<version>_macOS_Universal.zip`, containing one `.app` for
  Intel and Apple Silicon.
- `resolume-timecode_<version>_Windows_amd64.zip`.
- `checksums.txt`.

Releases are created as GitHub drafts. Publish them manually after checking the
assets. There is no DMG, Apple code signing, or notarization. README documents
the scoped `xattr` command required when Gatekeeper blocks the downloaded app.

Preferred release commands:

```bash
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin master
git push origin vX.Y.Z
```

If the tag push does not start the workflow, dispatch it explicitly:

```bash
gh workflow run .github/workflows/workflow.yml --ref master -f tag=vX.Y.Z
```

The manual workflow checks out the supplied tag. Never reuse or move a release
tag after publishing; create the next patch version instead.

## Current State (2026-06-27)

- Branch: `master`; remote: `calebweixun/resolume-timecode`.
- Latest release tag: `v1.0.7` at commit `251cbb1`.
- Release Action run `28313411696` completed successfully.
- The `v1.0.7` GitHub Release exists as a draft with macOS Universal, Windows
  amd64, and checksum assets; it still needs manual publication.
- macOS UPX compression was removed because UPX 5.2 rejects macOS binaries.
  Windows UPX compression remains enabled.
- The two macOS builds are merged before `.app` packaging to prevent parallel
  hooks from overwriting one another.
- While the server is running, the app actively polls Resolume transport
  position about every 110 ms and refreshes clip name/duration about once per
  second. Continuous OSC Output and manual Reset are therefore optional;
  Resolume must still accept OSC queries on its configured input port.
- A configured layer path such as `/composition/layers/1` follows that layer's
  playing clip. The app uses the `clips/playing` alias as a fallback and polls
  `clips/*/connected` to discover the exact active clip path. Discovery keeps
  the current clip stable during transitions and does not reset web clients.
  Resolume connected state codes `3` and `4` mean playing; `0` through `2` do
  not and must not lock the monitored clip.

## Change Discipline

- Preserve user changes and inspect `git status` before staging.
- Run `go test ./...` and `git diff --check` for Go or documentation changes.
- Run GoReleaser `check` after editing `.goreleaser.yml`.
- Do not claim a release succeeded until the Action conclusion and uploaded
  assets have both been verified.
- Keep `AGENTS.md` as the canonical shared project state; update it when the
  release process or current release changes.
