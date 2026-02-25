module fillmore-labs.com/errortype

go 1.25.0

toolchain go1.26.4

require (
	github.com/goccy/go-yaml v1.19.2
	github.com/golangci/plugin-module-register v0.1.2
	golang.org/x/tools v0.47.0
)

require (
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
)

tool (
	fillmore-labs.com/errortype/internal/cmd/bitmask
	golang.org/x/tools/cmd/stringer
)
