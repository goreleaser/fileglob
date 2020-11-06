package zglob

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobwas/glob"
	"github.com/spf13/afero"
)

func Glob(pattern string) ([]string, error) {
	return GlobWithFs(afero.NewOsFs(), pattern)
}

func GlobWithFs(fs afero.Fs, pattern string) ([]string, error) {
	var matches []string

	pattern = strings.TrimPrefix(pattern, "./")
	matcher, err := glob.Compile(pattern)
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

		if !matcher.Match(path) {
			return nil
		}

		if info.IsDir() {
			// a direct match on a directory implies that all files inside match
			filesInDir, err := filesInDirectory(fs, path)
			if err != nil {
				return err
			}

			matches = append(matches, filesInDir...)
			return filepath.SkipDir
		} else {
			matches = append(matches, path)
		}

		return nil
	})
}

func filesInDirectory(fs afero.Fs, dir string) ([]string, error) {
	var files []string

	return files, afero.Walk(fs, dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
}
