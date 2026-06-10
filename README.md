# pandemonium

A procedurally-generated, Wolfenstein-3D-style raycaster FPS written in Go with
[Ebiten](https://ebitengine.org/) v2, in the visual spirit of the original DOOM.

Every run drops you into a freshly generated maze of rooms and corridors. Walk it
with `WASD`, look around with the mouse (or arrow keys), dodge the demons, and find
the exit — which collapses the level and generates a brand new one. Levels are
deterministic from a seed, so a given seed always produces the same world.

```
go run ./cmd/pandemonium --seed 42
```

## Controls

| Input              | Action            |
| ------------------ | ----------------- |
| `W` / `S`          | Move forward/back |
| `A` / `D`          | Strafe left/right |
| Mouse / `←` `→`    | Turn              |
| `E`                | Interact (doors)  |
| `Esc`              | Quit              |

## CLI

```
pandemonium [flags]

  --seed int      world seed (0 = random, printed on start)
  --width int     level width in tiles  (default 48)
  --height int    level height in tiles (default 32)
```

The chosen seed and the current level number are printed to stdout on start and
whenever a new level is generated, so demos are reproducible.

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
