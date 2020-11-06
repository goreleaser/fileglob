package zglob

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
)

func Glob(pattern string) ([]string, error) {
	var matches []string
	g, err := glob.Compile(strings.TrimPrefix(pattern, "./"), '/')
	if err != nil {
		return matches, err
	}
	return matches, filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
 		if g.Match(path) {
 			matches = append(matches, path)
 		}
 		return nil
 	})
}
