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
	for archive, expected := range map[string][]string{
		"file-and-dir.tar":        {"dir", "small.txt"},
		"gnu-long-nul.tar":        {"0123456789"},
		"gnu-nil-sparse-data.tar": {"sparse.db"},
		"gnu-nil-sparse-hole.tar": {"sparse.db"},
		"gnu-utf8.tar":            {"☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹"},
		"gnu.tar":                 {"small.txt", "small2.txt"},
		"hardlink.tar":            {"file.txt", "hard.txt"},
		"nil-uid.tar":             {"P1050238.JPG.log"},
		"pax-global-records.tar":  {"GlobalHead.0.0", "file1", "file2", "file3", "file4", "global1"},
		"pax-multi-hdrs.tar":      {"bar"},
		"pax-nil-sparse-data.tar": {"sparse.db"},
		"pax-nil-sparse-hole.tar": {"sparse.db"},
		"pax-pos-size-file.tar":   {"foo"},
		"pax-records.tar":         {"file"},
		"star.tar":                {"small.txt", "small2.txt"},
		"ustar-file-devs.tar":     {"file"},
		"ustar-file-reg.tar":      {"foo"},
		"v7.tar":                  {"small.txt", "small2.txt"},
		"writer.tar":              {"small.txt", "small2.txt"},
		"xattrs.tar":              {"small.txt", "small2.txt"},

		// fs.ReadDir: list not sorted: badlink before badlink
		// "hdr-only.tar": {"badlink", "dir", "fifo", "file", "hardlink", "null", "sda", "symlink"},

		// sparse-posix-1.0: failed TestReader: ReadAll(small amounts)
		// "sparse-formats.tar": {"end", "sparse-gnu", "sparse-posix-0.0", "sparse-posix-0.1", "sparse-posix-1.0"},

		// .: Open: open .: file does not exist
		// "gnu-multi-hdrs.tar": nil,
		// "gnu-not-utf8.tar": nil,
		// "invalid-go17.tar": nil,
		// "pax-path-hdr.tar": nil,
		// "pax.tar": nil,
		// "trailing-slash.tar": nil,
		// "ustar.tar": nil,

		// archive/tar: invalid tar header
		// "issue10968.tar": nil,
		// "issue11169.tar": nil,
		// "issue12435.tar": nil,
		// "neg-size.tar": nil,
		// "pax-bad-hdr-file.tar": nil,
		// "pax-bad-mtime-file.tar": nil,
		// "pax-nul-path.tar": nil,
		// "pax-nul-xattrs.tar": nil,

		// unexpected EOF
		// "writer-big-long.tar": nil,
		// "writer-big.tar": nil,

		// slow (~ 45s)
		// "gnu-incremental.tar": {"test2", "test2/foo", "test2/sparse"},

		// killed (out of memory?
		// "gnu-sparse-big.tar": nil,
		// "pax-sparse-big.tar": nil,
	} {
		t.Run(archive, func(t *testing.T) {
			t.Log(archive)
			archive := "tar/testdata/" + archive
			expected := expected

			t.Parallel()
			f, err := os.Open(archive)
			require.NoError(t, err)

			tfs, err := NewFS(f)
			require.NoError(t, err)

			err = fstest.TestFS(tfs, expected...)
			require.NoError(t, err)
		})
	}
}
