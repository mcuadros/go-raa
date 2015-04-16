raa - Random Access Archive
===========================
[![Build Status](https://travis-ci.org/mcuadros/go-raa.png?branch=master)](https://travis-ci.org/mcuadros/go-raa) [![Coverage Status](https://coveralls.io/repos/mcuadros/go-raa/badge.svg?branch=master)](https://coveralls.io/r/mcuadros/go-raa?branch=master) [![GoDoc](http://godoc.org/github.com/mcuadros/go-raa?status.png)](http://godoc.org/github.com/mcuadros/go-raa) [![GitHub release](https://img.shields.io/github/release/mcuadros/go-raa.svg)](https://github.com/mcuadros/go-raa/releases)


raa is a file container, similar to tar or zip, focused on allowing constant-time random file access with linear memory consumption increase.

The library implements a very similar API to the go [os package](http://golang.org/pkg/os/#File), allowing full control over and low level acces to the contained files. raa is based on [boltdb](https://github.com/boltdb/bolt), a low-level key/value database for Go.

- [Library reference](http://godoc.org/github.com/mcuadros/go-raa)
- [Command-line interface](#cli)


Installation
------------

The recommended way to install raa

```
go get -u github.com/mcuadros/go-raa/...
```

Example
-------

Import the package:

```go
import "github.com/mcuadros/go-raa"
```

Create a new archive file respredented by a `Volume`:

```go
v, err = raa.NewVolume("example.raa")
if err != nil {
    panic(err)
}
```

Add a new file to your new `Volume`:

```go
f, _ := v.Create("/hello.txt")
defer f.Close()
f.WriteString("Hello World!")
```

And now you can read the file contained on the `Volume`:

```go
f, _ := v.Open("/hello.txt")
defer f.Close()
content, _ := ioutil.ReadAll(f)
fmt.Println(string(content))
//Output: Hello World!
```


<a name="cli"></a>Command-line interface
----------------------
raa cli interface, is a convinient command that helps you to creates and manipulates raa files.

Output from: `./raa --help`:

```
Usage:
  raa [OPTIONS] <command>

Help Options:
  -h, --help  Show this help message

Available commands:
  list    List the items contained on a file.
  pack    Create a new archive containing the specified items.
  stats   Display some stats about the file.
  unpack  Extract to disk from the archive.
```

License
-------

MIT, see [LICENSE](LICENSE)
