# Go CLI Development Standards

When building or refactoring Command Line Interfaces (CLIs) in Go, adhere to the following "Agent-First" and modern UX principles:

1. **Framework:** Use `github.com/spf13/cobra` for command routing. Always populate the `Short`, `Long`, and `Example` fields for every command to ensure high discoverability.
2. **Agent-First Interoperability:**
   - Always implement a persistent `--json` flag. When toggled, the command MUST output deterministic JSON (arrays or objects) suitable for `jq` or AI agent parsing.
   - Do not print human-readable text (like "Fetching data...") to `stdout` when `--json` is active.
3. **Security by Default:** Never output raw secrets, credentials, or full API keys by default in either text or JSON formats. Truncate them (e.g., `AIzaSy...abc12`) and require an explicit flag (e.g., `--full` or `--show-secrets`) to expose the raw string.
4. **Semantic Coloring:** Use `github.com/charmbracelet/lipgloss` for Tufte-inspired, semantic UI styling. 
   - Never hardcode colors. Use semantic tokens: `Accent` (headers), `Muted` (metadata/dates), `Pass` (success), `Warn` (risks), `Fail` (errors/danger).
   - `lipgloss` naturally respects `NO_COLOR` for CI/CD degradation.
5. **Configuration Fallbacks:** Implement robust configuration chains. For Google Cloud tools, the precedence should be: Explicit Flag -> Environment Variable -> `~/.env` (via `github.com/joho/godotenv`) -> Application Default Credentials.

## Go Quality & Linting

When finalizing a Go project or preparing a release, execute the following quality checks:
1. **Linting:** Always run `golangci-lint run ./...` to enforce Google Go style guidelines. Fix any issues related to unhandled errors, unchecked type assertions, or missing documentation.
2. **Licensing:** Use `github.com/google/addlicense` to prepend the Apache 2.0 license to all new or modified `.go` source files (e.g., `addlicense -c "Google LLC" -l apache .`).
3. **Standardize:** Ensure `go mod tidy` and `go fmt ./...` are run before any final build.

## Rendering Tables with Lipgloss and Printf
When constructing terminal tables using `fmt.Printf` (e.g., `%-35s`) alongside `lipgloss` styles, you MUST pad the raw string **before** applying the color.
Because `lipgloss` wraps the string in invisible ANSI escape codes, `Printf` will count those invisible characters toward the padding limit, breaking the table alignment.

**Incorrect:**
```go
fmt.Printf("%-35s | %-12s\n", StyleWarn.Render(val1), StyleMuted.Render(val2))
```

**Correct:**
```go
paddedVal1 := fmt.Sprintf("%-35s", val1)
paddedVal2 := fmt.Sprintf("%-12s", val2)
fmt.Printf("%s | %s\n", StyleWarn.Render(paddedVal1), StyleMuted.Render(paddedVal2))
```

## Google Cloud SDK Nuances

- **Protobuf `oneof` fields:** When interacting with Google Cloud Go SDKs (like `apikeyspb`), remember that `oneof` fields are not directly accessible on the struct. You must use the generated getter methods (e.g., `Restrictions.GetBrowserKeyRestrictions()`) to check for existence and cast the type.
- **Resource Lookups:** When a user requests a Google Cloud resource by a friendly name (like "Display Name"), but the API requires a system UUID or full resource path, first attempt a direct `Get` request. If it fails with a `NotFound` error, gracefully fallback to a `List` request and iterate to match the user's friendly string.
