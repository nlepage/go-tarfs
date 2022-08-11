package tarfs

import (
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestHackFS(t *testing.T) {
	f, err := os.Open("test.tar")
	require.NoError(t, err)

	tfs, err := NewFS(f)
	require.NoError(t, err)

	err = fstest.TestFS(tfs, "foo", "dir1/dir11")
	require.NoError(t, err)
}

func TestHackGnuNilSparseData(t *testing.T) {
	for _, s := range []string{
		"tar/testdata/gnu-nil-sparse-data.tar",
		"tar/testdata/gnu-nil-sparse-hole.tar",
		// "tar/testdata/gnu-sparse-big.tar", // out of memory
	} {
		t.Run(s, func(t *testing.T) {
			f, err := os.Open(s)
			require.NoError(t, err)

			tfs, err := NewFS(f)
			require.NoError(t, err)

			err = fstest.TestFS(tfs, "sparse.db")
			require.NoError(t, err)
		})
	}
}
