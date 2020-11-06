package zglob

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestGlob(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		matches, err := globInMemoryFs("./a/*/*", []string{
			"./c/file1.txt",
			"./a/nope/file1.txt",
			"./a/something",
			"./a/b/file1.txt",
			"./a/c/file2.txt",
			"./a/d/file1.txt",
		})
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file1.txt",
			"a/c/file2.txt",
			"a/d/file1.txt",
			"a/nope/file1.txt",
		}, matches)
	})

	t.Run("single file", func(t *testing.T) {
		matches, err := globInMemoryFs("a/b/*", []string{
			"./c/file1.txt",
			"./a/nope/file1.txt",
			"./a/b/file1.txt",
		})
		require.NoError(t, err)
		require.Equal(t, []string{"a/b/file1.txt"}, matches)
	})

	t.Run("double star", func(t *testing.T) {
		matches, err := globInMemoryFs("a/**/*", []string{
			"./a/nope.txt",
			"./a/d/file1.txt",
			"./a/e/f/file1.txt",
			"./a/b/file1.txt",
		})
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file1.txt",
			"a/d/file1.txt",
			"a/e/f",
			"a/e/f/file1.txt",
		}, matches)
	})

	t.Run("no matches", func(t *testing.T) {
		matches, err := globInMemoryFs("z/*", []string{
			"./a/nope.txt",
		})
		require.NoError(t, err)
		require.Empty(t, matches)
	})

	t.Run("empty folder", func(t *testing.T) {
		matches, err := globInMemoryFs("a*", []string{})
		require.NoError(t, err)
		require.Empty(t, matches)
	})
}

func globInMemoryFs(pattern string, files []string) ([]string, error) {
	var fs = afero.NewMemMapFs()
	if err := createFiles(fs, files); err != nil {
		return []string{}, err
	}
	return GlobWithFs(fs, pattern)
}

func createFiles(fs afero.Fs, files []string) error {
	for _, file := range files {
		if _, err := fs.Create(file); err != nil {
			return err
		}
	}
	return nil
}
