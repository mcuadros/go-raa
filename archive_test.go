package raa

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type FSSuite struct {
	a    *Archive
	file string
}

var _ = Suite(&FSSuite{})

const TestRAAFile = "foo.raa"

func (s *FSSuite) SetUpTest(c *C) {
	tempDir, err := ioutil.TempDir("/tmp", "raa")
	if err != nil {
		panic(err)
	}

	s.file = filepath.Join(tempDir, TestRAAFile)

	s.a, err = CreateArchive(s.file)
	if err != nil {
		panic(err)
	}
}

func (s *FSSuite) TestPah(c *C) {
	c.Assert(s.a.Path(), Equals, s.file)
}

func (s *FSSuite) TestArchive_ChdirAndGetcwd(c *C) {
	c.Assert(s.a.path, Equals, "/")

	err := s.a.Chdir("foo")
	c.Assert(err, IsNil)
	c.Assert(s.a.path, Equals, "/foo")

	err = s.a.Chdir("foo")
	c.Assert(err, IsNil)
	c.Assert(s.a.path, Equals, "/foo/foo")

	err = s.a.Chdir("..")
	c.Assert(err, IsNil)
	c.Assert(s.a.path, Equals, "/foo")

	err = s.a.Chdir("/bar")
	c.Assert(err, IsNil)

	path, err := s.a.Getwd()
	c.Assert(path, Equals, "/bar")
}

func (s *FSSuite) TestArchive_Chmod(c *C) {
	f, _ := s.a.Create("foo")
	f.WriteString("foo")
	f.Close()

	c.Assert(int(f.inode.Mode), Equals, 0666)

	err := s.a.Chmod("/foo", 0042)
	c.Assert(err, IsNil)

	f, _ = s.a.OpenFile("foo", os.O_RDONLY, 0)
	c.Assert(int(f.inode.Mode), Equals, 0042)
}

func (s *FSSuite) TestArchive_Chown(c *C) {
	f, _ := s.a.Create("foo")
	f.WriteString("foo")
	f.Close()

	c.Assert(f.inode.UserId, Equals, uint64(os.Getuid()))
	c.Assert(f.inode.GroupId, Equals, uint64(os.Getgid()))

	err := s.a.Chown("/foo", 42, 84)
	c.Assert(err, IsNil)

	f, _ = s.a.OpenFile("foo", os.O_RDONLY, 0)
	c.Assert(f.inode.UserId, Equals, uint64(42))
	c.Assert(f.inode.GroupId, Equals, uint64(84))
}

func (s *FSSuite) TestArchive_Rename(c *C) {
	f, _ := s.a.Create("foo")
	f.WriteString("foo")
	f.Close()

	err := s.a.Rename("/foo", "/bar")
	c.Assert(err, IsNil)

	f, err = s.a.Open("bar")
	c.Assert(err, IsNil)
	c.Assert(f.Name(), Equals, "/bar")
	c.Assert(f.buf.Len(), Equals, 3)

	_, err = s.a.Stat("foo")
	c.Assert(err, Not(IsNil))
}

func (s *FSSuite) TestArchive_RenameExists(c *C) {
	f, _ := s.a.Create("foo")
	f.Close()

	f, _ = s.a.Create("bar")
	f.Close()

	err := s.a.Rename("/foo", "/bar")
	c.Assert(err, Not(IsNil))
}

func (s *FSSuite) TestArchive_Truncate(c *C) {
	f, _ := s.a.Create("foo")
	f.WriteString("foo")
	f.Close()

	err := s.a.Truncate("/foo", 1)
	c.Assert(err, IsNil)

	f, _ = s.a.OpenFile("foo", os.O_RDONLY, 0)
	c.Assert(f.buf.Len(), Equals, 1)
}

func (s *FSSuite) TestArchive_Open(c *C) {
	f, err := s.a.Create("foo")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 0)

	f.Write([]byte("foo"))
	f.Close()

	f, err = s.a.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Name(), Equals, "/foo")
	c.Assert(f.buf.Len(), Equals, 3)
}

func (s *FSSuite) TestArchive_OpenFile(c *C) {
	f, err := s.a.OpenFile("foo", os.O_EXCL|os.O_CREATE, 0)
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 0)

	f.Write([]byte("foo"))
	f.Close()

	f, err = s.a.OpenFile("foo", os.O_EXCL, 0)
	c.Assert(err, FitsTypeOf, &os.PathError{})
}

func (s *FSSuite) TestArchive_Create(c *C) {
	f, err := s.a.Create("foo")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 0)

	f.Write([]byte("foo"))
	f.Close()

	f, err = s.a.Create("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Name(), Equals, "/foo")
	c.Assert(f.inode.Size, Equals, int64(0))
	c.Assert(f.buf.Len(), Equals, 0)
}

func (s *FSSuite) TestArchive_Stat(c *C) {
	f, _ := s.a.Create("foo")
	f.WriteString("foo")
	f.Close()

	fi, err := s.a.Stat("/foo")
	c.Assert(err, IsNil)
	c.Assert(fi.Name(), Equals, "foo")
}

func (s *FSSuite) TestArchive_Stat_NotFound(c *C) {
	_, err := s.a.Stat("/foo")
	c.Assert(err, FitsTypeOf, &os.PathError{})
}

func (s *FSSuite) TestArchive_Remove(c *C) {
	f, _ := s.a.Create("foo")
	f.Write([]byte("foo"))
	f.Close()

	err := s.a.Remove("foo")
	c.Assert(err, IsNil)

	f, err = s.a.Open("foo")
	c.Assert(err, Not(IsNil))
}

func (s *FSSuite) TestArchive_RemoveAll(c *C) {
	f, _ := s.a.Create("foo")
	f.Write([]byte("foo"))
	f.Close()

	f, _ = s.a.Create("foobar")
	f.Write([]byte("foo"))
	f.Close()

	f, _ = s.a.Create("foo/bar")
	f.Write([]byte("foo"))
	f.Close()

	err := s.a.RemoveAll("foo")
	c.Assert(err, IsNil)

	f, err = s.a.Open("foo")
	c.Assert(err, Not(IsNil))

	f, err = s.a.Open("foobar")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 3)

	f, err = s.a.Open("foo/bar")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 3)
}

func (s *FSSuite) TestArchive_Find(c *C) {
	f, _ := s.a.Create("foo")
	f.Write([]byte("foo"))
	f.Close()

	f, _ = s.a.Create("foo/qux")
	f.Write([]byte("foo"))
	f.Close()

	f, _ = s.a.Create("foo/bar")
	f.Write([]byte("foo"))
	f.Close()

	r := s.a.Find(func(name string) bool {
		return strings.HasPrefix(name, "/foo/")
	})

	c.Assert(r, HasLen, 2)
}
func (s *FSSuite) TearDownTest(c *C) {
	s.a.Close()
	if err := os.Remove(s.file); err != nil {
		panic(err)
	}
}
