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

// maxPathSeparators bounds the number of separators in a match.
// A negative result means unbounded or unknown, so the walk must not prune.
func maxPathSeparators(node *ast.Node) int {
	switch node.Kind {
	case ast.KindPattern, ast.KindAnyOf:
		var count int
		for _, child := range node.Children {
			n := maxPathSeparators(child)
			if n < 0 {
				return -1
			}
			if node.Kind == ast.KindAnyOf {
				count = max(count, n)
			} else {
				count += n
			}
		}
		return count
	case ast.KindText:
		return strings.Count(node.Value.(ast.Text).Text, separatorString)
	case ast.KindNothing, ast.KindAny, ast.KindSingle:
		return 0
	case ast.KindList:
		list := node.Value.(ast.List)
		if strings.ContainsRune(list.Chars, separatorRune) != list.Not {
			return 1
		}
		return 0
	case ast.KindRange:
		r := node.Value.(ast.Range)
		if (r.Lo <= separatorRune && separatorRune <= r.Hi) != r.Not {
			return 1
		}
		return 0
	default:
		return -1
	}
}
