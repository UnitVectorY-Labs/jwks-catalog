
# Commands for jwks-catalog
default:
  @just --list
# Build jwks-catalog with Go
build:
  go build ./...

# Run tests for jwks-catalog with Go
test:
  go clean -testcache
  go test ./...