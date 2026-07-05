# Security Policy

## Reporting Vulnerabilities

If you discover a security vulnerability in SvgForge, please report it responsibly.

## Security Considerations

### Input Validation
- SVG parsing uses Go's `encoding/xml` which provides built-in protection against XML attacks
- Path data validation checks for valid SVG path commands
- No shell execution or system calls from parsed content

### Script Elements
- The validator flags `<script>` elements in SVGs as potential security risks
- Consider sanitizing SVGs before rendering in web contexts

### File Operations
- Only reads files from specified paths
- No path traversal vulnerabilities in file operations

### Dependencies
- Minimal dependencies (only `github.com/spf13/cobra` for CLI)
- Regular dependency updates via automated maintenance

## Best Practices

1. **Validate untrusted SVGs** — Use `svgforge validate` before processing
2. **Remove scripts** — Check for and remove `<script>` elements
3. **Sanitize output** — When embedding SVGs in HTML, use proper escaping
4. **Use data URIs carefully** — Base64-encoded SVGs can hide malicious content
