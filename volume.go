package boltfs

import (
	"bytes"

	"github.com/boltdb/bolt"
)

type Volume struct {
	db *bolt.DB
}

var rootBucket = []byte("root")

func NewVolume(dbFile string) (*Volume, error) {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &Volume{db}, nil
}

func (v *Volume) Open(name string) (*File, error) {
	buf := bytes.NewBuffer(nil)
	v.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(rootBucket)
		if bucket == nil {
			return nil
		}

		buf.Write(bucket.Get([]byte(name)))
		return nil
	})

	return &File{
		name: name,
		buf:  buf,
		v:    v,
	}, nil
}

func (v *Volume) Close() error {
	return v.db.Close()
}

func (v *Volume) writeFile(f *File) error {
	return v.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(rootBucket)
		if err != nil {
			return err
		}

		if err = bucket.Put([]byte(f.Name()), f.buf.Bytes()); err != nil {
			return err
		}

		return nil
	})
}
