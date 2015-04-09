package boltfs

import (
	"bytes"
	"io"
	"os"

	. "gopkg.in/check.v1"
)

const longFile = "benchmark/fixtures/6133_files.tar"

func (s *FSSuite) TestFile_Write(c *C) {
	f := &File{buf: bytes.NewBuffer(nil)}
	n, err := f.Write([]byte{'F', 'O', 'O'})

	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)
}

func (s *FSSuite) TestFile_WriteLongFile(c *C) {
	osFile, err := os.Open(longFile)
	if err != nil {
		panic(err)
	}

	defer osFile.Close()

	fsFile, err := s.v.Open("foo")
	c.Assert(err, IsNil)

	n, err := io.Copy(fsFile, osFile)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, int64(26334208))
	fsFile.Close()

	fsFile, err = s.v.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(fsFile.buf.Len(), Equals, 26334208)
}

func (s *FSSuite) TestFile_Close(c *C) {
	v, _ := NewVolume(TestDBFile)
	defer v.Close()

	f, err := v.Open("foo")

	err = f.Close()
	c.Assert(err, IsNil)
}
