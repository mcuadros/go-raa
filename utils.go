package raa

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// AddFile adds a OS file to a Volume, returns the number of bytes written
func AddFile(v *Volume, from, to string) (int64, error) {
	src, err := os.Open(from)
	if err != nil {
		return -1, err
	}

	defer src.Close()
	dst, err := v.Create(to)
	if err != nil {
		return -1, err
	}

	fi, err := src.Stat()
	if err != nil {
		return -1, err
	}

	dst.Chmod(fi.Mode())
	dst.Chown(
		int(fi.Sys().(*syscall.Stat_t).Uid),
		int(fi.Sys().(*syscall.Stat_t).Gid),
	)

	defer dst.Close()

	return io.Copy(dst, src)
}

// AddFile adds a OS directory to a Volume, returns the number of files written
func AddDirectory(v *Volume, from, to string, recursive bool) (int, error) {
	return AddGlob(v, filepath.Join(from, "*"), to, recursive)
}

// AddGlob adds a OS files and directories to a Volume using a glob pattern,
// returns the number of files written
func AddGlob(v *Volume, pattern, to string, recursive bool) (int, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return -1, err
	}

	count := 0
	for _, file := range files {
		fi, _ := os.Stat(file)
		dst := filepath.Join(to, fi.Name())

		switch {
		case fi.Mode().IsRegular():
			if _, err := AddFile(v, file, dst); err != nil {
				return count, err
			}

			count++
		case fi.IsDir() && recursive:
			n, err := AddDirectory(v, file, dst, recursive)
			if n != -1 {
				count += n
			}

			if err != nil {
				return count, err
			}
		}
	}

	return count, nil
}

// AddTarContent add the contained files in a tar stream to the volume, returns
// the number of files copied to the Volume
func AddTarContent(v *Volume, file io.Reader, to string) (int, error) {
	reader := tar.NewReader(file)
	count := 0
	for {
		hdr, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return count, err
		}

		switch hdr.Typeflag {
		case tar.TypeReg:
			if err := readFileFromTar(v, reader, hdr, to); err != nil {
				return count, err
			}

			count++
		}
	}

	return count, nil
}

func readFileFromTar(v *Volume, reader *tar.Reader, h *tar.Header, to string) error {
	file, err := createFileFromTarHeader(v, h, to)
	defer file.Close()
	if err != nil {
		return err
	}

	if _, err = io.Copy(file, reader); err != nil {
		return err
	}

	return nil
}

func createFileFromTarHeader(v *Volume, h *tar.Header, to string) (*File, error) {
	file, err := v.Create(filepath.Join(to, h.Name))
	if err != nil {
		return nil, err
	}

	file.Chown(h.Uid, h.Gid)
	file.Chmod(os.FileMode(h.Mode))

	return file, nil
}
