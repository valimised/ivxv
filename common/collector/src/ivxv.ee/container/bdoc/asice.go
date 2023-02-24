package bdoc

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"io"
	"regexp"
	"strings"

	"ivxv.ee/safereader"
)

// metainf is the name of the folder containing meta information.
const metainf = "META-INF/"

// signatureRE is used to match names of signature files.
var signatureRE = regexp.MustCompile(`^` + metainf + `[^/]*signatures[^/]*\.xml$`)

// asiceFile is a file read from an ASiC-E container.
type asiceFile struct {
	name     string
	mimetype string // Only used for non-signatures.
	data     *bytes.Buffer
}

// signature returns true if the file is a signature.
func (f asiceFile) signature() bool {
	return signatureRE.MatchString(f.name)
}

// close releases any resources held by asiceFile.
func (f *asiceFile) close() {
	release(f.data)
	f.data = nil
}

// openASiCE opens an ASiC-E XAdES container and decompresses all the files in
// it. It checks that "mimetype" and "META-INF/manifest.xml" comply with all
// requirements, so these will not be returned.
//
// zipLimit is the maximum allowed length of the Zip archive and fileLimit is
// the maximum allowed decompressed length of a file in the archive.
//
// If readSigs is false, then signature files will be skipped: used if only the
// container contents are requested.
//
// All returned asiceFiles should be closed after use.
func openASiCE(r io.Reader, zipLimit, fileLimit int64, readSigs bool) (
	files map[string]*asiceFile, err error) {

	// Convert the stream to io.ReaderAt needed for zip.NewReader.
	rat, size, err := toReaderAt(r, zipLimit)
	if err != nil {
		return nil, ToReaderAtError{Err: err}
	}
	defer rat.close()

	// Fail fast: Check the ASiC-E magic number before reading the ZIP.
	if err = asiceMagic(rat); err != nil {
		return nil, NotASiCEError{Err: err}
	}

	// Read the ZIP directory and parse file headers.
	rzip, err := zip.NewReader(rat, size)
	if err != nil {
		return nil, ZIPReaderError{Err: err}
	}

	// Check the files and decompress data.
	files = make(map[string]*asiceFile)
	// Pass files as explicit argument so that it gets evaluated now and we
	// can later safely return nil. Do not do the same for err, because we
	// want to know its value on return, not now.
	defer func(files map[string]*asiceFile) {
		// Close all files on error.
		if err != nil {
			for _, file := range files {
				file.close()
			}
		}
	}(files)

	seen := make(map[string]struct{})
	var manifest *asiceFile
	var signatures int
	for i, file := range rzip.File {

		// Check that we have not seen this file yet.
		if _, ok := seen[file.Name]; ok {
			return nil, DuplicateFileNameError{FileName: file.Name}
		}
		seen[file.Name] = struct{}{}

		// asiceMagic checked that the first file in the archive was a
		// correct "mimetype" using the local file header: now check
		// that it is also first in the central directory and that the
		// directory header matches.
		if i == 0 {
			if err = asiceMagicCentral(file); err != nil {
				return nil, ASiCEMagicCentralError{Err: err}
			}
			continue
		}

		// Check that the file name is allowed.
		switch {
		case file.Name == metainf:
			continue
		case file.Name == metainf+"manifest.xml":
			// manifest.xml is not returned, so decompress here and
			// defer the recover.
			if manifest, err = decompress(file, fileLimit); err != nil {
				return nil, DecompressManifestError{Err: err}
			}
			defer manifest.close()
			continue
		case signatureRE.MatchString(file.Name):
			signatures++
			if !readSigs {
				continue // Skip signatures if not requested.
			}
		// No more META-INF/ files allowed below this case.
		case strings.HasPrefix(file.Name, metainf):
			return nil, UnknownMetaInfoFile{FileName: file.Name}

		case strings.Count(file.Name, "/") > 0:
			return nil, FileInSubfolderError{FileName: file.Name}
		}

		var asice *asiceFile
		if asice, err = decompress(file, fileLimit); err != nil {
			return nil, DecompressFileError{
				FileName: file.Name,
				Err:      err,
			}
		}
		files[file.Name] = asice
	}

	// The manifest, at least one signature, and at least one data file
	// must be present. If they are, then we can be certain that "mimetype"
	// is too, because we require it to be the first file.
	if manifest == nil {
		return nil, MissingManifestError{}
	}
	if signatures == 0 {
		return nil, NoSignaturesError{}
	}
	datafiles := len(files)
	if readSigs {
		datafiles -= signatures
	}
	if datafiles == 0 {
		return nil, NoDataFilesError{}
	}

	// Check that the manifest has references for all non-signature files,
	// and only those files, and retrieve their MIME types from it.
	if err = readManifest(manifest.data.Bytes(), files); err != nil {
		return nil, ManifestError{Err: err}
	}
	return files, nil
}

const (
	// magic is the name of the magic number file used to recognize ASiC
	// containers.
	magic = "mimetype"

	// mimetype is the mimetype used for ASiC-E containers.
	mimetype = "application/vnd.etsi.asic-e+zip"
)

