# CLAUDE.md - DiffGPT Go Project Guidelines

## Build Commands
- `make build` - Build the diffgpt binary to `build/diffgpt`
- `make run` - Build and run the application  
- `make install` - Install diffgpt globally using `go install`
- `make deps` - Update Go module dependencies with `go mod tidy`
- `make clean` - Remove the build directory

## Alternative Build / Lint / Test Commands
- Build: `go build .` or `CGO_ENABLED=0 go build -ldflags="-s -w" -o diffgpt .`
- Cross-compile (static): `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o diffgpt .`
- Lint: `go vet ./...`
- Test (all): `go test ./...`
- Test (single): `go test ./path/to/package -run TestName`

## Code Style Guidelines
- **Imports**: Standard lib first, third-party second, internal last (with newlines between)
- **Formatting**: Use gofmt; tabs for indentation; ~100 char line length
- **Types**: Clear definitions with descriptive field names; appropriate struct tags
- **Naming**:
  - Packages: lowercase, concise (git, config, llm)
  - Functions: CamelCase (public UpperCamelCase, private lowerCamelCase)
  - Constants: UPPERCASE_WITH_UNDERSCORES
- **Error Handling**: 
  - Check errors immediately
  - Wrap errors with context using `fmt.Errorf("context: %w", err)`
  - Return descriptive error messages
- **Comments**: Document function purpose and complex logic; explain why, not what
