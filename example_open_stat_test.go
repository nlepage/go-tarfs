package tarfs_test

import (
	"fmt"
	"os"

	"github.com/nlepage/go-tarfs"
)

// ExampleOpenAndStat demonstrates how to open a file from a tar and read file info
func Example_openAndStat() {
	tf, err := os.Open("test.tar")
	if err != nil {
		panic(err)
	}
	defer tf.Close()

	tfs, err := tarfs.New(tf)
	if err != nil {
		panic(err)
	}

	f, err := tfs.Open("dir1/dir11/file111")
	if err != nil {
		panic(err)
	}

	fi, err := f.Stat()
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
