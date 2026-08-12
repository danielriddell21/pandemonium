# How the world is generated

Every level is produced **deterministically from a seed** by `internal/world`:
the same `Config` always yields the same `Level`. A level is a row-major grid of
tiles, with per-tile floor and ceiling heights, light, theme, hazard and sky
data layered on top.

## The guarantee

`Generate` builds candidate levels until one passes two checks, then returns it:

- the **exit is reachable** from the spawn once doors are open (a flood fill), and
- every **keycard is obtainable** without first crossing the door it unlocks.

Each retry derives a distinct but deterministic sub-seed, so a `Config` that needs
several attempts still resolves to one fixed level.

## The pipeline

An ordinary level is assembled in stages, each driven by the level RNG so the
result is reproducible:

1. **Partition (BSP).** The grid is recursively split into sub-regions.
2. **Rooms & corridors.** A room is carved into each leaf region, sibling regions
   are joined with L-shaped corridors, and a few short dead-end stubs branch off.
3. **Spawn & exit.** The spawn goes in the first room; the exit is the *farthest*
   reachable cell from it (a distance field), maximising the journey. The exit
   becomes a **wall switch** beside the exit tile.
4. **Annotation.** Decision points are tagged — junctions, dead-end doors, and a
   **decoy exit** that resembles the real one — for the telemetry layer to read.
5. **Key gate.** Some levels lock the route behind a keycard, with the key placed
   where it can be fetched before the door.
6. **Items & barrels.** Consumables, the occasional backpack or power-up, and
   explosive barrels are scattered across the floor, clear of the spawn.
7. **Heights.** Floors and ceilings are sculpted in quarter-wall steps —
   staircases between rooms, a raised exit dais, varied ceiling heights, and a
   lift tile serving a ledge plain steps cannot reach.
8. **Hazards.** Some levels flood a small pool of damaging floor — radioactive
   slime, or, less often, molten lava — across flat tiles away from the spawn.
9. **Light & theme.** Each room gets its own brightness and wall tint, so the
   level reads as a series of distinct spaces; the spawn room stays bright.
10. **Low walls.** A few divider walls are dropped to a finite height, so you can
    see (but not walk) over them into the room beyond.
11. **Sky.** Roughly a third of rooms are opened to the sky: a tall, full-bright
    ceiling the renderer paints as open air.

Hazards, low walls and sky never change which tiles are walkable, so they can
never break the reachability guarantee.

## Arenas

Every fifth level is a **set-piece arena** instead of a maze (`Config.Arena`):
one large, open, sky-lit room with the spawn at one end, the exit switch at the
other, a megasphere-and-armour cache at its heart and a scatter of barrels. The
simulation fills it from the wide floor area, so it plays as a heavier, open
fight. With the flag off the same seed still yields the ordinary maze.

## What the simulation adds

Demons are not part of the level data; the simulation scatters them
deterministically from the level seed when it starts, scaling their number with
the floor area and the chosen difficulty. So a seed fixes the map exactly, while
difficulty changes only the threat upon it.
