package main

import (
	"bytes"
	"fmt"
	"go/build"
	"go/format"
	"text/template"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// header is added to the gen.tmpl template to identify files created by gen.
const header = "// Generated by ivxv.ee/cmd/gen. DO NOT EDIT!\n\n"

// funcMap contains the helper functions used in gen.tmpl.
var funcMap = template.FuncMap{
	"base": filepath.Base,
}

// Generate a Go source file which contains the template in gen.tmpl and parses
// it into the variable gentmpl. The variable name is given as an argument, so
// that it can easily be changed here, where it will also be used.
//go:generate ./gentmpl.sh gentmpl

// generate generates type declarations for the struct literals in pkg.
func generate(pkg *build.Package, literals []literal) (ok bool) {
	ok = true
	main, test := split(literals)
	if len(main) > 0 {
		ok = ok && genfile(pkg, *namep+".go", main)
	}
	if len(test) > 0 {
		ok = ok && genfile(pkg, *namep+"_test.go", test)
	}
	return
}

func genfile(pkg *build.Package, name string, literals []literal) bool {
	path := rel(filepath.Join(pkg.Dir, name))
	fp, err := open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return false
	}
	defer fp.Close()

	if *vp {
		fmt.Println("generating", path)
	}

	data := struct {
		Package  *build.Package
		Literals []literal
	}{
		pkg,
		literals,
	}

	// Write to both a memory buffer and the result file: this way have the
	// result in memory for formatting, but also any Go formatting errors
	// will refer to an actually existing file, greatly reducing debugging
	// complexity.
	var buf bytes.Buffer
	tee := io.MultiWriter(&buf, fp)
	if err = gentmpl.Execute(tee, data); err != nil {
		fmt.Fprintln(os.Stderr, "error: template execution:", err)
		return false
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: format generated code: %s:%v\n", path, err)
		return false
	}

	// Rewind, now writing formatted code.
	if _, err = fp.Seek(0, io.SeekStart); err != nil {
		fmt.Fprintln(os.Stderr, "error: rewind", path, "for formatting:", err)
		return false
	}

	if _, err = fp.Write(formatted); err != nil {
		fmt.Fprintf(os.Stderr, "error: write generated code %s: %v\n", path, err)
		return false
	}

	// Truncate fp in case it already contained a larger file when opened
	// or if the formatted code is shorter that unformatted.
	if err = fp.Truncate(int64(len(formatted))); err != nil {
		fmt.Fprintf(os.Stderr, "truncate %s to %d: %v\n", path, len(formatted), err)
		return false
	}

	return true
}

// split partitions the literals into two sets: ones from regular Go source
// files and ones from Go test files.
func split(literals []literal) (main []literal, test []literal) {
	for _, l := range literals {
		if strings.HasSuffix(l.Pos.Filename, "_test.go") {
			test = append(test, l)
		} else {
			main = append(main, l)
		}
	}
	return
}

// open attempts to create the path. If a file already exists there, then it
// will check if it is safe to overwrite.
func open(path string) (fp *os.File, err error) {
	// Attempt to open the path for reading and writing.
	var created bool
loop:
	fp, err = os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		// If it does not exist, attempt to create it, otherwise fail.
		if !os.IsNotExist(err) {
			return
		}

		fp, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			// If it now does exist, retry from beginning.
			if os.IsExist(err) {
				goto loop
			}
			return
		}
		created = true
	}

	// If we did not create the file, ensure that it contains our header,
	// so it is safe to overwrite, and rewind.
	if !created {
		hbuf := make([]byte, len(header))
		_, err := io.ReadFull(fp, hbuf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, fmt.Errorf("read %s header: %v", path, err)
		}
		if string(hbuf) != header {
			return nil, fmt.Errorf("%s already exists and is not generated, "+
				"not overwriting", path)
		}
		if _, err = fp.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("rewind %s: %v", path, err)
		}
	}

	return
}
