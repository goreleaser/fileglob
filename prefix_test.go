package fileglob

import (
	"fmt"
	"testing"

	"github.com/gobwas/glob/syntax/ast"
	"github.com/gobwas/glob/syntax/lexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticPrefix(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		pattern string
		prefix  string
	}{
		{"/foo/b*ar/baz", "/foo"},
		{"foo/bar", "foo/bar"},
		{"/foo/bar/{b,p}az", "/foo/bar"},
		{"*/foo", "."},
		{"./", "."},
		{"fo\\*o/bar/b*z", "fo*o/bar"},
		{"/\\{foo\\}/bar", "/{foo}/bar"},
		{"C:/Path/To/Some/File", "C:/Path/To/Some/File"},
	}

	for _, testCase := range testCases {
		prefix, err := staticPrefix(testCase.pattern)
		require.NoError(t, err)
		assert.Equal(t, testCase.prefix, prefix)
	}
}

func TestContainsMatchers(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		pattern          string
		containsMatchers bool
	}{
		{"/a/*/b", true},
		{"\\{a\\}/\\*/", false},
		{"a/b/c", false},
		{"", false},
		{"\\*/\\?", false},
		{"*/\\?", true},
		{"\\{*\\}", true},
		{"{a,b}/c", true},
		{"\\{\\\\[a-z]", true},
	}

	for _, testCase := range testCases {
		_, err := ast.Parse(lexer.NewLexer(testCase.pattern))
		require.NoError(t, err)

		assert.Equal(t, testCase.containsMatchers, ContainsMatchers(testCase.pattern),
			fmt.Sprintf("pattern: %s", testCase.pattern))
	}
}

func TestValidPattern(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		pattern string
		valid   bool
	}{
		{"/a/*/b", true},
		{"{a[", false},
		{"[*]", true},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.valid, ValidPattern(testCase.pattern) == nil,
			fmt.Sprintf("pattern: %s", testCase.pattern))
	}
}
