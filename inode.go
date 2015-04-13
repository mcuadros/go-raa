package raa

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path"
	"time"
)

var (
	InodeSignature      = []byte{'R', 'A', 'A'}
	WrongInodeSignature = errors.New("Wrong Inode signature")
)

const (
	InodeVersion int32 = 1
	InodeLength  int32 = 60
)

type Inode struct {
	Id           uint64
	Mode         os.FileMode
	UserId       uint64
	GroupId      uint64
	Size         int64
	ModifcatedAt time.Time
	CreatedAt    time.Time
}

// Write writes the byte representation of Inode
//
// Inode byte representation on LittleEndian have the following format:
// - 4-byte signature: The signature is: {'R', 'A', 'A'}
// - 4-byte lenght of the header, not includes the signature len
// - 4-byte version number
// - 8-byte inode id
// - 4-byte file mode
// - 8-byte user id
// - 8-byte group id
// - 8-byte file size
// - 8-byte modification timestamp
// - 8-byte creation timestamp
func (i *Inode) Write(w io.Writer) error {
	if _, err := w.Write(InodeSignature); err != nil {
		return err
	}

	var data = []interface{}{
		InodeLength,
		InodeVersion,
		i.Id,
		i.Mode,
		i.UserId,
		i.GroupId,
		i.Size,
		i.ModifcatedAt.Unix(),
		i.CreatedAt.Unix(),
	}

	for _, v := range data {
		if err := binary.Write(w, binary.LittleEndian, v); err != nil {
			return err
		}
	}

	return nil
}

// Read reads from a reader the byte representation of Inode and fills up the Inode
func (i *Inode) Read(r io.Reader) error {
	sig := make([]byte, 3)
	if _, err := r.Read(sig); err != nil {
		return err
	}

	if !bytes.Equal(sig, InodeSignature) {
		return WrongInodeSignature
	}

	var length int32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return err
	}

	var version int32
	if err := binary.Read(r, binary.LittleEndian, &version); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &i.Id); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &i.Mode); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &i.UserId); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &i.GroupId); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &i.Size); err != nil {
		return err
	}

	var modTs int64
	if err := binary.Read(r, binary.LittleEndian, &modTs); err != nil {
		return err
	}

	i.ModifcatedAt = time.Unix(modTs, 0)

	var creTs int64
	if err := binary.Read(r, binary.LittleEndian, &creTs); err != nil {
		return err
	}

	i.CreatedAt = time.Unix(creTs, 0)

	if leftover := length - InodeLength; leftover != 0 {
		raw := make([]byte, leftover)
		if _, err := r.Read(raw); err != nil {
			return err
		}
	}

	return nil
}

type FileInfo struct {
	name  string
	inode Inode
}

// Name returns base name of the file
func (fi *FileInfo) Name() string {
	return path.Base(fi.name)
}

// Size returns the length in bytes
func (fi *FileInfo) Size() int64 {
	return fi.inode.Size
}

// Mode returns the file mode bits
func (fi *FileInfo) Mode() os.FileMode {
	return fi.inode.Mode
}

// ModeTime returns the modification time
func (fi *FileInfo) ModTime() time.Time {
	return fi.inode.ModifcatedAt
}

// IsDir is present just for match the interface
func (fi *FileInfo) IsDir() bool {
	return false
}

// Sys returns the Inode value
func (fi *FileInfo) Sys() interface{} {
	return fi.inode
}
