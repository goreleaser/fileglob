package fileglob

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/caarlos0/testfs"
	"github.com/gobwas/glob"
	"github.com/stretchr/testify/require"
)

func TestGlob(t *testing.T) { // nolint:funlen
	t.Parallel()
	t.Run("real", func(t *testing.T) {
		t.Parallel()
		var w bytes.Buffer
		matches, err := Glob("*_test.go", WriteOptions(&w))
		require.NoError(t, err)
		require.Equal(t, []string{
			"glob_test.go",
			"prefix_test.go",
		}, matches)
		require.Equal(t, "&{fs:. matchDirectoriesDirectly:false prefix:./ pattern:*_test.go}", w.String())
	})

	t.Run("real with rootfs", func(t *testing.T) {
		t.Parallel()

		wd, err := os.Getwd()
		require.NoError(t, err)

		prefix := "/"
		if isWindows() {
			prefix = filepath.VolumeName(wd) + "/"
		}

		pattern := toNixPath(filepath.Join(wd, "*_test.go"))

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, WriteOptions(&w))
		require.NoError(t, err)
		require.Equal(t, []string{
			toNixPath(filepath.Join(wd, "glob_test.go")),
			toNixPath(filepath.Join(wd, "prefix_test.go")),
		}, matches)
		require.Equal(t, fmt.Sprintf("&{fs:%s matchDirectoriesDirectly:false prefix:%s pattern:%s}", prefix, prefix, pattern), w.String())
	})

	t.Run("real with rootfs direct file", func(t *testing.T) {
		t.Parallel()

		wd, err := os.Getwd()
		require.NoError(t, err)

		pattern := toNixPath(filepath.Join(wd, "prefix.go"))

		matches, err := Glob(pattern, MaybeRootFS)
		require.NoError(t, err)
		require.Equal(t, []string{
			toNixPath(filepath.Join(wd, "prefix.go")),
		}, matches)
	})

	t.Run("real with rootfs on relative path to parent disable globbing", func(t *testing.T) {
		t.Parallel()

		wd, err := os.Getwd()
		require.NoError(t, err)

		dir := filepath.Base(wd)

		prefix := "/"
		if isWindows() {
			prefix = filepath.VolumeName(wd) + "/"
		}

		pattern := "../" + dir + "/{file}["

		abs, err := filepath.Abs(pattern)
		require.NoError(t, err)
		abs = toNixPath(abs)

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, QuoteMeta, WriteOptions(&w))
		require.Error(t, err)
		require.True(t, strings.HasSuffix(err.Error(), "file does not exist"), "should have been file does not exist, got: "+err.Error())
		require.Empty(t, matches)
		require.Equal(t, fmt.Sprintf("&{fs:%s matchDirectoriesDirectly:false prefix:%s pattern:%s}", prefix, prefix, glob.QuoteMeta(abs)), w.String())
	})

	t.Run("real with rootfs on relative path to parent", func(t *testing.T) {
		t.Parallel()

		wd, err := os.Getwd()
		require.NoError(t, err)

		dir := filepath.Base(wd)

		prefix := "/"
		if isWindows() {
			prefix = filepath.VolumeName(wd) + "/"
		}

		pattern := "../" + dir + "/*_test.go"
		abs, err := filepath.Abs(pattern)
		require.NoError(t, err)

		abs = toNixPath(abs)

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, WriteOptions(&w))
		require.NoError(t, err)
		require.Equal(t, []string{
			toNixPath(filepath.Join(wd, "glob_test.go")),
			toNixPath(filepath.Join(wd, "prefix_test.go")),
		}, matches)
		require.Equal(t, fmt.Sprintf("&{fs:%s matchDirectoriesDirectly:false prefix:%s pattern:%s}", prefix, prefix, abs), w.String())
	})

	t.Run("real with rootfs on relative path", func(t *testing.T) {
		t.Parallel()

		pattern := "./*_test.go"

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, WriteOptions(&w))
		require.NoError(t, err)
		require.Equal(t, []string{
			"glob_test.go",
			"prefix_test.go",
		}, matches)
		require.Equal(t, "&{fs:. matchDirectoriesDirectly:false prefix:./ pattern:*_test.go}", w.String())
	})

	t.Run("real with rootfs on relative path match dir", func(t *testing.T) {
		t.Parallel()

		pattern := ".github"

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, MatchDirectoryAsFile, WriteOptions(&w))
		require.NoError(t, err)
		require.Equal(t, []string{
			".github",
		}, matches)
		require.Equal(t, "&{fs:. matchDirectoriesDirectly:true prefix:./ pattern:.github}", w.String())
	})

	t.Run("real with rootfs on relative path match dir", func(t *testing.T) {
		t.Parallel()

		pattern := ".github/workflows/"

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, MatchDirectoryAsFile, WriteOptions(&w))
		require.NoError(t, err)
		require.Equal(t, []string{
			".github/workflows",
		}, matches)
		require.Equal(t, "&{fs:. matchDirectoriesDirectly:true prefix:./ pattern:.github/workflows}", w.String())
	})

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		var w bytes.Buffer
		fsys := testFs(t, []string{
			"./c/file1.txt",
			"./a/nope/file1.txt",
			"./a/something",
			"./a/b/file1.txt",
			"./a/c/file2.txt",
			"./a/d/file1.txt",
		}, nil)
		matches, err := Glob("./a/*/*", WithFs(fsys), WriteOptions(&w))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file1.txt",
			"a/c/file2.txt",
			"a/d/file1.txt",
			"a/nope/file1.txt",
		}, matches)
		require.Equal(t, fmt.Sprintf("&{fs:%+v matchDirectoriesDirectly:false prefix:./ pattern:a/*/*}", fsys), w.String())
	})

	t.Run("single file", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/b/*", WithFs(testFs(t, []string{
			"./c/file1.txt",
			"./a/nope/file1.txt",
			"./a/b/file1.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{"a/b/file1.txt"}, matches)
	})

	t.Run("super asterisk", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/**/*", WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/d/file1.txt",
			"./a/e/f/file1.txt",
			"./a/b/file1.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file1.txt",
			"a/d/file1.txt",
			"a/e/f/file1.txt",
		}, matches)
	})

	t.Run("alternative matchers", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/{b,d}/file.txt", WithFs(testFs(t, []string{
			"a/b/file.txt",
			"a/c/file.txt",
			"a/d/file.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file.txt",
			"a/d/file.txt",
		}, matches)
	})

	t.Run("character list and range matchers", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("[!bc]/[a-z]/file[01].txt", WithFs(testFs(t, []string{
			"a/b/file0.txt",
			"a/c/file1.txt",
			"a/d/file2.txt",
			"a/0/file.txt",
			"b/b/file0.txt",
			"c/b/file0.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/file0.txt",
			"a/c/file1.txt",
		}, matches)
	})

	t.Run("nested matchers", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("{a,[0-9]b}.txt", WithFs(testFs(t, []string{
			"a.txt",
			"b.txt",
			"1b.txt",
			"ab.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"1b.txt",
			"a.txt",
		}, matches)
	})

	t.Run("single symbol wildcards", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a?.txt", WithFs(testFs(t, []string{
			"a.txt",
			"a1.txt",
			"ab.txt",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a1.txt",
			"ab.txt",
		}, matches)
	})

	t.Run("direct match", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/b/c", WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/b/c",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{"a/b/c"}, matches)
	})

	t.Run("direct match wildcard", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/b/c{a", QuoteMeta, WithFs(testFs(t, []string{
			"./a/nope.txt",
			"a/b/c{a",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{"a/b/c{a"}, matches)
	})

	t.Run("direct no match", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/b/d", WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/b/dc",
		}, nil)))
		require.EqualError(t, err, "matching \"./a/b/d\": file does not exist")
		require.True(t, errors.Is(err, os.ErrNotExist))
		require.Empty(t, matches)
	})

	t.Run("escaped direct no match", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/\\{b\\}", WithFs(testFs(t, nil, nil)))
		require.EqualError(t, err, "matching \"./a/{b}\": file does not exist")
		require.True(t, errors.Is(err, os.ErrNotExist))
		require.Empty(t, matches)
	})

	t.Run("direct no match escaped wildcards", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/b/c{a", QuoteMeta, WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/b/dc",
		}, nil)))
		require.EqualError(t, err, "matching \"./a/b/c{a\": file does not exist")
		require.Empty(t, matches)
	})

	t.Run("no matches", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("z/*", WithFs(testFs(t, []string{
			"./a/nope.txt",
		}, nil)))
		require.NoError(t, err)
		require.Empty(t, matches)
	})

	t.Run("empty folder", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a*", WithFs(testFs(t, nil, nil)))
		require.NoError(t, err)
		require.Empty(t, matches)
	})

	t.Run("escaped asterisk", func(t *testing.T) {
		t.Parallel()
		if isWindows() {
			t.Skip("can't create paths with * on Windows")
		}
		matches, err := Glob("a/\\*/b", WithFs(testFs(t, []string{
			"a/a/b",
			"a/*/b",
			"a/**/b",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/*/b",
		}, matches)
	})

	t.Run("escaped curly braces", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("\\{a,b\\}/c", WithFs(testFs(t, []string{
			"a/c",
			"b/c",
			"{a,b}/c",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"{a,b}/c",
		}, matches)
	})

	t.Run("invalid pattern", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("[*", WithFs(testFs(t, nil, nil)))
		require.EqualError(t, err, "compile glob pattern: unexpected end of input")
		require.Empty(t, matches)
	})

	t.Run("prefix is a file", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("ab/c/*", WithFs(testFs(t, []string{
			"ab/c",
			"ab/d",
			"ab/e",
		}, nil)))
		require.NoError(t, err)
		require.Empty(t, matches)
	})

	t.Run("match files in directories", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/{b,c}", WithFs(testFs(t, []string{
			"a/b/d",
			"a/b/e/f",
			"a/c",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b/d",
			"a/b/e/f",
			"a/c",
		}, matches)
	})

	t.Run("match directories directly", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/{b,c}", MatchDirectoryAsFile, WithFs(testFs(t, []string{
			"a/b/d",
			"a/b/e/f",
			"a/c",
		}, nil)))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b",
			"a/c",
		}, matches)
	})

	t.Run("match empty directory", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/{b,c}", MatchDirectoryAsFile, WithFs(testFs(t, []string{
			"a/b",
		}, []string{
			"a/c",
		})))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/b",
			"a/c",
		}, matches)
	})

	t.Run("pattern ending with star and subdir", func(t *testing.T) {
		t.Parallel()
		matches, err := Glob("a/*", MatchDirectoryIncludesContents, WithFs(testFs(t, []string{
			"./a/1.txt",
			"./a/2.txt",
			"./a/3.txt",
			"./a/b/4.txt",
		}, []string{
			"a",
			"a/b",
		})))
		require.NoError(t, err)
		require.Equal(t, []string{
			"a/1.txt",
			"a/2.txt",
			"a/3.txt",
			"a/b/4.txt",
		}, matches)
	})
}

func TestQuoteMeta(t *testing.T) {
	t.Parallel()
	matches, err := Glob("{a,b}/c", QuoteMeta, WithFs(testFs(t, []string{
		"a/c",
		"b/c",
		"{a,b}/c",
	}, nil)))
	require.NoError(t, err)
	require.Equal(t, []string{
		"{a,b}/c",
	}, matches)
}

func testFs(tb testing.TB, files, dirs []string) fs.FS {
	tb.Helper()

	tmpfs := testfs.New(tb)

	for _, file := range files {
		require.NoError(tb, tmpfs.MkdirAll(filepath.Dir(file), 0o764))
		require.NoError(tb, tmpfs.WriteFile(file, []byte(file), 0o654))
	}

	for _, dir := range dirs {
		require.NoError(tb, tmpfs.MkdirAll(dir, 0o764))
	}

	return tmpfs
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}
