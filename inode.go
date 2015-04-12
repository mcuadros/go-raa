package boltfs

import (
	"os"
	"path"
	"time"
)

type Inode struct {
	Id           uint64
	Name         string
	Mode         os.FileMode
	UserId       int
	GroupId      int
	Size         int64
	ModifcatedAt time.Time
	CreatedAt    time.Time
}

type FileInfo struct {
	Inode
}

func (i *FileInfo) Name() string {
	return path.Base(i.Inode.Name)
}

func (i *FileInfo) Size() int64 {
	return i.Inode.Size
}

func (i *FileInfo) Mode() os.FileMode {
	return i.Inode.Mode
}

func (i *FileInfo) ModTime() time.Time {
	return i.Inode.ModifcatedAt
}

func (i *FileInfo) IsDir() bool {
	return false
}

func (i *FileInfo) Sys() interface{} {
	return nil
}
