package raa

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"code.google.com/p/snappy-go/snappy"
	"github.com/mcuadros/bolt"
)

type Archive struct {
	path string
	db   *bolt.DB
}

var (
	rootBucket = []byte("root")

	stopError          = errors.New("stop")
	foundError         = errors.New("file already exist")
	notFoundError      = errors.New("no such file or directory")
	unableToReadHeader = errors.New("unable to read the file header")
)

// CreateArchive create an archive raa file
func CreateArchive(dbFile string) (*Archive, error) {
	if _, err := os.Stat(dbFile); err == nil {
		return nil, foundError
	}

	return newArchive(dbFile)
}

// OpenArchive open an archive raa file
func OpenArchive(dbFile string) (*Archive, error) {
	if _, err := os.Stat(dbFile); err != nil {
		return nil, notFoundError
	}

	return newArchive(dbFile)
}

func newArchive(dbFile string) (*Archive, error) {
	db, err := bolt.Open(dbFile, 0600, &bolt.Options{MinMmapSize: 2})
	if err != nil {
		return nil, err
	}

	return &Archive{path: "/", db: db}, nil
}

// Path returns the path to currently open volume file.
func (a *Archive) Path() string {
	return a.db.Path()
}

// Chdir changes the current working directory to the named directory.
func (a *Archive) Chdir(dir string) error {
	dir = filepath.Clean(dir)

	if !filepath.IsAbs(dir) {
		a.path = filepath.Join(a.path, dir)
		return nil
	}

	a.path = dir
	return nil
}

// Chmod changes the mode of the file to mode.
// If there is an error, it will be of type *PathError.
func (a *Archive) Chmod(name string, mode os.FileMode) error {
	f, err := a.Open(name)
	if err != nil {
		return err
	}

	f.Chmod(mode)
	return f.Close()
}

// Chown changes the numeric uid and gid of the named file.
// If there is an error, it will be of type *PathError.
func (a *Archive) Chown(name string, uid, gid int) error {
	f, err := a.Open(name)
	if err != nil {
		return err
	}

	f.Chown(uid, gid)
	return f.Close()
}

//func Chtimes(name string, atime time.Time, mtime time.Time) error

// Getwd returns a rooted path name corresponding to the
// current directory.
func (a *Archive) Getwd() (dir string, err error) {
	return a.path, nil
}

//func IsPermission(err error) bool
//func Lchown(name string, uid, gid int) error
//func Link(oldname, newname string) error
//func Mkdir(name string, perm FileMode) error
//func MkdirAll(path string, perm FileMode) error
//func Readlink(name string) (string, error)

// Remove removes the named file or directory.
// If there is an error, it will be of type *PathError.
func (a *Archive) Remove(name string) error {
	fname := a.getFullpath(name)
	key := []byte(fname)
	return a.iterateKeys(func(b *bolt.Bucket, k, v []byte) error {
		if bytes.Equal(k, key) {
			if err := b.DeleteBucket(k); err != nil {
				return err
			}

			return stopError
		}

		return nil
	})
}

// RemoveAll removes path and any children it contains.
// It removes everything it can but returns the first error
// it encounters.  If the path does not exist, RemoveAll
// returns nil (no error).
func (a *Archive) RemoveAll(path string) error {
	fname := a.getFullpath(path)
	key := []byte(fname)
	keySlash := []byte(filepath.Join(path, "/") + "/")
	return a.iterateKeys(func(b *bolt.Bucket, k, v []byte) error {
		if bytes.Equal(k, key) || bytes.HasPrefix(k, keySlash) {
			if err := b.DeleteBucket(k); err != nil {
				return err
			}

			return stopError
		}

		return nil
	})
}

func (a *Archive) iterateKeys(cb func(b *bolt.Bucket, k, v []byte) error) error {
	err := a.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(rootBucket)
		if err != nil {
			return err
		}

		b.ForEach(func(k, v []byte) error {
			return cb(b, k, v)
		})

		return nil
	})

	if err == stopError {
		return nil
	}

	return err
}

// Rename renames (moves) a file.
func (a *Archive) Rename(oldpath, newpath string) error {
	oldpath = a.getFullpath(oldpath)
	newpath = a.getFullpath(newpath)

	src, err := a.Open(oldpath)
	if err != nil {
		return err
	}

	dst, err := a.OpenFile(newpath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, src.inode.Mode)
	if err != nil {
		return err
	}

	defer dst.Close()
	defer a.Remove(oldpath)

	_, err = io.Copy(dst, src)
	return err
}

//func SameFile(fi1, fi2 FileInfo) bool
//func Symlink(oldname, newname string) error

// Truncate changes the size of the named file.
// If there is an error, it will be of type *PathError.
func (a *Archive) Truncate(name string, size int64) error {
	f, err := a.Open(name)
	if err != nil {
		return err
	}

	f.Truncate(size)
	return f.Close()
}

