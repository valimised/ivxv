/*
The voteunion application is used for finding the union of a set of exported
ballot boxes.
*/
package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"hash"
	"io"
	"os"

	"ivxv.ee/command/exit"
)

func usage() {
	// nolint: gas, if writing to stderr returns an error, then we do not
	// have any way to report it anyway.
	fmt.Fprintln(os.Stderr, "usage:", os.Args[0], `[options] <output> <input>...

voteunion takes a set of ballot boxes exported using voteexp and outputs a
ballot box which contains the union of votes in these inputs.

options:`)
	flag.PrintDefaults()
}

var qp = flag.Bool("q", false, "quiet mode")

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(exit.Usage)
	}

	code, err := zipunion(args[0], args[1:])
	if err != nil {
		// nolint: gas, if writing to stderr returns an error, then we
		// do not have any way to report it anyway.
		fmt.Fprintln(os.Stderr, "error:", err)
	}
	os.Exit(code)
}

func zipunion(output string, input []string) (code int, err error) {
	// Create output archive and setup defers to finalize the ZIP and close
	// the file pointer.
	fp, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return exit.CantCreate, fmt.Errorf("failed to create output: %v", err)
	}
	defer func() {
		if cerr := fp.Close(); err == nil && cerr != nil {
			code = exit.IOErr
			err = fmt.Errorf("failed to close output: %v", err)
		}
	}()

	u := union{
		w:        zip.NewWriter(fp),
		contents: make(map[string][]byte),
		hash:     sha256.New(),
	}
	u.sum = make([]byte, u.hash.Size())
	u.block = make([]byte, 256*u.hash.BlockSize())
	defer func() {
		if cerr := u.w.Close(); err == nil && cerr != nil {
			code = exit.Unavailable
			err = fmt.Errorf("failed to write central directory: %v", err)
		}
	}()

	// Open input archives and add their contents to output if not already
	// there.
	for _, in := range input {
		if !*qp {
			u.out.push("%s: ", in)
		}
		code, err = u.zipadd(in)
		if !*qp {
			u.out.pop()
		}
		if err != nil {
			return code, fmt.Errorf("failed to add %q: %v", in, err)
		}
	}
	return
}

type union struct {
	w        *zip.Writer
	contents map[string][]byte // file name to SHA-256 hash of the contents.
	out      progress

	// Reusable scratch variables.
	hash  hash.Hash
	sum   []byte
	block []byte
}

func (u *union) zipadd(path string) (code int, err error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return exit.NoInput, fmt.Errorf("failed to open input: %v", err)
	}
	defer r.Close() // nolint: errcheck, ignore close failure of read-only fd.

	count := len(r.File)
	for i, f := range r.File {
		if !*qp {
			u.out.push("%*d/%d %s",
				len(fmt.Sprint(count)), i+1, count, f.FileHeader.Name)
		}
		code, err = u.fileadd(f, path)
		if !*qp {
			u.out.pop()
		}
		if err != nil {
			return code, fmt.Errorf("failed to copy file %q: %v",
				f.FileHeader.Name, err)
		}
	}
	return
}

func (u *union) fileadd(f *zip.File, path string) (code int, err error) {
	// Open the archive file for reading.
	src, err := f.Open()
	if err != nil {
		return exit.DataErr, fmt.Errorf("failed to open file: %v", err)
	}
	defer src.Close() // nolint: errcheck, ignore close failure of read-only fd.

	// Copy the contents of the file to hash.
	u.hash.Reset()
	dst := u.hash.(io.Writer)

	name := f.FileHeader.Name
	existing, ok := u.contents[name]
	if !ok {
		// This file name is not in the output yet, so also copy the
		// contents to a new entry there.
		var o io.Writer
		o, err = u.w.CreateHeader(&f.FileHeader)
		if err != nil {
			return exit.Unavailable, fmt.Errorf("failed to create file: %v", err)
		}
		dst = io.MultiWriter(dst, o)
	}

	if _, err = io.CopyBuffer(dst, src, u.block); err != nil {
		return exit.Unavailable, fmt.Errorf("failed to copy contents of file: %v", err)
	}

	// Check or store the hash of the file contents.
	// XXX: Maybe we should also track some metadata, e.g., IsDir.
	u.sum = u.hash.Sum(u.sum[:0])
	if ok {
		// Compare hash to previous value.
		if !bytes.Equal(u.sum, existing) {
			return exit.DataErr, fmt.Errorf("file content mismatch: "+
				"hash is %x, previous was %x", u.sum, existing)
		}
	} else {
		// Store the new hash value.
		stored := make([]byte, len(u.sum))
		copy(stored, u.sum)
		u.contents[name] = stored

		if !*qp {
			u.out.hide()
			fmt.Printf("added %q file %q\n", path, name)
			u.out.show()
		}
	}
	return
}
