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
	"github.com/matryer/is"
)

func TestGlob(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()
	t.Run("real", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		var w bytes.Buffer
		matches, err := Glob("*_test.go", WriteOptions(&w))
		is.NoErr(err)
		is.Equal([]string{
			"glob_test.go",
			"prefix_test.go",
		}, matches)
		is.Equal("&{fs:. matchDirectoriesDirectly:false prefix:./ pattern:*_test.go}", w.String())
	})

	// https://github.com/goreleaser/fileglob/issues/54
	t.Run("real with absolute path and no options", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		wd, err := os.Getwd()
		is.NoErr(err)

		prefix := "/"
		if isWindows() {
			prefix = filepath.VolumeName(wd) + "/"
		}

		pattern := toNixPath(filepath.Join(wd, "*_test.go"))

		var w bytes.Buffer
		matches, err := Glob(pattern, WriteOptions(&w))
		is.NoErr(err)
		is.Equal([]string{
			toNixPath(filepath.Join(wd, "glob_test.go")),
			toNixPath(filepath.Join(wd, "prefix_test.go")),
		}, matches)
		is.Equal(fmt.Sprintf(
			"&{fs:%s matchDirectoriesDirectly:false prefix:%s pattern:%s}",
			prefix, prefix, pattern,
		), w.String())
	})

	// https://github.com/goreleaser/fileglob/issues/54
	t.Run("real with relative path to parent and no options", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		wd, err := os.Getwd()
		is.NoErr(err)

		matches, err := Glob("../" + filepath.Base(wd) + "/*_test.go")
		is.NoErr(err)
		is.Equal([]string{
			toNixPath(filepath.Join(wd, "glob_test.go")),
			toNixPath(filepath.Join(wd, "prefix_test.go")),
		}, matches)
	})

	t.Run("real with rootfs", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		wd, err := os.Getwd()
		is.NoErr(err)

		prefix := "/"
		if isWindows() {
			prefix = filepath.VolumeName(wd) + "/"
		}

		pattern := toNixPath(filepath.Join(wd, "*_test.go"))

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, WriteOptions(&w))
		is.NoErr(err)
		is.Equal([]string{
			toNixPath(filepath.Join(wd, "glob_test.go")),
			toNixPath(filepath.Join(wd, "prefix_test.go")),
		}, matches)
		is.Equal(fmt.Sprintf(
			"&{fs:%s matchDirectoriesDirectly:false prefix:%s pattern:%s}",
			prefix, prefix, pattern,
		), w.String())
	})

	t.Run("real with rootfs direct file", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		wd, err := os.Getwd()
		is.NoErr(err)

		pattern := toNixPath(filepath.Join(wd, "prefix.go"))

		matches, err := Glob(pattern, MaybeRootFS)
		is.NoErr(err)
		is.Equal([]string{
			toNixPath(filepath.Join(wd, "prefix.go")),
		}, matches)
	})

	t.Run("real with rootfs on relative path to parent disable globbing", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		wd, err := os.Getwd()
		is.NoErr(err)

		dir := filepath.Base(wd)

		prefix := "/"
		if isWindows() {
			prefix = filepath.VolumeName(wd) + "/"
		}

		pattern := "../" + dir + "/{file}["

		abs, err := filepath.Abs(pattern)
		is.NoErr(err)
		abs = toNixPath(abs)

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, QuoteMeta, WriteOptions(&w))
		is.True(err != nil)                                            // expected an error
		is.True(strings.HasSuffix(err.Error(), "file does not exist")) // should have been file does not exist
		is.Equal([]string{}, matches)
		is.Equal(fmt.Sprintf(
			"&{fs:%s matchDirectoriesDirectly:false prefix:%s pattern:%s}",
			prefix, prefix, glob.QuoteMeta(abs),
		), w.String())
	})

	t.Run("real with rootfs on relative path to parent", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		wd, err := os.Getwd()
		is.NoErr(err)

		dir := filepath.Base(wd)

		prefix := "/"
		if isWindows() {
			prefix = filepath.VolumeName(wd) + "/"
		}

		pattern := "../" + dir + "/*_test.go"
		abs, err := filepath.Abs(pattern)
		is.NoErr(err)

		abs = toNixPath(abs)

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, WriteOptions(&w))
		is.NoErr(err)
		is.Equal([]string{
			toNixPath(filepath.Join(wd, "glob_test.go")),
			toNixPath(filepath.Join(wd, "prefix_test.go")),
		}, matches)
		is.Equal(fmt.Sprintf("&{fs:%s matchDirectoriesDirectly:false prefix:%s pattern:%s}", prefix, prefix, abs), w.String())
	})

	t.Run("real with rootfs on relative path", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		pattern := "./*_test.go"

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, WriteOptions(&w))
		is.NoErr(err)
		is.Equal([]string{
			"glob_test.go",
			"prefix_test.go",
		}, matches)
		is.Equal("&{fs:. matchDirectoriesDirectly:false prefix:./ pattern:./*_test.go}", w.String())
	})

	t.Run("real with rootfs on relative path match dir", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		pattern := ".github"

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, MatchDirectoryAsFile, WriteOptions(&w))
		is.NoErr(err)
		is.Equal([]string{
			".github",
		}, matches)
		is.Equal("&{fs:. matchDirectoriesDirectly:true prefix:./ pattern:.github}", w.String())
	})

	t.Run("real with rootfs on relative path match dir", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		pattern := ".github/workflows/"

		var w bytes.Buffer
		matches, err := Glob(pattern, MaybeRootFS, MatchDirectoryAsFile, WriteOptions(&w))
		is.NoErr(err)
		is.Equal([]string{
			".github/workflows",
		}, matches)
		is.Equal("&{fs:. matchDirectoriesDirectly:true prefix:./ pattern:.github/workflows/}", w.String())
	})

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
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
		is.NoErr(err)
		is.Equal([]string{
			"a/b/file1.txt",
			"a/c/file2.txt",
			"a/d/file1.txt",
			"a/nope/file1.txt",
		}, matches)
		is.Equal(fmt.Sprintf("&{fs:%+v matchDirectoriesDirectly:false prefix:./ pattern:./a/*/*}", fsys), w.String())
	})

	t.Run("single file", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/b/*", WithFs(testFs(t, []string{
			"./c/file1.txt",
			"./a/nope/file1.txt",
			"./a/b/file1.txt",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{"a/b/file1.txt"}, matches)
	})

	t.Run("super asterisk", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/**/*", WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/d/file1.txt",
			"./a/e/f/file1.txt",
			"./a/b/file1.txt",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{
			"a/b/file1.txt",
			"a/d/file1.txt",
			"a/e/f/file1.txt",
		}, matches)
	})

	t.Run("alternative matchers", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/{b,d}/file.txt", WithFs(testFs(t, []string{
			"a/b/file.txt",
			"a/c/file.txt",
			"a/d/file.txt",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{
			"a/b/file.txt",
			"a/d/file.txt",
		}, matches)
	})

	t.Run("character list and range matchers", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("[!bc]/[a-z]/file[01].txt", WithFs(testFs(t, []string{
			"a/b/file0.txt",
			"a/c/file1.txt",
			"a/d/file2.txt",
			"a/0/file.txt",
			"b/b/file0.txt",
			"c/b/file0.txt",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{
			"a/b/file0.txt",
			"a/c/file1.txt",
		}, matches)
	})

	t.Run("nested matchers", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("{a,[0-9]b}.txt", WithFs(testFs(t, []string{
			"a.txt",
			"b.txt",
			"1b.txt",
			"ab.txt",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{
			"1b.txt",
			"a.txt",
		}, matches)
	})

	t.Run("single symbol wildcards", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a?.txt", WithFs(testFs(t, []string{
			"a.txt",
			"a1.txt",
			"ab.txt",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{
			"a1.txt",
			"ab.txt",
		}, matches)
	})

	t.Run("direct match", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/b/c", WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/b/c",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{"a/b/c"}, matches)
	})

	t.Run("direct match wildcard", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/b/c{a", QuoteMeta, WithFs(testFs(t, []string{
			"./a/nope.txt",
			"a/b/c{a",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{"a/b/c{a"}, matches)
	})

	t.Run("direct no match", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/b/d", WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/b/dc",
		}, nil)))
		is.True(err != nil) // expected an err
		is.Equal(err.Error(), "matching \"./a/b/d\": file does not exist")
		is.True(errors.Is(err, os.ErrNotExist))
		is.Equal([]string{}, matches)
	})

	t.Run("escaped direct no match", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/\\{b\\}", WithFs(testFs(t, nil, nil)))
		is.True(err != nil) // expected an err
		is.Equal(err.Error(), "matching \"./a/{b}\": file does not exist")
		is.True(errors.Is(err, os.ErrNotExist))
		is.Equal([]string{}, matches)
	})

	t.Run("direct no match escaped wildcards", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/b/c{a", QuoteMeta, WithFs(testFs(t, []string{
			"./a/nope.txt",
			"./a/b/dc",
		}, nil)))
		is.True(err != nil) // expected an err
		is.Equal(err.Error(), "matching \"./a/b/c{a\": file does not exist")
		is.Equal([]string{}, matches)
	})

	t.Run("no matches", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("z/*", WithFs(testFs(t, []string{
			"./a/nope.txt",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{}, matches)
	})

	t.Run("empty folder", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a*", WithFs(testFs(t, nil, nil)))
		is.NoErr(err)
		is.Equal(nil, matches)
	})

	t.Run("escaped asterisk", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		if isWindows() {
			t.Skip("can't create paths with * on Windows")
		}
		matches, err := Glob("a/\\*/b", WithFs(testFs(t, []string{
			"a/a/b",
			"a/*/b",
			"a/**/b",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{
			"a/*/b",
		}, matches)
	})

	t.Run("escaped curly braces", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("\\{a,b\\}/c", WithFs(testFs(t, []string{
			"a/c",
			"b/c",
			"{a,b}/c",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{
			"{a,b}/c",
		}, matches)
	})

	t.Run("invalid pattern", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("[*", WithFs(testFs(t, nil, nil)))
		is.True(err != nil) // expected an error
		is.Equal(err.Error(), "compile glob pattern: unexpected end of input")
		is.Equal(nil, matches)
	})

	t.Run("prefix is a file", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("ab/c/*", WithFs(testFs(t, []string{
			"ab/c",
			"ab/d",
			"ab/e",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{}, matches)
	})

	t.Run("match files in directories", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/{b,c}", WithFs(testFs(t, []string{
			"a/b/d",
			"a/b/e/f",
			"a/c",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{
			"a/b/d",
			"a/b/e/f",
			"a/c",
		}, matches)
	})

	t.Run("match directories directly", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/{b,c}", MatchDirectoryAsFile, WithFs(testFs(t, []string{
			"a/b/d",
			"a/b/e/f",
			"a/c",
		}, nil)))
		is.NoErr(err)
		is.Equal([]string{
			"a/b",
			"a/c",
		}, matches)
	})

	t.Run("match empty directory", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/{b,c}", MatchDirectoryAsFile, WithFs(testFs(t, []string{
			"a/b",
		}, []string{
			"a/c",
		})))
		is.NoErr(err)
		is.Equal([]string{
			"a/b",
			"a/c",
		}, matches)
	})

	t.Run("pattern ending with star and subdir", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)
		matches, err := Glob("a/*", MatchDirectoryIncludesContents, WithFs(testFs(t, []string{
			"./a/1.txt",
			"./a/2.txt",
			"./a/3.txt",
			"./a/b/4.txt",
		}, []string{
			"a",
			"a/b",
		})))
		is.NoErr(err)
		is.Equal([]string{
			"a/1.txt",
			"a/2.txt",
			"a/3.txt",
			"a/b/4.txt",
		}, matches)
	})

	t.Run("symlinks", func(t *testing.T) {
		t.Parallel()
		var fsPath string
		if testFS, ok := testFs(t, []string{"./a/file"}, nil).(testfs.FS); ok {
			fsPath = testFS.Path()
		}

		t.Run("good", func(t *testing.T) {
			is := is.New(t)
			workingSymlink := filepath.Join(fsPath, "b")
			is.NoErr(os.Symlink("a", workingSymlink))
			matches, err := Glob(workingSymlink)
			is.NoErr(err)
			is.Equal([]string{
				workingSymlink,
			}, matches)
		})

		t.Run("broken", func(t *testing.T) {
			is := is.New(t)
			brokenSymlink := filepath.Join(fsPath, "c")
			is.NoErr(os.Symlink("non-existent", brokenSymlink))

			matches, err := Glob(brokenSymlink)
			is.NoErr(err)
			is.Equal([]string{
				brokenSymlink,
			}, matches)
		})
	})
}

