package main

import (
	"archive/zip"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ivxv.ee/command/exit"
)

func TestUnion(t *testing.T) {
	tests := []struct {
		input    []string // paths to input archives
		expected string   // path to archive containing expected output
		err      string   // expected error
	}{
		{[]string{"foo.zip"}, "foo.zip", ""},
		{[]string{"foo.zip", "foo.zip"}, "foo.zip", ""},
		{[]string{"foo.zip", "bar.zip"}, "foobar.zip", ""},
		{[]string{"foo.zip", "foobar.zip"}, "foobar.zip", ""},
		{[]string{"foobar.zip", "bar.zip"}, "foobar.zip", ""},

		{[]string{"dir.zip"}, "dir.zip", ""},
		{[]string{"dirfoo.zip", "bar.zip"}, "dirfoobar.zip", ""},

		{[]string{"foo.zip", "foo2.zip"}, "mismatch",
			`failed to add "testdata/foo2.zip": failed to copy file "foo": ` +
				"file content mismatch: hash is " +
				"7d865e959b2466918c9863afca942d0fb89d7c9ac0c99bafc3749504ded97730, " +
				"previous was " +
				"b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"},
	}

	*qp = true // Run tests in quiet mode.
	for _, test := range tests {
		t.Run(strings.Join(test.input, " "), func(t *testing.T) {
			output := filepath.Join("testdata", test.expected+".tmp")
			defer func() {
				if err := os.Remove(output); err != nil {
					t.Error("failed to remove output:", err)
				}
			}()

			var input []string
			for _, in := range test.input {
				input = append(input, filepath.Join("testdata", in))
			}

			code, err := zipunion(output, input)
			if err != nil {
				if code == exit.OK {
					t.Error("exit code OK with non-nil error")
				}
				if len(test.err) == 0 {
					t.Fatal("union error:", err)
				}
				if err.Error() != test.err {
					t.Fatalf("unexpected error: got %q, want %q",
						err, test.err)
				}
				return
			}
			if code != exit.OK {
				t.Error("exit code", code, "with nil error")
			}
			if len(test.err) > 0 {
				t.Fatalf("expected error %q, got nil", test.err)
			}

			zipequal(t, filepath.Join("testdata", test.expected), output)
		})
	}
}

func zipequal(t *testing.T, expected, result string) {
	e, err := zip.OpenReader(expected)
	if err != nil {
		t.Fatal("open expected error:", err)
	}
	defer e.Close()

	r, err := zip.OpenReader(result)
	if err != nil {
		t.Fatal("open result error:", err)
	}
	defer r.Close()

	if len(e.File) != len(r.File) {
		for _, f := range r.File {
			t.Log(f.FileHeader.Name)
		}
		t.Fatal("expected", len(e.File), "file(s), got", len(r.File))
	}
	for i, ef := range e.File {
		rf := r.File[i]
		if ef.FileHeader.Name != rf.FileHeader.Name {
			t.Fatalf("expected %q, got %q",
				ef.FileHeader.Name, rf.FileHeader.Name)
		}

		ep, err := ef.Open()
		if err != nil {
			t.Fatal("open expected file error:", err)
		}
		defer ep.Close()
		ec, err := ioutil.ReadAll(ep)
		if err != nil {
			t.Fatal("read expected file error:", err)
		}

		rp, err := rf.Open()
		if err != nil {
			t.Fatal("open result file error:", err)
		}
		defer rp.Close()
		rc, err := ioutil.ReadAll(rp)
		if err != nil {
			t.Fatal("read result file error:", err)
		}

		if !bytes.Equal(ec, rc) {
			t.Fatalf("%q content mismatch: expected %q, got %q",
				ef.FileHeader.Name, ec, rc)
		}
	}
}