// OpenFile is the generalized open call; most users will use Open
// or Create instead.  It opens the named file with specified flag
// (O_RDONLY etc.) and perm, (0666 etc.) if applicable.  If successful,
// methods on the returned File can be used for I/O.
// If there is an error, it will be of type *PathError.
func (a *Archive) OpenFile(name string, flag int, perm os.FileMode) (file *File, err error) {
	//TODO: Implement O_APPEND
	fname := a.getFullpath(name)

	f := newFile(a, fname, flag, perm)
	if flag&os.O_TRUNC != 0 {
		//We dont read the file if should be truncated
		return f, nil
	}

	if err := a.readFile(f, []byte(fname)); err != nil {
		switch err {
		case notFoundError:
			if flag&os.O_CREATE != 0 {
				err = nil
			}
		case foundError:
			if flag&os.O_EXCL == 0 {
				err = nil
			}
		}

		if err != nil {
			return nil, &os.PathError{"open", name, err}
		}
	}

	return f, nil
}

func (a *Archive) readFile(f *File, name []byte) error {
	return a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(rootBucket)
		if b == nil {
			return notFoundError
		}

		blocks := b.Bucket(name)
		if blocks == nil {
			return notFoundError
		}

		blocks.ForEach(func(k, v []byte) error {
			if bytes.Equal(k, BlockInode) {
				buf := bytes.NewBuffer(v)
				if err := f.inode.Read(buf); err != nil {
					if err == io.EOF {
						return notFoundError
					}

					return err
				}

				return nil
			}

			dec, err := snappy.Decode(nil, v)
			if err != nil {
				return err
			}

			buf := bytes.NewBuffer(dec)
			if _, err := io.Copy(f.buf, buf); err != nil {
				return err
			}

			return nil
		})

		return foundError
	})
}

// Open opens the named file for reading.  If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func (a *Archive) Open(name string) (file *File, err error) {
	return a.OpenFile(name, os.O_RDONLY, 0)
}

// Create creates the named file mode 0666 (before umask), truncating
// it if it already exists.  If successful, methods on the returned
// File can be used for I/O; the associated file descriptor has mode
// O_RDWR.
// If there is an error, it will be of type *PathError.
func (a *Archive) Create(name string) (file *File, err error) {
	return a.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

//func Pipe() (r *File, w *File, err error)
//func Lstat(name string) (fi FileInfo, err error)

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError.
func (a *Archive) Stat(name string) (os.FileInfo, error) {
	fname := a.getFullpath(name)

	i := &Inode{}
	err := a.readInode(i, []byte(fname))
	if err != nil && err != foundError {
		return nil, &os.PathError{"stat", fname, err}
	}

	return &FileInfo{fname, *i}, nil
}

func (a *Archive) readInode(i *Inode, name []byte) error {
	return a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(rootBucket)
		if b == nil {
			return notFoundError
		}

		blocks := b.Bucket(name)
		if blocks == nil {
			return notFoundError
		}

		buf := bytes.NewBuffer(blocks.Get(BlockInode))
		if err := i.Read(buf); err != nil {
			if err == io.EOF {
				return notFoundError
			}

			return err
		}

		return foundError
	})
}

// Find return the names of the files matching with the function matcher
func (a *Archive) Find(matcher func(string) bool) []string {
	r := make([]string, 0)
	a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(rootBucket)
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			name := string(k)
			if matcher(name) {
				r = append(r, name)
			}
		}

		return nil
	})

	return r
}

// Close the Volumen and releases all database resources.
func (a *Archive) Close() error {
	return a.db.Close()
}

const BlockPattern = "block.%d"

var BlockInode = []byte("block.inode")

func (a *Archive) writeFile(f *File) error {
	return a.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(rootBucket)
		if err != nil {
			return err
		}

		blocks, err := b.CreateBucketIfNotExists([]byte(f.name))
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(nil)
		if err := f.inode.Write(buf); err != nil {
			return err
		}

		if err := blocks.Put(BlockInode, buf.Bytes()); err != nil {
			return err
		}

		if err = a.writeFileBlocks(blocks, f); err != nil {
			return err
		}

		return nil
	})
}

func (a *Archive) writeFileBlocks(b *bolt.Bucket, f *File) error {
	r := bytes.NewReader(f.buf.Bytes())
	current := 0
	next := true
	for next {
		buf := bytes.NewBuffer(nil)
		if _, err := io.CopyN(buf, r, int64(f.inode.BlockSize)); err != nil {
			if err == io.EOF {
				next = false
			} else {
				return err
			}
		}

		name := fmt.Sprintf(BlockPattern, current)
		enc, err := snappy.Encode(nil, buf.Bytes())
		if err != nil {
			return err
		}

		if err := b.Put([]byte(name), enc); err != nil {
			return err
		}

		current++
		buf.Reset()
	}

	return nil
}

func (a *Archive) getFullpath(name string) string {
	return filepath.Join(a.path + name)
}
