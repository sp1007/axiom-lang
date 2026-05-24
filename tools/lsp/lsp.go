// Package lsp implements the AXIOM Language Server Protocol server.
// It provides IDE features including diagnostics, hover, and go-to-definition.
package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/axiom-lang/axiom/compiler/ast"
	"github.com/axiom-lang/axiom/compiler/diagnostics"
	"github.com/axiom-lang/axiom/compiler/lexer"
	"github.com/axiom-lang/axiom/compiler/parser"
	"github.com/axiom-lang/axiom/compiler/types"
	"github.com/axiom-lang/axiom/compiler/sema"
)

// Global state for document tracking
var (
	filesMu sync.RWMutex
	files   = make(map[string]string)
	
	// Cache for type checking results per file URI
	cacheMu sync.RWMutex
	pools   = make(map[string]*ast.InternPool)
	trees   = make(map[string]*ast.AstTree)
	symtabs = make(map[string]*sema.SymbolTable)
	ttables = make(map[string]*types.TypeTable)
	infers  = make(map[string]*sema.InferenceEngine)

	writeMu   sync.Mutex
	lspWriter io.Writer
)

// LSP JSON-RPC structures
type rawMessage struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *LspError   `json:"error,omitempty"`
}

type LspError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Notification struct {
	Jsonrpc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

type Position struct {
	Line      int `json:"line"`      // 0-based
	Character int `json:"character"` // 0-based
}

type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

type Diagnostic struct {
	Range    Range  `json:"range"`
	Severity int    `json:"severity"` // 1 = Error, 2 = Warning
	Code     string `json:"code,omitempty"`
	Source   string `json:"source,omitempty"`
	Message  string `json:"message"`
}

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type TextDocumentContentChangeEvent struct {
	Text string `json:"text"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type HoverParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type DefinitionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

type MarkupContent struct {
	Kind  string `json:"kind"` // "markdown"
	Value string `json:"value"`
}

type HoverResult struct {
	Contents MarkupContent `json:"contents"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

type ServerCapabilities struct {
	TextDocumentSync           int  `json:"textDocumentSync"` // 1 = Full sync
	HoverProvider              bool `json:"hoverProvider"`
	DefinitionProvider         bool `json:"definitionProvider"`
}

// StartServer launches the LSP server on standard input/output.
func StartServer() int {
	log.SetOutput(os.Stderr)
	log.Println("LSP: Starting AXIOM language server...")

	// Redirect os.Stdout to os.Stderr to prevent any debug prints from corrupting the LSP stream
	originalStdout := os.Stdout
	os.Stdout = os.Stderr
	lspWriter = originalStdout

	reader := bufio.NewReader(os.Stdin)

	for {
		// Read headers
		var contentLength int
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					log.Println("LSP: EOF reached on stdin, shutting down")
					return 0
				}
				log.Printf("LSP: error reading header: %v", err)
				return 1
			}
			line = strings.TrimSpace(line)
			if line == "" {
				// Empty line indicates end of headers
				break
			}
			if strings.HasPrefix(line, "Content-Length:") {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					val, err := strconv.Atoi(strings.TrimSpace(parts[1]))
					if err == nil {
						contentLength = val
					}
				}
			}
		}

		if contentLength <= 0 {
			continue
		}

		// Read payload body
		body := make([]byte, contentLength)
		_, err := io.ReadFull(reader, body)
		if err != nil {
			log.Printf("LSP: error reading message body: %v", err)
			return 1
		}

		// Process message
		var msg rawMessage
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Printf("LSP: failed to unmarshal message: %v", err)
			continue
		}

		if msg.Method != "" {
			if msg.ID != nil {
				// Request
				handleRequest(&msg)
			} else {
				// Notification
				handleNotification(&msg)
			}
		}
	}
}

func writeResponse(id json.RawMessage, result interface{}, lspErr *LspError) {
	resp := Response{
		Jsonrpc: "2.0",
		ID:      id,
		Result:  result,
		Error:   lspErr,
	}
	bytesVal, err := json.Marshal(resp)
	if err != nil {
		log.Printf("LSP: failed to marshal response: %v", err)
		return
	}
	writeMessage(bytesVal)
}

func sendNotification(method string, params interface{}) {
	notif := Notification{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
	}
	bytesVal, err := json.Marshal(notif)
	if err != nil {
		log.Printf("LSP: failed to marshal notification: %v", err)
		return
	}
	writeMessage(bytesVal)
}

