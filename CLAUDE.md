# CLAUDE.md - DiffGPT Go Project Guidelines

## Build / Lint / Test Commands
- Build: `go build .` or `go build -ldflags="-s -w" -o diffgpt .`
- Cross-compile: `GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o diffgpt .`
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
