package boltfs

import (
	"bytes"

	. "gopkg.in/check.v1"
)

func (s *FSSuite) TestFile_Write(c *C) {
	f := &File{buf: bytes.NewBuffer(nil)}
	n, err := f.Write([]byte{'F', 'O', 'O'})

	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)
}

func (s *FSSuite) TestFile_Close(c *C) {
	v, _ := NewVolume(TestDBFile)
	defer v.Close()

	f, err := v.Open("foo")

	err = f.Close()
	c.Assert(err, IsNil)
}