func writeMessage(bytesVal []byte) {
	writeMu.Lock()
	defer writeMu.Unlock()
	fmt.Fprintf(lspWriter, "Content-Length: %d\r\n\r\n%s", len(bytesVal), string(bytesVal))
}

func handleRequest(msg *rawMessage) {
	var id interface{}
	_ = json.Unmarshal(msg.ID, &id)

	switch msg.Method {
	case "initialize":
		result := InitializeResult{
			Capabilities: ServerCapabilities{
				TextDocumentSync:           1, // Full document sync
				HoverProvider:              true,
				DefinitionProvider:         true,
			},
		}
		writeResponse(msg.ID, result, nil)

	case "shutdown":
		writeResponse(msg.ID, struct{}{}, nil)

	case "textDocument/hover":
		var params HoverParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			writeResponse(msg.ID, nil, &LspError{Code: -32602, Message: "Invalid params"})
			return
		}
		
		filesMu.RLock()
		content, ok := files[params.TextDocument.URI]
		filesMu.RUnlock()
		
		if !ok {
			writeResponse(msg.ID, nil, nil)
			return
		}

		cacheMu.RLock()
		pool := pools[params.TextDocument.URI]
		tree := trees[params.TextDocument.URI]
		symtab := symtabs[params.TextDocument.URI]
		tt := ttables[params.TextDocument.URI]
		infer := infers[params.TextDocument.URI]
		cacheMu.RUnlock()

		if pool == nil || tree == nil || symtab == nil || tt == nil {
			writeResponse(msg.ID, nil, nil)
			return
		}

		result := handleHover(content, pool, tree, symtab, tt, infer, params.Position.Line, params.Position.Character)
		writeResponse(msg.ID, result, nil)

	case "textDocument/definition":
		var params DefinitionParams
		if err := json.Unmarshal(msg.Params, &params); err != nil {
			writeResponse(msg.ID, nil, &LspError{Code: -32602, Message: "Invalid params"})
			return
		}
		
		filesMu.RLock()
		content, ok := files[params.TextDocument.URI]
		filesMu.RUnlock()
		
		if !ok {
			writeResponse(msg.ID, nil, nil)
			return
		}

		cacheMu.RLock()
		pool := pools[params.TextDocument.URI]
		tree := trees[params.TextDocument.URI]
		symtab := symtabs[params.TextDocument.URI]
		tt := ttables[params.TextDocument.URI]
		infer := infers[params.TextDocument.URI]
		cacheMu.RUnlock()

		if pool == nil || tree == nil || symtab == nil || tt == nil {
			writeResponse(msg.ID, nil, nil)
			return
		}

		result := handleDefinition(content, pool, tree, symtab, tt, infer, params.Position.Line, params.Position.Character, params.TextDocument.URI)
		writeResponse(msg.ID, result, nil)

	default:
		writeResponse(msg.ID, nil, &LspError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", msg.Method)})
	}
}

func handleNotification(msg *rawMessage) {
	switch msg.Method {
	case "initialized":
		log.Println("LSP: client initialized successfully")

	case "exit":
		log.Println("LSP: exiting server process")
		os.Exit(0)

	case "textDocument/didOpen":
		var params DidOpenTextDocumentParams
		if err := json.Unmarshal(msg.Params, &params); err == nil {
			filesMu.Lock()
			files[params.TextDocument.URI] = params.TextDocument.Text
			filesMu.Unlock()
			publishDiagnosticsFor(params.TextDocument.URI, params.TextDocument.Text)
		}

	case "textDocument/didChange":
		var params DidChangeTextDocumentParams
		if err := json.Unmarshal(msg.Params, &params); err == nil {
			if len(params.ContentChanges) > 0 {
				filesMu.Lock()
				files[params.TextDocument.URI] = params.ContentChanges[0].Text
				filesMu.Unlock()
				publishDiagnosticsFor(params.TextDocument.URI, params.ContentChanges[0].Text)
			}
		}

	case "textDocument/didSave":
		var params DidSaveTextDocumentParams
		if err := json.Unmarshal(msg.Params, &params); err == nil {
			filesMu.RLock()
			content := files[params.TextDocument.URI]
			filesMu.RUnlock()
			publishDiagnosticsFor(params.TextDocument.URI, content)
		}

	case "textDocument/didClose":
		var params DidCloseTextDocumentParams
		if err := json.Unmarshal(msg.Params, &params); err == nil {
			filesMu.Lock()
			delete(files, params.TextDocument.URI)
			filesMu.Unlock()

			cacheMu.Lock()
			delete(pools, params.TextDocument.URI)
			delete(trees, params.TextDocument.URI)
			delete(symtabs, params.TextDocument.URI)
			delete(ttables, params.TextDocument.URI)
			delete(infers, params.TextDocument.URI)
			cacheMu.Unlock()
		}
	}
}

