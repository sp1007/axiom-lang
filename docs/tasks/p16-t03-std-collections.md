# p16-t03: std.collections — Array, Slice, HashMap, Set

## Purpose
Implement core collection types: dynamic `Array[T]`, `Slice[T]`, `HashMap[K, V]`, and `HashSet[T]` as the fundamental data structures for AXIOM programs.

## Context
AXIOM needs generic, efficient collection types backed by the AxAlloc allocator. These collections form the foundation of most programs. `Array[T]` is a growable vector; `HashMap[K, V]` uses open addressing with Robin Hood hashing for cache-friendly performance.

## Inputs
- AXIOM generics from p05-t01 and p05-t02
- AxAlloc from p14 (backing allocator)
- Hash trait/interface for HashMap key requirements

## Outputs
- `stdlib/collections/array.ax` — dynamic array
- `stdlib/collections/slice.ax` — slice (view into array)
- `stdlib/collections/hashmap.ax` — Robin Hood hash map
- `stdlib/collections/hashset.ax` — hash set (wraps HashMap)

## Dependencies
- p05-t02: monomorphization — generates concrete Array[i32], Array[str], etc.
- p14-t01: axalloc — backing allocator for growable arrays
- p04-t08: overload-resolution — Hash interface

## Detailed Requirements

```axiom
# stdlib/collections/array.ax
type Array[T]:
    var ptr: *T
    var len: u32
    var cap: u32

    fn new() -> Array[T]
    fn with_capacity(cap: u32) -> Array[T]
    fn push(mut self, val: T)           # amortized O(1)
    fn pop(mut self) -> Option[T]
    fn get(self, idx: u32) -> Option[T]
    fn set(mut self, idx: u32, val: T)
    fn len(self) -> u32
    fn is_empty(self) -> bool
    fn slice(self, start: u32, end: u32) -> Slice[T]
    fn iter(self) -> ArrayIter[T]
    fn sort(mut self) where T: Ord

# stdlib/collections/hashmap.ax
type HashMap[K: Hash + Eq, V]:
    var buckets: Array[Option[(K, V)]]
    var count: u32
    var load_factor: f32    # default 0.75

    fn new() -> HashMap[K, V]
    fn insert(mut self, key: K, val: V)
    fn get(self, key: K) -> Option[V]
    fn get_mut(mut self, key: K) -> Option[*V]
    fn remove(mut self, key: K) -> Option[V]
    fn contains_key(self, key: K) -> bool
    fn len(self) -> u32
    fn iter(self) -> HashMapIter[K, V]
    fn keys(self) -> KeyIter[K]
    fn values(self) -> ValueIter[V]
```

Array growth: double capacity when `len == cap`. Initial capacity: 8.

HashMap: Robin Hood open addressing. Probe sequence: linear. Load factor: 0.75.

Hash interface:
```axiom
interface Hash:
    fn hash(self) -> u64

# Built-in implementations for primitives
impl Hash for i32: fn hash(self) -> u64 { ... }
impl Hash for str:  fn hash(self) -> u64 { ... }  # FNV-1a
```

## Implementation Steps

1. Create `stdlib/collections/array.ax` — growable Array[T].
2. Implement growth strategy: double-on-full with AxAlloc realloc.
3. Create `stdlib/collections/hashmap.ax` — Robin Hood open addressing.
4. Implement FNV-1a hash for string keys.
5. Create `stdlib/collections/hashset.ax` — wraps HashMap[T, ()].
6. Write comprehensive tests for all collection types.

## Test Plan
- `TestArrayPushPop`: push 1000 elements, pop all → correct LIFO order
- `TestArrayGrowth`: push beyond capacity → no panic, data preserved
- `TestHashMapInsertGet`: insert 1000 k/v pairs, get all → correct values
- `TestHashMapCollision`: hash collision handled via Robin Hood
- `TestHashSetContains`: insert + contains + remove cycle

## Validation Checklist
- [ ] Array bounds check on get/set (panic on out-of-range)
- [ ] HashMap handles load > 0.75 by resizing
- [ ] HashMap iteration stable (no skip/duplicate under read-only)
- [ ] Monomorphization produces separate code for Array[i32] vs Array[str]

## Acceptance Criteria
- HashMap with 1M string keys: lookup < 100ns average

## Definition of Done
- [ ] All collection types implemented
- [ ] All tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Robin Hood deletion complex | Use tombstone deletion for MVP; switch to backshift deletion later |

## Future Follow-up Tasks
- BTreeMap[K,V] for ordered iteration
- Deque[T] for efficient front/back operations
