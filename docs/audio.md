# How the audio works

There are no sound files. Every effect and the musical backing are **synthesised
in code** as raw PCM — 44.1 kHz, 16-bit, stereo — by `internal/audio`, which is
pure and deterministic: the same build produces byte-identical samples. The
Ebiten playback that turns PCM into sound is the only part that touches a device,
and it lives in `internal/gui`.

## Sound effects

Each cue is a short waveform shaped by an amplitude envelope:

- **fire** — a noise burst over a falling chirp;
- **hit** — a low thud with a noise edge, when a demon dies;
- **door** — a slow rising-then-falling rumble;
- **pickup** — a quick two-note blip;
- **secret** — a little major arpeggio;
- **death** — a tone sliding down toward a groan;
- **exit** — a rising fifth as the level ends.

Firing is player-driven and plays at full volume; the rest are triggered by
**observations** from the simulation and attenuated by distance from the player,
fading to silence beyond a fixed range, so a fight across the level sounds
distant.

## The ambient and music beds

Two looping layers play under everything:

- a low **ambient drone**, and
- a **music bed**: a sparse, minor-mode pad — a low triad pulsing under a high
  shimmer.

Both are built so a whole number of wave cycles fits the buffer, so they loop
without a click. The music is generated in **darkness bands**: as a run goes
deeper the bed flattens its pitch, slows its pulse, drops its minor third and
fades its shimmer, and the ambient quietens — so the soundscape grows colder the
further in you get. The bed is swapped to its next darker band at a level
boundary, where the tally screen masks the seam.

## Playback and settings

The app holds one reusable player per effect plus the two looping beds. It tracks
the player's position to pan and attenuate effects, counts levels to advance the
ambient and music, and applies the user's **sound on/off** and **effect / ambient
volume** settings on top of the engine's own levels. On a machine with no audio
device the engine simply falls back to silence.
