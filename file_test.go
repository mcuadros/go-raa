package raa

import (
	"bytes"
	"os"

	. "gopkg.in/check.v1"
)

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
	f, _ := s.a.Create("foo")
	err := f.Chdir()
	c.Assert(err, FitsTypeOf, &os.PathError{})
}

func (s *FSSuite) TestFile_Stat(c *C) {
	f, err := s.a.Create("foo")
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
	f, err := s.a.Create("foo")
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
	f, _ := s.a.OpenFile("foo", os.O_WRONLY|os.O_SYNC|os.O_CREATE, 0)
	f.WriteString("foo")

	r, _ := s.a.OpenFile("foo", os.O_RDONLY, 0)
	c.Assert(r.buf.String(), Equals, "foo")

	n, err := f.WriteString("bar")
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)

	r, _ = s.a.OpenFile("foo", os.O_RDONLY, 0)
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
	length := 26334208
	fsFile, err := s.a.Create("foo")
	c.Assert(err, IsNil)

	n, err := fsFile.Write(bytes.Repeat([]byte("f"), length))
	c.Assert(err, IsNil)
	c.Assert(n, Equals, length)

	fsFile.Close()

	fsFile, err = s.a.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(fsFile.inode.Size, Equals, int64(length))
	c.Assert(fsFile.buf.Len(), Equals, length)
}

func (s *FSSuite) TestFile_Read(c *C) {
	f, _ := s.a.Create("foo")
	f.WriteString("foo")
	defer f.Close()

	content := make([]byte, 3)
	n, err := f.Read(content)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)
	c.Assert(string(content), Equals, "foo")
}

func (s *FSSuite) TestFile_ReadInClosed(c *C) {
	f, _ := s.a.Create("foo")
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
	f, err := s.a.Create("foo")
	c.Assert(err, IsNil)

	err = f.Close()
	c.Assert(err, IsNil)
}
