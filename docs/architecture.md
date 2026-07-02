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

The world stays a grid, but every tile carries a **floor and ceiling height** in
wall units, sculpted in quarter-wall steps by the generator (staircases between
rooms, a raised exit platform, lift tiles serving high ledges). The simulation
gives every body a height: small rises are climbed, drops fall under gravity, and
lift platforms carry whoever stands on them. The renderer walks each screen
column's ray boundary-by-boundary, painting floors, ceilings and textured step
faces inside a shrinking clip window until a wall closes the column — so the
vertical feel comes from the same single pass that draws the walls.

The Ebiten front-end lives in `internal/app` — the only package that imports
Ebiten — which drives the loop, reads input, and uploads each rendered frame. It
is a small state machine: a **title** screen (which idles into a bot-driven
attract demo), **play**, the between-levels **intermission** tally, a **pause**
menu, and a **settings** screen. Player options (sound, volumes, sensitivity,
field of view, crosshair, debug messages) live in `Settings`, persisted as JSON
under the user config directory and applied to the renderer and audio engine.

The attract demo and the documentation clips share one brain: `internal/sim/bot`
is a pure, headless pilot that routes to the exit and fights as it goes, so the
title screen and the capture tool play the game the same way a person would.

`cmd/pandemonium` is the composition root that wires the layers together.
