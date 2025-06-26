package tarfs

import (
	"archive/tar"
	"bytes"
	"io"
	"io/fs"
	"path"
	"strings"
)

type tarfs struct {
	entries map[string]fs.DirEntry
}

var _ fs.FS = &tarfs{}

// New creates a new tar fs.FS from r.
// If r implements io.ReaderAt:
// - files content are not stored in memory
// - r must stay opened while using the fs.FS
func New(r io.Reader) (fs.FS, error) {
	tfs := &tarfs{make(map[string]fs.DirEntry)}
	tfs.entries["."] = newDirEntry(fs.FileInfoToDirEntry(fakeDirFileInfo(".")))

	ra, isReaderAt := r.(readReaderAt)
	if !isReaderAt {
		buf, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		ra = bytes.NewReader(buf)
	}

	var cr readCounterIface
	if rs, isReadSeeker := ra.(io.ReadSeeker); isReadSeeker {
		cr = &readSeekCounter{ReadSeeker: rs}
	} else {
		cr = &readCounter{Reader: ra}
	}

	tr := tar.NewReader(cr)

	var (
		bodyEnd, headerStart, headerEnd int64
		blk                             block
	)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		headerEnd = cr.Count()

		holes, _ := reconstructSparse(io.NewSectionReader(ra, headerStart, headerEnd-headerStart), h, &blk)

		switch h.Typeflag {
		case tar.TypeLink, tar.TypeSymlink, tar.TypeChar, tar.TypeBlock, tar.TypeDir, tar.TypeFifo,
			tar.TypeCont, tar.TypeXHeader, tar.TypeXGlobalHeader,
			tar.TypeGNULongName, tar.TypeGNULongLink:
			// They have size for name.
			bodyEnd = headerEnd
		default:
			// In testdata tars there's typeflag value not defined in archive/tar
			// nor in https://www.gnu.org/software/tar/manual/html_node/Standard.html
			bodyEnd = headerEnd + h.Size
			if holes != nil {
				// reverse-caluculating size
				var holeSize int64
				for _, hole := range holes {
					holeSize += hole.Length
				}
				bodyEnd -= holeSize
			}
		}

		de := fs.FileInfoToDirEntry(h.FileInfo())
		name := path.Clean(h.Name)
		if h.Typeflag != tar.TypeXGlobalHeader && name != "." {
			if h.FileInfo().IsDir() {
				tfs.append(name, newDirEntry(de))
			} else {
				tfs.append(name, &regEntry{de, name, ra, headerStart})
			}
		}

		headerStart = bodyEnd + ((-bodyEnd) & (blockSize - 1))
	}

	return tfs, nil
}

func (tfs *tarfs) append(name string, e fs.DirEntry) {
	tfs.entries[name] = e

	dir := path.Dir(name)

	if parent, ok := tfs.entries[dir]; ok {
		parent := parent.(*dirEntry)
		parent.append(e)
		return
	}

	parent := newDirEntry(fs.FileInfoToDirEntry(fakeDirFileInfo(path.Base(dir))))

	tfs.append(dir, parent)

	parent.append(e)
}

func (tfs *tarfs) Open(name string) (fs.File, error) {
	const op = "open"

	e, err := tfs.get(op, name)
	if err != nil {
		return nil, err
	}

	return e.open()
}

var _ fs.ReadDirFS = &tarfs{}

func (tfs *tarfs) ReadDir(name string) ([]fs.DirEntry, error) {
	e, err := tfs.get("readdir", name)
	if err != nil {
		return nil, err
	}

	return e.readdir(name)
}

var _ fs.ReadFileFS = &tarfs{}

func (tfs *tarfs) ReadFile(name string) ([]byte, error) {
	e, err := tfs.get("readfile", name)
	if err != nil {
		return nil, err
	}

	return e.readfile(name)
}

var _ fs.StatFS = &tarfs{}

func (tfs *tarfs) Stat(name string) (fs.FileInfo, error) {
	e, err := tfs.get("stat", name)
	if err != nil {
		return nil, err
	}

	return e.Info()
}

var _ fs.GlobFS = &tarfs{}

func (tfs *tarfs) Glob(pattern string) (matches []string, _ error) {
	for name := range tfs.entries {
		match, err := path.Match(pattern, name)
		if err != nil {
			return nil, err
		}
		if match {
			matches = append(matches, name)
		}
	}
	return
}

var _ fs.SubFS = &tarfs{}

func (tfs *tarfs) Sub(dir string) (fs.FS, error) {
	const op = "sub"

	if dir == "." {
		return tfs, nil
	}

	e, err := tfs.get(op, dir)
	if err != nil {
		return nil, err
	}

	subfs := &tarfs{make(map[string]fs.DirEntry)}

	subfs.entries["."] = e

	prefix := dir + "/"
	for name, file := range tfs.entries {
		if strings.HasPrefix(name, prefix) {
			subfs.entries[strings.TrimPrefix(name, prefix)] = file
		}
	}

	return subfs, nil
}

func (tfs *tarfs) get(op, path string) (entry, error) {
	if !fs.ValidPath(path) {
		return nil, newErr(op, path, fs.ErrInvalid)
	}

	e, ok := tfs.entries[path]
	if !ok {
		return nil, newErrNotExist(op, path)
	}

	return e.(entry), nil
}
