# Architecture

The project is built around three deliberately decoupled core layers, with the
Ebiten front-end on top and a few supporting packages alongside. Dependencies
only ever point one direction — **render → sim → world** — so the world generator
and the simulation can be exercised headlessly, with no graphics in sight.

1. **`internal/world` — generation.** Pure Go, zero rendering knowledge, no Ebiten
   import. A level is a 2D grid of tiles (walls, floors, doors, a spawn and an exit)
   produced deterministically from a seed. Generation uses recursive **BSP** grid
   splitting: the map is repeatedly cut into sub-regions, a room is carved into each
   leaf, and sibling regions are joined with corridors. A flood-fill reachability
   check guarantees the exit is reachable from the spawn; if it isn't, the level is
   regenerated. Fully unit-testable in isolation (see [worldgen.md](worldgen.md)).

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

The Ebiten front-end lives in `internal/gui` — the only package that imports
Ebiten — which drives the loop, reads input, and uploads each rendered frame. It
is a small state machine: a **title** screen (which idles into a bot-driven
attract demo), **play**, the between-levels **intermission** tally, a **pause**
menu, and a **settings** screen. Player options (difficulty, sound, volumes,
sensitivity, field of view, crosshair, debug messages) live in `Settings`,
persisted as JSON under the user config directory and applied to the renderer,
audio engine and level builder. Difficulty (`sim.Skill`) scales the threat —
demon count and the damage they deal — without touching the generated geometry,
so a seed yields the same map at every skill.

The attract demo and the documentation clips share one brain: `internal/sim/bot`
is a pure, headless pilot that routes to the exit and fights as it goes, so the
title screen and the capture tool play the game the same way a person would.

A few supporting packages sit alongside these, all pure and observing inward:

- **`internal/audio`** synthesises every sound effect and the ambient/music bed
  as raw PCM — no files, deterministic. The Ebiten playback that turns PCM into
  sound lives in `internal/gui`, which also pans and attenuates effects by
  distance and darkens the bed as the run deepens (see [audio.md](audio.md)).
- **`internal/telemetry`** attaches to the simulation as an observer and turns its
  observations into a cumulative run profile and per-event stream.
- **`internal/status`** subscribes to that telemetry and posts the on-screen
  status messages, choosing lines from a scripted source by how the run is going.
- **`internal/hud`** is the small overlay those messages are posted to and that
  the renderer reads back when drawing a frame.

`cmd/pandemonium` is the composition root that wires the layers together.