func publishDiagnosticsFor(uri string, content string) {
	log.Printf("LSP: analyzing and type checking file %s", uri)
	lspDiags, pool, tree, symtab, tt, infer, err := runAnalysis(uri, content)
	if err != nil {
		log.Printf("LSP: analysis failed for %s: %v", uri, err)
		return
	}

	// Update in-memory caches for Hoover and Definition resolution
	cacheMu.Lock()
	pools[uri] = pool
	trees[uri] = tree
	symtabs[uri] = symtab
	ttables[uri] = tt
	infers[uri] = infer
	cacheMu.Unlock()

	// Send diagnostics notification
	sendNotification("textDocument/publishDiagnostics", PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: lspDiags,
	})
}

// runAnalysis runs the full front-end parser/type-checker pipeline on the in-memory source.
func runAnalysis(uri string, content string) ([]Diagnostic, *ast.InternPool, *ast.AstTree, *sema.SymbolTable, *types.TypeTable, *sema.InferenceEngine, error) {
	src := []byte(content)
	tokens, lt, lexErrs := lexer.Lex(src)
	var allDiags []diagnostics.Diagnostic
	allDiags = append(allDiags, lexErrs...)

	pool := ast.NewInternPool(1024)
	tree, parseErrs := parser.Parse(tokens, src, pool)
	allDiags = append(allDiags, parseErrs...)

	symtab := sema.NewSymbolTable(pool)
	tt := types.NewTypeTable()
	var infer *sema.InferenceEngine

	if !hasFatal(parseErrs) {
		resolver := sema.NewNameResolver(tree, pool, symtab, tt, nil)
		resolveErrs := resolver.Resolve()
		allDiags = append(allDiags, resolveErrs...)

		if !hasFatal(allDiags) {
			infer = sema.NewInferenceEngine(tree, symtab, tt, nil)
			inferErrs := infer.Infer()
			allDiags = append(allDiags, inferErrs...)

			if !hasFatal(inferErrs) {
				tc := sema.NewTypeChecker(tree, pool, symtab, tt, infer)
				typeErrs := tc.Check()
				allDiags = append(allDiags, typeErrs...)

				effects := sema.NewEffectChecker(tree, pool, symtab, tt, infer)
				effectErrs := effects.Check()
				allDiags = append(allDiags, effectErrs...)
			}
		}
	}

	var lspDiags []Diagnostic
	for _, d := range allDiags {
		line := d.Pos.Line
		col := d.Pos.Col

		if (line == 0 || col == 0) && lt != nil {
			line, col = lt.LineCol(d.Pos.Offset)
		}

		l := int(line) - 1
		c := int(col) - 1
		if l < 0 {
			l = 0
		}
		if c < 0 {
			c = 0
		}

		lenVal := 1
		if lt != nil {
			for _, t := range tree.Tokens {
				if t.Offset == d.Pos.Offset {
					lenVal = int(t.Len)
					break
				}
			}
		}

		severity := 1 // Error
		if d.Severity == diagnostics.SeverityWarning {
			severity = 2 // Warning
		}

		lspDiags = append(lspDiags, Diagnostic{
			Range: Range{
				Start: Position{Line: l, Character: c},
				End:   Position{Line: l, Character: c + lenVal},
			},
			Severity: severity,
			Code:     fmt.Sprintf("E%d", d.Code),
			Source:   "axc",
			Message:  d.Message,
		})
	}

	return lspDiags, pool, tree, symtab, tt, infer, nil
}

func hasFatal(diags []diagnostics.Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == diagnostics.SeverityError {
			return true
		}
	}
	return false
}

func OffsetFromLineCol(content string, line, col int) int {
	currLine := 0
	currCol := 0
	for i, b := range []byte(content) {
		if currLine == line && currCol == col {
			return i
		}
		if b == '\n' {
			currLine++
			currCol = 0
		} else {
			currCol++
		}
	}
	return len(content)
}

