package tv

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/katzenpost/trunnel/fault"
	"github.com/katzenpost/trunnel/inspect"
	"github.com/katzenpost/trunnel/internal/test"
	"github.com/katzenpost/trunnel/parse"
)

func TestFiles(t *testing.T) {
	dirs := []string{
		"../testdata/tor",
		"../testdata/trunnel",
	}
	for _, dir := range dirs {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			groups, err := test.LoadFileGroups(dir)
			require.NoError(t, err)
			for _, group := range groups {
				t.Run(strings.Join(group, ","), func(t *testing.T) {
					VerifyGroup(t, group)
				})
			}
		})
	}
}

func VerifyGroup(t *testing.T, filenames []string) {
	fs, err := parse.Files(filenames)
	require.NoError(t, err)

	c, err := GenerateFiles(fs, WithSelector(RandomSampleSelector(16)))
	if err == fault.ErrNotImplemented {
		t.Log(err)
		t.SkipNow()
	}
	require.NoError(t, err)

	r, err := inspect.NewResolverFiles(fs)
	require.NoError(t, err)

	for _, s := range r.Structs() {
		if s.Extern() {
			continue
		}
		t.Run(s.Name, func(t *testing.T) {
			num := len(c.Vectors(s.Name))
			t.Logf("%d test vectors for %s", num, s.Name)
			assert.True(t, num > 0)
		})
	}
}
