package main

import "testing"

func TestZIPVersion(t *testing.T) {
	const comment = `Version: start of text

Some commentary about the archive.

Version: mid-text

Some more commentary about the archive.

Version:    left-trim and Unicode ✔️
version: lowercase
Version:no space
Version: end of text`
	const expected = `["start of text","mid-text","left-trim and Unicode ✔️","end of text"]`

	version, err := containerVersion(zipContainer{comment: comment})
	if err != nil {
		t.Fatal(err)
	}
	if version != expected {
		t.Errorf("unexpected ZIP version: got %s, want %s", version, expected)
	}
}
