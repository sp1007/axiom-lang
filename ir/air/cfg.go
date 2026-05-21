package air

// ---------------------------------------------------------------------------
// BasicBlock — a contiguous sequence of instructions ending with a terminator.
// ---------------------------------------------------------------------------

// BasicBlock represents a node in the control-flow graph.
// Instructions are stored by index into the parent AirFunc.Insts slice.
type BasicBlock struct {
	ID        uint32   // unique block ID within the function
	Instrs    []uint32 // instruction indices into AirFunc.Insts
	Succs     []uint32 // successor block IDs
	Preds     []uint32 // predecessor block IDs
	LoopDepth uint8    // nesting depth (0 = not in a loop)
	IsEntry   bool     // true if this is the function entry block
	IsExit    bool     // true if this block ends with a return
}

// ---------------------------------------------------------------------------
// AirFunc — one function in the AIR module.
// ---------------------------------------------------------------------------

// AirFunc is the AIR representation of a single function. It owns all
// basic blocks and instructions. The instruction array is flat; blocks
// reference into it via index slices.
type AirFunc struct {
	SymID    uint32       // symbol table ID
	Name     uint32       // interned name
	Params   []uint32     // parameter TypeIDs
	RetType  uint32       // return type ID
	Blocks   []BasicBlock // basic blocks (index 0 is always the entry)
	Insts    []AirInst    // flat instruction array
	Extras   []uint32     // extra operands for variadic instructions (call args, phi edges, etc.)
	IsAsync  bool         // true if this is an async function
	IsExtern bool         // true if this is an extern declaration (no body)
}

// ---------------------------------------------------------------------------
// AirModule — top-level container for a translation unit.
// ---------------------------------------------------------------------------

// AirModule collects all functions in one compilation unit.
// The TypeTable and InternPool are external (not owned by AirModule).
type AirModule struct {
	Funcs []AirFunc
}

// ---------------------------------------------------------------------------
// AirFuncBuilder — incremental construction of an AirFunc.
// ---------------------------------------------------------------------------

// AirFuncBuilder provides a convenient API for emitting instructions
// into basic blocks while maintaining SSA value numbering.
type AirFuncBuilder struct {
	name     uint32
	retType  uint32
	blocks   []BasicBlock
	insts    []AirInst
	extras   []uint32
	curBlock int    // index into blocks, -1 if none
	nextReg  uint32 // next available SSA register ID
}

// NewAirFuncBuilder creates a builder for a function with the given
// interned name and return type. An entry block (ID 0) is created
// automatically and set as the current insertion point.
func NewAirFuncBuilder(name uint32, retType uint32) *AirFuncBuilder {
	b := &AirFuncBuilder{
		name:     name,
		retType:  retType,
		curBlock: -1,
		nextReg:  1, // register 0 is reserved (NoValue)
	}
	entryID := b.NewBlock()
	b.blocks[entryID].IsEntry = true
	b.SwitchTo(entryID)
	return b
}

// NewBlock allocates a new empty basic block and returns its ID.
func (b *AirFuncBuilder) NewBlock() uint32 {
	id := uint32(len(b.blocks))
	b.blocks = append(b.blocks, BasicBlock{ID: id})
	return id
}

// SwitchTo sets the active block for subsequent Emit calls.
func (b *AirFuncBuilder) SwitchTo(blockID uint32) {
	b.curBlock = int(blockID)
}

// CurrentBlock returns the ID of the block that Emit will append to.
// Returns -1 (as uint32 max) if no block is active.
func (b *AirFuncBuilder) CurrentBlock() uint32 {
	if b.curBlock < 0 {
		return ^uint32(0)
	}
	return uint32(b.curBlock)
}

// Emit appends an instruction to the current block and returns the
// instruction index within AirFunc.Insts.
func (b *AirFuncBuilder) Emit(inst AirInst) uint32 {
	idx := uint32(len(b.insts))
	b.insts = append(b.insts, inst)
	if b.curBlock >= 0 && b.curBlock < len(b.blocks) {
		b.blocks[b.curBlock].Instrs = append(b.blocks[b.curBlock].Instrs, idx)
	}
	return idx
}

// EmitExtra appends an extra operand (for variadic instructions) and
// returns its index in the extras slice.
func (b *AirFuncBuilder) EmitExtra(val uint32) uint32 {
	idx := uint32(len(b.extras))
	b.extras = append(b.extras, val)
	return idx
}

// SetExtra updates an existing extra operand at the given index.
func (b *AirFuncBuilder) SetExtra(idx uint32, val uint32) {
	if int(idx) < len(b.extras) {
		b.extras[idx] = val
	}
}

// FreshReg allocates a new SSA register ID.
func (b *AirFuncBuilder) FreshReg() uint32 {
	r := b.nextReg
	b.nextReg++
	return r
}

// AddEdge records a control-flow edge from src block to dst block,
// updating both successor and predecessor lists.
func (b *AirFuncBuilder) AddEdge(src, dst uint32) {
	if int(src) < len(b.blocks) && int(dst) < len(b.blocks) {
		b.blocks[src].Succs = appendUnique(b.blocks[src].Succs, dst)
		b.blocks[dst].Preds = appendUnique(b.blocks[dst].Preds, src)
	}
}

// Build finalizes the builder and returns the completed AirFunc.
// After Build, the builder should not be reused.
func (b *AirFuncBuilder) Build() *AirFunc {
	// Mark exit blocks.
	for i := range b.blocks {
		blk := &b.blocks[i]
		if len(blk.Instrs) > 0 {
			lastIdx := blk.Instrs[len(blk.Instrs)-1]
			if int(lastIdx) < len(b.insts) && b.insts[lastIdx].Opcode == OpReturn {
				blk.IsExit = true
			}
		}
	}

	return &AirFunc{
		Name:    b.name,
		RetType: b.retType,
		Blocks:  b.blocks,
		Insts:   b.insts,
		Extras:  b.extras,
	}
}

// ---------------------------------------------------------------------------
// CFG traversal utilities
// ---------------------------------------------------------------------------

// PostOrder returns block IDs in post-order (children before parents).
// This is the standard depth-first post-order useful for dataflow analysis.
func (f *AirFunc) PostOrder() []uint32 {
	if len(f.Blocks) == 0 {
		return nil
	}
	visited := make([]bool, len(f.Blocks))
	result := make([]uint32, 0, len(f.Blocks))

	var dfs func(id uint32)
	dfs = func(id uint32) {
		if visited[id] {
			return
		}
		visited[id] = true
		for _, succ := range f.Blocks[id].Succs {
			dfs(succ)
		}
		result = append(result, id)
	}
	dfs(0) // start from entry block
	return result
}

// ReversePostOrder returns block IDs in reverse post-order.
// This is the natural iteration order for forward dataflow problems.
func (f *AirFunc) ReversePostOrder() []uint32 {
	po := f.PostOrder()
	n := len(po)
	rpo := make([]uint32, n)
	for i := 0; i < n; i++ {
		rpo[i] = po[n-1-i]
	}
	return rpo
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func appendUnique(s []uint32, v uint32) []uint32 {
	for _, e := range s {
		if e == v {
			return s
		}
	}
	return append(s, v)
}