func TestUnclosedAlternatives(t *testing.T) {
	t.Parallel()
	for _, pattern := range []string{
		"{",
		"{foo",
		"{a,",
		"{a,{b,c}",
		"{a,b}{c",
		"{a/b,c/d",
		"missing/{a,",
		"*/{foo",
		`{a\}`,
		"{a[}]",
		`\\{foo`,
	} {
		t.Run(pattern, func(t *testing.T) {
			t.Parallel()
			t.Run("ValidPattern", func(t *testing.T) {
				t.Parallel()
				is := is.New(t)
				err := ValidPattern(pattern)
				is.True(err != nil) // unfinished alternatives must be invalid
				is.Equal("unexpected end of input: unclosed alternative group", err.Error())
			})

			t.Run("ContainsMatchers", func(t *testing.T) {
				t.Parallel()
				is.New(t).True(!ContainsMatchers(pattern))
			})

			t.Run("Glob", func(t *testing.T) {
				t.Parallel()
				is := is.New(t)
				matches, err := Glob(pattern, WithFs(testFs(t, []string{
					"a/contents.txt",
					"foo/contents.txt",
					"b",
					"c",
				}, nil)))
				t.Logf("Glob(%q) = %q, %v", pattern, matches, err)
				is.True(err != nil) // reject before returning matches or a filesystem error
				is.Equal("compile glob pattern: unexpected end of input: unclosed alternative group", err.Error())
				is.Equal(nil, matches)
			})
		})
	}
}

