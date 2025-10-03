package tar

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"tivi.io/core/common/files"
)

type Exclusions []string

func NewExclusions(s ...string) Exclusions {
	exc := make(Exclusions, len(s))
	i := 0
	for _, v := range s {
		exc[i] = v
		i++
	}
	return exc
}

func (e Exclusions) Contains(x string) bool {
	for _, v := range e {
		if v == x {
			return true
		}
	}
	return false
}

func CompressFile(path string, tarball *tar.Writer, info os.FileInfo, prefix string) error {

	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	header.Name = strings.TrimPrefix(path, prefix)

	if err := tarball.WriteHeader(header); err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(tarball, file)
	return err

}

func CompressDirectory(inputDir string, tarball *tar.Writer, exc Exclusions) (err error) {
	return filepath.Walk(inputDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			extension := filepath.Ext(path)

			if exc.Contains(extension) {
				return nil
			}

			return CompressFile(path, tarball, info, inputDir)
		})
}

func Compress(path string, buf io.Writer, exc Exclusions) (err error) {

	tarball := tar.NewWriter(buf)
	defer func() {
		err = tarball.Close()
	}()

	if files.IsDir(path) == nil {
		return CompressDirectory(path, tarball, exc)
	}

	var info os.FileInfo
	info, err = os.Stat(path)
	if err != nil {
		return
	}
	// TODO - exclusion of the single file
	return CompressFile(path, tarball, info, filepath.Dir(path))
}

func Create(inputDir string, tarName string, exc Exclusions) (err error) {

	tarfile, err := os.Create(tarName)
	if err != nil {
		return err
	}
	defer func() {
		err = tarfile.Close()
	}()

	return Compress(inputDir, tarfile, exc)
}

// Extract extracts the data as tar into the specified destination.
func Extract(data io.Reader, dest string) (err error) {
	// If results directory exists, assume it is already extracted
	var fi os.FileInfo
	if fi, err = os.Stat(dest); err == nil && fi.IsDir() {
		return nil
	}
	defer func() {
		// Cleanup the extracted directory in case of an error
		if err != nil {
			if tmpErr := os.RemoveAll(dest); tmpErr != nil {
				err = fmt.Errorf("failed to cleanup after error: %w", err)
			}
		}
	}()

	if err = os.MkdirAll(dest, 0700); err != nil {
		return err
	}

	tr := tar.NewReader(data)
	for {
		var hdr *tar.Header
		if hdr, err = tr.Next(); err == io.EOF {
			break
		} else if err != nil {
			return
		}
		path := filepath.Join(dest, hdr.Name)
		if hdr.FileInfo().IsDir() {
			if err = os.MkdirAll(path, 0700); err != nil {
				return
			}
		} else if err = files.WriteFile(tr, path); err != nil {
			return
		}
	}
	return nil
}

func ExtractToDir(src string, dest string) (err error) {

	srcfile, err := os.Open(src)
	if err != nil {
		return err
	}

	defer func() {
		err = srcfile.Close()
	}()

	return Extract(srcfile, dest)
}
