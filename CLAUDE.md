# pandemonium — Claude Code instructions

pandemonium is a procedurally generated, DOOM-style raycaster game: each level is generated from a seed, and reaching the exit generates the next. Correctness and simplicity over cleverness — prefer a well-tested, focused function over a clever abstraction. If the standard library does it, use it.

## Before every commit

* Run `just ci` autonomously (lint + test + build). All must pass.
* When adding a function or package, write unit tests alongside the code in the same commit.
* The Ebiten window lives in `internal/gui` and is the only package that imports Ebiten. It is ungated — this is a pure game with no headless mode — but stays behind the `gui.Run`/`gui.Available` seam.
* Never commit until the user explicitly confirms. Propose changes as diffs, run `just ci` autonomously, then stop and wait before `git commit`.

## Code quality

* Run `just lint` before proposing a diff. Fix all lint errors before committing.
* Prefer early returns over nesting.
* Do not add error handling or fallbacks for scenarios that cannot happen.
* Comments in `internal/` and `cmd/`: only *inside* functions, and only for non-obvious logic — never a doc comment on a declaration, a file or package header, or a `doc.go`. (`//go:build`, `//go:generate`, and `//nolint` are directives, not comments, and stay.)
* Public library code (module root or `pkg/…`, where present) keeps full godoc — a doc comment on every exported symbol and the package, enforced by the revive `exported` rule.

## Balance gate

* `internal/eval` plays the bot (`internal/sim/bot`) through seeded campaigns and checks the result against bands — a floor *and* a ceiling per metric — in `eval.Bands()`. That is the only place pass marks live; `TestGate` runs them, so `just test`, `just ci` and CI all share one gate. `just eval` prints the table.
* A red gate means the game changed, not the bands. Find the seed, trace the cause, fix it in `sim`/`world`/`bot`; only move a band when the design itself moved, and say so in the commit.
* Report the gate's table before and after any change to `sim`, `world` or `bot`.
* Keep the gate under ~10s.

## Level invariants

Generation guarantees these and `internal/world` tests them over many seeds; keep them true:

* The exit and every keycard are reachable from the spawn, and reachable without wading through a pool.
* Barrels are solid and stand only where they can be walked around on all sides.
* An alarm panel faces exactly one floor cell and never sits beside the spawn or the exit.
* No demon starts where it can see the spawn from within its notice range (`internal/sim` tests this one).

## Conventions

* The CLI entrypoint and Ebiten GUI structure is shared across the tool family and documented in [CONVENTIONS.md](CONVENTIONS.md). Keep new code consistent with it.
