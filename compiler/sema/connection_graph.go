package sema

import (
	"encoding/json"
	"fmt"
	"strings"
)

// EdgeKind classifies the relationship between two nodes in the ConnectionGraph.
type EdgeKind uint8

const (
	// EdgeOwns indicates that the source node owns the target node.
	// When the source is destroyed, the target is also destroyed.
	EdgeOwns EdgeKind = iota
	// EdgeBorrows indicates that the source borrows (references) the target.
	// The target must outlive the source.
	EdgeBorrows
	// EdgeFlowsTo indicates that data flows from source to target (assignment, return).
	EdgeFlowsTo
	// EdgeEscapesTo indicates that the source escapes to the target scope/global.
	// This prevents stack allocation of the source.
	EdgeEscapesTo
	// EdgeReusedBy indicates CTGC alias reuse: source's memory is reused by target.
	EdgeReusedBy
)

// String returns a human-readable name for the edge kind.
func (ek EdgeKind) String() string {
	switch ek {
	case EdgeOwns:
		return "Owns"
	case EdgeBorrows:
		return "Borrows"
	case EdgeFlowsTo:
		return "FlowsTo"
	case EdgeEscapesTo:
		return "EscapesTo"
	case EdgeReusedBy:
		return "ReusedBy"
	default:
		return fmt.Sprintf("EdgeKind(%d)", ek)
	}
}

// CGNode represents a value or reference in the ConnectionGraph.
// Each variable, temporary, or allocation point becomes a node.
type CGNode struct {
	ID       uint32 `json:"id"`
	SymID    uint32 `json:"sym_id"`    // 0 for temporary/anonymous nodes
	TypeID   uint32 `json:"type_id"`
	IsRef    bool   `json:"is_ref"`    // true if this is a reference node (not a value node)
	Lifetime uint32 `json:"lifetime"`  // scope depth where this node is alive
}

// CGEdge represents a directed relationship between two nodes.
type CGEdge struct {
	From uint32   `json:"from"`
	To   uint32   `json:"to"`
	Kind EdgeKind `json:"kind"`
}

// ConnectionGraph models ownership relationships between values as a directed
// graph with typed edges. It enables the compiler to statically reason about
// memory lifetimes. Each function has its own ConnectionGraph instance.
//
// The graph supports O(V+E) escape detection via DFS and provides efficient
// edge queries through adjacency lists.
type ConnectionGraph struct {
	Nodes     []CGNode           `json:"nodes"`
	Edges     []CGEdge           `json:"edges"`
	adjOut    [][]uint32         // outgoing edge indices per node
	adjIn     [][]uint32         // incoming edge indices per node
	symToNode map[uint32]uint32  // SymID → NodeID
}

// NewConnectionGraph creates a new empty ConnectionGraph.
func NewConnectionGraph() *ConnectionGraph {
	return &ConnectionGraph{
		symToNode: make(map[uint32]uint32),
	}
}

// AddValueNode adds a value node to the graph.
// Returns the node ID.
func (cg *ConnectionGraph) AddValueNode(symID, typeID, lifetime uint32) uint32 {
	id := uint32(len(cg.Nodes))
	cg.Nodes = append(cg.Nodes, CGNode{
		ID:       id,
		SymID:    symID,
		TypeID:   typeID,
		IsRef:    false,
		Lifetime: lifetime,
	})
	cg.adjOut = append(cg.adjOut, nil)
	cg.adjIn = append(cg.adjIn, nil)
	if symID != 0 {
		cg.symToNode[symID] = id
	}
	return id
}

// AddRefNode adds a reference node that points to targetNodeID.
// Returns the node ID of the reference.
func (cg *ConnectionGraph) AddRefNode(targetNodeID uint32) uint32 {
	var typeID, lifetime uint32
	if int(targetNodeID) < len(cg.Nodes) {
		typeID = cg.Nodes[targetNodeID].TypeID
		lifetime = cg.Nodes[targetNodeID].Lifetime
	}
	id := uint32(len(cg.Nodes))
	cg.Nodes = append(cg.Nodes, CGNode{
		ID:       id,
		SymID:    0,
		TypeID:   typeID,
		IsRef:    true,
		Lifetime: lifetime,
	})
	cg.adjOut = append(cg.adjOut, nil)
	cg.adjIn = append(cg.adjIn, nil)
	// Add a Borrows edge from the ref to the target
	cg.AddEdge(id, targetNodeID, EdgeBorrows)
	return id
}

// AddEdge adds a directed edge between two nodes.
func (cg *ConnectionGraph) AddEdge(from, to uint32, kind EdgeKind) {
	edgeIdx := uint32(len(cg.Edges))
	cg.Edges = append(cg.Edges, CGEdge{From: from, To: to, Kind: kind})

	// Grow adjacency lists if necessary
	for uint32(len(cg.adjOut)) <= from {
		cg.adjOut = append(cg.adjOut, nil)
	}
	for uint32(len(cg.adjIn)) <= to {
		cg.adjIn = append(cg.adjIn, nil)
	}

	cg.adjOut[from] = append(cg.adjOut[from], edgeIdx)
	cg.adjIn[to] = append(cg.adjIn[to], edgeIdx)
}

// NodeOfSym returns the node ID for a given symbol ID.
// Returns (nodeID, true) if found, (0, false) otherwise.
func (cg *ConnectionGraph) NodeOfSym(symID uint32) (uint32, bool) {
	id, ok := cg.symToNode[symID]
	return id, ok
}

