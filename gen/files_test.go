package gen

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/katzenpost/trunnel/ast"
	"github.com/katzenpost/trunnel/internal/test"
	"github.com/katzenpost/trunnel/parse"
)

func TestFilesBuild(t *testing.T) {
	dirs := []string{
		"testdata/valid",
		"../testdata/tor",
		"../testdata/trunnel",
	}
	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			groups, err := test.LoadFileGroups(dir)
			require.NoError(t, err)
			for _, group := range groups {
				t.Run(strings.Join(group, ","), func(t *testing.T) {
					Build(t, group)
				})
			}
		})
	}
}

func Build(t *testing.T, filenames []string) {
	// Match how the CLI actually invokes the generator: parse all files
	// in the group into a single []*ast.File, then call Marshallers once
	// to emit one Go source file for the whole package. Calling
	// Marshallers per-file would emit one source per input file with
	// duplicate package-level declarations (e.g. MaxParseSize), which
	// is not a usage pattern the generator supports.
	asts := make([]*ast.File, 0, len(filenames))
	for _, filename := range filenames {
		f, err := parse.File(filename)
		require.NoError(t, err)
		asts = append(asts, f)
	}

	src, err := Marshallers("pkg", asts)
	require.NoError(t, err)

	output, err := test.Build([][]byte{src})
	if err != nil {
		t.Fatal(string(output))
	}
}
