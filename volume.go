package raa

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
)

type Volume struct {
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

//NewVolume create or open a Volume
func NewVolume(dbFile string) (*Volume, error) {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &Volume{path: "/", db: db}, nil
}

// Chdir changes the current working directory to the named directory.
func (v *Volume) Chdir(dir string) error {
	dir = filepath.Clean(dir)

	if !filepath.IsAbs(dir) {
		v.path = filepath.Join(v.path, dir)
		return nil
	}

	v.path = dir
	return nil
}

// Chmod changes the mode of the file to mode.
// If there is an error, it will be of type *PathError.
func (v *Volume) Chmod(name string, mode os.FileMode) error {
	f, err := v.Open(name)
	if err != nil {
		return err
	}

	f.Chmod(mode)
	return f.Close()
}

// Chown changes the numeric uid and gid of the named file.
// If there is an error, it will be of type *PathError.
func (v *Volume) Chown(name string, uid, gid int) error {
	f, err := v.Open(name)
	if err != nil {
		return err
	}

	f.Chown(uid, gid)
	return f.Close()
}

//func Chtimes(name string, atime time.Time, mtime time.Time) error

// Getwd returns a rooted path name corresponding to the
// current directory.
func (v *Volume) Getwd() (dir string, err error) {
	return v.path, nil
}

//func IsPermission(err error) bool
//func Lchown(name string, uid, gid int) error
//func Link(oldname, newname string) error
//func Mkdir(name string, perm FileMode) error
//func MkdirAll(path string, perm FileMode) error
//func Readlink(name string) (string, error)

// Remove removes the named file or directory.
// If there is an error, it will be of type *PathError.
func (v *Volume) Remove(name string) error {
	fname := v.getFullpath(name)
	key := []byte(fname)
	return v.iterateKeys(func(b *bolt.Bucket, k, v []byte) error {
		if bytes.Equal(k, key) {
			if err := b.Delete(k); err != nil {
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
func (v *Volume) RemoveAll(path string) error {
	fname := v.getFullpath(path)
	key := []byte(fname)
	keySlash := []byte(filepath.Join(path, "/") + "/")
	return v.iterateKeys(func(b *bolt.Bucket, k, v []byte) error {
		if bytes.Equal(k, key) || bytes.HasPrefix(k, keySlash) {
			if err := b.Delete(k); err != nil {
				return err
			}

			return stopError
		}

		return nil
	})
}

func (v *Volume) iterateKeys(cb func(b *bolt.Bucket, k, v []byte) error) error {
	err := v.db.Update(func(tx *bolt.Tx) error {
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
func (v *Volume) Rename(oldpath, newpath string) error {
	oldpath = v.getFullpath(oldpath)
	newpath = v.getFullpath(newpath)

	src, err := v.Open(oldpath)
	if err != nil {
		return err
	}

	dst, err := v.OpenFile(newpath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, src.inode.Mode)
	if err != nil {
		return err
	}

	defer dst.Close()
	defer v.Remove(oldpath)

	_, err = io.Copy(dst, src)
	return err
}

//func SameFile(fi1, fi2 FileInfo) bool
//func Symlink(oldname, newname string) error

// Truncate changes the size of the named file.
// If there is an error, it will be of type *PathError.
func (v *Volume) Truncate(name string, size int64) error {
	f, err := v.Open(name)
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
func (v *Volume) OpenFile(name string, flag int, perm os.FileMode) (file *File, err error) {
	//TODO: Implement O_APPEND
	fname := v.getFullpath(name)

	f := newFile(v, fname, flag, perm)
	if flag&os.O_TRUNC != 0 {
		//We dont read the file if should be truncated
		return f, nil
	}

	if err := v.readFile(f, []byte(fname)); err != nil {
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

func (v *Volume) readFile(f *File, name []byte) error {
	return v.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(rootBucket)
		if b == nil {
			return notFoundError
		}

		buf := bytes.NewBuffer(b.Get(name))
		if err := v.readFileInode(b, f, buf); err != nil {
			return err
		}

		if err := v.readFileContent(b, f, buf); err != nil {
			return err
		}

		return foundError
	})
}

func (v *Volume) readFileInode(b *bolt.Bucket, f *File, c *bytes.Buffer) error {
	var length int64
	if err := binary.Read(c, binary.LittleEndian, &length); err != nil {
		if err == io.EOF {
			return notFoundError
		}

		return err
	}

	header := make([]byte, length)
	n, err := c.Read(header)
	if err != nil && err != io.EOF {
		return err
	}

	if int64(n) != length {
		return unableToReadHeader
	}

	dec := gob.NewDecoder(bytes.NewBuffer(header))
	return dec.Decode(&f.inode)
}

func (v *Volume) readFileContent(b *bolt.Bucket, f *File, c *bytes.Buffer) error {
	if _, err := io.Copy(f.buf, c); err != nil {
		return err
	}

	return nil
}

// Open opens the named file for reading.  If successful, methods on
// the returned file can be used for reading; the associated file
// descriptor has mode O_RDONLY.
// If there is an error, it will be of type *PathError.
func (v *Volume) Open(name string) (file *File, err error) {
	return v.OpenFile(name, os.O_RDONLY, 0)
}

// Create creates the named file mode 0666 (before umask), truncating
// it if it already exists.  If successful, methods on the returned
// File can be used for I/O; the associated file descriptor has mode
// O_RDWR.
// If there is an error, it will be of type *PathError.
func (v *Volume) Create(name string) (file *File, err error) {
	return v.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

//func Pipe() (r *File, w *File, err error)
//func Lstat(name string) (fi FileInfo, err error)

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError.
func (v *Volume) Stat(name string) (os.FileInfo, error) {
	f, err := v.Open(v.getFullpath(name))
	if err != nil {
		return nil, err
	}

	return f.Stat()
}

// Find return the names of the files matching with the function matcher
func (v *Volume) Find(matcher func(string) bool) []string {
	r := make([]string, 0)
	v.db.View(func(tx *bolt.Tx) error {
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
func (v *Volume) Close() error {
	return v.db.Close()
}

func (v *Volume) writeFile(f *File) error {
	return v.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(rootBucket)
		if err != nil {
			return err
		}

		name := []byte(f.inode.Name)
		buf := bytes.NewBuffer(nil)
		if v.writeFileInode(b, f, buf); err != nil {
			return err
		}

		if v.writeFileContent(b, f, buf); err != nil {
			return err
		}

		if err = b.Put(name, buf.Bytes()); err != nil {
			return err
		}

		return nil
	})
}

func (v *Volume) writeFileInode(b *bolt.Bucket, f *File, c *bytes.Buffer) error {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(f.inode); err != nil {
		return err
	}

	if err := binary.Write(c, binary.LittleEndian, int64(buf.Len())); err != nil {
		return err
	}

	if _, err := buf.WriteTo(c); err != nil {
		return err
	}

	return nil
}

func (v *Volume) writeFileContent(b *bolt.Bucket, f *File, c *bytes.Buffer) error {
	if _, err := c.Write(f.buf.Bytes()); err != nil {
		return err
	}

	return nil
}

func (v *Volume) getFullpath(name string) string {
	return filepath.Join(v.path + name)
}
