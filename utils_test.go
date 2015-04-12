package raa

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

func (s *FSSuite) TestAddFile(c *C) {
	src, err := ioutil.TempFile("/tmp/", "perms_raa")
	c.Assert(err, IsNil)

	src.WriteString("foo")
	src.Chmod(0555)
	src.Close()

	n, err := AddFile(s.v, src.Name(), "/bar")
	c.Assert(n, Equals, int64(3))
	c.Assert(err, IsNil)

	dst, err := s.v.Open("/bar")
	dst.Close()
	c.Assert(err, IsNil)
	c.Assert(int(dst.inode.Mode), Equals, 0555)
}

func (s *FSSuite) TestAddGlob(c *C) {
	dir := makeDirFixture()

	n, err := AddGlob(s.v, filepath.Join(dir, "*"), "foo", false)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	dst, err := s.v.Open("/foo/qux")
	defer dst.Close()
	c.Assert(err, IsNil)
	c.Assert(dst.String(), Equals, "qux")

	dst, err = s.v.Open("/foo/bar")
	defer dst.Close()
	c.Assert(err, IsNil)
	c.Assert(dst.String(), Equals, "bar")
}

func (s *FSSuite) TestAddGlob_Rescursive(c *C) {
	dir := makeDirFixture()

	n, err := AddGlob(s.v, filepath.Join(dir, "*"), "foo", true)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)

	dst, err := s.v.Open("/foo/baz/baz")
	defer dst.Close()
	c.Assert(err, IsNil)
	c.Assert(dst.String(), Equals, "baz")
}

func makeDirFixture() string {
	dir, err := ioutil.TempDir("/tmp/", "fixture")
	if err != nil {
		panic(err)
	}

	makeFileFixture(filepath.Join(dir, "qux"), "qux")
	makeFileFixture(filepath.Join(dir, "bar"), "bar")

	if err := os.Mkdir(filepath.Join(dir, "baz"), 0766); err != nil {
		panic(err)
	}

	makeFileFixture(filepath.Join(dir, "baz/baz"), "baz")

	return dir
}

func makeFileFixture(path, content string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	f.WriteString(content)
	f.Close()
}
