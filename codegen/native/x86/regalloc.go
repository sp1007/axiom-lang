package x86

// --------------------------------------------------------------------------
// p11-t05: Linear Scan Register Allocation
//
// Allocates physical registers to virtual registers using the linear scan
// algorithm. Intervals sorted by start point are processed left-to-right.
// When no free register is available, the interval with the furthest end
// point is spilled to the stack.
// --------------------------------------------------------------------------

// RegAllocation records the physical register or spill slot for each VReg.
type RegAllocation struct {
	VReg     uint32
	Phys     PhysReg // physical register (RegNone if spilled)
	Spilled  bool
	SpillIdx int     // spill slot index (if spilled)
}

// RegAllocResult contains the full register allocation output.
type RegAllocResult struct {
	Allocs     map[uint32]RegAllocation // VReg → allocation
	SpillCount int                       // number of spill slots needed
}

// LinearScanAlloc performs linear scan register allocation.
// availRegs is the list of registers available for allocation.
func LinearScanAlloc(intervals []LiveInterval, availRegs []PhysReg) RegAllocResult {
	result := RegAllocResult{
		Allocs: make(map[uint32]RegAllocation, len(intervals)),
	}

	if len(intervals) == 0 || len(availRegs) == 0 {
		return result
	}

	// Active intervals (currently live and assigned to a register)
	var active []activeEntry

	// Free register pool
	freeRegs := make([]PhysReg, len(availRegs))
	copy(freeRegs, availRegs)

	for _, interval := range intervals {
		// Expire old intervals
		active, freeRegs = expireOld(active, freeRegs, interval.Start)

		if len(freeRegs) > 0 {
			// Allocate a free register
			reg := freeRegs[len(freeRegs)-1]
			freeRegs = freeRegs[:len(freeRegs)-1]

			result.Allocs[interval.VReg] = RegAllocation{
				VReg: interval.VReg,
				Phys: reg,
			}

			// Insert into active list, maintaining sorted-by-End order
			entry := activeEntry{interval: interval, reg: reg}
			active = insertActive(active, entry)
		} else {
			// Spill: find the active interval with the furthest end
			if len(active) > 0 && active[len(active)-1].interval.End > interval.End {
				// Spill the active interval with furthest end
				spilled := active[len(active)-1]
				active = active[:len(active)-1]

				// Free the spilled register for current interval
				reg := spilled.reg
				result.Allocs[interval.VReg] = RegAllocation{
					VReg: interval.VReg,
					Phys: reg,
				}

				// Mark spilled interval
				result.Allocs[spilled.interval.VReg] = RegAllocation{
					VReg:     spilled.interval.VReg,
					Phys:     RegNone,
					Spilled:  true,
					SpillIdx: result.SpillCount,
				}
				result.SpillCount++

				entry := activeEntry{interval: interval, reg: reg}
				active = insertActive(active, entry)
			} else {
				// Spill the current interval
				result.Allocs[interval.VReg] = RegAllocation{
					VReg:     interval.VReg,
					Phys:     RegNone,
					Spilled:  true,
					SpillIdx: result.SpillCount,
				}
				result.SpillCount++
			}
		}
	}

	return result
}

// expireOld removes intervals that have ended before pos, freeing their registers.
func expireOld(active []activeEntry, freeRegs []PhysReg, pos int) ([]activeEntry, []PhysReg) {
	i := 0
	for i < len(active) {
		if active[i].interval.End < pos {
			freeRegs = append(freeRegs, active[i].reg)
			active = append(active[:i], active[i+1:]...)
		} else {
			i++
		}
	}
	return active, freeRegs
}

type activeEntry struct {
	interval LiveInterval
	reg      PhysReg
}

// insertActive inserts an entry into the active list sorted by End.
func insertActive(active []activeEntry, entry activeEntry) []activeEntry {
	active = append(active, entry)
	// Insertion sort by End
	for i := len(active) - 1; i > 0 && active[i].interval.End < active[i-1].interval.End; i-- {
		active[i], active[i-1] = active[i-1], active[i]
	}
	return active
}
