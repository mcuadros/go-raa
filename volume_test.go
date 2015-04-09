package boltfs

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type FSSuite struct {
	v *Volume
}

var _ = Suite(&FSSuite{})

const TestDBFile = "/tmp/foo.db"

func (s *FSSuite) SetUpTest(c *C) {
	var err error
	s.v, err = NewVolume(TestDBFile)
	if err != nil {
		panic(err)
	}
}

func (s *FSSuite) TestVolume_Open(c *C) {
	f, err := s.v.Open("foo")

	c.Assert(err, IsNil)
	c.Assert(f.Name(), Equals, "foo")
	c.Assert(f.buf.Len(), Equals, 0)
}

func (s *FSSuite) TestVolume_OpenNew(c *C) {
	f, err := s.v.Open("foo")
	c.Assert(err, IsNil)

	f.Write([]byte("foo"))
	f.Close()

	f, err = s.v.Open("foo")
	c.Assert(err, IsNil)
	c.Assert(f.Name(), Equals, "foo")
	c.Assert(f.buf.Len(), Equals, 3)
}

func (s *FSSuite) TearDownTest(c *C) {
	s.v.Close()
}
