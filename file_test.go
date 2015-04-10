package boltfs

import (
	"io"
	"os"

	. "gopkg.in/check.v1"
)

const longFile = "benchmark/fixtures/6133_files.tar"

func (s *FSSuite) TestFile_Chdir(c *C) {
	f, _ := s.v.Create("foo")
	err := f.Chdir()
	c.Assert(err, FitsTypeOf, &os.PathError{})
}

func (s *FSSuite) TestFile_Stat(c *C) {
	f, err := s.v.Create("foo")
	f.WriteString("foo")
	c.Assert(err, IsNil)

	fi, err := f.Stat()
	c.Assert(err, IsNil)
	c.Assert(fi.Name(), Equals, "foo")
}

func (s *FSSuite) TestFile_Write(c *C) {
	f := newFile(nil, "", os.O_WRONLY, 0)
	n, err := f.Write([]byte{'F', 'O', 'O'})

	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)
	c.Assert(f.hdr.Size, Equals, int64(3))
}

func (s *FSSuite) TestFile_WriteInClosed(c *C) {
	f, err := s.v.Create("foo")
	f.Close()

	_, err = f.Write([]byte{'F', 'O', 'O'})
	c.Assert(err, FitsTypeOf, &os.PathError{})
}

func (s *FSSuite) TestFile_WriteInNonWritale(c *C) {
	f := newFile(nil, "", os.O_RDONLY, 0)

	_, err := f.Write([]byte{'F', 'O', 'O'})
	c.Assert(err, FitsTypeOf, &os.PathError{})
}

func (s *FSSuite) TestFile_WriteInSynced(c *C) {
	f, _ := s.v.OpenFile("foo", os.O_WRONLY|os.O_SYNC|os.O_CREATE, 0)
	f.WriteString("foo")

	r, _ := s.v.OpenFile("foo", os.O_RDONLY, 0)
	c.Assert(r.buf.String(), Equals, "foo")

	f.WriteString("bar")
	r, _ = s.v.OpenFile("foo", os.O_RDONLY, 0)
	c.Assert(r.buf.String(), Equals, "foobar")
}

func (s *FSSuite) TestFile_WriteString(c *C) {
	f := newFile(nil, "", os.O_WRONLY, 0)
	n, err := f.WriteString("foo")

	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)
	c.Assert(f.hdr.Size, Equals, int64(3))
}

func (s *FSSuite) TestFile_WriteLongFile(c *C) {
	osFile, err := os.Open(longFile)
	if err != nil {
		panic(err)
	}

	defer osFile.Close()

	fsFile, err := s.v.Create("foo")
	c.Assert(err, IsNil)

	n, err := io.Copy(fsFile, osFile)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, int64(26334208))

	fsFile.Close()

	fsFile, err = s.v.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(fsFile.hdr.Size, Equals, int64(26334208))
	c.Assert(fsFile.buf.Len(), Equals, 26334208)
}

func (s *FSSuite) TestFile_ReadInClosed(c *C) {
	f, _ := s.v.Create("foo")
	f.Close()

	_, err := f.Read(nil)
	c.Assert(err, FitsTypeOf, &os.PathError{})
}

func (s *FSSuite) TestFile_ReadInNonReadable(c *C) {
	f := newFile(nil, "", os.O_WRONLY, 0)

	_, err := f.Read(nil)
	c.Assert(err, FitsTypeOf, &os.PathError{})
}

func (s *FSSuite) TestFile_Close(c *C) {
	v, _ := NewVolume(TestDBFile)
	defer v.Close()

	f, err := v.Open("foo")

	err = f.Close()
	c.Assert(err, IsNil)
}
