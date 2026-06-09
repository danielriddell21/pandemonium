binary := "pandemonium"
bin_dir := "bin"

default: build

# Compile the binary into ./bin
build:
    go build -o {{bin_dir}}/{{binary}} ./cmd/pandemonium

# Build and run
run:
    go run ./cmd/pandemonium

# Run all tests
test:
    go test ./...

# Run go vet
vet:
    go vet ./...

# Run golangci-lint
lint:
    golangci-lint run

# Run govulncheck
vulncheck:
    govulncheck ./...

# Tidy module dependencies
tidy:
    go mod tidy

# Fetch Freedoom assets
assets:
    ./scripts/fetch-assets.sh

# Remove build artifacts
clean:
    rm -rf {{bin_dir}} dist
