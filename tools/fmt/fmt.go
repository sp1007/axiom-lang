package fmt

import (
	"bytes"
	"os"
	"sort"
	"strings"
)

// FmtTokenType represents the lexical category of format-preserving tokens.
type FmtTokenType int

const (
	TokenEOF FmtTokenType = iota
	TokenNewline
	TokenWhitespace
	TokenComment
	TokenString
	TokenChar
	TokenNumber
	TokenOperator
	TokenPunctuation
	TokenIdent
)

// FmtToken is a format-preserving token capturing exactly raw slice bytes.
type FmtToken struct {
	Type FmtTokenType
	Text string
}

// Formatter implements the AXIOM canonical, zero-config formatting engine.
type Formatter struct {
	IndentWidth int // 4
	MaxWidth    int // 100
}

// NewFormatter creates a new Formatter with standard canonical settings.
func NewFormatter() *Formatter {
	return &Formatter{
		IndentWidth: 4,
		MaxWidth:    100,
	}
}

// Format takes raw AXIOM source bytes, formats it, and returns the styled output.
func (f *Formatter) Format(src []byte) ([]byte, error) {
	tokens := scanTokens(src)
	lines := splitIntoLines(tokens)
	formattedLines := f.formatLines(lines)
	
	var buf bytes.Buffer
	for i, line := range formattedLines {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(line)
	}
	
	// Canonical ending: exactly one trailing newline
	res := buf.Bytes()
	if len(res) > 0 && res[len(res)-1] != '\n' {
		res = append(res, '\n')
	}
	return res, nil
}

// FormatFile formats an AXIOM file in place.
func (f *Formatter) FormatFile(path string) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	formatted, err := f.Format(src)
	if err != nil {
		return err
	}
	if bytes.Equal(src, formatted) {
		return nil // No change needed
	}
	return os.WriteFile(path, formatted, 0644)
}

// Check returns true if the file needs formatting without editing it.
func (f *Formatter) Check(path string) (bool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	formatted, err := f.Format(src)
	if err != nil {
		return false, err
	}
	return !bytes.Equal(src, formatted), nil
}

// --------------------------------------------------------------------------
// Lexical Scanning (Comment & Whitespace preserving)
// --------------------------------------------------------------------------

