package bdoc

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"testing"

	"ivxv.ee/errors"
)

func TestASiCEMagic(t *testing.T) {
	tests := []struct {
		name   string
		header *zip.FileHeader
		data   []byte
		err    error
	}{
		{"short", nil, []byte("too short"), new(ReadMagicFileHeaderError)},

		{"non-ZIP", nil, make([]byte, 30+len(magic)+len(mimetype)), new(NoZIPMagicError)},

		{"magic length", &zip.FileHeader{Name: "short"}, nil, new(MagicLengthError)},

		{"magic", &zip.FileHeader{Name: "MIMETYPE"}, nil, new(NotASiCMagicError)},

		{"compressed", &zip.FileHeader{
			Name:   magic,
			Method: zip.Deflate,
		}, nil, new(CompressedMIMETypeError)},

		{"extra", &zip.FileHeader{
			Name:  magic,
			Extra: make([]byte, 10),
		}, nil, new(ExtraFieldError)},

		{"MIME type length", &zip.FileHeader{
			Name:           magic,
			CompressedSize: 20,
		}, make([]byte, 20), new(MIMETypeLengthError)},

		{"MIME type", &zip.FileHeader{
			Name:           magic,
			CompressedSize: uint32(len(mimetype)),
		}, make([]byte, len(mimetype)), new(NotASiCEMIMETypeError)},

		{"OK", &zip.FileHeader{
			Name:           magic,
			CompressedSize: uint32(len(mimetype)),
		}, []byte(mimetype), nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer

			if test.header != nil {
				z := zip.NewWriter(&buf)
				w, err := z.CreateHeader(test.header)
				if err != nil {
					t.Fatal("failed to create file header:", err)
				}
				if _, err = w.Write(test.data); err != nil {
					t.Fatal("failed to write file data:", err)
				}
				if err = z.Close(); err != nil {
					t.Fatal("failed to close container:", err)
				}

				// Since the zip package uses data descriptors,
				// we need to write the compressed size
				// ourself.
				binary.LittleEndian.PutUint32(buf.Bytes()[18:22],
					test.header.CompressedSize)
			} else {
				buf.Write(test.data)
			}

			err := asiceMagic(bytes.NewReader(buf.Bytes()))
			if err != test.err && errors.CausedBy(err, test.err) == nil {
				t.Errorf("expected error %T, got %v", test.err, err)
			}
		})
	}
}

func TestReadManifest(t *testing.T) {
	tests := []struct {
		name     string
		manifest string
		files    map[string]*asiceFile
		err      error
	}{
		{"XML", "not xml", nil, new(ManifestXMLError)},

		{"duplicate", `<manifest:manifest
				xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0">
				<manifest:file-entry
					manifest:full-path="foo"
					manifest:media-type="bar"/>
				<manifest:file-entry
					manifest:full-path="foo"
					manifest:media-type="bar"/>
			</manifest:manifest>`,
			map[string]*asiceFile{"foo": {name: "foo"}},
			new(DuplicateManifestEntryError)},

		{"root", `<manifest:manifest
				xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0">
				<manifest:file-entry
					manifest:full-path="/"
					manifest:media-type="foobar"/>
			</manifest:manifest>`, nil, new(ManifestRootMIMETypeError)},

		{"extra", `<manifest:manifest
				xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0">
				<manifest:file-entry
					manifest:full-path="foo"
					manifest:media-type="bar"/>
			</manifest:manifest>`, nil, new(ExtraManifestEntryError)},

		{"signature", `<manifest:manifest
				xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0">
				<manifest:file-entry
					manifest:full-path="META-INF/signatures0.xml"
					manifest:media-type="foobar"/>
			</manifest:manifest>`,
			map[string]*asiceFile{"META-INF/signatures0.xml": {
				name: "META-INF/signatures0.xml"}},
			new(SignatureInManifestError)},

		{"missing", `<manifest:manifest
				xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0"/>`,
			map[string]*asiceFile{"foo": {name: "foo"}},
			new(MissingManifestEntryError)},

		{"OK", `<manifest:manifest
				xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0">
				<manifest:file-entry
					manifest:full-path="/"
					manifest:media-type="application/vnd.etsi.asic-e+zip"/>
				<manifest:file-entry
					manifest:full-path="foo"
					manifest:media-type="bar"/>
			</manifest:manifest>`,
			map[string]*asiceFile{"foo": {name: "foo"}},
			nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := readManifest([]byte(test.manifest), test.files)
			if err != test.err && errors.CausedBy(err, test.err) == nil {
				t.Errorf("expected error %T, got %v", test.err, err)
			}
		})
	}
}
