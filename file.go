package boltfs

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
)

type File struct {
	hdr      tar.Header
	buf      *bytes.Buffer
	v        *Volume
	IsClosed bool
}

func newFile(name string, v *Volume) *File {
	return &File{
		hdr: tar.Header{Name: name},
		buf: bytes.NewBuffer(nil),
		v:   v,
	}
}

//func (f *File) Chdir() error
//func (f *File) Chmod(mode FileMode) error
//func (f *File) Chown(uid, gid int) error

func (f *File) Close() error {
	return f.Sync()
}

//func (f *File) Fd() uintptr

// Name returns the name of the file as presented to Open.
func (f *File) Name() string {
	return f.hdr.Name
}

// Read reads up to len(b) bytes from the File.
func (f *File) Read(b []byte) (int, error) {
	n, err := f.buf.Read(b)
	if err != nil {
		err = &os.PathError{"read", f.hdr.Name, err}
	}

	return n, err
}

//func (f *File) ReadAt(b []byte, off int64) (n int, err error)
//unc (f *File) Readdir(n int) (fi []FileInfo, err error)
//func (f *File) Readdirnames(n int) (names []string, err error)
//func (f *File) Seek(offset int64, whence int) (ret int64, err error)

// Stat returns a FileInfo describing the named file.
func (f *File) Stat() (os.FileInfo, error) {
	return f.hdr.FileInfo(), nil
}

// Sync commits the current contents of the file to stable storage.
func (f *File) Sync() error {
	return f.v.writeFile(f)
}

// Truncate changes the size of the file.
func (f *File) Truncate(size int64) error {
	f.buf.Truncate(int(size))

	return nil
}

// Write writes len(b) bytes to the File.
// It returns the number of bytes written and an error, if any.
// Write returns a non-nil error when n != len(b).
func (f *File) Write(b []byte) (int, error) {
	n, err := f.buf.Write(b)
	f.hdr.Size += int64(n)

	if err != nil {
		err = &os.PathError{"write", f.hdr.Name, err}
	}

	if n != len(b) {
		return n, io.ErrShortWrite
	}

	return n, err
}

//func (f *File) WriteAt(b []byte, off int64) (n int, err error) {}

// WriteString is like Write, but writes the contents of string s rather than
// a slice of bytes.
func (f *File) WriteString(s string) (int, error) {
	return f.Write([]byte(s))
}
