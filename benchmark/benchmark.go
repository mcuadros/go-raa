package benchmark

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/mcuadros/go-raa"
	"github.com/mcuadros/go-rtar"
)

const FixturePattern = "/tmp/%d_%s_%s_files.%s"

func openDbAndReadFile(files, minSize, maxSize uint64, names []string) {
	randomFile := names[rand.Intn(len(names))]

	a, err := raa.OpenArchive(getFixtureFilename(files, minSize, maxSize, "raa"))
	if err != nil {
		panic(err)
	}

	file, err := a.Open(randomFile)
	ifErrPanic(err)

	buf := bytes.NewBuffer(nil)

	s, _ := file.Stat()

	n, err := io.Copy(buf, file)

	if s.Size() != n {
		panic("ws")
	}

	a.Close()
}

func openTarAndReadFile(files, minSize, maxSize uint64, names []string) {
	randomFile := names[rand.Intn(len(names))]
	file, err := os.Open(getFixtureFilename(files, minSize, maxSize, "tar"))
	if err != nil {
		panic(err)
	}

	defer file.Close()
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
			found = true
			break
		}
	}

	if !found {
		panic("Cannot find file: " + randomFile)
	}
}

func openTarMappedAndReadFile(files, minSize, maxSize uint64, names []string) {
	randomFile := names[rand.Intn(len(names))]
	file, err := os.Open(getFixtureFilename(files, minSize, maxSize, "tar"))
	if err != nil {
		panic(err)
	}

	defer file.Close()

	m, err := os.Open(getFixtureFilename(files, minSize, maxSize, "tar.map"))
	if err != nil {
		panic(err)
	}

	defer m.Close()

	tar, _ := rat.NewReader(file, m)
	content, err := tar.ReadFile(randomFile)
	ifErrPanic(err)

	bytes.NewBuffer(content)
}

func getFixtureFilename(files, minSize, maxSize uint64, ext string) string {
	return fmt.Sprintf(FixturePattern,
		files,
		humanize.Bytes(minSize),
		humanize.Bytes(maxSize),
		ext,
	)
}

func getFixtureRandomSize(minSize, maxSize uint64) int {
	return int(minSize) + rand.Intn(int(maxSize-minSize))
}

func fixtureGenerator(files, minSize, maxSize uint64) []string {
	f, err := os.Create(getFixtureFilename(files, minSize, maxSize, "tar"))
	if err != nil {
		ifErrPanic(err)
	}

	m, err := os.Create(getFixtureFilename(files, minSize, maxSize, "tar.map"))
	if err != nil {
		ifErrPanic(err)
	}

	defer f.Close()
	defer m.Close()

	result := make([]string, files)
	// Create a new tar archive.
	tw := rat.NewWriter(f, m)
	for i := 0; i < int(files); i++ {
		size := getFixtureRandomSize(minSize, maxSize)
		fname := fmt.Sprintf("file_%d_%d.foo", i, size)
		result[i] = fname

		hdr := &tar.Header{
			Name:     fname,
			Size:     int64(size),
			Typeflag: tar.TypeReg,
		}

		ifErrPanic(tw.WriteHeader(hdr))

		_, err := tw.Write(bytes.Repeat([]byte("f"), size))
		ifErrPanic(err)
	}

	ifErrPanic(tw.Close())

	buildVolumeFromTar(files, minSize, maxSize)
	return result
}

func buildVolumeFromTar(files, minSize, maxSize uint64) {
	file, err := os.Open(getFixtureFilename(files, minSize, maxSize, "tar"))
	ifErrPanic(err)
	defer file.Close()

	a, err := raa.CreateArchive(getFixtureFilename(files, minSize, maxSize, "raa"))
	ifErrPanic(err)
	defer a.Close()

	_, err = raa.AddTarContent(a, file, "/")
	ifErrPanic(err)
}

func ifErrPanic(err error) {
	if err != nil && err != io.EOF {
		panic(err)
	}
}