func scanTokens(src []byte) []FmtToken {
	var tokens []FmtToken
	n := len(src)
	i := 0

	for i < n {
		b := src[i]

		// 1. Newline
		if b == '\n' {
			tokens = append(tokens, FmtToken{Type: TokenNewline, Text: "\n"})
			i++
			continue
		}

		// 2. Line Comment
		if b == '/' && i+1 < n && src[i+1] == '/' {
			start := i
			i += 2
			for i < n && src[i] != '\n' {
				i++
			}
			tokens = append(tokens, FmtToken{Type: TokenComment, Text: string(src[start:i])})
			continue
		}

		// 3. Whitespace (spaces/tabs, excluding newlines)
		if b == ' ' || b == '\t' || b == '\r' {
			start := i
			for i < n && (src[i] == ' ' || src[i] == '\t' || src[i] == '\r') {
				i++
			}
			tokens = append(tokens, FmtToken{Type: TokenWhitespace, Text: string(src[start:i])})
			continue
		}

		// 4. String Literal
		if b == '"' {
			start := i
			i++
			escaped := false
			for i < n {
				if escaped {
					escaped = false
				} else if src[i] == '\\' {
					escaped = true
				} else if src[i] == '"' {
					i++
					break
				}
				i++
			}
			tokens = append(tokens, FmtToken{Type: TokenString, Text: string(src[start:i])})
			continue
		}

		// 5. Char Literal
		if b == '\'' {
			start := i
			i++
			escaped := false
			for i < n {
				if escaped {
					escaped = false
				} else if src[i] == '\\' {
					escaped = true
				} else if src[i] == '\'' {
					i++
					break
				}
				i++
			}
			tokens = append(tokens, FmtToken{Type: TokenChar, Text: string(src[start:i])})
			continue
		}

		// 6. Number Literal
		if isDigit(b) {
			start := i
			// Hex/Octal/Binary detection
			if b == '0' && i+1 < n {
				next := src[i+1]
				if next == 'x' || next == 'X' {
					i += 2
					for i < n && (isHexDigit(src[i]) || src[i] == '_') {
						i++
					}
					tokens = append(tokens, FmtToken{Type: TokenNumber, Text: string(src[start:i])})
					continue
				} else if next == 'o' || next == 'O' {
					i += 2
					for i < n && (isOctalDigit(src[i]) || src[i] == '_') {
						i++
					}
					tokens = append(tokens, FmtToken{Type: TokenNumber, Text: string(src[start:i])})
					continue
				} else if next == 'b' || next == 'B' {
					i += 2
					for i < n && (src[i] == '0' || src[i] == '1' || src[i] == '_') {
						i++
					}
					tokens = append(tokens, FmtToken{Type: TokenNumber, Text: string(src[start:i])})
					continue
				}
			}
			// Decimal or Float
			for i < n && (isDigit(src[i]) || src[i] == '_') {
				i++
			}
			if i < n && src[i] == '.' && i+1 < n && isDigit(src[i+1]) {
				i++ // consume '.'
				for i < n && (isDigit(src[i]) || src[i] == '_') {
					i++
				}
			}
			if i < n && (src[i] == 'e' || src[i] == 'E') {
				i++
				if i < n && (src[i] == '+' || src[i] == '-') {
					i++
				}
				for i < n && isDigit(src[i]) {
					i++
				}
			}
			tokens = append(tokens, FmtToken{Type: TokenNumber, Text: string(src[start:i])})
			continue
		}

		// 7. Operators & Punctuation
		if opText, length := matchOperatorOrPunct(src[i:]); length > 0 {
			tp := TokenOperator
			if isPunctuation(opText) {
				tp = TokenPunctuation
			}
			tokens = append(tokens, FmtToken{Type: tp, Text: opText})
			i += length
			continue
		}

		// 8. Identifiers / Keywords
		if isIdentStart(b) {
			start := i
			for i < n && isIdentPart(src[i]) {
				i++
			}
			tokens = append(tokens, FmtToken{Type: TokenIdent, Text: string(src[start:i])})
			continue
		}

		// Fallback single-character token
		tokens = append(tokens, FmtToken{Type: TokenIdent, Text: string(src[i : i+1])})
		i++
	}

	return tokens
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func isOctalDigit(b byte) bool {
	return b >= '0' && b <= '7'
}

func isIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == '@'
}

func isIdentPart(b byte) bool {
	return isIdentStart(b) || isDigit(b)
}

func matchOperatorOrPunct(slice []byte) (string, int) {
	// Match multi-character operators first
	prefixes := []string{
		"**", "==", "!=", "<=", ">=", "<<", ">>", ":=", "+=", "-=", "*=", "/=", "%=", "->", "..",
	}
	for _, p := range prefixes {
		if len(slice) >= len(p) && string(slice[:len(p)]) == p {
			return p, len(p)
		}
	}
	// Single characters
	single := []string{
		"+", "-", "*", "/", "%", "<", ">", "=", "&", "|", "^", "~", "!", ".", ",", ":", ";", "(", ")", "[", "]", "{", "}",
	}
	for _, s := range single {
		if len(slice) >= len(s) && string(slice[:len(s)]) == s {
			return s, len(s)
		}
	}
	return "", 0
}

func isPunctuation(text string) bool {
	switch text {
	case ".", ",", ":", ";", "(", ")", "[", "]", "{", "}", "..":
		return true
	}
	return false
}

// --------------------------------------------------------------------------
// Line Splitting
// --------------------------------------------------------------------------

type Line struct {
	Tokens []FmtToken
}

func splitIntoLines(tokens []FmtToken) []Line {
	var lines []Line
	var current []FmtToken
	for _, t := range tokens {
		if t.Type == TokenNewline {
			lines = append(lines, Line{Tokens: current})
			current = nil
		} else {
			current = append(current, t)
		}
	}
	if len(current) > 0 {
		lines = append(lines, Line{Tokens: current})
	}
	return lines
}

// --------------------------------------------------------------------------
// Formatting Engine
// --------------------------------------------------------------------------

