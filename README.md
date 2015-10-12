# go-getter

go-getter is a library for Go (golang) for downloading things from various
different sources using a URL as the primary form of input of where to
download this thing.

The power of this library is being flexible in being able to download
from a number of different sources (file paths, Git, HTTP, Mercurial, etc.)
using a single string as input. This removes the burden of knowing how to
download from a variety of sources from the implementer.

This library is used by [Terraform](https://terraform.io) for
downloading modules, [Otto](https://ottoproject.io) for dependencies and
Appfile imports, and [Nomad](https://nomadproject.io) for downloading
binaries.

## Installation and Usage

Package documentation can be found on
[GoDoc](http://godoc.org/github.com/hashicorp/go-getter).

Installation can be done with a normal `go get`:

```
$ go get github.com/hashicorp/go-getter
```
