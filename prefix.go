package fileglob

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob/syntax/ast"
	"github.com/gobwas/glob/syntax/lexer"
)

// ValidPattern determines whether a pattern is valid. It returns the parser
// error if the pattern is invalid and nil otherwise.
func ValidPattern(pattern string) error {
	_, err := ast.Parse(lexer.NewLexer(pattern))
	return err // nolint:wrapcheck
}

// ContainsMatchers determines whether the pattern contains any type of glob
// matcher. It will also return false if the pattern is an invalid expression.
func ContainsMatchers(pattern string) bool {
	rootNode, err := ast.Parse(lexer.NewLexer(pattern))
	if err != nil {
		return false
	}

	containsMatchers, _ := containsMatchers(rootNode)
	return containsMatchers
}

func containsMatchers(node *ast.Node) (result bool, staticText string) {
	// nolint:exhaustive
	switch node.Kind {
	case ast.KindPattern:
		text := ""

		for _, child := range node.Children {
			cm, staticText := containsMatchers(child)
			if cm {
				return true, ""
			}

			text += staticText
		}

		return false, text
	case ast.KindText:
		return false, node.Value.(ast.Text).Text
	case ast.KindNothing:
		return false, ""
	default:
		return true, ""
	}
}

// staticPrefix returns the file path inside the pattern up
// to the first path element that contains a wildcard.
func staticPrefix(pattern string) (string, error) {
	parts := strings.Split(pattern, string(filepath.Separator))

	prefix := ""
	if len(pattern) > 0 && rune(pattern[0]) == filepath.Separator {
		prefix = string(filepath.Separator)
	}

	for _, part := range parts {
		if part == "" {
			continue
		}

		rootNode, err := ast.Parse(lexer.NewLexer(part))
		if err != nil {
			return "", fmt.Errorf("parse glob pattern: %w", err)
		}

		cm, staticText := containsMatchers(rootNode)
		if cm {
			break
		}

		prefix = filepath.Join(prefix, staticText)
	}

	if prefix == "" {
		prefix = "."
	}

	return prefix, nil
}
