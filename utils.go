package raa

import (
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
