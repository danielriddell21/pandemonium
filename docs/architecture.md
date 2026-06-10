# Architecture

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

3. **`internal/render` — rendering.** The pure CPU rasteriser that reads simulation
   state and produces a frame buffer. It owns no game logic. This is where the
   raycaster lives (see [raycaster.md](raycaster.md)).

The Ebiten front-end lives in `internal/app` — the only package that imports
Ebiten — which drives the loop, reads input, and uploads each rendered frame.
`cmd/pandemonium` is the composition root that wires the layers together.
