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
  --width int     internal render width  (default 640)
  --height int    internal render height (default 400)
```

The chosen seed and the current level number are printed to stdout on start and
whenever a new level is generated, so demos are reproducible.

## Architecture

The project is split into three deliberately decoupled layers. Dependencies only
ever point one direction — **render → sim → world** — so the world generator and
the simulation can be exercised headlessly, with no graphics in sight.

1. **`internal/world` — generation.** Pure Go, zero rendering knowledge, no Ebiten
   import. A level is a 2D grid of tiles (walls, floors, doors, a spawn and an exit)
   produced deterministically from a seed. Generation uses recursive **BSP** grid
   splitting: the map is repeatedly cut into sub-regions, a room is carved into each
   leaf, and sibling regions are joined with corridors. A flood-fill reachability
   check guarantees the exit is reachable from the spawn; if it isn't, the level is
   regenerated. Fully unit-testable in isolation.

2. **`internal/sim` — simulation.** Player position and facing, movement with solid
   grid collision, billboarded entities, and interaction. State advances one fixed
   step at a time via an explicit `Tick`. Depends only on `world`; knows nothing
   about how anything is drawn.

3. **`internal/render` — rendering.** The Ebiten front-end. It owns no game logic;
   it reads simulation state and draws it. This is where the raycaster lives.

## How the raycaster works

The renderer draws the 3D view one **vertical screen column at a time**.

1. **Camera.** The player has a position and a facing angle. From the angle we
   derive a *direction* vector and a perpendicular *camera plane* vector whose
   length sets the field of view.

2. **Per-column ray.** For each column `x` across the screen, we build a ray that
   points from the player through that column of the camera plane.

3. **DDA traversal.** Each ray walks the grid with a **Digital Differential
   Analyzer**: instead of sampling at fixed steps, it jumps exactly from one grid
   line to the next, always advancing to whichever of the next vertical/horizontal
   cell boundaries is closer. This finds the first solid wall cell the ray hits with
   no missed thin walls and no wasted samples.

4. **No fisheye.** Using the raw ray length would bow straight walls outward at the
   screen edges. Instead we use the **perpendicular distance** — the hit distance
   projected onto the camera direction — which keeps walls flat.

5. **Wall slice.** The wall's on-screen height is inversely proportional to that
   perpendicular distance: closer walls are taller. Each column is drawn as a single
   vertical strip, **distance-shaded** so farther walls fade toward black for the
   dim, moody look. Walls hit on a north/south face are shaded slightly differently
   from east/west faces to give edges definition.

6. **Sprites.** Demons are **billboards** — flat images always facing the camera.
   After the walls are drawn, sprites are transformed into camera space, sorted
   **far-to-near**, and drawn. A per-column depth buffer recorded during the wall
   pass lets sprite columns be hidden correctly behind nearer walls.

## Development

```
make build    # compile the binary into ./bin
make run      # build and run
make test     # run all tests (headless: world, sim, telemetry)
make lint     # golangci-lint
```

## Assets

Wall and sprite art come from [Freedoom](https://freedoom.github.io/)
(BSD-3-Clause), a free, drop-in, Doom-compatible asset set — **not** the original
DOOM data, which is not freely licensed. See [`NOTICE`](./NOTICE) for attribution.
`scripts/fetch-assets.sh` pulls the needed files; if assets are absent the game
falls back to flat placeholder colours so it always runs.

## License

See [`NOTICE`](./NOTICE) for third-party asset attribution.
