package boltfs

import (
	. "gopkg.in/check.v1"
)

func (s *FSSuite) TestFileInfo_Name(c *C) {
	f := &FileInfo{Inode{Name: "/foo/bar"}}
	c.Assert(f.Name(), Equals, "bar")
}
