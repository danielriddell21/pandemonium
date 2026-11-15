# Contributing to pandemonium

## Requirements

* [Go](https://go.dev) (stable — version from `go.mod`)
* [just](https://github.com/casey/just)
* [golangci-lint](https://golangci-lint.run/welcome/install/) (for `just lint`)
* [gremlins](https://github.com/go-gremlins/gremlins) (for mutation testing)

## Development workflow

```
just build      # build the binary
just run        # build and run
just test       # run unit tests (includes the balance gate)
just eval       # run the balance gate alone and print its table
just lint       # golangci-lint
just fmt        # golangci-lint fmt (gofumpt + goimports)
just ci         # lint + test + build
just bench      # run benchmarks
just fuzz       # run fuzz tests
just vulncheck  # govulncheck
just tidy       # go mod tidy
just assets     # regenerate bundled assets
just demos      # regenerate demo assets
just clean      # remove build artifacts
```

Run `just --list` to see every recipe. Run `just ci` (lint + test + build) before each commit. CI runs the same gate on every push to `trunk` and every pull request targeting `trunk`.

`golangci-lint` must be built with at least the Go version in `go.mod`, or it refuses to load the config. Build it with the project's own toolchain: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0`.

## Balance gate

`internal/eval` plays a bot through seeded campaigns — maze floors and boss floors, kit carried between them — and checks clear rate, deaths, damage taken, alarms and pace against bands in `eval.Bands()`. Each band has a floor and a ceiling: the game must be neither unwinnable nor toothless. It also replays every level with the demons removed and requires all of them to be walkable. The gate is an ordinary test, so `just test` and CI run it; `just eval` prints the table. When it fails, trace the seed rather than widening the band.

## Conventions

The CLI entrypoint and Ebiten GUI structure is shared across the tool family
(unum is the CLI reference; rubix/vivarium the GUI references). See
[CONVENTIONS.md](CONVENTIONS.md).

## Project layout

```
cmd/pandemonium/   entry point
internal/cli/      cobra root + game entrypoint
internal/gui/      Ebiten window + game loop (the only package importing Ebiten)
internal/          implementation packages (render, sim, world, eval, …)
tools/             developer tooling
scripts/           helper scripts
docs/              documentation
```

## Commit style

```
type(scope): short imperative description
```

Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`. No period at the end of the subject line; keep it under 72 characters.

## Releases

Releases are cut by maintainers only. Every push to `trunk` that passes CI is tagged as a release candidate and built with letsgo; un-ticking "pre-release" on a candidate promotes it to the stable version and updates the Homebrew cask, which needs the tap app credentials configured as repository secrets.
