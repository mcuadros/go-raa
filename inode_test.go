package raa

import (
	"bytes"
	"time"

	. "gopkg.in/check.v1"
)

func (s *FSSuite) TestInode_WriteRead(c *C) {
	buf := bytes.NewBuffer(nil)
	i := &Inode{
		Id:           42,
		Mode:         0042,
		UserId:       42 * 2,
		GroupId:      42 * 3,
		Size:         42 * 4,
		ModifcatedAt: time.Now(),
		CreatedAt:    time.Unix(42, 42),
	}

	err := i.Write(buf)
	c.Assert(err, IsNil)
	c.Assert(buf.Len(), Equals, int(InodeLength)+len(InodeSignature))

	leftover := []byte{'F', 'O', 'O'}
	buf.Write(leftover)

	o := &Inode{}
	err = o.Read(buf)
	c.Assert(err, IsNil)

	c.Assert(i.Id, Equals, o.Id)
	c.Assert(i.Mode, Equals, o.Mode)
	c.Assert(i.UserId, Equals, o.UserId)
	c.Assert(i.GroupId, Equals, o.GroupId)
	c.Assert(i.Size, Equals, o.Size)
	c.Assert(i.ModifcatedAt.Unix(), Equals, o.ModifcatedAt.Unix())
	c.Assert(i.CreatedAt.Unix(), Equals, o.CreatedAt.Unix())

	c.Assert(buf.String(), Equals, "FOO")
}

func (s *FSSuite) TestFileInfo_Name(c *C) {
	f := &FileInfo{"/foo/bar", Inode{}}
	c.Assert(f.Name(), Equals, "bar")
}