// asiceMagic checks the magic number "mimetype" file as described in clause A.1
// of the ASiC specification and that it specifies the ASiC-E MIME type.
func asiceMagic(rat io.ReaderAt) error {
	buf := make([]byte, 30+len(magic)+len(mimetype))
	if n, err := rat.ReadAt(buf, 0); err != nil {
		return ReadMagicFileHeaderError{Header: buf[:n], Err: err}
	}
	le := binary.LittleEndian

	if marker := string(buf[:4]); marker != "PK\x03\x04" {
		return NoZIPMagicError{Marker: marker}
	}

	if length := le.Uint16(buf[26:28]); int(length) != len(magic) {
		return MagicLengthError{Length: length}
	}
	if filename := string(buf[30 : 30+len(magic)]); filename != magic {
		return NotASiCMagicError{FileName: filename}
	}

	if method := le.Uint16(buf[8:10]); method != zip.Store {
		return CompressedMIMETypeError{Method: method}
	}
	if length := le.Uint16(buf[28:30]); length != 0 {
		return ExtraFieldError{Length: length}
	}

	if length := le.Uint32(buf[18:22]); int(length) != len(mimetype) {
		return MIMETypeLengthError{Length: length}
	}
	if mime := string(buf[30+len(magic):]); mime != mimetype {
		return NotASiCEMIMETypeError{MIMEType: mime}
	}
	return nil
}

// asiceMagicCentral checks that the ZIP central directory header contains the
// same information as the local header checked in asiceMagic.
func asiceMagicCentral(file *zip.File) error {
	// Check the file name.
	if file.Name != magic {
		return MIMETypeCentralNameError{Name: file.Name}
	}

	// Check if the central directory is referring to the file at the start
	// of the archive. We cannot check the header offset, but data offset
	// 30 + file name length is the same, since we have no extra data.
	offset, err := file.DataOffset()
	if err != nil {
		return GetMIMETypeOffsetError{Err: err}
	}
	if offset != int64(30+len(magic)) {
		return MIMETypeCentralOffsetError{Offset: offset}
	}

	// Check directory header compression, data size, and extra length --
	// file name length was already checked by comparing the file name.
	if file.Method != zip.Store {
		return MIMETypeCentralMethodError{Method: file.Method}
	}
	if file.CompressedSize != uint32(len(mimetype)) {
		return MIMETypeCentralSizeError{Size: file.CompressedSize}
	}
	if len(file.Extra) != 0 {
		return MIMETypeCentralExtraError{Extra: file.Extra}
	}
	return nil
}

// decompress decompresses the ZIP file into an asiceFile. Note that mimetype
// is not set for these files.
func decompress(file *zip.File, limit int64) (*asiceFile, error) {
	fd, err := file.Open()
	if err != nil {
		return nil, OpenZIPFileError{Err: err}
	}
	defer fd.Close()

	asice := &asiceFile{name: file.Name, data: buffer()}
	size, err := asice.data.ReadFrom(safereader.New(fd, limit))
	if err != nil {
		asice.close()
		return nil, ReadZIPFileError{Err: err}
	}

	if file.UncompressedSize64 != uint64(size) {
		asice.close()
		return nil, UncompressedZIPFileSizeError{
			Declared: file.UncompressedSize64,
			Actual:   size,
		}
	}
	return asice, nil
}

// Specified in section 17.7.2 of the Open Document Format for Office
// Applications (OpenDocument) v1.0 (page 687 of
// https://docs.oasis-open.org/office/v1.0/OpenDocument-v1.0-os.pdf).
type manifest struct {
	XMLElement  c14n `xmlx:"urn:oasis:names:tc:opendocument:xmlns:manifest:1.0 manifest"`
	FileEntries []fileEntry
}

// Specified in section 17.7.3 of the Open Document Format for Office
// Applications (OpenDocument) v1.0 (page 687 of
// https://docs.oasis-open.org/office/v1.0/OpenDocument-v1.0-os.pdf).
type fileEntry struct {
	XMLElement c14n   `xmlx:"urn:oasis:names:tc:opendocument:xmlns:manifest:1.0 file-entry"`
	FullPath   string `xmlx:"urn:oasis:names:tc:opendocument:xmlns:manifest:1.0 full-path,attr"`
	MediaType  string `xmlx:"urn:oasis:names:tc:opendocument:xmlns:manifest:1.0 media-type,attr"`
}

// readManifest parses the OpenDocument manifest and fills the MIME types for
// all non-signature files. The file entries in the manifest must match exactly
// the set of non-signature files plus the "/" element, which must have the
// same MIME type as specified in "mimetype".
//
// Although BDOC 2.1.2:2014 references OpenDocument v1.2, the official
// implementations actually only produce v1.0, so we only support that.
func readManifest(data []byte, files map[string]*asiceFile) error {
	var m manifest
	if err := parseXML(data, &m); err != nil {
		return ManifestXMLError{Data: data, Err: err}
	}

	seen := make(map[string]struct{})
	for _, entry := range m.FileEntries {
		// Check that we have not seen this entry yet.
		if _, ok := seen[entry.FullPath]; ok {
			return DuplicateManifestEntryError{Path: entry.FullPath}
		}
		seen[entry.FullPath] = struct{}{}

		// The OpenDocument should contain a root element, but it is
		// not mandatory.
		if entry.FullPath == "/" {
			if entry.MediaType != mimetype {
				return ManifestRootMIMETypeError{MIMEType: entry.MediaType}
			}
			continue
		}

		file, ok := files[entry.FullPath]
		if !ok {
			return ExtraManifestEntryError{Path: entry.FullPath}
		}
		if file.signature() {
			return SignatureInManifestError{Path: entry.FullPath}
		}
		file.mimetype = entry.MediaType
	}

	for _, file := range files {
		if !file.signature() && len(file.mimetype) == 0 {
			return MissingManifestEntryError{Path: file.name}
		}
	}
	return nil
}
