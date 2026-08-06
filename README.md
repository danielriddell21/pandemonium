# pandemonium

> *pandemonium* — all the demons, in one place.

[![CI](https://github.com/danielriddell21/pandemonium/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/pandemonium/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/danielriddell21/pandemonium/graph/badge.svg)](https://codecov.io/gh/danielriddell21/pandemonium)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_pandemonium&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_pandemonium)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

A procedurally-generated, Wolfenstein-3D-style raycaster FPS written in Go with [Ebiten](https://ebitengine.org/) v2, in the visual spirit of the original DOOM.

Every run generates a fresh maze of rooms and corridors — staircases, lifts, raised platforms, the odd open-sky courtyard. Walk it with `WASD`, look with the mouse, dodge the demons, and reach the exit to collapse the level and spawn a new one. Levels are deterministic from their seed.

```sh
go run ./cmd/pandemonium --seed 42
```

![pandemonium gameplay](docs/demos/hero.mp4)

## Install

### Homebrew (macOS)
```sh
brew install --cask danielriddell21/tap/pandemonium
```

### From source
On Linux/Windows, build from source:

```sh
go build ./cmd/pandemonium
```

<details>
<summary>Linux: OpenGL/X11 libraries</summary>

The window needs OpenGL/X11. On Debian/Ubuntu:

```sh
sudo apt install libgl1-mesa-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev
```
</details>

## Documentation

Full documentation lives in the [pandemonium wiki](https://github.com/danielriddell21/pandemonium/wiki) — features, controls, CLI flags, settings, and how the raycaster, world generation and audio work.
