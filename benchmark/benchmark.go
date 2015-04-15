package benchmark

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/mcuadros/go-raa"
)

const FixtureTarPattern = "fixtures/%d_files.tar"
const FixtureDbParttern = "fixtures/%d_files.raa"

func buildVolumeFromTar(numFiles int) []string {
	file, err := os.Open(fmt.Sprintf(FixtureTarPattern, numFiles))
	ifErrPanic(err)
	defer file.Close()

	v, err := raa.NewVolume(fmt.Sprintf(FixtureDbParttern, numFiles))
	ifErrPanic(err)
	defer v.Close()

	_, err = raa.AddTarContent(v, file, "/")
	ifErrPanic(err)

	return v.Find(func(string) bool { return true })
}

func openDbAndReadFile(files int, names []string) {
	randomFile := names[rand.Intn(len(names))]

	v, err := raa.NewVolume(fmt.Sprintf(FixtureDbParttern, files))
	if err != nil {
		panic(err)
	}

	file, err := v.Open(randomFile)
	ifErrPanic(err)

	buf := bytes.NewBuffer(nil)

	s, _ := file.Stat()

	n, err := io.Copy(buf, file)

	if s.Size() != n {
		panic("ws")
	}

	v.Close()
}

func openTarAndReadFile(files int, names []string) {
	randomFile := names[rand.Intn(len(names))]
	file, err := os.Open(fmt.Sprintf(FixtureTarPattern, files))
	if err != nil {
		panic(err)
	}

	tar := tar.NewReader(file)
	found := false
	for {
		hdr, err := tar.Next()
		if err == io.EOF {
			break
		}
		ifErrPanic(err)

		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, tar)
		ifErrPanic(err)

		if hdr.Name == randomFile[1:] {
			found = true
			break
		}
	}

	if !found {
		panic("Cannot find file: " + randomFile)
	}
}

func ifErrPanic(err error) {
	if err != nil && err != io.EOF {
		panic(err)
	}
}
