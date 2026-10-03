package fileglob

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob/syntax/ast"
	"github.com/gobwas/glob/syntax/lexer"
)

// ValidPattern determines whether a pattern is valid. It returns the parser
// error if the pattern is invalid and nil otherwise. Alternative groups must
// be closed.
func ValidPattern(pattern string) error {
	_, err := parsePattern(pattern)
	return err //nolint:wrapcheck
}

// ContainsMatchers determines whether the pattern contains any type of glob
// matcher. It will also return false if the pattern is an invalid expression.
func ContainsMatchers(pattern string) bool {
	rootNode, err := parsePattern(pattern)
	if err != nil {
		return false
	}

	_, isStatic := staticText(rootNode)
	return !isStatic
}

func parsePattern(pattern string) (*ast.Node, error) {
	return ast.Parse(&patternLexer{Lexer: lexer.NewLexer(pattern)})
}

// patternLexer rejects unclosed alternatives in complete patterns.
// staticPrefix uses the upstream lexer because it parses partial path elements.
type patternLexer struct {
	ast.Lexer
	depth int
}

func (l *patternLexer) Next() lexer.Token {
	token := l.Lexer.Next()
	switch token.Type {
	case lexer.TermsOpen:
		l.depth++
	case lexer.TermsClose:
		l.depth--
	case lexer.EOF:
		if l.depth > 0 {
			return lexer.Token{
				Type: lexer.Error,
				Raw:  "unexpected end of input: unclosed alternative group",
			}
		}
	}
	return token
}

// staticText returns the static string matcher represented by the AST unless
// it contains dynamic matchers (wildcards, etc.). In this case the ok return
// value is false.
func staticText(node *ast.Node) (text string, ok bool) {
	//nolint:exhaustive
	switch node.Kind {
	case ast.KindPattern:
		var sb strings.Builder

		for _, child := range node.Children {
			childText, ok := staticText(child)
			if !ok {
				return "", false
			}

			sb.WriteString(childText)
		}

		return sb.String(), true
	case ast.KindText:
		t, ok := node.Value.(ast.Text)
		if !ok {
			return "", false
		}
		return t.Text, true
	case ast.KindNothing:
		return "", true
	default:
		return "", false
	}
}

// staticPrefix returns the file path inside the pattern up
// to the first path element that contains a wildcard.
func staticPrefix(pattern string) (string, error) {
	parts := strings.Split(pattern, separatorString)

	//nolint:prealloc
	var prefixPath []string
	for _, part := range parts {
		if part == "" {
			continue
		}

		rootNode, err := ast.Parse(lexer.NewLexer(part))
		if err != nil {
			return "", fmt.Errorf("parse glob pattern: %w", err)
		}

		staticPart, ok := staticText(rootNode)
		if !ok {
			break
		}

		prefixPath = append(prefixPath, staticPart)
	}
	prefix := strings.Join(prefixPath, separatorString)
	if len(pattern) > 0 && rune(pattern[0]) == separatorRune && !strings.HasPrefix(prefix, separatorString) {
		prefix = separatorString + prefix
	}

	if prefix == "" {
		prefix = "."
	}

	return prefix, nil
}
