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