func (f *Formatter) formatLines(lines []Line) []string {
	// First pass: sort and group imports
	lines = sortAndGroupImports(lines)

	var formatted []string
	braceDepth := 0

	for lineIdx := 0; lineIdx < len(lines); lineIdx++ {
		origLine := lines[lineIdx]
		
		// 1. Detect empty line
		if isEmptyLine(origLine) {
			// Skip multiple consecutive blank lines inside function bodies (collapsing)
			if len(formatted) > 0 && formatted[len(formatted)-1] == "" {
				continue
			}
			formatted = append(formatted, "")
			continue
		}

		// Strip leading/trailing whitespaces from the line tokens
		contentTokens := stripLeadingTrailingWhitespace(origLine.Tokens)

		// Calculate indentation level for the current line
		closeCount := countLeadingCloseBraces(contentTokens)
		level := braceDepth - closeCount
		if level < 0 {
			level = 0
		}

		// Update braceDepth for the next line
		braceDepth += getBraceBalance(contentTokens)
		if braceDepth < 0 {
			braceDepth = 0
		}

		// 3. Format the statement content
		formattedContent := f.formatTokenSpacing(contentTokens)

		// 4. Prepend canonical indentation
		indentStr := strings.Repeat(" ", level*f.IndentWidth)
		formattedLine := indentStr + formattedContent
		
		// Remove trailing spaces
		formattedLine = strings.TrimRight(formattedLine, " \t\r")
		formatted = append(formatted, formattedLine)
	}

	// Post pass: align inline comments on consecutive lines
	formatted = f.alignInlineComments(formatted)

	// Clean up leading/trailing empty lines
	for len(formatted) > 0 && formatted[0] == "" {
		formatted = formatted[1:]
	}
	for len(formatted) > 0 && formatted[len(formatted)-1] == "" {
		formatted = formatted[:len(formatted)-1]
	}

	return formatted
}

func countLeadingCloseBraces(tokens []FmtToken) int {
	count := 0
	for _, t := range tokens {
		if t.Type == TokenWhitespace {
			continue
		}
		if t.Type == TokenPunctuation && t.Text == "}" {
			count++
		} else {
			break
		}
	}
	return count
}

func getBraceBalance(tokens []FmtToken) int {
	balance := 0
	for _, t := range tokens {
		if t.Type == TokenPunctuation {
			if t.Text == "{" {
				balance++
			} else if t.Text == "}" {
				balance--
			}
		}
	}
	return balance
}

func isEmptyLine(l Line) bool {
	for _, t := range l.Tokens {
		if t.Type != TokenWhitespace {
			return false
		}
	}
	return true
}

func countLeadingIndent(l Line) int {
	indent := 0
	for _, t := range l.Tokens {
		if t.Type == TokenWhitespace {
			for _, char := range t.Text {
				if char == '\t' {
					indent += 4 // Convert tabs to 4 spaces
				} else {
					indent++
				}
			}
		} else {
			break
		}
	}
	return indent
}

func stripLeadingTrailingWhitespace(tokens []FmtToken) []FmtToken {
	start := 0
	for start < len(tokens) && tokens[start].Type == TokenWhitespace {
		start++
	}
	end := len(tokens)
	for end > start && tokens[end-1].Type == TokenWhitespace {
		end--
	}
	return tokens[start:end]
}

