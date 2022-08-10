package tarfs_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/nlepage/go-tarfs"
)

// ExampleHTTPFileServer demonstrates how to serve the contents of a tar file using HTTP
func Example_httpFileServer() {
	tf, err := os.Open("test.tar")
	if err != nil {
		panic(err)
	}
	defer tf.Close()

	tfs, err := tarfs.NewFS(tf)
	if err != nil {
		panic(err)
	}

	srv := httptest.NewServer(http.FileServer(http.FS(tfs)))
	defer srv.Close()

	res, err := srv.Client().Get(srv.URL + "/dir1/dir11/file111")
	if err != nil {
		panic(err)
	}

	if _, err := io.Copy(os.Stdout, res.Body); err != nil {
		panic(err)
	}
	res.Body.Close()

	// Output:
	// file111
}
