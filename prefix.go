package zglob

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob/syntax/ast"
	"github.com/gobwas/glob/syntax/lexer"
)

// staticPrefix returns the filepath up to the
// first path element that contains a wildcard.
func staticPrefix(pattern string) (string, error) {
	parts := strings.Split(pattern, string(filepath.Separator))

	prefix := ""
	if len(pattern) > 0 && pattern[0] == filepath.Separator {
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

		nChildren := len(rootNode.Children)

		if nChildren > 1 {
			break
		}

		candidate := rootNode
		if len(rootNode.Children) == 1 {
			candidate = rootNode.Children[0]
		}

		v, ok := candidate.Value.(ast.Text)
		if !ok {
			break
		}

		prefix = path.Join(prefix, v.Text)
	}

	if prefix == "" {
		prefix = "."
	}

	return prefix, nil
}
