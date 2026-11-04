package fileglob

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob/syntax/ast"
	"github.com/gobwas/glob/syntax/lexer"
)

// ValidPattern determines whether a pattern is valid. It returns the parser
// error if the pattern is invalid and nil otherwise.
func ValidPattern(pattern string) error {
	_, err := ast.Parse(lexer.NewLexer(pattern))
	return err //nolint:wrapcheck
}

// ContainsMatchers determines whether the pattern contains any type of glob
// matcher. It will also return false if the pattern is an invalid expression.
func ContainsMatchers(pattern string) bool {
	rootNode, err := ast.Parse(lexer.NewLexer(pattern))
	if err != nil {
		return false
	}

	_, isStatic := staticText(rootNode)
	return !isStatic
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
// to the first path element that contains a glob matcher.
func staticPrefix(pattern string) (string, error) {
	rootNode, err := ast.Parse(lexer.NewLexer(pattern))
	if err != nil {
		return "", fmt.Errorf("parse glob pattern: %w", err)
	}

	var prefix string
	for _, child := range rootNode.Children {
		text, ok := staticText(child)
		if !ok {
			// Only complete path components before the first matcher are static.
			prefix = prefix[:strings.LastIndex(prefix, separatorString)+1]
			break
		}

		prefix += text
	}

	rooted := strings.HasPrefix(prefix, separatorString)
	prefix = strings.Join(strings.FieldsFunc(prefix, func(r rune) bool {
		return r == separatorRune
	}), separatorString)
	if rooted {
		prefix = separatorString + prefix
	}

	if prefix == "" {
		prefix = "."
	}

	return prefix, nil
}
