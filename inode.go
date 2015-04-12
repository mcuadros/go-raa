package raa

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

// Name returns base name of the file
func (fi *FileInfo) Name() string {
	return path.Base(fi.Inode.Name)
}

// Size returns the length in bytes
func (fi *FileInfo) Size() int64 {
	return fi.Inode.Size
}

// Mode returns the file mode bits
func (fi *FileInfo) Mode() os.FileMode {
	return fi.Inode.Mode
}

// ModeTime returns the modification time
func (fi *FileInfo) ModTime() time.Time {
	return fi.Inode.ModifcatedAt
}

// IsDir is present just for match the interface
func (fi *FileInfo) IsDir() bool {
	return false
}

// Sys returns the Inode value
func (fi *FileInfo) Sys() interface{} {
	return fi.Inode
}
