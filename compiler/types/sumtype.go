package types

// VariantInfo describes a single variant within a SumType (tagged union).
type VariantInfo struct {
	NameID      uint32 // interned variant name (e.g., "Ok", "Err")
	PayloadType TypeID // TypeID of payload (0 if unit variant)
	Tag         uint8  // numeric tag value (0, 1, 2, ...)
}

// SumType holds the full definition of a tagged union type.
type SumType struct {
	Variants      []VariantInfo
	GenericParams []uint32 // TypeIDs of generic parameters if this is a generic definition
}
