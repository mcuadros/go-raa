package boltfs

import (
	"io/ioutil"
	"os"
	"path/filepath"
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

func (s *FSSuite) TestVolume_Open(c *C) {
	f, err := s.v.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 0)

	f.Write([]byte("foo"))
	f.Close()

	f, err = s.v.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Name(), Equals, "/foo")
	c.Assert(f.buf.Len(), Equals, 3)
}

func (s *FSSuite) TestVolume_Remove(c *C) {
	f, _ := s.v.Open("foo")
	f.Write([]byte("foo"))
	f.Close()

	err := s.v.Remove("foo")
	c.Assert(err, IsNil)

	f, err = s.v.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 0)
}

func (s *FSSuite) TestVolume_RemoveAll(c *C) {
	f, _ := s.v.Open("foo")
	f.Write([]byte("foo"))
	f.Close()

	f, _ = s.v.Open("foobar")
	f.Write([]byte("foo"))
	f.Close()

	f, _ = s.v.Open("foo/bar")
	f.Write([]byte("foo"))
	f.Close()

	err := s.v.RemoveAll("foo")
	c.Assert(err, IsNil)

	f, err = s.v.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 0)

	f, err = s.v.Open("foobar")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 3)

	f, err = s.v.Open("foo/bar")
	c.Assert(err, IsNil)
	c.Assert(f.buf.Len(), Equals, 3)
}

func (s *FSSuite) TearDownTest(c *C) {
	s.v.Close()
	if err := os.Remove(s.file); err != nil {
		panic(err)
	}
}
