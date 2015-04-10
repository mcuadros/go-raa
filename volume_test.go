package boltfs

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
	v    *Volume
	file string
}

var _ = Suite(&FSSuite{})

const TestDBFile = "foo.db"

func (s *FSSuite) SetUpTest(c *C) {
	tempDir, err := ioutil.TempDir("/tmp", "boltfs")
	if err != nil {
		panic(err)
	}

	s.file = filepath.Join(tempDir, TestDBFile)

	s.v, err = NewVolume(s.file)
	if err != nil {
		panic(err)
	}
}

func (s *FSSuite) TestVolume_ChdirAndGetcwd(c *C) {
	c.Assert(s.v.path, Equals, "/")

	err := s.v.Chdir("foo")
	c.Assert(err, IsNil)
	c.Assert(s.v.path, Equals, "/foo")

	err = s.v.Chdir("foo")
	c.Assert(err, IsNil)
	c.Assert(s.v.path, Equals, "/foo/foo")

	err = s.v.Chdir("..")
	c.Assert(err, IsNil)
	c.Assert(s.v.path, Equals, "/foo")

	err = s.v.Chdir("/bar")
	c.Assert(err, IsNil)

	path, err := s.v.Getwd()
	c.Assert(path, Equals, "/bar")
}

func (s *FSSuite) TestVolume_Rename(c *C) {
	f, _ := s.v.Create("foo")
	f.WriteString("foo")
	f.Close()

	err := s.v.Rename("/foo", "/bar")
	c.Assert(err, IsNil)

	f, err = s.v.Open("bar")
	c.Assert(err, IsNil)
	c.Assert(f.Name(), Equals, "/bar")
	c.Assert(f.buf.Len(), Equals, 3)

	_, err = s.v.Stat("foo")
	c.Assert(err, Not(IsNil))
}

func (s *FSSuite) TestVolume_RenameExists(c *C) {
	f, _ := s.v.Create("foo")
	f.Close()

	f, _ = s.v.Create("bar")
	f.Close()

	err := s.v.Rename("/foo", "/bar")
	c.Assert(err, Not(IsNil))
}

func (s *FSSuite) TestVolume_Open(c *C) {
	f, err := s.v.Create("foo")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 0)

	f.Write([]byte("foo"))
	f.Close()

	f, err = s.v.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Name(), Equals, "/foo")
	c.Assert(f.buf.Len(), Equals, 3)
}

func (s *FSSuite) TestVolume_OpenFile(c *C) {
	f, err := s.v.OpenFile("foo", os.O_EXCL|os.O_CREATE, 0)
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 0)

	f.Write([]byte("foo"))
	f.Close()

	f, err = s.v.OpenFile("foo", os.O_EXCL, 0)
	c.Assert(err, FitsTypeOf, &os.PathError{})
}

func (s *FSSuite) TestVolume_Create(c *C) {
	f, err := s.v.Create("foo")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 0)

	f.Write([]byte("foo"))
	f.Close()

	f, err = s.v.Create("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Name(), Equals, "/foo")
	c.Assert(f.hdr.Size, Equals, int64(0))
	c.Assert(f.buf.Len(), Equals, 0)
}

func (s *FSSuite) TestVolume_Stat(c *C) {
	f, _ := s.v.Create("foo")
	f.WriteString("foo")
	f.Close()

	st, err := s.v.Stat("/foo")
	c.Assert(err, IsNil)
	c.Assert(st.Name(), Equals, "foo")
}

func (s *FSSuite) TestVolume_Remove(c *C) {
	f, _ := s.v.Create("foo")
	f.Write([]byte("foo"))
	f.Close()

	err := s.v.Remove("foo")
	c.Assert(err, IsNil)

	f, err = s.v.Open("foo")
	c.Assert(err, Not(IsNil))
}

func (s *FSSuite) TestVolume_RemoveAll(c *C) {
	f, _ := s.v.Create("foo")
	f.Write([]byte("foo"))
	f.Close()

	f, _ = s.v.Create("foobar")
	f.Write([]byte("foo"))
	f.Close()

	f, _ = s.v.Create("foo/bar")
	f.Write([]byte("foo"))
	f.Close()

	err := s.v.RemoveAll("foo")
	c.Assert(err, IsNil)

	f, err = s.v.Open("foo")
	c.Assert(err, Not(IsNil))

	f, err = s.v.Open("foobar")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 3)

	f, err = s.v.Open("foo/bar")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 3)
}

func (s *FSSuite) TestVolume_Find(c *C) {
	f, _ := s.v.Create("foo")
	f.Write([]byte("foo"))
	f.Close()

	f, _ = s.v.Create("foo/qux")
	f.Write([]byte("foo"))
	f.Close()

	f, _ = s.v.Create("foo/bar")
	f.Write([]byte("foo"))
	f.Close()

	r := s.v.Find(func(name string) bool {
		return strings.HasPrefix(name, "/foo/")
	})

	c.Assert(r, HasLen, 2)
}
func (s *FSSuite) TearDownTest(c *C) {
	s.v.Close()
	if err := os.Remove(s.file); err != nil {
		panic(err)
	}
}
