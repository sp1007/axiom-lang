package main

import (
	"fmt"
	"os"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
)

func main() {
	data, err := os.ReadFile("std/collections.ax")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	tokens, _, _ := lexer.Lex(data)
	pool := ast.NewInternPool(1024)
	tree, _ := parser.Parse(tokens, data, pool)

	dumpNode(tree, 1342, "")
}

func findNodesByText(tree *ast.AstTree, text string) {
	fmt.Printf("Searching for nodes with text %q:\n", text)
	for i := 0; i < len(tree.Nodes); i++ {
		t := string(tree.NodeText(uint32(i)))
		if t == text {
			fmt.Printf("  Node %d: Kind=%v, TokenIdx=%d, Payload=%d, Flags=%d, Text=%q\n",
				i, tree.Nodes[i].Kind, tree.Nodes[i].TokenIdx, tree.Nodes[i].Payload, tree.Nodes[i].Flags, t)
			// Print parent
			var parent uint32 = 0
			for p := 0; p < len(tree.Nodes); p++ {
				c := tree.Nodes[p].FirstChild
				for c != 0 {
					if c == uint32(i) {
						parent = uint32(p)
						break
					}
					c = tree.Nodes[c].NextSibling
				}
				if parent != 0 {
					break
				}
			}
			if parent != 0 {
				fmt.Printf("    Parent Node %d: Kind=%v\n", parent, tree.Nodes[parent].Kind)
			}
		}
	}
}

func findParentAndDump(tree *ast.AstTree, target uint32) {
	var parent uint32 = 0
	for i := 0; i < len(tree.Nodes); i++ {
		node := tree.Nodes[i]
		child := node.FirstChild
		for child != 0 {
			if child == target {
				parent = uint32(i)
				break
			}
			child = tree.Nodes[child].NextSibling
		}
		if parent != 0 {
			break
		}
	}
	if parent != 0 {
		fmt.Printf("Parent of Node %d is Node %d:\n", target, parent)
		dumpNode(tree, parent, "  ")
	} else {
		fmt.Printf("Node %d has no parent (is root?)\n", target)
		dumpNode(tree, target, "  ")
	}
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