func TestGlobValidAlternatives(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		pattern string
		matches []string
	}{
		{"{a/b,c/d}.txt", []string{"a/b.txt", "c/d.txt"}},
		{"root/{a/b,c/d}.txt", []string{"root/a/b.txt", "root/c/d.txt"}},
		{"{a/{b,c},d/e}.txt", []string{"a/b.txt", "a/c.txt", "d/e.txt"}},
		{"{a,{b,c}}.txt", []string{"a.txt", "b.txt", "c.txt"}},
		{`\{`, []string{"{"}},
		{`\{foo`, []string{"{foo"}},
		{`\{a,`, []string{"{a,"}},
		{"[{].txt", []string{"{.txt"}},
		{"[}].txt", []string{"}.txt"}},
		{"[{a].txt", []string{"a.txt", "{.txt"}},
		{"{a,[}]}.txt", []string{"a.txt", "}.txt"}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.pattern, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			matches, err := Glob(testCase.pattern, WithFs(testFs(t, []string{
				"a/b.txt",
				"a/c.txt",
				"c/d.txt",
				"d/e.txt",
				"root/a/b.txt",
				"root/c/d.txt",
				"root/c/nope.txt",
				"a.txt",
				"b.txt",
				"c.txt",
				"nope.txt",
				"{",
				"{foo",
				"{a,",
				"{.txt",
				"}.txt",
			}, nil)))
			is.NoErr(err)
			is.Equal(testCase.matches, matches)
		})
	}
}

func TestQuoteMeta(t *testing.T) {
	t.Parallel()
	for _, pattern := range []string{"{a,b}/c", "{", "{foo", "{a,", "{a/b,c/d"} {
		t.Run(pattern, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			matches, err := Glob(pattern, QuoteMeta, WithFs(testFs(t, []string{
				"a/c",
				"b/c",
				pattern,
			}, nil)))
			is.NoErr(err)
			is.Equal([]string{pattern}, matches)
		})
	}
}

func testFs(tb testing.TB, files, dirs []string) fs.FS {
	tb.Helper()

	tmpfs := testfs.New(tb)
	is := is.New(tb)

	for _, file := range files {
		is.NoErr(tmpfs.MkdirAll(filepath.Dir(file), 0o764))
		is.NoErr(tmpfs.WriteFile(file, []byte(file), 0o654))
	}

	for _, dir := range dirs {
		is.NoErr(tmpfs.MkdirAll(dir, 0o764))
	}

	return tmpfs
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}
