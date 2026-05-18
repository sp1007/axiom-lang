# p17-t06: Documentation Generator

## Purpose
Implement `axc doc` — a documentation generator that extracts doc strings from AXIOM source, generates HTML documentation, and serves it locally for browsing.

## Context
Documentation is a product feature. `axc doc` extracts `##` doc comments from AXIOM source, renders them as HTML with syntax-highlighted code examples, and produces a browsable documentation website. The output is static HTML usable on GitHub Pages or any web host.

## Inputs
- AXIOM source with `##` doc comments above declarations
- AST from parser (p03)
- TypeInfo from type checker (p04) for type signatures

## Outputs
- `tools/doc/doc.go` — doc generator
- `doc/` directory with HTML output
- `axc doc [--serve]` subcommand

## Dependencies
- p03: parser — extracts doc comments from AST
- p04: type checker — provides type signatures for functions
- p16-t02: std-string — string manipulation for HTML generation

## Detailed Requirements

Doc comment syntax:
```axiom
## Compute the sum of two integers.
##
## Example:
##   let result = add(1, 2)
##   assert_eq(result, 3)
fn add(a: i32, b: i32) -> i32:
    a + b
```

Generated HTML structure:
```
doc/
  index.html          — module index
  stdlib/
    string.html       — std.string module docs
    math.html         — std.math module docs
  mymodule/
    mymodule.html     — user module docs
```

```go
type DocItem struct {
    Name    string
    Kind    string    // "fn", "type", "const"
    TypeSig string
    DocText string    // markdown-rendered
    SourceLoc string  // file:line
    Examples []string // code examples from ## Example: blocks
}

func GenerateDocs(modules []string, outputDir string) error
func ServeDocsLocal(dir string, port int) error
```

HTML generation: no external templates — generate HTML directly.

Markdown rendering in doc strings:
- `##` paragraphs → `<p>` tags
- `## Example:` blocks → `<pre><code class="axiom">...</code></pre>` with syntax highlighting

Search: generate `search_index.json` with all symbol names for client-side search.

`axc doc --serve`: generate docs, then serve on `http://localhost:8080`.

## Implementation Steps

1. Create `tools/doc/doc.go`.
2. Extract `##` doc comments from AST (attached to following declaration).
3. Parse type signatures from TypeInfo.
4. Generate HTML per module.
5. Generate `index.html` with module listing and search index.
6. Implement `--serve` using net/http.
7. Write tests for doc extraction and HTML output.

## Test Plan
- `TestDocExtract`: function with `##` doc → DocItem with correct text
- `TestDocHTMLOutput`: generated HTML contains function name and signature
- `TestDocExample`: `## Example:` block → syntax-highlighted code block
- `TestDocServe`: `--serve` → HTTP 200 on localhost:8080
- `TestDocNoComment`: function without ## → appears in docs with just signature

## Validation Checklist
- [ ] All public symbols documented (warning if missing doc comment)
- [ ] HTML is valid (no unclosed tags)
- [ ] Search index includes all symbol names
- [ ] Source links point to correct file:line

## Acceptance Criteria
- `axc doc stdlib/` generates complete stdlib docs browsable in browser

## Definition of Done
- [ ] `tools/doc/doc.go` implemented
- [ ] Stdlib docs generated and browsable

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Doc comment alignment with AST nodes after reformatting | Attach doc comments to AST nodes during parse, not post-process |

## Future Follow-up Tasks
- Online docs hosting at docs.axiom-lang.org
- Cross-reference links between types (click on `i32` → primitive docs)
