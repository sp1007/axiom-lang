package stdlib

// --------------------------------------------------------------------------
// p16-t12: Result[T] and Option[T] Types
//
// Core error-handling types for AXIOM. These are generic sum types:
//   Option[T] = Some(T) | None
//   Result[T, E] = Ok(T) | Err(E)
//
// The compiler generates specialized code for each instantiation.
// --------------------------------------------------------------------------

// StdType represents a standard library type known to the compiler.
type StdType uint16

const (
	TypeNone StdType = iota

	// p16-t12: Core sum types
	TypeOption       // Option[T]
	TypeResult       // Result[T, E]

	// p16-t02: String types
	TypeString       // str (owned, heap-allocated)
	TypeStringSlice  // &str (borrowed string slice)
	TypeStringBuilder // StringBuilder

	// p16-t03: Collection types
	TypeVec          // Vec[T] (dynamic array)
	TypeHashMap      // HashMap[K, V]
	TypeHashSet      // HashSet[T]
	TypeDeque        // Deque[T] (double-ended queue)
	TypeBTreeMap     // BTreeMap[K, V]
	TypeLinkedList   // LinkedList[T]

	// p16-t15: Iterator types
	TypeIterator     // Iterator[T] (trait)
	TypeRange        // Range (start..end)

	// p16-t21: Memory types
	TypeBox          // Box[T] (heap pointer with ownership)
	TypeRc           // Rc[T] (reference-counted)
	TypeArc          // Arc[T] (atomic reference-counted)
	TypeSlice        // Slice[T] (&[T])

	// p16-t04: IO types
	TypeReader       // Reader (trait)
	TypeWriter       // Writer (trait)
	TypeFile         // File

	// p16-t08: Sync types
	TypeMutex        // Mutex[T]
	TypeRwLock       // RwLock[T]
	TypeChannel      // Channel[T]
	TypeAtomic       // Atomic[T]

	// p16-t10: Time types
	TypeDuration     // Duration
	TypeInstant      // Instant
	TypeSystemTime   // SystemTime

	// p16-t09: JSON types
	TypeJsonValue    // JsonValue

	// p16-t13: Log types
	TypeLogLevel     // LogLevel
	TypeLogger       // Logger

	// p16-t14: OS types
	TypePathBuf      // PathBuf

	// p16-t06: Net types
	TypeIpAddr       // IpAddr
	TypeTcpStream    // TcpStream
	TypeTcpListener  // TcpListener
	TypeUdpSocket    // UdpSocket

	// p16-t07: Process types
	TypeCommand      // Command
	TypeChild        // Child

	// p16-t16: Random types
	TypeRng          // Rng

	// p16-t22: SIMD types
	TypeF32x4        // f32x4
	TypeF64x4        // f64x4
	TypeI32x4        // i32x4

	// p16-t23: Compiler types
	TypeTypeInfo     // TypeInfo

	// p16-t24: Quantum types
	TypeQReg         // QReg

	// p16-t25: GPU types
	TypeGpuDevice    // GpuDevice
	TypeGpuBuffer    // GpuBuffer[T]
	TypeGpuKernel    // GpuKernel
)

// StdFunc represents a standard library function known to the compiler.
type StdFunc uint16

const (
	FuncNone StdFunc = iota

	// p16-t01: Testing/Assert
	FuncAssert      // assert(condition)
	FuncAssertEq    // assert_eq(a, b)
	FuncAssertNe    // assert_ne(a, b)
	FuncPanic       // panic(message)

	// p16-t02: String
	FuncStrLen      // str.len()
	FuncStrConcat   // str + str
	FuncStrSlice    // str[start..end]
	FuncStrContains // str.contains(sub)
	FuncStrSplit    // str.split(sep)
	FuncStrTrim     // str.trim()
	FuncStrToUpper  // str.to_upper()
	FuncStrToLower  // str.to_lower()

	// p16-t05: Math
	FuncAbs         // abs(x)
	FuncMin         // min(a, b)
	FuncMax         // max(a, b)
	FuncSqrt        // sqrt(x)
	FuncPow         // pow(base, exp)
	FuncFloor       // floor(x)
	FuncCeil        // ceil(x)

	// p16-t11: Formatting
	FuncPrint       // print(...)
	FuncPrintln     // println(...)
	FuncFormat      // format("...", ...)
	FuncSprint      // sprint(...) -> str

	// p16-t03: Collections
	FuncVecNew      // Vec::new()
	FuncVecPush     // vec.push(item)
	FuncVecPop      // vec.pop()
	FuncVecLen      // vec.len()
	FuncVecGet      // vec[index]
	FuncMapNew      // HashMap::new()
	FuncMapInsert   // map.insert(key, value)
	FuncMapGet      // map.get(key)
	FuncMapContains // map.contains_key(key)
)

