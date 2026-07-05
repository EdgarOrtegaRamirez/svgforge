# SvgForge

A comprehensive SVG processing toolkit in Go with a CLI tool and library API.

## Features

- **Parse** — Parse SVG documents into a rich AST
- **Optimize** — Remove redundancy, simplify paths, strip editor metadata
- **Query** — Find elements using CSS-like selectors (tag, class, ID, attributes, pseudo-classes)
- **Diff** — Compare two SVGs structurally with detailed change reports
- **Convert** — Export to data URIs, inline HTML, CSS sprites, formatted output
- **Validate** — Check for issues, accessibility, security risks, and best practices
- **Transform** — Apply scale, rotate, translate, skew transformations
- **Stats** — Comprehensive analysis of element counts, features, and complexity

## Quick Start

```bash
# Install
go install github.com/EdgarOrtegaRamirez/svgforge@latest

# Or build from source
git clone https://github.com/EdgarOrtegaRamirez/svgforge.git
cd svgforge
go build -o svgforge .
```

## CLI Usage

```bash
# Parse and display SVG structure
svgforge parse input.svg

# Optimize SVG (remove redundancy)
svgforge optimize input.svg > optimized.svg

# Query elements with CSS selectors
svgforge query input.svg "rect.active"
svgforge query input.svg "g > circle"
svgforge query input.svg "#myId"

# Compare two SVGs
svgforge diff before.svg after.svg

# Show statistics
svgforge stats input.svg

# Validate SVG
svgforge validate input.svg

# Convert to data URI
svgforge convert input.svg --to datauri

# Convert to formatted output
svgforge convert input.svg --to formatted

# Apply transformations
svgforge transform input.svg --scale 2 --rotate 45 --translate-x 10
```

## Library API

```go
package main

import (
    "fmt"
    "github.com/EdgarOrtegaRamirez/svgforge/internal/parser"
    "github.com/EdgarOrtegaRamirez/svgforge/internal/optimizer"
    "github.com/EdgarOrtegaRamirez/svgforge/internal/query"
    "github.com/EdgarOrtegaRamirez/svgforge/internal/stats"
)

func main() {
    // Parse SVG
    p := parser.New()
    doc, err := p.ParseFile("input.svg")
    if err != nil {
        panic(err)
    }

    // Optimize
    opts := optimizer.DefaultOptions()
    optimizer.Optimize(doc, opts)

    // Query elements
    results, _ := query.QueryString(doc.Elements[0], "rect.fill-red")
    fmt.Printf("Found %d matching elements\n", len(results))

    // Get statistics
    s := stats.Analyze(doc)
    fmt.Printf("Total elements: %d\n", s.TotalElements)
}
```

## CSS-Like Selector Syntax

| Selector | Description |
|----------|-------------|
| `rect` | Match all `<rect>` elements |
| `*` | Match all elements |
| `#myId` | Match element with id="myId" |
| `.active` | Match elements with class "active" |
| `rect.active` | Match `<rect>` with class "active" |
| `rect#main.active` | Match `<rect>` with id="main" and class "active" |
| `g rect` | Match `<rect>` that is a descendant of `<g>` |
| `g > rect` | Match `<rect>` that is a direct child of `<g>` |
| `g + rect` | Match `<rect>` adjacent sibling of `<g>` |
| `[fill]` | Match elements with "fill" attribute |
| `[fill='red']` | Match elements where fill="red" |
| `[class~='active']` | Match elements where class contains "active" |
| `[id^='test']` | Match elements where id starts with "test" |
| `:first-child` | Match first child elements |
| `:last-child` | Match last child elements |
| `:nth-child(2)` | Match second child elements |
| `:empty` | Match empty elements |

## Architecture

```
svgforge/
├── cmd/                    # CLI entry point (Cobra)
├── internal/
│   ├── models/             # SVG AST data structures
│   ├── parser/             # SVG XML parser
│   ├── optimizer/          # SVG optimization engine
│   ├── query/              # CSS-like selector engine
│   ├── diff/               # Structural SVG diffing
│   ├── convert/            # Format conversion
│   ├── stats/              # Statistics and analysis
│   ├── validate/           # Validation engine
│   └── transform/          # Affine transformations
└── tests/                  # Test suites
```

## Testing

```bash
# Run all tests
go test ./...

# Run specific test suite
go test ./tests/parser/
go test ./tests/query/
```

## License

MIT
