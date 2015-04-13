package benchmark

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/mcuadros/raa"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type FSSuite struct{}

var _ = Suite(&FSSuite{})

const RandomSeed = 42

var files78 []string
var files6133 []string
var files820 []string

func (s *FSSuite) SetUpSuite(c *C) {
	rand.Seed(RandomSeed)
	files78 = buildVolumeFromTar(78)
	files6133 = buildVolumeFromTar(6133)
	files820 = buildVolumeFromTar(820)
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromTar_78(c *C) {
	for i := 0; i < c.N; i++ {
		openTarAndReadFile(78, files78)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromTar_1k(c *C) {
	for i := 0; i < c.N; i++ {
		openTarAndReadFile(820, files820)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromTar_6k(c *C) {
	for i := 0; i < c.N; i++ {
		openTarAndReadFile(6133, files6133)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromDb_78(c *C) {
	for i := 0; i < c.N; i++ {
		openDbAndReadFile(78, files78)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromDb_1k(c *C) {
	for i := 0; i < c.N; i++ {
		openDbAndReadFile(820, files820)
	}
}

func (s *FSSuite) BenchmarkReadingRandomFilesFromDb_6k(c *C) {
	for i := 0; i < c.N; i++ {
		openDbAndReadFile(6133, files6133)
	}
}

func (s *FSSuite) BenchmarkFindingFilesFromDb_6k(c *C) {
	v, err := raa.NewVolume(fmt.Sprintf(FixtureDbParttern, 6133))
	if err != nil {
		panic(err)
	}

	for i := 0; i < c.N; i++ {
		randomFile := files6133[rand.Intn(len(files6133))]

		r := v.Find(func(name string) bool {
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
