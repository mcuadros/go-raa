package boltfs

import (
	"bytes"
)

type File struct {
	name string
	buf  *bytes.Buffer
	v    *Volume
}

//func (f *File) Chdir() error
//func (f *File) Chmod(mode FileMode) error
//func (f *File) Chown(uid, gid int) error

func (f *File) Close() error {
	return f.v.writeFile(f)
}

//func (f *File) Fd() uintptr

func (f *File) Name() string {
	return f.name
}

func (f *File) Read(b []byte) (n int, err error) {
	return f.buf.Read(b)
}

//func (f *File) ReadAt(b []byte, off int64) (n int, err error)
//unc (f *File) Readdir(n int) (fi []FileInfo, err error)
//func (f *File) Readdirnames(n int) (names []string, err error)
//func (f *File) Seek(offset int64, whence int) (ret int64, err error)
//func (f *File) Stat() (fi FileInfo, err error)
//func (f *File) Sync() (err error)

func (f *File) Truncate(size int64) error {
	f.buf.Truncate(int(size))

	return nil
}

func (f *File) Write(b []byte) (n int, err error) {
	return f.buf.Write(b)
}

//func (f *File) WriteAt(b []byte, off int64) (n int, err error) {}

func (f *File) WriteString(s string) (ret int, err error) {
	return f.buf.WriteString(s)
}
