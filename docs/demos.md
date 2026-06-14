# Demos

Short clips of the game in action. Each is generated from a fixed seed, so the
runs are reproducible.

## Exploration

![exploration](demos/exploration.gif)

Walking a procedurally generated level from the spawn through its rooms and
corridors toward the exit.

## Looking around

![looking around](demos/looking.gif)

A full turn on the spot, showing the per-column raycasting and distance shading
as the walls sweep past.

## Combat

![combat](demos/combat.gif)

Closing on a billboarded demon — taking damage as it reaches you (watch the
health gauge) — then gunning it down with the pistol (note the viewmodel and
muzzle flash). Demons animate and collapse when killed; sprites always face the
camera and are occluded correctly by nearer walls.

## Tally

![tally](demos/tally.gif)

Reaching the exit shows a level-complete tally — kills, items and secrets as
percentages, plus the time taken — then generates a fresh level to continue into.
