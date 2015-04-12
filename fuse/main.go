// Hellofs implements a simple "hello world" file system.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mcuadros/boltfs"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
	"bazil.org/fuse/fuseutil"
	"golang.org/x/net/context"
)

var volume *boltfs.Volume

var Usage = func() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	volume, _ = boltfs.NewVolume("78_files.db")

	flag.Usage = Usage
	flag.Parse()

	if flag.NArg() != 1 {
		Usage()
		os.Exit(2)
	}
	mountpoint := flag.Arg(0)

	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("helloworld"),
		fuse.Subtype("hellofs"),
		fuse.LocalVolume(),
		fuse.VolumeName("Hello world!"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	err = fs.Serve(c, FS{})
	if err != nil {
		log.Fatal(err)
	}

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}

// FS implements the hello world file system.
type FS struct{}

func (FS) Root() (fs.Node, error) {
	return Dir{}, nil
}

// Dir implements both Node and Handle for the root directory.
type Dir struct{}

func (Dir) Attr(a *fuse.Attr) {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
}

func (Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if file, err := volume.Open(name); err == nil {
		return &File{file}, nil
	}

	return nil, fuse.ENOENT
}

func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	files := volume.Find(func(string) bool { return true })

	dir := make([]fuse.Dirent, 0)
	for n, name := range files {
		_, file := filepath.Split(name)
		dir = append(dir, fuse.Dirent{Inode: uint64(n), Name: file, Type: fuse.DT_File})
	}

	return dir, nil
}

type File struct {
	*boltfs.File
}

func (f *File) Attr(a *fuse.Attr) {
	stats, _ := f.Stat()
	a.Inode = 2
	a.Mode = stats.Mode().Perm()
	a.Size = uint64(stats.Size())
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	return f, nil
}

func (f *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	b, err := ioutil.ReadAll(f.File)
	fuseutil.HandleRead(req, resp, b)

	return err
}

func (f *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	fmt.Println(string(req.Data))
	f.File.Write(req.Data)
	return nil
}

func (f *File) Flush(ctx context.Context, req *fuse.FlushRequest) error {
	return f.File.Sync()
}

func (f *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return f.File.Close()
}

func (f *File) Setattr(ctx context.Context, req *fuse.SetattrRequest, resp *fuse.SetattrResponse) error {
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	return ioutil.ReadAll(f.File)
}