func findNodeAtOffset(tree *ast.AstTree, offset int) uint32 {
	if tree == nil {
		return 0
	}
	var bestNode uint32 = 0
	var smallestLen uint32 = 100000000

	for i, node := range tree.Nodes {
		if i == 0 || node.Kind == ast.NodeInvalid || node.Kind == ast.NodeError {
			continue
		}
		if int(node.TokenIdx) >= len(tree.Tokens) {
			continue
		}
		tok := tree.Tokens[node.TokenIdx]
		start := int(tok.Offset)
		end := start + int(tok.Len)

		if offset >= start && offset <= end {
			if tok.Len < uint16(smallestLen) {
				smallestLen = uint32(tok.Len)
				bestNode = uint32(i)
			}
		}
	}
	return bestNode
}

func declLocation(tree *ast.AstTree, nodeIdx uint32, uri string) *Location {
	tok := tree.Tokens[tree.Nodes[nodeIdx].TokenIdx]
	line, col := LineColFromOffset(tree.Source, tok.Offset)
	return &Location{
		URI: uri,
		Range: Range{
			Start: Position{Line: int(line) - 1, Character: int(col) - 1},
			End:   Position{Line: int(line) - 1, Character: int(col) - 1 + int(tok.Len)},
		},
	}
}

func LineColFromOffset(src []byte, offset uint32) (line, col uint32) {
	line = 1
	lineStart := uint32(0)
	for i := uint32(0); i < offset && i < uint32(len(src)); i++ {
		if src[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}
	col = offset - lineStart + 1
	return
}

func handleDefinition(content string, pool *ast.InternPool, tree *ast.AstTree, symtab *sema.SymbolTable, tt *types.TypeTable, infer *sema.InferenceEngine, line, col int, uri string) interface{} {
	offset := OffsetFromLineCol(content, line, col)
	nodeIdx := findNodeAtOffset(tree, offset)
	if nodeIdx == 0 {
		return nil
	}

	node := tree.Node(nodeIdx)

	// Case 1: Simple symbol resolve
	if node.Kind == ast.NodeIdent {
		symIdx := node.Payload
		if symIdx != 0 && int(symIdx) < len(symtab.Symbols) {
			sym := symtab.SymbolAt(symIdx)
			if sym.DeclNode != 0 {
				return declLocation(tree, sym.DeclNode, uri)
			}
		}
	}

	// Case 2: Field reference
	if node.Kind == ast.NodeIdent && infer != nil {
		for parentIdx, parent := range tree.Nodes {
			if parent.Kind == ast.NodeFieldExpr {
				children := tree.Children(uint32(parentIdx))
				if len(children) == 2 && children[1] == nodeIdx {
					left := children[0]
					leftType := infer.TypeOf(left)
					if leftType != types.TypeUnknown {
						entry := tt.Entry(leftType)
						if entry.Kind == types.KindPointer {
							leftType = tt.PointerElem(leftType)
							entry = tt.Entry(leftType)
						}
						if entry.Kind == types.KindGenericInst {
							for idx := 0; idx < tt.Count(); idx++ {
								e := tt.Entry(types.TypeID(idx))
								if (e.Kind == types.KindStruct || e.Kind == types.KindSum) && e.NameID == entry.NameID {
									entry = e
									break
								}
							}
						}
						if entry.Kind == types.KindStruct {
							var structDeclNode uint32 = 0
							for sIdx, sNode := range tree.Nodes {
								if sNode.Kind == ast.NodeStructDecl && sNode.Payload != 0 {
									sym := symtab.SymbolAt(sNode.Payload)
									if sym.NameID == entry.NameID {
										structDeclNode = uint32(sIdx)
										break
									}
								}
							}

							if structDeclNode != 0 {
								fieldNameID := parent.Payload
								fieldChildren := tree.Children(structDeclNode)
								for _, fc := range fieldChildren {
									fcNode := tree.Node(fc)
									if fcNode.Kind == ast.NodeFieldDecl {
										symIdx := fcNode.Payload
										if symIdx != 0 && int(symIdx) < len(symtab.Symbols) {
											sym := symtab.SymbolAt(symIdx)
											if sym.NameID == fieldNameID {
												return declLocation(tree, fc, uri)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func handleHover(content string, pool *ast.InternPool, tree *ast.AstTree, symtab *sema.SymbolTable, tt *types.TypeTable, infer *sema.InferenceEngine, line, col int) interface{} {
	offset := OffsetFromLineCol(content, line, col)
	nodeIdx := findNodeAtOffset(tree, offset)
	if nodeIdx == 0 {
		return nil
	}

	node := tree.Node(nodeIdx)
	var typeID types.TypeID = types.TypeUnknown
	var name string = ""
	var details string = ""

	// Case 1: Simple symbol lookup
	if node.Kind == ast.NodeIdent {
		symIdx := node.Payload
		if symIdx != 0 && int(symIdx) < len(symtab.Symbols) {
			sym := symtab.SymbolAt(symIdx)
			typeID = types.TypeID(sym.TypeID)
			name = string(pool.GetBytes(sym.NameID))
			details = fmt.Sprintf("(%s) %s", sym.Kind.String(), name)
		}
	}

	// Case 2: Field access lookup
	if node.Kind == ast.NodeIdent && infer != nil {
		for parentIdx, parent := range tree.Nodes {
			if parent.Kind == ast.NodeFieldExpr {
				children := tree.Children(uint32(parentIdx))
				if len(children) == 2 && children[1] == nodeIdx {
					left := children[0]
					leftType := infer.TypeOf(left)
					if leftType != types.TypeUnknown {
						entry := tt.Entry(leftType)
						if entry.Kind == types.KindPointer {
							leftType = tt.PointerElem(leftType)
							entry = tt.Entry(leftType)
						}
						structType := leftType
						if entry.Kind == types.KindGenericInst {
							for idx := 0; idx < tt.Count(); idx++ {
								e := tt.Entry(types.TypeID(idx))
								if (e.Kind == types.KindStruct || e.Kind == types.KindSum) && e.NameID == entry.NameID {
									structType = types.TypeID(idx)
									entry = e
									break
								}
							}
						}
						if entry.Kind == types.KindStruct {
							structInfo := tt.StructInfo(structType)
							fieldNameID := parent.Payload
							for _, field := range structInfo.Fields {
								if field.NameID == fieldNameID {
									typeID = field.TypeID
									name = string(pool.GetBytes(fieldNameID))
									details = fmt.Sprintf("(field) %s.%s", string(pool.GetBytes(entry.NameID)), name)
									break
								}
							}
						}
					}
				}
			}
		}
	}

	if typeID != types.TypeUnknown {
		typeStr := FormatType(tt, pool, typeID)
		mdValue := fmt.Sprintf("```axiom\n%s: %s\n```", details, typeStr)
		return &HoverResult{
			Contents: MarkupContent{
				Kind:  "markdown",
				Value: mdValue,
			},
		}
	}

	return nil
}

func FormatType(tt *types.TypeTable, pool *ast.InternPool, id types.TypeID) string {
	if id.IsPrimitive() {
		return id.String()
	}
	if int(id) >= tt.Count() {
		return "unknown"
	}
	entry := tt.Entry(id)
	switch entry.Kind {
	case types.KindPrimitive:
		return id.String()
	case types.KindPointer:
		return "*" + FormatType(tt, pool, tt.PointerElem(id))
	case types.KindSlice:
		return "[" + FormatType(tt, pool, tt.SliceElem(id)) + "]"
	case types.KindArray:
		return fmt.Sprintf("[%s; %d]", FormatType(tt, pool, tt.ArrayElem(id)), tt.ArrayLength(id))
	case types.KindRef:
		return "&" + FormatType(tt, pool, types.TypeID(entry.Extra))
	case types.KindStruct, types.KindSum, types.KindInterface:
		if entry.NameID != 0 {
			return string(pool.GetBytes(entry.NameID))
		}
		return "anonymous"
	case types.KindGenericInst:
		name := string(pool.GetBytes(entry.NameID))
		args := tt.GenericInstArgs(id)
		argStrs := make([]string, len(args))
		for i, arg := range args {
			argStrs[i] = FormatType(tt, pool, arg)
		}
		return fmt.Sprintf("%s[%s]", name, strings.Join(argStrs, ", "))
	case types.KindFunction:
		fInfo := tt.FuncInfo(id)
		paramStrs := make([]string, len(fInfo.Params))
		for i, param := range fInfo.Params {
			paramStrs[i] = FormatType(tt, pool, param)
		}
		return fmt.Sprintf("fn(%s) -> %s", strings.Join(paramStrs, ", "), FormatType(tt, pool, fInfo.Return))
	case types.KindGeneric:
		if entry.NameID != 0 {
			return string(pool.GetBytes(entry.NameID))
		}
		return "generic"
	default:
		return "unknown"
	}
}
