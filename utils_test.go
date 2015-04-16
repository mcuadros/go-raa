package raa

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "gopkg.in/check.v1"
)

func (s *FSSuite) TestAddFile(c *C) {
	src, err := ioutil.TempFile("/tmp/", "perms_raa")
	c.Assert(err, IsNil)

	src.WriteString("foo")
	src.Chmod(0555)
	src.Close()

	n, err := AddFile(s.a, src.Name(), "/bar")
	c.Assert(n, Equals, int64(3))
	c.Assert(err, IsNil)

	dst, err := s.a.Open("/bar")
	dst.Close()
	c.Assert(err, IsNil)
	c.Assert(int(dst.inode.Mode), Equals, 0555)
}

func (s *FSSuite) TestAddGlob(c *C) {
	dir := makeDirFixture()

	n, err := AddGlob(s.a, filepath.Join(dir, "*"), "foo", false)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 2)

	dst, err := s.a.Open("/foo/qux")
	defer dst.Close()
	c.Assert(err, IsNil)
	c.Assert(dst.String(), Equals, "qux")

	dst, err = s.a.Open("/foo/bar")
	defer dst.Close()
	c.Assert(err, IsNil)
	c.Assert(dst.String(), Equals, "bar")
}

func (s *FSSuite) TestAddGlob_Rescursive(c *C) {
	dir := makeDirFixture()

	n, err := AddGlob(s.a, filepath.Join(dir, "*"), "foo", true)
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 3)

	dst, err := s.a.Open("/foo/baz/baz")
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

func (s *FSSuite) TestAddTarContent(c *C) {
	f, err := os.Open(smallFileTar)
	c.Assert(err, IsNil)

	n, err := AddTarContent(s.a, f, "/")
	c.Assert(err, IsNil)
	c.Assert(n, Equals, 61)

	s.a.Close()

	v, err := CreateArchive(s.file)
	if err != nil {
		panic(err)
	}

	AssertVolumeAgainstTar(c, v, smallFileTar, 61)
}

func AssertVolumeAgainstTar(c *C, a *Archive, tar string, files int) {
	count := 0
	dir := extractTarToDir(tar)
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		rel, _ := filepath.Rel(dir, path)
		if !f.Mode().IsRegular() {
			return nil
		}

		count++
		assertFileAgaintsSource(c, a, path, rel)
		return nil
	})

	c.Assert(err, IsNil)
	c.Assert(count, Equals, files)
}

func assertFileAgaintsSource(c *C, a *Archive, o, f string) {
	orig, err := os.Open(o)
	defer orig.Close()
	c.Assert(err, IsNil)

	file, err := a.Open(f)
	defer file.Close()
	c.Assert(err, IsNil)

	oc, _ := ioutil.ReadAll(orig)
	fc, _ := ioutil.ReadAll(file)
	c.Assert(oc, DeepEquals, fc)

	ofi, _ := orig.Stat()
	ffi, _ := file.Stat()
	c.Assert(ofi.Size(), Equals, ffi.Size())
	c.Assert(ofi.Mode(), Equals, ffi.Mode())
}

func extractTarToDir(file string) string {
	dir, err := ioutil.TempDir("/tmp/", "tar")
	if err != nil {
		panic(err)
	}

	path, _ := os.Getwd()
	cmd := exec.Command("tar", "-xf", filepath.Join(path, file))
	cmd.Dir = dir

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	err = cmd.Wait()
	if err != nil {
		panic(err)
	}

	return dir
}
