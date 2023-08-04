package tarfs

import (
	"io"
	"io/fs"
	"path"
	"sort"
	"time"

	"github.com/nlepage/go-tarfs/tar"
)

type hackfs struct {
	files   map[string]hackFile
	folders map[string]*folder
}

func NewFS(r io.Reader) (fs.FS, error) {
	tfs := hackfs{
		files:   make(map[string]hackFile),
		folders: make(map[string]*folder),
	}

	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		name := path.Clean(h.Name)
		if name == "." {
			continue
		}
		if !fs.ValidPath(name) {
			continue
		}

		fi := h.FileInfo()
		if fi.IsDir() {
			if f := tfs.folders[name]; f != nil {
				f.FileInfo = fi
			} else {
				tfs.folders[name] = &folder{
					FileInfo: fi,
				}
				dir := path.Dir(name)
				tfs.addFoldersAndParents(dir, dirEntryFromFileInfo{fi})
			}
			continue
		}

		sr, err := tr.SectionReader()
		if err != nil {
			return nil, err
		}

		tfs.files[name] = hackFile{
			SectionReader: sr,
			FileInfo:      fi,
		}

		dir := path.Dir(name)
		tfs.addFoldersAndParents(dir, dirEntryFromFileInfo{fi})
	}

	// sort DirEntries
	for _, folder := range tfs.folders {
		sort.Slice(folder.DirEntries, func(i, j int) bool {
			return folder.DirEntries[i].Name() < folder.DirEntries[j].Name()
		})
	}

	return tfs, nil
}

type dirEntryFromFileInfo struct {
	fi fs.FileInfo
}

var _ fs.DirEntry = dirEntryFromFileInfo{}

func (de dirEntryFromFileInfo) Name() string {
	return de.fi.Name()
}
func (de dirEntryFromFileInfo) IsDir() bool {
	return de.fi.IsDir()
}
func (de dirEntryFromFileInfo) Type() fs.FileMode {
	return de.fi.Mode() & fs.ModeType
}
func (de dirEntryFromFileInfo) Info() (fs.FileInfo, error) {
	return de.fi, nil
}

type dirEntryFromFolder struct {
	name string
}

var _ fs.DirEntry = dirEntryFromFolder{}

func (de dirEntryFromFolder) Name() string {
	return de.name
}
func (de dirEntryFromFolder) IsDir() bool {
	return true
}
func (de dirEntryFromFolder) Type() fs.FileMode {
	return fs.ModeDir
}
func (de dirEntryFromFolder) Info() (fs.FileInfo, error) {
	return de, nil
}
func (dirEntryFromFolder) Mode() fs.FileMode {
	return fs.ModeDir
}
func (dirEntryFromFolder) Size() int64 {
	return 0
}
func (dirEntryFromFolder) ModTime() time.Time {
	return time.Time{}
}
func (dirEntryFromFolder) Sys() interface{} {
	return nil
}

func (tfs *hackfs) addFoldersAndParents(dir string, de fs.DirEntry) {
	f := tfs.folders[dir]
	if f != nil {
		f.DirEntries = append(f.DirEntries, de)
		return
	}

	parent, name := path.Split(dir)

	dirEntry := dirEntryFromFolder{name: name}
	tfs.folders[dir] = &folder{
		FileInfo:   dirEntry,
		DirEntries: []fs.DirEntry{de},
	}

	if dir == "." {
		return
	}
	tfs.addFoldersAndParents(parent, dirEntry)
}

func (tfs hackfs) Open(name string) (fs.File, error) {
	if f, ok := tfs.files[name]; ok {
		return hackFile{
			// make a new one, so that new Read ar at the start
			SectionReader: io.NewSectionReader(f, 0, f.SectionReader.Size()),
			FileInfo:      f.FileInfo,
		}, nil
	}
	folder := tfs.folders[name]
	if folder == nil {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	return &openedFolder{
		folder: folder,
	}, nil
}

type hackFile struct {
	*io.SectionReader
	FileInfo fs.FileInfo
}

var _ fs.File = &hackFile{}

func (f hackFile) Stat() (fs.FileInfo, error) {
	return f.FileInfo, nil
}

func (hackFile) Close() error {
	return nil
}

type folder struct {
	FileInfo   fs.FileInfo
	DirEntries []fs.DirEntry
}

type openedFolder struct {
	folder        *folder
	dirEntriesPos int
}

var _ fs.ReadDirFile = &openedFolder{}

func (f openedFolder) Stat() (fs.FileInfo, error) {
	return f.folder.FileInfo, nil
}
func (f openedFolder) Read(p []byte) (n int, err error) {
	return 0, &fs.PathError{Op: "read", Path: f.folder.FileInfo.Name(), Err: ErrDir}
}
func (openedFolder) Close() error {
	return nil
}
func (f *openedFolder) ReadDir(n int) ([]fs.DirEntry, error) {
	if f.dirEntriesPos >= len(f.folder.DirEntries) {
		if n <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}

	start := f.dirEntriesPos
	if n > 0 && f.dirEntriesPos+n <= len(f.folder.DirEntries) {
		f.dirEntriesPos += n
		return f.folder.DirEntries[start:f.dirEntriesPos], nil
	}

	f.dirEntriesPos = len(f.folder.DirEntries)
	return f.folder.DirEntries[start:], nil
}