// formatTokenSpacing formats operators, parentheses, commas, type annotations spacing.
func (f *Formatter) formatTokenSpacing(tokens []FmtToken) string {
	if len(tokens) == 0 {
		return ""
	}

	// Filter out original whitespaces first to build the spacing canonical layout cleanly
	var filtered []FmtToken
	for _, t := range tokens {
		if t.Type != TokenWhitespace {
			filtered = append(filtered, t)
		}
	}
	tokens = filtered
	if len(tokens) == 0 {
		return ""
	}

	var sb strings.Builder

	for j := 0; j < len(tokens); j++ {
		t := tokens[j]

		// 1. Inline comment
		if t.Type == TokenComment {
			// Space before comment if preceded by code on the same line
			if sb.Len() > 0 && !strings.HasSuffix(sb.String(), " ") {
				sb.WriteByte(' ')
			}
			sb.WriteString(t.Text)
			continue
		}

		// 2. Binary Operators (e.g. +, -, *, ==, =, :=, ->, etc.)
		if t.Type == TokenOperator {
			// Check if unary or binary operator
			isUnary := false
			if t.Text == "-" || t.Text == "!" || t.Text == "~" {
				// Detect if unary based on preceding token
				if j == 0 {
					isUnary = true
				} else {
					prev := tokens[j-1]
					if prev.Type == TokenOperator || prev.Type == TokenPunctuation {
						isUnary = true
					} else if prev.Type == TokenIdent && isUnaryKeyword(prev.Text) {
						isUnary = true
					}
				}
			}

			if isUnary {
				// No space after unary operator
				// Ensure space before if preceded by an identifier/number/string
				if sb.Len() > 0 && needsSpaceBeforeUnary(sb.String()) {
					sb.WriteByte(' ')
				}
				sb.WriteString(t.Text)
			} else {
				// Binary operator: exactly one space before and after
				s := sb.String()
				if sb.Len() > 0 && !strings.HasSuffix(s, " ") && !strings.HasSuffix(s, "(") && !strings.HasSuffix(s, "[") {
					sb.WriteByte(' ')
				}
				sb.WriteString(t.Text)
				// Put space after unless followed by EOF or a comment
				if j+1 < len(tokens) && tokens[j+1].Type != TokenComment {
					sb.WriteByte(' ')
				}
			}
			continue
		}

		// 3. Punctuation
		if t.Type == TokenPunctuation {
			switch t.Text {
			case ":":
				// Type annotation: no space before, one space after
				// Trim any trailing space first
				s := sb.String()
				if strings.HasSuffix(s, " ") {
					trimmed := strings.TrimRight(s, " ")
					sb.Reset()
					sb.WriteString(trimmed)
				}
				sb.WriteString(":")
				if j+1 < len(tokens) && tokens[j+1].Type != TokenComment && tokens[j+1].Type != TokenNewline {
					sb.WriteByte(' ')
				}
			case ",":
				// Comma: no space before, one space after
				s := sb.String()
				if strings.HasSuffix(s, " ") {
					trimmed := strings.TrimRight(s, " ")
					sb.Reset()
					sb.WriteString(trimmed)
				}
				sb.WriteString(",")
				if j+1 < len(tokens) && tokens[j+1].Type != TokenComment {
					sb.WriteByte(' ')
				}
			case ";":
				// Semicolon: no space before, one space after
				s := sb.String()
				if strings.HasSuffix(s, " ") {
					trimmed := strings.TrimRight(s, " ")
					sb.Reset()
					sb.WriteString(trimmed)
				}
				sb.WriteString(";")
				if j+1 < len(tokens) && tokens[j+1].Type != TokenComment {
					sb.WriteByte(' ')
				}
			case ".", "..":
				// Dot or Range/Relative import
				s := sb.String()
				prev := getPrevNonWhitespaceToken(tokens[:j])
				if prev == "import" {
					if !strings.HasSuffix(s, " ") {
						sb.WriteByte(' ')
					}
					sb.WriteString(t.Text)
				} else {
					if strings.HasSuffix(s, " ") {
						trimmed := strings.TrimRight(s, " ")
						sb.Reset()
						sb.WriteString(trimmed)
					}
					sb.WriteString(t.Text)
				}
			case "(":
				// Function call: no space before if preceded by an identifier (e.g. `foo(a)`)
				// Space before if preceded by keyword (e.g. `if (cond)`)
				if sb.Len() > 0 {
					prevText := getPrevNonWhitespaceToken(tokens[:j])
					if prevText != "" && !isIdentifierChar(prevText[len(prevText)-1]) {
						// Do nothing, no space
					} else if prevText != "" && isKeywordPrecedingParenthesis(prevText) {
						if !strings.HasSuffix(sb.String(), " ") {
							sb.WriteByte(' ')
						}
					} else {
						// Strip trailing space to ensure no space before '(' for function calls
						s := sb.String()
						if strings.HasSuffix(s, " ") {
							trimmed := strings.TrimRight(s, " ")
							sb.Reset()
							sb.WriteString(trimmed)
						}
					}
				}
				sb.WriteString("(")
			case ")":
				// Close parenthesis: no space before
				s := sb.String()
				if strings.HasSuffix(s, " ") && !strings.HasSuffix(s, "(") {
					trimmed := strings.TrimRight(s, " ")
					sb.Reset()
					sb.WriteString(trimmed)
				}
				sb.WriteString(")")
			case "[":
				// Generics or indexing: no space before
				if sb.Len() > 0 {
					s := sb.String()
					if strings.HasSuffix(s, " ") {
						trimmed := strings.TrimRight(s, " ")
						sb.Reset()
						sb.WriteString(trimmed)
					}
				}
				sb.WriteString("[")
			case "]":
				// Close bracket: no space before
				s := sb.String()
				if strings.HasSuffix(s, " ") && !strings.HasSuffix(s, "[") {
					trimmed := strings.TrimRight(s, " ")
					sb.Reset()
					sb.WriteString(trimmed)
				}
				sb.WriteString("]")
			case "{":
				// Struct literal / block: space before, space after
				if sb.Len() > 0 && !strings.HasSuffix(sb.String(), " ") {
					sb.WriteByte(' ')
				}
				sb.WriteString("{")
				if j+1 < len(tokens) && tokens[j+1].Text != "}" && tokens[j+1].Type != TokenComment {
					sb.WriteByte(' ')
				}
			case "}":
				// Close brace: space before
				s := sb.String()
				if sb.Len() > 0 && !strings.HasSuffix(s, " ") && !strings.HasSuffix(s, "{") {
					sb.WriteByte(' ')
				}
				sb.WriteString("}")
			default:
				sb.WriteString(t.Text)
			}
			continue
		}

		// 4. Identifiers, numbers, strings
		if sb.Len() > 0 {
			s := sb.String()
			prevToken := getPrevNonWhitespaceToken(tokens[:j])
			// Space between identifiers, keywords, numbers
			if !strings.HasSuffix(s, " ") && !strings.HasSuffix(s, "(") && !strings.HasSuffix(s, "[") && !strings.HasSuffix(s, ".") && !strings.HasSuffix(s, "!") && !strings.HasSuffix(s, "~") {
				if isIdentifierChar(s[len(s)-1]) || t.Type == TokenString || t.Type == TokenChar || prevToken == "," || prevToken == ":" {
					sb.WriteByte(' ')
				}
			}
		}
		sb.WriteString(t.Text)
	}

	return sb.String()
}

