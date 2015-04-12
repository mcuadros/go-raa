package benchmark

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/mcuadros/raa"
)

const FixtureTarPattern = "fixtures/%d_files.tar"
const FixtureDbParttern = "fixtures/%d_files.db"

func buildVolumeFromTar(files int) []string {
	result := make([]string, 0)

	file, err := os.Open(fmt.Sprintf(FixtureTarPattern, files))
	if err != nil {
		panic(err)
	}

	v, err := raa.NewVolume(fmt.Sprintf(FixtureDbParttern, files))
	if err != nil {
		panic(err)
	}

	tar := tar.NewReader(file)
	cur := 0
	for {
		hdr, err := tar.Next()
		if err == io.EOF {
			break
		}

		ifErrPanic(err)

		file, err := v.Create(hdr.Name)
		ifErrPanic(err)

		_, err = io.Copy(file, tar)
		ifErrPanic(err)
		file.Close()

		if !hdr.FileInfo().IsDir() {
			result = append(result, hdr.Name)
		}

		cur++
	}

	v.Close()
	return result
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

	//ifErrPanic(err)

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

		if hdr.Name == randomFile {
			//fmt.Printf("Contents of %s:\n", hdr.Name)
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
