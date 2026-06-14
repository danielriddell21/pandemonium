# pandemonium

[![CI](https://github.com/danielriddell21/pandemonium/actions/workflows/ci.yaml/badge.svg)](https://github.com/danielriddell21/pandemonium/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/danielriddell21/pandemonium/graph/badge.svg)](https://codecov.io/gh/danielriddell21/pandemonium)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=danielriddell21_pandemonium&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=danielriddell21_pandemonium)
[![Go 1.26](https://img.shields.io/badge/go-1.26-blue)](https://go.dev)
[![MIT License](https://img.shields.io/badge/licence-MIT-green)](LICENSE)

A procedurally-generated, Wolfenstein-3D-style raycaster FPS written in Go with
[Ebiten](https://ebitengine.org/) v2, in the visual spirit of the original DOOM.

Every run drops you into a freshly generated maze of rooms and corridors — with
staircases, raised platforms and lifts sculpted into the terrain, and the odd
courtyard open to the sky. Walk it with
`WASD`, look around with the mouse (or arrow keys), dodge the demons, and find
the exit — which collapses the level and generates a brand new one. Levels are
deterministic from a seed, so a given seed always produces the same world.

```
go run ./cmd/pandemonium --seed 42
```

![pandemonium gameplay](docs/demos/hero.mp4)

## Install

### Homebrew (macOS)
```sh
brew install --cask danielriddell21/tap/pandemonium
```

On Linux/Windows, build from source (`go build ./cmd/pandemonium`).

<details>
<summary>Linux: OpenGL/X11 libraries</summary>

The window needs OpenGL/X11. On Debian/Ubuntu:

```sh
sudo apt install libgl1-mesa-dev libxrandr-dev libxcursor-dev libxinerama-dev libxi-dev
```
</details>

## Features

- **Procedural worlds** — BSP rooms and corridors carved fresh each level and
  deterministic from a seed, with terrain sculpted in quarter-wall steps:
  staircases, raised platforms, lifts, open-sky courtyards, and a **set-piece
  arena** every fifth level.
- **Arsenal** — fists, pistol, shotgun, chaingun and rocket launcher (hitscan and
  splash), with ammo, a backpack, and auto-switch when a weapon runs dry.
- **Bestiary** — melee chargers, fireball imps, hitscan gunners, fast pinkies and
  armoured barons, plus exploding barrels and monster infighting.
- **Pickups & power-ups** — health, armour, soulsphere, megasphere, berserk,
  invulnerability and a radiation suit; keycards behind locked doors; hidden
  secrets to find.
- **Hazards** — radioactive slime and molten lava floors (a radsuit shrugs off
  slime, but not lava).
- **Four difficulty levels** scaling the threat without changing the map.
- **Pure-CPU column raycaster** with per-tile floor/ceiling heights, see-over low
  walls, per-sector lighting and distance shading — and an automap.
- **All procedural, no assets** — every texture and sound is synthesised in code,
  including an ambient/music bed that darkens as a run deepens, and positional
  sound effects.
- **Title, pause and settings menus**, a self-playing attract demo, and run
  history kept across sessions.

The world, simulation and renderer are pure and headless-testable; only the
Ebiten front-end touches the screen.

## Controls

| Input              | Action            |
| ------------------ | ----------------- |
| `W` / `S` (or `↑` `↓`) | Move forward/back |
| `A` / `D`          | Strafe left/right |
| Mouse / `←` `→`    | Turn              |
| Left-click / `Ctrl` / `F` | Attack     |
| `1`–`5` / wheel    | Switch weapon (fists, pistol, shotgun, chaingun, rockets) |
| `E` / `Space`      | Interact (doors, switches) |
| `Tab`              | Toggle automap    |
| `F11`              | Toggle fullscreen |
| `F12`              | Screenshot (HUD-free PNG in the working dir) |
| `Enter` / `Space`  | Next level (on the tally screen) |
| `Esc`              | Pause / back (quit from the menu) |

The game runs fullscreen by default with the mouse captured for the view, so the
cursor stays inside the game while you play; `F11` toggles fullscreen and the
pause/menu screens release the cursor. The game opens on a title screen — which
shows your run history (runs played and the deepest level reached) and, left
idle, plays a short demo of itself. `Esc` during play opens a pause menu
(resume, settings, quit) rather than quitting outright.

## CLI

```
pandemonium [flags]

  --seed int      world seed (0 = random, printed on start)
  --width int     level width in tiles  (default 48)
  --height int    level height in tiles (default 32)
```

The chosen seed and the current level number are printed to stdout on start and
whenever a new level is generated, so demos are reproducible.

## Settings

Open the settings screen from the title or pause menu (`←` `→` adjust the
highlighted row): difficulty (easy through nightmare — it scales how many demons
spawn and how hard they hit, leaving the map itself unchanged), sound on/off,
sound-effect and ambient volume, mouse sensitivity, field of view, a crosshair
toggle, on-screen debug messages, and fullscreen.
Changes are saved to a JSON file under your user config directory
(`~/.config/pandemonium/settings.json` on Linux) and reloaded on the next run;
your run history is kept alongside it in `records.json`.
Turn **sound** off there to run silently on a machine with no audio device; the
game also falls back to silence on its own if the audio engine can't start.

## Development

```
just build    # compile the binary into ./bin
just run      # build and run
just test     # run all tests
just lint     # golangci-lint
```

## Documentation

- [Architecture](docs/architecture.md) — the layered, headless-testable design.
- [How the raycaster works](docs/raycaster.md) — the column renderer.
- [How the world is generated](docs/worldgen.md) — the generation pipeline.
- [How the audio works](docs/audio.md) — the procedural synthesis.
- [Demos](docs/demos.md) — a tour of the features in motion.
