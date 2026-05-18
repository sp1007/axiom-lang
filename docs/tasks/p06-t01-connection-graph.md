# p06-t01: Connection Graph Data Structure

## Purpose
Implement the Connection Graph — the central data structure for ownership tracking, escape analysis, and Compile-Time GC. The Connection Graph models ownership relationships between values as a directed graph with typed edges, enabling the compiler to statically reason about memory lifetimes.

## Context
The Connection Graph is inspired by Escape Analysis in the Java HotSpot JVM and Vale's region-based system. Each value in a function is a node; edges represent how values relate to each other. The key insight: if a value has no `EscapesTo` edges pointing outside its scope, it can be stack-allocated and freed automatically. The graph is built during semantic analysis and queried by subsequent passes.

## Inputs
- Typed, name-resolved AST
- SymbolTable (p04-t01) — each symbol becomes a potential graph node

## Outputs
- `compiler/sema/connection_graph.go` — ConnectionGraph data structure
- `CGNode`, `CGEdge`, `EdgeKind` types

## Dependencies
- p04-t04: name-resolver — symbol resolution needed to identify value nodes
- p04-t05: type-inference-hm — TypeIDs needed for node types

## Subsystems Affected
- Ownership analysis (p06-t02): reads/writes connection graph
- Escape analysis (p06-t04): queries EscapesTo edges
- CTGC (p06-t05): reads destroy injection points from graph
- AIR metadata (p09-t03): ownership info stored in AIR metadata

## Detailed Requirements

1. `EdgeKind` enum: `Owns`, `Borrows`, `FlowsTo`, `EscapesTo`, `ReusedBy`
2. `CGNode` struct:
   ```go
   type CGNode struct {
       ID       uint32
       SymID    uint32  // 0 for temporary nodes
       TypeID   uint32
       IsRef    bool    // true if this is a reference node (not a value node)
       Lifetime uint32  // scope depth where this node is alive
   }
   ```
3. `CGEdge` struct: `{From:uint32, To:uint32, Kind:EdgeKind}`
4. `ConnectionGraph` struct:
   ```go
   type ConnectionGraph struct {
       Nodes    []CGNode
       Edges    []CGEdge
       adjOut   [][]uint32  // outgoing edge indices per node
       adjIn    [][]uint32  // incoming edge indices per node
       symToNode map[uint32]uint32  // SymID → NodeID
   }
   ```
5. API:
   - `AddValueNode(symID, typeID, lifetime uint32) uint32`
   - `AddRefNode(targetNodeID uint32) uint32`
   - `AddEdge(from, to uint32, kind EdgeKind)`
   - `NodeOfSym(symID uint32) uint32`
   - `OutEdges(nodeID uint32, kind EdgeKind) []uint32`
   - `InEdges(nodeID uint32, kind EdgeKind) []uint32`
   - `Escapes(nodeID uint32) bool` — has any EscapesTo edge
   - `DominatedBy(nodeID, scopeID uint32) bool`
6. ConnectionGraph is per-function (created fresh for each function body).
7. Serializable to JSON for `.axmeta` output and debugging.

## Implementation Steps

1. Create `compiler/sema/connection_graph.go`.
2. Implement `ConnectionGraph` with adjacency lists.
3. Implement `AddValueNode`, `AddRefNode`, `AddEdge`.
4. Implement `Escapes(nodeID)`: walk outgoing edges, return true if any EscapesTo found via DFS.
5. Implement `OutEdges(nodeID, kind)`: filter adjacency list by edge kind.
6. Add `String() string` method for debug printing.
7. Add JSON marshaling for `.axmeta` export.
8. Write unit tests: `TestCGAddNodes`, `TestCGAddEdges`, `TestCGEscapes`, `TestCGSerialize`.

## Test Plan

- `TestCGSimple`: add two nodes, one Owns edge, verify OutEdges
- `TestCGEscapeDetection`: node with EscapesTo edge → `Escapes()=true`; node with only Owns edges → `Escapes()=false`
- `TestCGTransitiveEscape`: A Owns B, B EscapesTo global → A considered to escape
- `TestCGSerialization`: graph to JSON and back → equal

## Validation Checklist

- [ ] All edge kinds supported
- [ ] `Escapes()` correctly detects direct and transitive escape
- [ ] Per-function graph correctly reset between functions
- [ ] JSON serialization preserves all node/edge data
- [ ] O(V+E) traversal for escape detection

## Acceptance Criteria

- ConnectionGraph correctly models the ownership of a 10-variable function
- Escape detection runs in O(V+E) per query
- Graph serializes to valid JSON

## Definition of Done

- [ ] `compiler/sema/connection_graph.go` implemented
- [ ] Unit tests pass
- [ ] JSON serialization working

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Cycle detection needed for escape DFS | Use visited set in DFS to avoid infinite loops |
| Graph grows too large for big functions | Limit to 10K nodes per function; emit warning if exceeded |

## Future Follow-up Tasks

- p06-t02: ownership rules checker populates the graph
- p06-t04: escape analysis queries EscapesTo edges
- p11-t14: axmeta-writer exports the graph to .axmeta
