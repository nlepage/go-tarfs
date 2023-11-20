package tarfs

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFS(t *testing.T) {
	require := require.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	err = fstest.TestFS(tfs, "bar", "foo", "dir1", "dir1/dir11", "dir1/dir11/file111", "dir1/file11", "dir1/file12", "dir2", "dir2/dir21", "dir2/dir21/file211", "dir2/dir21/file212")
	require.NoError(err)
}

func TestOpenInvalid(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	for _, name := range []string{"/foo", "./foo", "foo/", "foo/../foo", "foo//bar"} {
		_, err := tfs.Open(name)
		assert.ErrorIsf(err, fs.ErrInvalid, "when tarfs.Open(%#v)", name)
	}
}

func TestOpenNotExist(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	for _, name := range []string{"baz", "qwe", "foo/bar", "file11"} {
		_, err := tfs.Open(name)
		assert.ErrorIsf(err, fs.ErrNotExist, "when tarfs.Open(%#v)", name)
	}
}

func TestOpenThenStat(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	for _, file := range []struct {
		path  string
		name  string
		isDir bool
	}{
		{"foo", "foo", false},
		{"bar", "bar", false},
		{"dir1", "dir1", true},
		{"dir1/file11", "file11", false},
		{".", ".", true},
	} {
		f, err := tfs.Open(file.path)
		if !assert.NoErrorf(err, "when tarfs.Open(%#v)", file.path) {
			continue
		}

		fi, err := f.Stat()
		if !assert.NoErrorf(err, "when file{%#v}.Stat()", file.path) {
			continue
		}

		assert.Equalf(file.name, fi.Name(), "file{%#v}.Stat().Name()", file.path)
		assert.Equalf(file.isDir, fi.IsDir(), "file{%#v}.Stat().IsDir()", file.path)
	}
}

func TestOpenThenReadAll(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	for _, file := range []struct {
		path    string
		content []byte
	}{
		{"foo", []byte("foo")},
		{"bar", []byte("bar")},
		{"dir1/file11", []byte("file11")},
	} {
		f, err := tfs.Open(file.path)
		if !assert.NoErrorf(err, "when tarfs.Open(%#v)", file.path) {
			continue
		}

		content, err := io.ReadAll(f)
		if !assert.NoErrorf(err, "when io.ReadAll(file{%#v})", file.path) {
			continue
		}

		assert.Equalf(file.content, content, "content of %#v", file.path)
	}
}

func TestOpenThenSeekAfterEnd(t *testing.T) {
	require := require.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	r, err := tfs.Open("foo")
	require.NoError(err, "when tarfs.Open(foo)")

	rs := r.(io.ReadSeeker)

	abs, err := rs.Seek(10, io.SeekStart)
	require.NoError(err, "when ReadSeeker.Seek(10, io.SeekStart)")
	require.Equal(int64(10), abs, "when ReadSeeker.Seek(10, io.SeekStart)")

	b := make([]byte, 0, 1)
	_, err = rs.Read(b)
	require.ErrorIs(err, io.EOF, "when ReadSeeker.Read([]byte)")
}

func TestReadDir(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	for _, dir := range []struct {
		name       string
		entriesLen int
	}{
		{".", 4},
		{"dir1", 3},
		{"dir2/dir21", 2},
	} {
		entries, err := fs.ReadDir(tfs, dir.name)
		if !assert.NoErrorf(err, "when fs.ReadDir(tfs, %#v)", dir.name) {
			continue
		}

		assert.Equalf(dir.entriesLen, len(entries), "len(entries) for %#v", dir.name)
	}
}

func TestReadDirNotDir(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	for _, name := range []string{"foo", "dir1/file12"} {
		_, err := fs.ReadDir(tfs, name)
		assert.ErrorIsf(err, ErrNotDir, "when tarfs.ReadDir(tfs, %#v)", name)
	}
}

func TestReadFile(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	for _, file := range []struct {
		path    string
		content string
	}{
		{"bar", "bar"},
		{"dir1/dir11/file111", "file111"},
		{"dir1/file11", "file11"},
		{"dir1/file12", "file12"},
		{"dir2/dir21/file211", "file211"},
		{"dir2/dir21/file212", "file212"},
		{"foo", "foo"},
	} {
		b, err := fs.ReadFile(tfs, file.path)
		if !assert.NoErrorf(err, "when fs.ReadFile(tfs, %#v)", file.path) {
			continue
		}

		assert.Equalf(file.content, string(b), "in %#v", file.path)
	}
}

func TestStat(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	for _, file := range []struct {
		path  string
		name  string
		isDir bool
	}{
		{"dir1/dir11/file111", "file111", false},
		{"foo", "foo", false},
		{"dir2/dir21", "dir21", true},
		{".", ".", true},
	} {
		fi, err := fs.Stat(tfs, file.path)
		if !assert.NoErrorf(err, "when fs.Stat(tfs, %#v)", file.path) {
			continue
		}

		assert.Equalf(file.name, fi.Name(), "FileInfo{%#v}.Name()", file.path)

		assert.Equalf(file.isDir, fi.IsDir(), "FileInfo{%#v}.IsDir()", file.path)
	}
}

func TestGlob(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	for pattern, expected := range map[string][]string{
		"*/*2*":   {"dir1/file12", "dir2/dir21"},
		"*":       {"bar", "dir1", "dir2", "foo", "."},
		"*/*/*":   {"dir1/dir11/file111", "dir2/dir21/file211", "dir2/dir21/file212"},
		"*/*/*/*": nil,
	} {
		actual, err := fs.Glob(tfs, pattern)
		if !assert.NoErrorf(err, "when fs.Glob(tfs, %#v)", pattern) {
			continue
		}

		assert.ElementsMatchf(expected, actual, "matches for pattern %#v", pattern)
	}
}

