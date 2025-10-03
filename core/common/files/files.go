package files

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

// Copy copies src file to the dest location.
func Copy(src, dest string) (err error) {
	var in *os.File
	if in, err = os.Open(src); err != nil {
		return
	}
	defer in.Close()
	return WriteFile(in, dest)
}

// IsFile checks that the path exists, is accessible and is a file.
func IsFile(path string) error {
	fi, err := os.Stat(path)
	if err == nil && fi != nil && fi.IsDir() {
		err = fmt.Errorf("Path is a directory, not a file: %s", path)
	}
	return err
}

// IsDir checks if the filesystem path exists and is a directory.
func IsDir(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return &os.PathError{Op: "isdir", Path: path, Err: syscall.ENOTDIR}
	}
	return nil
}

// CleanDir removes the directory at the filesystem path if it exists and then
// creates it and any missing parents with permission bits 0700 (before umask).
func CleanDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return os.MkdirAll(path, 0700)
}

// WriteFile reads the data from the reader and writes it into the specified file.
func WriteFile(data io.Reader, path string) (err error) {
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return
	}
	var out *os.File
	if out, err = os.Create(path); err != nil {
		return
	}
	defer out.Close()
	_, err = io.Copy(out, data)
	return
}

// LineByLine calls the specified function for all lines in the file
func LineByLine(path string, f func(string) error) (err error) {
	var file *os.File
	if file, err = os.Open(path); err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err = f(scanner.Text()); err != nil {
			return
		}
	}
	return scanner.Err()
}
