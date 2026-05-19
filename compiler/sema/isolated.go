package sema

// IsolatedVerifier checks that values of type Isolated[T] have no external
// references — their entire object subgraph is self-contained.
// This guarantees safe zero-copy message passing between actors.
//
// The verifier operates on a populated ConnectionGraph by:
// 1. Collecting the subgraph reachable from a value node via Owns/FlowsTo edges
// 2. Checking all incoming edges to subgraph nodes
// 3. Flagging any incoming edge whose source is outside the subgraph (external ref)
type IsolatedVerifier struct {
	cg *ConnectionGraph
}

// NewIsolatedVerifier creates a new IsolatedVerifier.
func NewIsolatedVerifier(cg *ConnectionGraph) *IsolatedVerifier {
	return &IsolatedVerifier{cg: cg}
}

// VerifyIsolated checks whether the value at nodeID is provably isolated.
// Returns (true, nil) if isolated, or (false, violatingNodeIDs) if not.
// A value is isolated if no nodes outside its owned subgraph have incoming
// edges pointing into it.
func (iv *IsolatedVerifier) VerifyIsolated(nodeID uint32) (bool, []uint32) {
	// 1. Collect the subgraph: all nodes reachable via Owns/FlowsTo from nodeID
	subgraph := make(map[uint32]bool)
	iv.collectSubgraph(nodeID, subgraph)

	// 2. For each node in subgraph, check incoming edges
	var violators []uint32
	for nodeInSubgraph := range subgraph {
		for _, edgeIdx := range iv.incomingEdgeIndices(nodeInSubgraph) {
			edge := &iv.cg.Edges[edgeIdx]
			// External edge: source is NOT in the subgraph
			if !subgraph[edge.From] {
				// Any external reference is a violation
				violators = append(violators, edge.From)
			}
		}
	}

	if len(violators) > 0 {
		return false, unique(violators)
	}
	return true, nil
}

// IsFreshlyAllocated returns true if the node has no incoming Borrows edges.
// A freshly allocated value with no borrows taken is automatically isolated.
func (iv *IsolatedVerifier) IsFreshlyAllocated(nodeID uint32) bool {
	borrows := iv.cg.InEdges(nodeID, EdgeBorrows)
	return len(borrows) == 0
}

// collectSubgraph performs DFS from nodeID, following Owns and FlowsTo edges,
// and adds all reachable nodes to the subgraph set.
func (iv *IsolatedVerifier) collectSubgraph(nodeID uint32, subgraph map[uint32]bool) {
	if subgraph[nodeID] {
		return // already visited
	}
	subgraph[nodeID] = true

	if int(nodeID) >= len(iv.cg.adjOut) {
		return
	}

	for _, edgeIdx := range iv.cg.adjOut[nodeID] {
		edge := &iv.cg.Edges[edgeIdx]
		if edge.Kind == EdgeOwns || edge.Kind == EdgeFlowsTo {
			iv.collectSubgraph(edge.To, subgraph)
		}
	}
}

// incomingEdgeIndices returns the raw edge indices for incoming edges to nodeID.
func (iv *IsolatedVerifier) incomingEdgeIndices(nodeID uint32) []uint32 {
	if int(nodeID) >= len(iv.cg.adjIn) {
		return nil
	}
	return iv.cg.adjIn[nodeID]
}

// unique deduplicates a slice of uint32.
func unique(ids []uint32) []uint32 {
	seen := make(map[uint32]bool, len(ids))
	result := make([]uint32, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}
