package boltfs

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"path/filepath"

	"github.com/boltdb/bolt"
)

type Volume struct {
	db *bolt.DB
}

var rootBucket = []byte("root")
var stopError = errors.New("stop")

func NewVolume(dbFile string) (*Volume, error) {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &Volume{db}, nil
}

//func Chdir(dir string) error
//func Chmod(name string, mode FileMode) error
//func Chown(name string, uid, gid int) error
//func Chtimes(name string, atime time.Time, mtime time.Time) error
//func Getwd() (dir string, err error)
//func IsPermission(err error) bool
//func Lchown(name string, uid, gid int) error
//func Link(oldname, newname string) error
//func Mkdir(name string, perm FileMode) error
//func MkdirAll(path string, perm FileMode) error
//func Readlink(name string) (string, error)

func (v *Volume) Remove(name string) error {
	key := []byte(name)
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

func (v *Volume) RemoveAll(path string) error {
	key := []byte(path)
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

//func Rename(oldpath, newpath string) error
//func SameFile(fi1, fi2 FileInfo) bool
//func Symlink(oldname, newname string) error
//func Truncate(name string, size int64) error
//func Create(name string) (file *File, err error)

func (v *Volume) Open(name string) (*File, error) {
	file, err := v.readFile(name)
	if err != nil || file != nil {
		return file, err
	}

	return newFile(name, v), nil
}

func (v *Volume) readFile(name string) (*File, error) {
	f := newFile(name, v)
	err := v.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(rootBucket)
		if b == nil {
			return nil
		}

		buf := bytes.NewBuffer(nil)
		buf.Write(b.Get([]byte(name)))

		r := tar.NewReader(buf)
		hdr, err := r.Next()
		if hdr == nil {
			return nil
		}

		if err != nil {
			return err
		}

		f.hdr = *hdr
		if _, err := io.Copy(f.buf, r); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return f, nil
}

//func OpenFile(name string, flag int, perm FileMode) (file *File, err error)
//func Pipe() (r *File, w *File, err error)
//func Lstat(name string) (fi FileInfo, err error)
//func Stat(name string) (fi FileInfo, err error)

func (v *Volume) Close() error {
	return v.db.Close()
}

func (v *Volume) writeFile(f *File) error {
	return v.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(rootBucket)
		if err != nil {
			return err
		}

		content, err := v.getHeaderBytes(f)
		if err != nil {
			return err
		}

		content = append(content, f.buf.Bytes()...)

		if err = b.Put([]byte(f.Name()), content); err != nil {
			return err
		}

		return nil
	})
}

func (v *Volume) getHeaderBytes(f *File) ([]byte, error) {
	b := bytes.NewBuffer(nil)
	w := tar.NewWriter(b)
	if err := w.WriteHeader(&f.hdr); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