// StdTypeInfo holds metadata about a stdlib type.
type StdTypeInfo struct {
	Type       StdType
	Name       string
	Module     string
	IsGeneric  bool
	TypeParams int // number of type parameters
}

// Registry of standard library types.
var StdTypes = []StdTypeInfo{
	{TypeOption, "Option", "std::option", true, 1},
	{TypeResult, "Result", "std::result", true, 2},
	{TypeString, "String", "std::string", false, 0},
	{TypeStringSlice, "str", "std::string", false, 0},
	{TypeStringBuilder, "StringBuilder", "std::string", false, 0},
	{TypeVec, "Vec", "std::collections", true, 1},
	{TypeHashMap, "HashMap", "std::collections", true, 2},
	{TypeHashSet, "HashSet", "std::collections", true, 1},
	{TypeDeque, "Deque", "std::collections", true, 1},
	{TypeBTreeMap, "BTreeMap", "std::collections", true, 2},
	{TypeLinkedList, "LinkedList", "std::collections", true, 1},
	{TypeIterator, "Iterator", "std::iter", true, 1},
	{TypeRange, "Range", "std::iter", false, 0},
	{TypeBox, "Box", "std::mem", true, 1},
	{TypeRc, "Rc", "std::mem", true, 1},
	{TypeArc, "Arc", "std::mem", true, 1},
	{TypeSlice, "Slice", "std::mem", true, 1},
	{TypeReader, "Reader", "std::io", false, 0},
	{TypeWriter, "Writer", "std::io", false, 0},
	{TypeFile, "File", "std::io", false, 0},
	{TypeMutex, "Mutex", "std::sync", true, 1},
	{TypeRwLock, "RwLock", "std::sync", true, 1},
	{TypeChannel, "Channel", "std::sync", true, 1},
	{TypeAtomic, "Atomic", "std::sync", true, 1},
	{TypeDuration, "Duration", "std::time", false, 0},
	{TypeInstant, "Instant", "std::time", false, 0},
	{TypeSystemTime, "SystemTime", "std::time", false, 0},
	{TypeJsonValue, "JsonValue", "std::json", false, 0},
	{TypeLogLevel, "LogLevel", "std::log", false, 0},
	{TypeLogger, "Logger", "std::log", false, 0},
	{TypePathBuf, "PathBuf", "std::os", false, 0},
	{TypeIpAddr, "IpAddr", "std::net", false, 0},
	{TypeTcpStream, "TcpStream", "std::net", false, 0},
	{TypeTcpListener, "TcpListener", "std::net", false, 0},
	{TypeUdpSocket, "UdpSocket", "std::net", false, 0},
	{TypeCommand, "Command", "std::process", false, 0},
	{TypeChild, "Child", "std::process", false, 0},
	{TypeRng, "Rng", "std::random", false, 0},
	{TypeF32x4, "f32x4", "std::arch", false, 0},
	{TypeF64x4, "f64x4", "std::arch", false, 0},
	{TypeI32x4, "i32x4", "std::arch", false, 0},
	{TypeTypeInfo, "TypeInfo", "std::compiler", false, 0},
	{TypeQReg, "QReg", "std::quantum", false, 0},
	{TypeGpuDevice, "GpuDevice", "std::gpu", false, 0},
	{TypeGpuBuffer, "GpuBuffer", "std::gpu", true, 1},
	{TypeGpuKernel, "GpuKernel", "std::gpu", false, 0},
}

// LookupType returns the StdTypeInfo for a given name and module.
func LookupType(name, module string) *StdTypeInfo {
	for i := range StdTypes {
		if StdTypes[i].Name == name && StdTypes[i].Module == module {
			return &StdTypes[i]
		}
	}
	return nil
}

// LookupTypeByName returns the first StdTypeInfo matching the name.
func LookupTypeByName(name string) *StdTypeInfo {
	for i := range StdTypes {
		if StdTypes[i].Name == name {
			return &StdTypes[i]
		}
	}
	return nil
}
