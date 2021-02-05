# go-tarfs

[![Go Reference](https://pkg.go.dev/badge/github.com/nlepage/go-tarfs.svg)](https://pkg.go.dev/github.com/nlepage/go-tarfs)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/nlepage/go-tarfs?sort=semver)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/nlepage/go-tarfs/Go)
[![License Unlicense](https://img.shields.io/github/license/nlepage/go-tarfs)](https://github.com/nlepage/go-tarfs/blob/master/LICENSE)

> Read a tar file contents using go1.16 io/fs abstraction

## Usage

⚠️ go-tarfs needs go>=1.16

Install:
```sh
go get github.com/nlepage/go-tarfs
```

Use:
```go
package main

import (
    "os"

    tarfs "github.com/nlepage/go-tarfs"
)

func main() {
    tf, err := os.Open("path/to/archive.tar")
	if err != nil {
		panic(err)
	}
	defer tf.Close()

	tfs, err := tarfs.New(tf)
	if err != nil {
		panic(err)
	}

	f, err := tfs.Open("path/to/some/file")
	if err != nil {
		panic(err)
	}
	// defer f.Close() isn't necessary, it is a noop
    
	// use f...
}
```

More information at [pkg.go.dev/github.com/nlepage/go-tarfs](https://pkg.go.dev/github.com/nlepage/go-tarfs#section-documentation)

## Caveats

For now, no effort is done to support symbolic links.

## Author

👤 **Nicolas Lepage**

* Website: https://nicolas.lepage.dev/
* Twitter: [@njblepage](https://twitter.com/njblepage)
* Github: [@nlepage](https://github.com/nlepage)

## 🤝 Contributing

Contributions, issues and feature requests are welcome!

Feel free to check [issues page](https://github.com/nlepage/go-tarfs/issues).

## Show your support

Give a ⭐️ if this project helped you!

## 📝 License

This project is [unlicensed](https://github.com/nlepage/go-tarfs/blob/master/LICENSE), it is free and unencumbered software released into the public domain.
