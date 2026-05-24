package main

import (
	"fmt"
	"os"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
)

func main() {
	data, err := os.ReadFile("scratch/self_linked_concatenated.ax")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	tokens, _, _ := lexer.Lex(data)
	pool := ast.NewInternPool(1024)
	tree, _ := parser.Parse(tokens, data, pool)

	nodeIdx := uint32(1071)
	dumpNode(tree, nodeIdx, "")
}

func dumpNode(tree *ast.AstTree, nodeIdx uint32, indent string) {
	if nodeIdx == 0 {
		return
	}
	node := tree.Nodes[nodeIdx]
	text := tree.NodeText(nodeIdx)
	fmt.Printf("%sNode %d: Kind=%v, TokenIdx=%d, Payload=%d, Flags=%d, Text=%q\n",
		indent, nodeIdx, node.Kind, node.TokenIdx, node.Payload, node.Flags, string(text))

	child := node.FirstChild
	for child != 0 {
		cNode := tree.Nodes[child]
		dumpNode(tree, child, indent+"  ")
		child = cNode.NextSibling
	}
}
