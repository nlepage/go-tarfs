package tarfs_test

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/nlepage/go-tarfs"
)

// Example_stat demonstrates how to read a file info from within a tar
func Example_stat() {
	tf, err := os.Open("test.tar")
	if err != nil {
		panic(err)
	}
	defer tf.Close()

	tfs, err := tarfs.New(tf, tarfs.DisableSeek(true))
	if err != nil {
		panic(err)
	}

	fi, err := fs.Stat(tfs, "dir1/dir11/file111")
	if err != nil {
		panic(err)
	}

	fmt.Println(fi.Name())
	fmt.Println(fi.IsDir())
	fmt.Println(fi.Size())

	// Output:
	// file111
	// false
	// 7
}
