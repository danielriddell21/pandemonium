# pandemonium

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
| `Enter` / `Space`  | Next level (on the tally screen) |
| `Esc`              | Pause / back (quit from the menu) |

The game opens on a title screen — which shows your run history (runs played and
the deepest level reached) and, left idle, plays a short demo of itself. `Esc`
during play opens a pause menu (resume, settings, quit) rather than quitting
outright.

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
toggle, and on-screen debug messages.
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

- [Architecture](docs/architecture.md) — the three decoupled layers.
- [How the raycaster works](docs/raycaster.md).
- [Demos](docs/demos.md) — more gameplay clips.