func TestSubThenReadDir(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	for _, dir := range []struct {
		name       string
		entriesLen int
	}{
		{".", 4},
		{"dir1", 3},
		{"dir2/dir21", 2},
	} {
		subfs, err := fs.Sub(tfs, dir.name)
		if !assert.NoErrorf(err, "when fs.Sub(tfs, %#v)", dir.name) {
			continue
		}

		entries, err := fs.ReadDir(subfs, ".")
		if !assert.NoErrorf(err, "when fs.ReadDir(subfs, %#v)", dir.name) {
			continue
		}

		assert.Equalf(dir.entriesLen, len(entries), "len(entries) for %#v", dir.name)
	}
}

func TestSubThenReadFile(t *testing.T) {
	require := require.New(t)

	f, err := os.Open("test.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	name := "dir2"

	subfs, err := fs.Sub(tfs, name)
	require.NoErrorf(err, "when fs.Sub(tfs, %#v)", name)

	name = "dir21/file211"
	content := "file211"

	b, err := fs.ReadFile(subfs, name)
	require.NoErrorf(err, "when fs.ReadFile(subfs, %#v)", name)

	require.Equalf(content, string(b), "in %#v", name)
}

func TestReadOnDir(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	tf, err := os.Open("test.tar")
	require.NoError(err)
	defer tf.Close()

	tfs, err := New(tf)
	require.NoError(err)

	var dirs = []string{"dir1", "dir2/dir21", "."}

	for _, name := range dirs {
		f, err := tfs.Open(name)
		if !assert.NoErrorf(err, "when fs.ReadFile(subfs, %#v)", name) {
			continue
		}

		_, err = f.Read(make([]byte, 1))
		assert.ErrorIsf(err, ErrDir, "when file{%#v}.Read()", name)

		_, err = fs.ReadFile(tfs, name)
		assert.ErrorIsf(err, ErrDir, "fs.ReadFile(tfs, %#v)", name)
	}
}

func TestWithDotDirInArchive(t *testing.T) {
	require := require.New(t)

	f, err := os.Open("test-with-dot-dir.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	err = fstest.TestFS(tfs, "bar", "foo", "dir1", "dir1/dir11", "dir1/dir11/file111", "dir1/file11", "dir1/file12", "dir2", "dir2/dir21", "dir2/dir21/file211", "dir2/dir21/file212")
	require.NoError(err)
}

func TestWithNoDirEntriesInArchive(t *testing.T) {
	require := require.New(t)

	f, err := os.Open("test-no-directory-entries.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	err = fstest.TestFS(tfs, "bar", "foo", "dir1", "dir1/dir11", "dir1/dir11/file111", "dir1/file11", "dir1/file12", "dir2", "dir2/dir21", "dir2/dir21/file211", "dir2/dir21/file212")
	require.NoError(err)
}

func TestSparse(t *testing.T) {
	require, assert := require.New(t), assert.New(t)

	f, err := os.Open("test-sparse.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	err = fstest.TestFS(tfs, "file1", "file2")
	assert.NoError(err)

	if file1Actual, err := fs.ReadFile(tfs, "file1"); assert.NoError(err, "fs.ReadFile(tfs, \"file1\")") {
		file1Expected := make([]byte, 1000000)
		copy(file1Expected, []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
		copy(file1Expected[999990:], []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
		assert.Equal(file1Expected, file1Actual, "fs.ReadFile(tfs, \"file1\")")
	}

	if file2Actual, err := fs.ReadFile(tfs, "file2"); assert.NoError(err, "fs.ReadFile(tfs, \"file2\")") {
		assert.Equal([]byte("file2"), file2Actual, "fs.ReadFile(tfs, \"file2\")")
	}
}

func TestIgnoreGlobalHeader(t *testing.T) {
	require := require.New(t)

	// This file was created by initializing a git repository,
	// committing a few files, and running: `git archive HEAD`
	f, err := os.Open("test-with-global-header.tar")
	require.NoError(err)
	defer f.Close()

	tfs, err := New(f)
	require.NoError(err)

	err = fstest.TestFS(tfs, "bar", "dir1", "dir1/file11")
	require.NoError(err)
}

func TestVariousTarTypes(t *testing.T) {
	assert := assert.New(t)

	for _, file := range []struct {
		path      string
		expecteds []string
	}{
		{"file-and-dir.tar", []string{"small.txt"}},
		{"gnu.tar", []string{"small.txt", "small2.txt"}},
		// {"gnu-incremental.tar", []string{"test2/foo", "test2/sparse"}},
		{"gnu-long-nul.tar", []string{"0123456789"}},
		{"gnu-multi-hdrs.tar", []string{"GNU2/GNU2/long-path-name"}},
		{"gnu-nil-sparse-data.tar", []string{"sparse.db"}},
		{"gnu-nil-sparse-hole.tar", []string{"sparse.db"}},
		{"gnu-not-utf8.tar", []string{"hi\200\201\202\203bye"}},
		// gnu-sparse-big.tar: too big
		{"gnu-utf8.tar", []string{"☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹☺☻☹"}},
		{"hardlink.tar", []string{"file.txt", "hard.txt"}},
		// {"hdr-only.tar", []string{"file"}},
	} {
		func() {
			f, err := os.Open(filepath.Join(runtime.GOROOT(), "src/archive/tar/testdata", file.path))
			if !assert.NoError(err) {
				return
			}
			defer f.Close()

			tfs, err := New(f)
			if !assert.NoError(err) {
				return
			}
			assert.NoError(err)

			err = fstest.TestFS(tfs, file.expecteds...)
			if !assert.NoError(err) {
				return
			}
			assert.NoError(err)
		}()
	}
}
