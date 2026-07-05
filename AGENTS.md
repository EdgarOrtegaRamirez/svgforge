# AGENTS.md — SvgForge

## Project Overview

SvgForge is a comprehensive SVG processing toolkit in Go with a CLI tool and library API. It provides parsing, optimizing, querying, diffing, converting, validating, transforming, and analyzing SVG documents.

## Build & Test

```bash
# Build
cd /root/workspace/svgforge
go build -o svgforge .

# Run all tests
go test ./...

# Run specific test suite
go test ./tests/parser/
go test ./tests/query/
go test ./tests/models/

# Lint
go vet ./...
```

## Architecture

- `internal/models/` — SVG AST data structures (SVGDocument, Element, ViewBox, etc.)
- `internal/parser/` — XML-based SVG parser that produces the AST
- `internal/optimizer/` — SVG optimization (remove redundancy, simplify paths, strip editor metadata)
- `internal/query/` — CSS-like selector engine (tag, class, ID, attributes, combinators, pseudo-classes)
- `internal/diff/` — Structural SVG diffing with change tracking
- `internal/convert/` — Format conversion (data URIs, inline HTML, CSS sprites, formatted output)
- `internal/stats/` — Statistics and analysis (element counts, features, complexity)
- `internal/validate/` — Validation (issues, accessibility, security, best practices)
- `internal/transform/` — Affine transformations (scale, rotate, translate, skew)
- `cmd/root.go` — Cobra CLI with 8 subcommands

## Key Design Decisions

1. **XML-based parsing** — Uses Go's `encoding/xml` for robust SVG parsing with lenient mode
2. **CSS-like selectors** — Full selector syntax including combinators, attribute selectors, and pseudo-classes
3. **Structural diffing** — Compares SVGs by element tree structure, not line-by-line
4. **Composable transforms** — Matrix-based affine transformations with SVG attribute output
5. **Zero external deps** beyond `github.com/spf13/cobra`

## Common Tasks

### Add a new query pseudo-class
1. Add handling in `internal/query/query.go` `matchesPseudo()` function
2. Add tests in `tests/query/query_test.go`
3. Update README selector documentation

### Add a new validation rule
1. Add validation function in `internal/validate/validate.go`
2. Call it from `validateElements()` or `validateCommonIssues()`
3. Add tests in `tests/validate/validate_test.go`

### Add a new optimization pass
1. Add optimization function in `internal/optimizer/optimizer.go`
2. Call it from `optimizeElement()` or `optimizeElements()`
3. Add tests in `tests/optimizer/optimizer_test.go`

## Testing Strategy

- `tests/models_test.go` — Unit tests for AST data structures (28 tests)
- `tests/parser_test.go` — Parser tests with various SVG inputs (8 tests)
- `tests/query_test.go` — Selector parsing and matching tests (18 tests)
- `tests/optimizer_test.go` — Optimization tests (6 tests)
- `tests/diff_test.go` — Diff comparison tests (7 tests)
- `tests/stats_test.go` — Statistics tests (7 tests)
- `tests/validate_test.go` — Validation tests (6 tests)
- `tests/transform_test.go` — Transformation tests (12 tests)
- `tests/convert_test.go` — Conversion tests (7 tests)

## CI

GitHub Actions workflow at `.github/workflows/ci.yml` runs `go test ./...` and `go vet ./...` on push to main.
