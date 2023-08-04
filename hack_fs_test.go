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
	defer f.Close()

	tfs, err := NewFS(f)
	require.NoError(t, err)

	err = fstest.TestFS(tfs, "foo", "dir1/dir11")
	require.NoError(t, err)
}

func TestHackFSTarTestdata(t *testing.T) {
	for archive, expected := range map[string][]string{
		// TODO: fix following failing test
		// sparse-posix-1.0: failed TestReader: ReadAll(small amounts)
		// "sparse-formats.tar": {"end", "sparse-gnu", "sparse-posix-0.0", "sparse-posix-0.1", "sparse-posix-1.0"},

		//*
		"file-and-dir.tar":        {"dir", "small.txt"},
		"gnu-long-nul.tar":        {"0123456789"},
		"gnu-multi-hdrs.tar":      {"GNU2", "GNU2/GNU2", "GNU2/GNU2/long-path-name"},
		"gnu-nil-sparse-data.tar": {"sparse.db"},
		"gnu-nil-sparse-hole.tar": {"sparse.db"},
		"gnu-utf8.tar":            {"☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹"},
		"gnu.tar":                 {"small.txt", "small2.txt"},
		"hardlink.tar":            {"file.txt", "hard.txt"},
		"hdr-only.tar":            {"badlink", "dir", "fifo", "file", "hardlink", "null", "sda", "symlink"},
		"invalid-go17.tar":        {"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/foo"},
		"nil-uid.tar":             {"P1050238.JPG.log"},
		"pax-global-records.tar":  {"GlobalHead.0.0", "file1", "file2", "file3", "file4", "global1"},
		"pax-multi-hdrs.tar":      {"bar"},
		"pax-nil-sparse-data.tar": {"sparse.db"},
		"pax-nil-sparse-hole.tar": {"sparse.db"},
		"pax-pos-size-file.tar":   {"foo"},
		"pax-records.tar":         {"file"},
		"pax.tar":                 {"a", "a/123456789101112131415161718192021222324252627282930313233343536373839404142434445464748495051525354555657585960616263646566676869707172737475767778798081828384858687888990919293949596979899100", "a/b"},
		"star.tar":                {"small.txt", "small2.txt"},
		"trailing-slash.tar":      {"123456789", "123456789/123456789", "123456789/123456789/123456789", "123456789/123456789/123456789/123456789", "123456789/123456789/123456789/123456789/123456789", "123456789/123456789/123456789/123456789/123456789/123456789", "123456789/123456789/123456789/123456789/123456789/123456789/123456789", "123456789/123456789/123456789/123456789/123456789/123456789/123456789/123456789", "123456789/123456789/123456789/123456789/123456789/123456789/123456789/123456789/123456789", "123456789/123456789/123456789/123456789/123456789/123456789/123456789/123456789/123456789/123456789"},
		"ustar-file-devs.tar":     {"file"},
		"ustar-file-reg.tar":      {"foo"},
		"ustar.tar":               {"longname", "longname/longname", "longname/longname/longname", "longname/longname/longname/longname", "longname/longname/longname/longname/longname", "longname/longname/longname/longname/longname/longname", "longname/longname/longname/longname/longname/longname/longname", "longname/longname/longname/longname/longname/longname/longname/longname", "longname/longname/longname/longname/longname/longname/longname/longname/longname"},
		"v7.tar":                  {"small.txt", "small2.txt"},
		"writer.tar":              {"small.txt", "small2.txt"},
		"xattrs.tar":              {"small.txt", "small2.txt"},
		//*/

		// slow (~ 45s)
		// "gnu-incremental.tar": {"test2", "test2/foo", "test2/sparse"},

		// killed (out of memory?)
		// "gnu-sparse-big.tar": nil,
		// "pax-sparse-big.tar": nil,

		// Errors below are probably expected
		// archive with no valid path name
		// "gnu-not-utf8.tar": nil,

		// empty archive
		// "pax-path-hdr.tar": nil,

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
	} {
		t.Run(archive, func(t *testing.T) {
			t.Log(archive)
			archive := "tar/testdata/" + archive
			expected := expected

			t.Parallel()
			f, err := os.Open(archive)
			require.NoError(t, err)
			defer f.Close()

			tfs, err := NewFS(f)
			require.NoError(t, err)

			err = fstest.TestFS(tfs, expected...)
			require.NoError(t, err)
		})
	}
}
