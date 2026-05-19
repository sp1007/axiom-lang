// Package ast defines the AXIOM abstract syntax tree node types.
// AST nodes reference types only by TypeID (uint32), never by
// pointer to TypeInfo, to avoid circular imports with the types package.
package ast
