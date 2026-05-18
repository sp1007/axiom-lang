# p16-t15: std.iter — Iterator Protocol

## Purpose
Implement the AXIOM iterator protocol (`Iterator` interface) and adapter library (`map`, `filter`, `reduce`, `enumerate`, `zip`, `take`, `collect`) enabling lazy, composable iteration over collections.

## Context
AXIOM's `for x in collection` syntax desugars to iterator protocol calls. `std.iter` defines the `Iterator` interface and adapter chain so collections can compose iteration logic without intermediate allocations. Lazy evaluation means `map(filter(...))` doesn't create intermediate arrays.

## Inputs
- AXIOM `for` loop desugaring from p03/p04
- `std.collections` Array, HashMap from p16-t03
- Generic interface system from p05-t04

## Outputs
- `stdlib/iter/iter.ax` — Iterator interface + core adapters
- `stdlib/iter/collect.ax` — `collect()` into Array, HashMap, str

## Dependencies
- p05-t04: structural-duck-typing — Iterator satisfied by any type with `next()`
- p16-t03: std-collections — Array has iter() → ArrayIter
- p03-t04: parser-statements — `for x in y` desugaring

## Detailed Requirements

```axiom
# stdlib/iter/iter.ax

interface Iterator:
    type Item
    fn next(mut self) -> Option[Self::Item]

# Core adapters (lazy, no intermediate allocation)
fn map[I: Iterator, B](iter: I, f: fn(I::Item) -> B) -> MapIter[I, B]
fn filter[I: Iterator](iter: I, pred: fn(I::Item) -> bool) -> FilterIter[I]
fn filter_map[I: Iterator, B](iter: I, f: fn(I::Item) -> Option[B]) -> FilterMapIter[I, B]
fn flat_map[I: Iterator, B: Iterator](iter: I, f: fn(I::Item) -> B) -> FlatMapIter[I, B]
fn enumerate[I: Iterator](iter: I) -> EnumerateIter[I]
fn zip[A: Iterator, B: Iterator](a: A, b: B) -> ZipIter[A, B]
fn take[I: Iterator](iter: I, n: u32) -> TakeIter[I]
fn skip[I: Iterator](iter: I, n: u32) -> SkipIter[I]
fn chain[I: Iterator](a: I, b: I) -> ChainIter[I]
fn peekable[I: Iterator](iter: I) -> PeekableIter[I]

# Terminal operations
fn collect[I: Iterator](iter: I) -> Array[I::Item]
fn collect_map[I: Iterator[Item=(K,V)], K: Hash+Eq, V](iter: I) -> HashMap[K, V]
fn count[I: Iterator](iter: I) -> u32
fn sum[I: Iterator[Item: Add]](iter: I) -> I::Item
fn product[I: Iterator[Item: Mul]](iter: I) -> I::Item
fn fold[I: Iterator, B](iter: I, init: B, f: fn(B, I::Item) -> B) -> B
fn reduce[I: Iterator](iter: I, f: fn(I::Item, I::Item) -> I::Item) -> Option[I::Item]
fn find[I: Iterator](iter: I, pred: fn(I::Item) -> bool) -> Option[I::Item]
fn any[I: Iterator](iter: I, pred: fn(I::Item) -> bool) -> bool
fn all[I: Iterator](iter: I, pred: fn(I::Item) -> bool) -> bool
fn for_each[I: Iterator](iter: I, f: fn(I::Item))
```

`for x in collection` desugaring:
```axiom
for x in arr:
    body(x)
# →
var _iter = arr.iter()
loop:
    match _iter.next():
        Some(x) -> body(x)
        None    -> break
```

## Implementation Steps

1. Create `stdlib/iter/iter.ax` — Iterator interface + all adapter types.
2. Implement each adapter as a lazy struct implementing Iterator.
3. Implement terminal operations (`collect`, `count`, `fold`, etc.).
4. Create `stdlib/iter/collect.ax` — collect into Array, HashMap, str.
5. Wire `for x in y` desugaring in compiler to iterator protocol.
6. Write tests chaining multiple adapters.

## Test Plan
- `TestMapFilter`: [1..10].iter().filter(|x| x%2==0).map(|x| x*x) → [4,16,36,64,100]
- `TestEnumerate`: iter.enumerate() → yields (0,a), (1,b), ...
- `TestCollect`: map iter → collect into Array → correct
- `TestZip`: zip([1,2,3], ["a","b","c"]) → [(1,"a"),(2,"b"),(3,"c")]
- `TestFold`: fold([1..5], 0, |acc,x| acc+x) = 15

## Validation Checklist
- [ ] Adapters are lazy (no allocation until terminal op)
- [ ] `for` loop desugaring produces correct Iterator calls
- [ ] Infinite iterators (take + chain) don't allocate infinitely
- [ ] collect() size-hints for pre-allocation when available

## Acceptance Criteria
- `.iter().filter().map().collect()` on 1M elements: no intermediate allocations

## Definition of Done
- [ ] All adapter types implemented
- [ ] `for` loop desugaring verified
- [ ] Adapter chain tests pass

## Risks & Mitigations
| Risk | Mitigation |
|------|-----------|
| Associated type `Item` not yet supported in type system | Add associated type support to interface system in p05 extension |

## Future Follow-up Tasks
- Parallel iterator (`par_iter()`) using work-stealing
- Async iterator for async streams