// OutEdges returns the target node IDs of outgoing edges from nodeID
// filtered by the given edge kind.
func (cg *ConnectionGraph) OutEdges(nodeID uint32, kind EdgeKind) []uint32 {
	if int(nodeID) >= len(cg.adjOut) {
		return nil
	}
	var targets []uint32
	for _, edgeIdx := range cg.adjOut[nodeID] {
		edge := &cg.Edges[edgeIdx]
		if edge.Kind == kind {
			targets = append(targets, edge.To)
		}
	}
	return targets
}

// InEdges returns the source node IDs of incoming edges to nodeID
// filtered by the given edge kind.
func (cg *ConnectionGraph) InEdges(nodeID uint32, kind EdgeKind) []uint32 {
	if int(nodeID) >= len(cg.adjIn) {
		return nil
	}
	var sources []uint32
	for _, edgeIdx := range cg.adjIn[nodeID] {
		edge := &cg.Edges[edgeIdx]
		if edge.Kind == kind {
			sources = append(sources, edge.From)
		}
	}
	return sources
}

// AllOutEdges returns all outgoing edges from nodeID (any kind).
func (cg *ConnectionGraph) AllOutEdges(nodeID uint32) []CGEdge {
	if int(nodeID) >= len(cg.adjOut) {
		return nil
	}
	var edges []CGEdge
	for _, edgeIdx := range cg.adjOut[nodeID] {
		edges = append(edges, cg.Edges[edgeIdx])
	}
	return edges
}

// Escapes returns true if nodeID has any EscapesTo edge, either directly
// or transitively through Owns/FlowsTo chains.
// Uses DFS with a visited set to handle cycles in O(V+E).
func (cg *ConnectionGraph) Escapes(nodeID uint32) bool {
	visited := make(map[uint32]bool)
	return cg.escapeDFS(nodeID, visited)
}

func (cg *ConnectionGraph) escapeDFS(nodeID uint32, visited map[uint32]bool) bool {
	if nodeID == 0 {
		return true
	}
	if visited[nodeID] {
		return false
	}
	visited[nodeID] = true

	if int(nodeID) >= len(cg.adjOut) {
		return false
	}

	for _, edgeIdx := range cg.adjOut[nodeID] {
		edge := &cg.Edges[edgeIdx]
		switch edge.Kind {
		case EdgeEscapesTo:
			return true
		case EdgeOwns, EdgeFlowsTo:
			// Transitive: if what we own/flow to escapes, we escape.
			if cg.escapeDFS(edge.To, visited) {
				return true
			}
		}
	}
	return false
}

// DominatedBy returns true if nodeID's lifetime is within scopeDepth.
// A node is dominated by a scope if its lifetime (scope depth) is >= the scope depth.
func (cg *ConnectionGraph) DominatedBy(nodeID, scopeDepth uint32) bool {
	if int(nodeID) >= len(cg.Nodes) {
		return false
	}
	return cg.Nodes[nodeID].Lifetime >= scopeDepth
}

// NodeCount returns the number of nodes in the graph.
func (cg *ConnectionGraph) NodeCount() int {
	return len(cg.Nodes)
}

// EdgeCount returns the number of edges in the graph.
func (cg *ConnectionGraph) EdgeCount() int {
	return len(cg.Edges)
}

// Reset clears the graph for reuse (e.g., between function bodies).
func (cg *ConnectionGraph) Reset() {
	cg.Nodes = cg.Nodes[:0]
	cg.Edges = cg.Edges[:0]
	cg.adjOut = cg.adjOut[:0]
	cg.adjIn = cg.adjIn[:0]
	for k := range cg.symToNode {
		delete(cg.symToNode, k)
	}
}

// String returns a human-readable debug representation of the graph.
func (cg *ConnectionGraph) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ConnectionGraph: %d nodes, %d edges\n", len(cg.Nodes), len(cg.Edges)))

	sb.WriteString("Nodes:\n")
	for _, n := range cg.Nodes {
		refStr := ""
		if n.IsRef {
			refStr = " [ref]"
		}
		sb.WriteString(fmt.Sprintf("  N%d: sym=%d type=%d lifetime=%d%s\n", n.ID, n.SymID, n.TypeID, n.Lifetime, refStr))
	}

	sb.WriteString("Edges:\n")
	for _, e := range cg.Edges {
		sb.WriteString(fmt.Sprintf("  N%d --%s--> N%d\n", e.From, e.Kind, e.To))
	}
	return sb.String()
}

// MarshalJSON implements JSON serialization for the ConnectionGraph.
func (cg *ConnectionGraph) MarshalJSON() ([]byte, error) {
	type jsonGraph struct {
		Nodes []CGNode `json:"nodes"`
		Edges []CGEdge `json:"edges"`
	}
	return json.Marshal(jsonGraph{Nodes: cg.Nodes, Edges: cg.Edges})
}

// UnmarshalJSON implements JSON deserialization for the ConnectionGraph.
func (cg *ConnectionGraph) UnmarshalJSON(data []byte) error {
	type jsonGraph struct {
		Nodes []CGNode `json:"nodes"`
		Edges []CGEdge `json:"edges"`
	}
	var jg jsonGraph
	if err := json.Unmarshal(data, &jg); err != nil {
		return err
	}

	cg.Reset()
	// Re-add nodes
	for _, n := range jg.Nodes {
		cg.Nodes = append(cg.Nodes, n)
		cg.adjOut = append(cg.adjOut, nil)
		cg.adjIn = append(cg.adjIn, nil)
		if n.SymID != 0 {
			cg.symToNode[n.SymID] = n.ID
		}
	}
	// Re-add edges
	for _, e := range jg.Edges {
		cg.AddEdge(e.From, e.To, e.Kind)
	}
	return nil
}
