package boltfs

import (
	"bytes"
	"errors"
	"io"
	"os"
	"time"
)

var (
	NotDirectoryErr = errors.New("not a directory")
	ClosedFileErr   = errors.New("cannot read/write on a closed file")
	NonReadableErr  = errors.New("cannot read from a O_WRONLY file")
	NonWritableErr  = errors.New("cannot write from on a not O_WRONLY or O_RDWR file")
)

type File struct {
	inode Inode
	flag  int
	buf   *bytes.Buffer
	v     *Volume

	isClosed   bool
	isWritable bool
	isReadable bool
	isSync     bool
}

func newFile(v *Volume, name string, flag int, mode os.FileMode) *File {
	return &File{
		inode: Inode{
			Name:         name,
			Mode:         mode,
			ModifcatedAt: time.Now(),
			CreatedAt:    time.Now(),
		},
		flag: flag,
		buf:  bytes.NewBuffer(nil),
		v:    v,

		isReadable: isReadable(flag),
		isWritable: isWritable(flag),
		isSync:     isSync(flag),
	}
}

// Chdir changes the current working directory to the file,
// which must be a directory.
// If there is an error, it will be of type *PathError.
func (f *File) Chdir() error {
	if true {
		return &os.PathError{"chdir", f.inode.Name, NotDirectoryErr}
	}

	return f.v.Chdir(f.inode.Name)
}

// Chmod changes the mode of the file to mode.
func (f *File) Chmod(mode os.FileMode) error {
	f.inode.Mode = mode

	return nil
}

// Chown changes the numeric uid and gid of the named file.
func (f *File) Chown(uid, gid int) error {
	f.inode.UserId = uid
	f.inode.GroupId = gid

	return nil
}

// Close closes the File, rendering it unusable for I/O.
// It returns an error, if any.
func (f *File) Close() error {
	f.isClosed = true
	return f.Sync()
}

//func (f *File) Fd() uintptr

// Name returns the name of the file as presented to Open.
func (f *File) Name() string {
	return f.inode.Name
}

// Read reads up to len(b) bytes from the File.
func (f *File) Read(b []byte) (int, error) {
	if f.isClosed {
		return 0, &os.PathError{"read", f.inode.Name, ClosedFileErr}
	}

	if !f.isReadable {
		return 0, &os.PathError{"read", f.inode.Name, NonReadableErr}
	}

	n, err := f.buf.Read(b)
	if err == io.EOF || err == nil {
		return n, err
	}

	return n, &os.PathError{"read", f.inode.Name, err}
}

//func (f *File) ReadAt(b []byte, off int64) (n int, err error)
//func (f *File) Readdir(n int) (fi []FileInfo, err error)
//func (f *File) Readdirnames(n int) (names []string, err error)
//func (f *File) Seek(offset int64, whence int) (ret int64, err error)

// Stat returns a FileInfo describing the named file.
func (f *File) Stat() (os.FileInfo, error) {
	return &FileInfo{f.inode}, nil
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
	if f.isClosed {
		return 0, &os.PathError{"read", f.inode.Name, ClosedFileErr}
	}

	if !f.isWritable {
		return 0, &os.PathError{"read", f.inode.Name, NonWritableErr}
	}

	n, err := f.buf.Write(b)
	f.inode.Size += int64(n)

	if err != nil {
		err = &os.PathError{"write", f.inode.Name, err}
	}

	if n != len(b) {
		return n, io.ErrShortWrite
	}

	if f.isSync {
		if err := f.Sync(); err != nil {
			return n, err
		}
	}

	return n, err
}

//func (f *File) WriteAt(b []byte, off int64) (n int, err error) {}

// WriteString is like Write, but writes the contents of string s rather than
// a slice of bytes.
func (f *File) WriteString(s string) (int, error) {
	return f.Write([]byte(s))
}

func isWritable(flag int) bool {
	if flag&os.O_WRONLY != 0 || flag&os.O_RDWR != 0 {
		return true
	}

	return false
}

func isReadable(flag int) bool {
	return flag&os.O_WRONLY == 0
}

func isSync(flag int) bool {
	return flag&os.O_SYNC != 0
}
