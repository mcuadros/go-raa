package raa

import (
	"bytes"
	"io"
	"os"

	. "gopkg.in/check.v1"
)

const bigFileTar = "benchmark/fixtures/6133_files.tar"
const smallFileTar = "benchmark/fixtures/78_files.tar"

func (s *FSSuite) TestNewFile(c *C) {
	f := newFile(nil, "foo", os.O_WRONLY, 0042)
	c.Assert(f.name, Equals, "foo")
	c.Assert(f.flag, Equals, os.O_WRONLY)
	c.Assert(f.inode.BlockSize, Equals, DefaultBlockSize)
	c.Assert(int(f.inode.Mode), Equals, 0042)
	c.Assert(f.inode.UserId, Equals, uint64(os.Getuid()))
	c.Assert(f.inode.GroupId, Equals, uint64(os.Getgid()))
	c.Assert(f.inode.CreatedAt.Unix(), Not(Equals), 0)
	c.Assert(f.inode.CreatedAt.Unix(), Equals, f.inode.ModifcatedAt.Unix())
	c.Assert(f.buf.Len(), Equals, 0)

	c.Assert(f.isReadable, Equals, false)
	c.Assert(f.isWritable, Equals, true)
	c.Assert(f.isSync, Equals, false)

	f = newFile(nil, "foo", os.O_RDWR, 0042)
	c.Assert(f.isReadable, Equals, true)
	c.Assert(f.isWritable, Equals, true)
	c.Assert(f.isSync, Equals, false)

	f = newFile(nil, "foo", os.O_RDONLY, 0042)
	c.Assert(f.isReadable, Equals, true)
	c.Assert(f.isWritable, Equals, false)
	c.Assert(f.isSync, Equals, false)

	f = newFile(nil, "foo", os.O_SYNC, 0042)
	c.Assert(f.isReadable, Equals, true)
	c.Assert(f.isWritable, Equals, false)
	c.Assert(f.isSync, Equals, true)
}

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
	c.Assert(f.inode.Size, Equals, int64(3))
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

	n, err := f.WriteString("bar")
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)

	r, _ = s.v.OpenFile("foo", os.O_RDONLY, 0)
	c.Assert(r.buf.String(), Equals, "foobar")
}

func (s *FSSuite) TestFile_WriteString(c *C) {
	f := newFile(nil, "", os.O_WRONLY, 0)
	n, err := f.WriteString("foo")

	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)
	c.Assert(f.inode.Size, Equals, int64(3))
}

func (s *FSSuite) TestFile_WriteLongFile(c *C) {
	osFile, err := os.Open(bigFileTar)
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
	c.Assert(fsFile.inode.Size, Equals, int64(26334208))
	c.Assert(fsFile.buf.Len(), Equals, 26334208)
}

func (s *FSSuite) TestFile_Read(c *C) {
	f, _ := s.v.Create("foo")
	f.WriteString("foo")
	defer f.Close()

	content := make([]byte, 3)
	n, err := f.Read(content)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)
	c.Assert(string(content), Equals, "foo")
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
	v, err := NewVolume(TestRAAFile)
	if err != nil {
		panic(err)
	}

	defer v.Close()

	f, err := v.Create("foo")
	c.Assert(err, IsNil)

	err = f.Close()
	c.Assert(err, IsNil)
}
