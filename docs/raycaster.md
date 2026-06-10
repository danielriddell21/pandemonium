# How the raycaster works

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
