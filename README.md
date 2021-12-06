# go-tarfs

[![Go Reference](https://pkg.go.dev/badge/github.com/nlepage/go-tarfs.svg)](https://pkg.go.dev/github.com/nlepage/go-tarfs)
![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/nlepage/go-tarfs?sort=semver)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/nlepage/go-tarfs/Go)
[![License Unlicense](https://img.shields.io/github/license/nlepage/go-tarfs)](https://github.com/nlepage/go-tarfs/blob/master/LICENSE)
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-3-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->

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

## Show your support

Give a ⭐️ if this project helped you!

## Author

👤 **Nicolas Lepage**

* Website: https://nicolas.lepage.dev/
* Twitter: [@njblepage](https://twitter.com/njblepage)
* Github: [@nlepage](https://github.com/nlepage)

## Contributors ✨

Thanks goes to these wonderful people ([emoji key](https://allcontributors.org/docs/en/emoji-key)):

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<table>
  <tr>
    <td align="center"><a href="https://blog.cugu.eu/"><img src="https://avatars.githubusercontent.com/u/653777?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Jonas Plum</b></sub></a><br /><a href="https://github.com/nlepage/go-tarfs/commits?author=cugu" title="Tests">⚠️</a> <a href="https://github.com/nlepage/go-tarfs/commits?author=cugu" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/ix64"><img src="https://avatars.githubusercontent.com/u/13902388?v=4?s=100" width="100px;" alt=""/><br /><sub><b>MengYX</b></sub></a><br /><a href="https://github.com/nlepage/go-tarfs/issues?q=author%3Aix64" title="Bug reports">🐛</a> <a href="https://github.com/nlepage/go-tarfs/commits?author=ix64" title="Code">💻</a></td>
    <td align="center"><a href="https://github.com/adyatlov"><img src="https://avatars.githubusercontent.com/u/1386270?v=4?s=100" width="100px;" alt=""/><br /><sub><b>Andrey Dyatlov</b></sub></a><br /><a href="https://github.com/nlepage/go-tarfs/issues?q=author%3Aadyatlov" title="Bug reports">🐛</a> <a href="https://github.com/nlepage/go-tarfs/commits?author=adyatlov" title="Code">💻</a> <a href="https://github.com/nlepage/go-tarfs/commits?author=adyatlov" title="Tests">⚠️</a></td>
  </tr>
</table>

<!-- markdownlint-restore -->
<!-- prettier-ignore-end -->

<!-- ALL-CONTRIBUTORS-LIST:END -->

This project follows the [all-contributors](https://github.com/all-contributors/all-contributors) specification. Contributions of any kind welcome!

## 📝 License

This project is [unlicensed](https://github.com/nlepage/go-tarfs/blob/master/LICENSE), it is free and unencumbered software released into the public domain.
