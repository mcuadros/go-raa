package benchmark

import (
	"math/rand"
	"testing"

	"github.com/mcuadros/go-raa"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type FSSuite struct{}

var _ = Suite(&FSSuite{})

const RandomSeed = 42

var files5 []string
var files78 []string
var files6133 []string
var files820 []string
var filesBig60 []string

const (
	kb = 1024
	mb = kb * 1024
	gb = mb * 1024
)

func (s *FSSuite) SetUpSuite(c *C) {
	rand.Seed(RandomSeed)
	files5 = fixtureGenerator(5, kb, 100*kb)
	files78 = fixtureGenerator(100, kb, 100*kb)
	files820 = fixtureGenerator(1000, kb, 100*kb)
	files6133 = fixtureGenerator(6000, kb, 100*kb)
	filesBig60 = fixtureGenerator(60, mb, 20*mb)

}

func (s *FSSuite) BenchmarkReadingRandomFilesFromTar_5(c *C) {
	for i := 0; i < c.N; i++ {
		openTarAndReadFile(5, kb, 100*kb, files5)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromTar_60(c *C) {
	for i := 0; i < c.N; i++ {
		openTarAndReadFile(60, mb, 20*mb, filesBig60)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromTar_100(c *C) {
	for i := 0; i < c.N; i++ {
		openTarAndReadFile(100, kb, 100*kb, files78)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromTar_1k(c *C) {
	for i := 0; i < c.N; i++ {
		openTarAndReadFile(1000, kb, 100*kb, files820)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromTar_6k(c *C) {
	for i := 0; i < c.N; i++ {
		openTarAndReadFile(6000, kb, 100*kb, files6133)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromDb_6(c *C) {
	for i := 0; i < c.N; i++ {
		openDbAndReadFile(60, mb, 20*mb, filesBig60)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromDb_5(c *C) {
	for i := 0; i < c.N; i++ {
		openDbAndReadFile(5, kb, 100*kb, files5)
	}
}
func (s *FSSuite) BenchmarkReadingRandomFilesFromDb_100(c *C) {
	for i := 0; i < c.N; i++ {
		openDbAndReadFile(100, kb, 100*kb, files78)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromDb_1k(c *C) {
	for i := 0; i < c.N; i++ {
		openDbAndReadFile(1000, kb, 100*kb, files820)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromDb_6k(c *C) {
	for i := 0; i < c.N; i++ {
		openDbAndReadFile(6000, kb, 100*kb, files6133)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromMapped_6(c *C) {
	for i := 0; i < c.N; i++ {
		openTarMappedAndReadFile(60, mb, 20*mb, filesBig60)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromMapped_5(c *C) {
	for i := 0; i < c.N; i++ {
		openTarMappedAndReadFile(5, kb, 100*kb, files5)
	}
}
func (s *FSSuite) BenchmarkReadingRandomFilesFromMapped_100(c *C) {
	for i := 0; i < c.N; i++ {
		openTarMappedAndReadFile(100, kb, 100*kb, files78)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromMapped_1k(c *C) {
	for i := 0; i < c.N; i++ {
		openTarMappedAndReadFile(1000, kb, 100*kb, files820)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromMapped_6k(c *C) {
	for i := 0; i < c.N; i++ {
		openTarMappedAndReadFile(6000, kb, 100*kb, files6133)
	}
}

func (s *FSSuite) BenchmarkFindingFilesFromDb_6k(c *C) {
	a, err := raa.OpenArchive(getFixtureFilename(6000, kb, 100*kb, "raa"))
	if err != nil {
		panic(err)
	}

	for i := 0; i < c.N; i++ {
		randomFile := files6133[rand.Intn(len(files6133))]

		r := a.Find(func(name string) bool {
			if name == randomFile[1:] {
				return true
			}

			return false
		})

		if len(r) != 0 {
			panic("not found")
		}
	}
}
