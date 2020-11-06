package zglob

import (
	"os"
	"strings"

	"github.com/gobwas/glob"
	"github.com/spf13/afero"
)

func Glob(pattern string) ([]string, error) {
	return GlobWithFs(afero.NewOsFs(), pattern)
}

func GlobWithFs(fs afero.Fs, pattern string) ([]string, error) {
	var matches []string
	g, err := glob.Compile(strings.TrimPrefix(pattern, "./"), '/')
	if err != nil {
		return matches, err
	}
	return matches, afero.Walk(fs, ".", func(path string, info os.FileInfo, err error) error {
 		if g.Match(path) {
 			matches = append(matches, path)
 		}
 		return nil
 	})
}
