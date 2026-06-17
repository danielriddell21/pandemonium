# How the raycaster works

The renderer draws the 3D view one **vertical screen column at a time**, and the
world is a grid where every tile carries its own **floor and ceiling height** — so
the same pass that draws walls also draws staircases, raised platforms, lifts and
varied ceilings.

1. **Camera.** The player has a position, a facing angle, and an eye height. From
   the angle we derive a *direction* vector and a perpendicular *camera plane*
   vector whose length sets the field of view.

2. **Per-column ray.** For each column `x` across the screen, we build a ray that
   points from the player through that column of the camera plane.

3. **Boundary walk (DDA).** Each ray walks the grid with a **Digital Differential
   Analyzer**, jumping exactly from one grid line to the next. Unlike a classic
   raycaster it does **not** stop at the first solid cell: it visits each tile
   boundary in order, carrying a vertical **clip window** (the screen rows still
   unpainted) and the height of the tile it is leaving.

4. **No fisheye.** We use the **perpendicular distance** to each boundary — the hit
   distance projected onto the camera direction — which keeps straight walls flat
   instead of bowing them outward at the screen edges.

5. **Projection.** A world height `z` at perpendicular distance `d` maps to screen
   row `h/2 + (eyeZ − z)·h/d`. A wall one unit tall at distance `d` therefore spans
   `h/d` rows — closer surfaces are taller — and raising the eye or a floor simply
   shifts where the horizon and each surface land.

6. **What each boundary paints.** Crossing from tile A into tile B at distance `d`:
   - **Floor and ceiling** of A are cast up to their projected far edges, sampling
     the floor/ceiling textures at the world point recovered for each row (the
     per-column form of textured floor casting).
   - If B is **solid**, the remaining clip window is drawn as a textured wall slice
     and the column closes (its distance is recorded for sprite occlusion).
   - Otherwise B is a **step**: where B's floor rises above A's (or its ceiling
     drops below A's) a textured step face is drawn, the clip window tightens to
     B's floor/ceiling, and the walk continues into B. The column closes early if
     the window pinches shut.
   Walls and step faces are **distance-shaded** (and north/south faces a touch
   darker than east/west) for the dim, moody look.

7. **Sprites.** Demons, items and projectiles are **billboards** — flat images
   always facing the camera. After the geometry pass they are transformed into
   camera space, sorted **far-to-near**, and drawn. Their vertical placement uses
   the same projection, so a sprite stands on its tile's floor (or a projectile
   floats at its own height). The per-column depth buffer from the geometry pass
   hides sprite columns correctly behind nearer walls.
