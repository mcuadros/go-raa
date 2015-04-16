package raa

import (
	"bytes"
	"encoding/hex"
	"time"

	. "gopkg.in/check.v1"
)

const inodeWithLeftover = "52414148000000010000002a0000000000000054000000220000007e00000000000000a800000000000000d20000000000000011e62f55000000002a000000000000006c6566746f766572"

func (s *FSSuite) TestInode_WriteRead(c *C) {
	buf := bytes.NewBuffer(nil)
	i := getInodeFixture()

	err := i.Write(buf)
	c.Assert(err, IsNil)
	c.Assert(buf.Len(), Equals, int(InodeLength)+len(InodeSignature))

	leftover := []byte{'F', 'O', 'O'}
	buf.Write(leftover)

	o := &Inode{}
	err = o.Read(buf)
	c.Assert(err, IsNil)

	c.Assert(i.Id, Equals, o.Id)
	c.Assert(i.BlockSize, Equals, o.BlockSize)
	c.Assert(i.Mode, Equals, o.Mode)
	c.Assert(i.UserId, Equals, o.UserId)
	c.Assert(i.GroupId, Equals, o.GroupId)
	c.Assert(i.Size, Equals, o.Size)
	c.Assert(i.ModifcatedAt.Unix(), Equals, o.ModifcatedAt.Unix())
	c.Assert(i.CreatedAt.Unix(), Equals, o.CreatedAt.Unix())

	c.Assert(buf.String(), Equals, "FOO")
}

func (s *FSSuite) TestInode_WriteReadWithLeftover(c *C) {
	raw, _ := hex.DecodeString(inodeWithLeftover)
	buf := bytes.NewBuffer(raw)

	i := getInodeFixture()
	o := &Inode{}
	err := o.Read(buf)
	c.Assert(err, IsNil)

	c.Assert(i.Id, Equals, o.Id)
	c.Assert(i.BlockSize, Equals, o.BlockSize)
	c.Assert(i.Mode, Equals, o.Mode)
	c.Assert(i.UserId, Equals, o.UserId)
	c.Assert(i.GroupId, Equals, o.GroupId)
	c.Assert(i.Size, Equals, o.Size)
	c.Assert(int64(1429202449), Equals, o.ModifcatedAt.Unix())
	c.Assert(i.CreatedAt.Unix(), Equals, o.CreatedAt.Unix())

	c.Assert(buf.String(), Equals, "")
}

func (s *FSSuite) TestFileInfo_Name(c *C) {
	f := &FileInfo{"/foo/bar", Inode{}}
	c.Assert(f.Name(), Equals, "bar")
}

func (s *FSSuite) TestFileInfo_ModeTime(c *C) {
	i := getInodeFixture()
	f := &FileInfo{"", *i}
	c.Assert(f.ModTime().Unix(), Equals, i.ModifcatedAt.Unix())
}

func (s *FSSuite) TestFileInfo_Sys(c *C) {
	i := getInodeFixture()
	f := &FileInfo{"", *i}
	c.Assert(f.Sys().(Inode).Id, Equals, i.Id)
}

func (s *FSSuite) TestFileInfo_IsDir(c *C) {
	f := &FileInfo{"", Inode{}}
	c.Assert(f.IsDir(), Equals, false)
}

func getInodeFixture() *Inode {
	return &Inode{
		Id:           42,
		BlockSize:    42 * 2,
		Mode:         0042,
		UserId:       42 * 3,
		GroupId:      42 * 4,
		Size:         42 * 5,
		ModifcatedAt: time.Now(),
		CreatedAt:    time.Unix(42, 42),
	}
}
