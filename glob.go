package fileglob

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/spf13/afero"
)

// Options allowed to be passed to Glob.
type Options struct {
	fs afero.Fs
}

// WithFs allows to provide another afero.Fs implementation to Glob.
func WithFs(fs afero.Fs) Options {
	return Options{fs: fs}
}

// QuoteMeta returns a string that quotes all glob pattern meta characters
// inside the argument text; For example, QuoteMeta(`{foo*}`) returns `\{foo\*\}`.
func QuoteMeta(pattern string) string {
	return glob.QuoteMeta(pattern)
}

// Glob returns all files that match the given pattern in the current directory.
func Glob(pattern string, opts ...Options) ([]string, error) {
	var options = compileOptions(opts)
	pattern = strings.TrimPrefix(pattern, "./")

	var fs = options.fs
	var matches []string

	matcher, err := glob.Compile(pattern, filepath.Separator)
	if err != nil {
		return matches, err
	}

	prefix, err := staticPrefix(pattern)
	if err != nil {
		return nil, fmt.Errorf("cannot determine static prefix: %w", err)
	}

	prefixInfo, err := fs.Stat(prefix)
	if os.IsNotExist(err) {
		// if the prefix does not exist, the whole
		// glob pattern won't match anything
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("stat prefix: %w", err)
	}

	if !prefixInfo.IsDir() {
		// if the prefix is a file, it either has to be
		// the only match, or nothing matches at all
		if matcher.Match(prefix) {
			return []string{prefix}, nil
		}

		return []string{}, nil
	}

	return matches, afero.Walk(fs, prefix, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if matcher.Match(path) && !info.IsDir() {
			matches = append(matches, path)
		}

		return nil
	})
}

func compileOptions(opts []Options) Options {
	var options = Options{
		fs: afero.NewOsFs(),
	}
	for _, opt := range opts {
		if opt.fs != nil {
			options.fs = opt.fs
		}
	}
	return options
}