func isUnaryKeyword(text string) bool {
	return text == "not" || text == "return" || text == "spawn" || text == "in" || text == "let" || text == "mut"
}

func needsSpaceBeforeUnary(s string) bool {
	if len(s) == 0 {
		return false
	}
	last := s[len(s)-1]
	return last != '(' && last != '[' && last != ' ' && last != ',' && last != ':'
}

func getPrevNonWhitespaceToken(tokens []FmtToken) string {
	for k := len(tokens) - 1; k >= 0; k-- {
		if tokens[k].Type != TokenWhitespace {
			return tokens[k].Text
		}
	}
	return ""
}

func isIdentifierChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '@'
}

func isKeywordPrecedingParenthesis(text string) bool {
	switch text {
	case "if", "while", "for", "match", "elif", "return", "and", "or":
		return true
	}
	return false
}

// --------------------------------------------------------------------------
// Import Sorting and Grouping
// --------------------------------------------------------------------------

func sortAndGroupImports(lines []Line) []Line {
	// Locate consecutive top-level import statement blocks
	var result []Line
	var imports []Line
	inImportBlock := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if isEmptyLine(line) {
			if inImportBlock {
				// Don't add empty line to import block yet
				continue
			}
			result = append(result, line)
			continue
		}

		// Find first non-whitespace token
		var firstToken FmtToken
		found := false
		for _, t := range line.Tokens {
			if t.Type != TokenWhitespace {
				firstToken = t
				found = true
				break
			}
		}

		if found && firstToken.Type == TokenIdent && firstToken.Text == "import" {
			imports = append(imports, line)
			inImportBlock = true
		} else {
			if inImportBlock {
				// Flush imports
				result = append(result, groupAndSortImportLines(imports)...)
				imports = nil
				inImportBlock = false
			}
			result = append(result, line)
		}
	}

	if len(imports) > 0 {
		result = append(result, groupAndSortImportLines(imports)...)
	}

	return result
}

