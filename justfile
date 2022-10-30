PROJECT_NAME := "spok"
PROJECT_PATH := "github.com/FollowTheProcess/spok"
PROJECT_BIN := "./bin"
PROJECT_ENTRY_POINT := "./cmd/spok"
COVERAGE_DATA := "coverage.out"
COVERAGE_HTML := "coverage.html"
GORELEASER_DIST := "dist"
COMMIT_SHA := `git rev-parse HEAD`
VERSION_LDFLAG := PROJECT_PATH + "/cli/cmd.version"
COMMIT_LDFLAG := PROJECT_PATH + "/cli/cmd.commit"
SPOK_CACHE := ".spok"

# By default print the list of recipes
_default:
    @just --list

# Show justfile variables
show:
    @just --evaluate

# Tidy up dependencies in go.mod and go.sum
tidy:
    go mod tidy

# Run go generate on all packages
generate:
    go generate ./...

# Compile the project binary
build: tidy generate fmt
    go build -ldflags="-X {{ VERSION_LDFLAG }}=dev -X {{ COMMIT_LDFLAG }}={{ COMMIT_SHA }}" -o {{ PROJECT_BIN }}/{{ PROJECT_NAME }} {{ PROJECT_ENTRY_POINT }}

# Run go fmt on all project files
fmt:
    go fmt ./...

# Run all project tests (unit and integration)
test *flags: fmt
    SPOK_INTEGRATION_TEST=true gotest -race ./... {{ flags }}

# Run all project unit tests
unit *flags: fmt
    gotest -race ./... {{ flags }}

# Run all project benchmarks
bench: fmt
    go test ./... -run=None -benchmem -bench .

# Generate and view a CPU/Memory profile
pprof pkg type:
    go test ./{{ pkg }} -cpuprofile cpu.pprof -memprofile mem.pprof -bench .
    go tool pprof -http=:8000 {{ type }}.pprof

# Trace the program
trace:
    go tool trace trace.out

# Lint the project and auto-fix errors if possible
lint: fmt
    golangci-lint run --fix

# Calculate test coverage and render the html
cover:
    SPOK_INTEGRATION_TEST=true go test -race -cover -covermode=atomic -coverprofile={{ COVERAGE_DATA }} ./...
    go tool cover -html={{ COVERAGE_DATA }} -o {{ COVERAGE_HTML }}
    open {{ COVERAGE_HTML }}

# Remove build artifacts and other project clutter
clean:
    go clean ./...
    rm -rf {{ COVERAGE_DATA }} {{ COVERAGE_HTML }} {{ PROJECT_BIN }} {{ GORELEASER_DIST }} {{ SPOK_CACHE }}

# Run all tests and linting in one go
check: test lint

# Print lines of code (for fun)
sloc:
    find . -name "*.go" | xargs wc -l | sort -nr | head

# Install the project on your machine
install: uninstall build
    cp {{ PROJECT_BIN }}/{{ PROJECT_NAME }} $GOBIN/

# Uninstall the project from your machine
uninstall:
    rm -rf $GOBIN/{{ PROJECT_NAME }}
