boltfs [![Build Status](https://travis-ci.org/mcuadros/boltfs.png?branch=master)](https://travis-ci.org/mcuadros/boltfs) [![GoDoc](http://godoc.org/github.com/mcuadros/boltfs?status.png)](http://godoc.org/github.com/mcuadros/boltfs) [![GitHub release](https://img.shields.io/github/release/mcuadros/boltfs.svg)](https://github.com/mcuadros/boltfs/releases)
==============================

boltsfs is a file container, similar to tar or zip, focused on allowing constant-time random file access with linear memory consumption increase.

The library implements a very similar API to the go [os package](http://golang.org/pkg/os/#File), allowing full control over,and low level acces to the contained files. boltfs is based on [boltdb](https://github.com/boltdb/bolt), a low-level key/value database for Go.



Installation
------------

The recommended way to install boltfs

```
go get github.com/mcuadros/boltfs
```

Example
-------

Import the package:

```go
import  "github.com/mcuadros/boltfs"
```

Create a new archive file respredented by a `Volume`:

```go
v, err = boltfs.NewVolume("example.archive.bfs")
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
var content []byte
f.Read(content)
fmt.Println(string(content))
//Output: Hello World!
```


License
-------

MIT, see [LICENSE](LICENSE)