type importInfo struct {
	Text string
	Line Line
	Path string
}

func groupAndSortImportLines(lines []Line) []Line {
	if len(lines) == 0 {
		return nil
	}

	var stdImports []importInfo
	var thirdImports []importInfo
	var localImports []importInfo

	for _, l := range lines {
		// Reconstruct import text
		var sb strings.Builder
		for _, t := range l.Tokens {
			if t.Type != TokenWhitespace {
				sb.WriteString(t.Text)
			}
		}
		// Extract path: "importstd.io" -> "std.io"
		impStr := sb.String()
		path := strings.TrimPrefix(impStr, "import")

		info := importInfo{Text: impStr, Line: l, Path: path}

		if strings.HasPrefix(path, "std.") || path == "std" {
			stdImports = append(stdImports, info)
		} else if strings.HasPrefix(path, ".") || strings.HasPrefix(path, "local.") || path == "local" {
			localImports = append(localImports, info)
		} else {
			thirdImports = append(thirdImports, info)
		}
	}

	// Sort alphabetically
	sortImports(stdImports)
	sortImports(thirdImports)
	sortImports(localImports)

	var sortedLines []Line

	// Add stdlib group
	for _, imp := range stdImports {
		sortedLines = append(sortedLines, imp.Line)
	}
	// Separator if needed
	if len(stdImports) > 0 && (len(thirdImports) > 0 || len(localImports) > 0) {
		sortedLines = append(sortedLines, Line{Tokens: nil}) // empty line
	}

	// Add third-party group
	for _, imp := range thirdImports {
		sortedLines = append(sortedLines, imp.Line)
	}
	if len(thirdImports) > 0 && len(localImports) > 0 {
		sortedLines = append(sortedLines, Line{Tokens: nil}) // empty line
	}

	// Add local group
	for _, imp := range localImports {
		sortedLines = append(sortedLines, imp.Line)
	}

	// Add exactly 1 trailing empty line after the whole import block
	sortedLines = append(sortedLines, Line{Tokens: nil})

	return sortedLines
}

func sortImports(imports []importInfo) {
	sort.Slice(imports, func(i, j int) bool {
		return imports[i].Path < imports[j].Path
	})
}

// --------------------------------------------------------------------------
// Comment Alignment
// --------------------------------------------------------------------------

func (f *Formatter) alignInlineComments(lines []string) []string {
	var result []string
	n := len(lines)
	i := 0

	for i < n {
		if !hasInlineComment(lines[i]) {
			result = append(result, lines[i])
			i++
			continue
		}

		// Find block of consecutive lines with inline comments
		start := i
		for i < n && hasInlineComment(lines[i]) {
			i++
		}
		end := i

		// Align this block [start, end)
		// 1. Calculate comment column for each line, and find the maximum column
		maxCol := 0
		type lineCommentInfo struct {
			CodePart    string
			CommentPart string
		}
		infos := make([]lineCommentInfo, end-start)

		for k := start; k < end; k++ {
			line := lines[k]
			idx := strings.Index(line, "//")
			code := strings.TrimRight(line[:idx], " \t")
			comment := line[idx:]
			infos[k-start] = lineCommentInfo{CodePart: code, CommentPart: comment}

			col := len(code)
			if col > maxCol {
				maxCol = col
			}
		}

		// Ensure alignment column is at least 40, and aligned nicely
		alignCol := maxCol + 2
		if alignCol < 40 {
			alignCol = 40
		}

		// 2. Re-assemble aligned lines
		for _, info := range infos {
			spacesNeeded := alignCol - len(info.CodePart)
			if spacesNeeded < 1 {
				spacesNeeded = 1 // At least one space separation
			}
			aligned := info.CodePart + strings.Repeat(" ", spacesNeeded) + info.CommentPart
			result = append(result, aligned)
		}
	}

	return result
}

func hasInlineComment(line string) bool {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "//") {
		return false // Top-level comment line, not inline
	}
	return strings.Contains(line, "//")
}
