binary := "pandemonium"
bin_dir := "bin"

# list available recipes
default:
    @just --list

# compile the binary into ./bin
[group('build')]
build:
    go build -o {{bin_dir}}/{{binary}} ./cmd/pandemonium

# run all tests
[group('test')]
test:
    go test ./...

# run golangci-lint
[group('dev')]
lint:
    golangci-lint run

# format the code
[group('dev')]
fmt:
    golangci-lint fmt

# tidy module dependencies
[group('dev')]
tidy:
    go mod tidy

# full gate: lint + test + build. all must pass before committing
[group('dev')]
ci: lint test build

# build and run
[group('run')]
run:
    go run ./cmd/pandemonium

# benchmark the renderer (Frame cost)
[group('test')]
bench:
    go test -run=^$ -bench=. -benchmem ./internal/render

# fuzz the world generator for a fixed time
[group('test')]
fuzz:
    go test -run=^$ -fuzz=FuzzGenerate -fuzztime=30s ./internal/world

# run govulncheck
[group('dev')]
vulncheck:
    govulncheck ./...

# fetch Freedoom assets
[group('run')]
assets:
    ./scripts/fetch-assets.sh

# regenerate the documentation demo clips and stills (needs ffmpeg)
[group('run')]
demos:
    go run ./tools/demogen

# remove build artifacts
[group('dev')]
clean:
    rm -rf {{bin_dir}} dist
