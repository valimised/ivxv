package zip

import (
	"archive/zip"
	"bytes"
	"crypto/subtle"
	"fmt"
	"io"
	"os"
	"testing"
)

// endOfCentralDirectory locates the end of [end of central directory record] segment in a zip stream r.
// Input files represent a collection of [local file header 1] [encryption header 1] [file data 1] ...
// [local file header n] [encryption header n] [file data n] segments of a zip stream.
func endOfCentralDirectory(r io.ReaderAt, files []*zip.File) (offset int64, err error) {
	var off int64 // offset from beginning of [local file header 1] till [file data n] end

	for i := range files {
		if off, err = files[i].DataOffset(); err != nil { // offset [local file header] [encryption header]
			return offset, fmt.Errorf("failed to offset [local file header] [encryption header]: %v", err)
		}

		off += int64(files[i].CompressedSize64) // offset [local file header] [encryption header] [file data]

		if offset < off { // filter the largest offset
			offset = off
		}
	}

	// [data descriptor n]
	if offset, err = readDataDescriptorN(r, offset); err != nil {
		return
	}

	// [archive decryption header]
	if offset, err = readArchiveDecryptionHeader(r, offset); err != nil {
		return
	}

	// [archive extra data record]
	if offset, err = readArchiveExtraDataRecord(r, offset); err != nil {
		return
	}

	// [central directory header 1] ... [central directory header n]
	if offset, err = readCentralDirectoryHeader(r, len(files), offset); err != nil {
		return
	}

	// [zip64 end of central directory record]
	if offset, err = readZip64EndOfCentralDirectoryRecord(r, offset); err != nil {
		return
	}

	// [zip64 end of central directory locator]
	if offset, err = readZip64EndOfCentralDirectoryLocator(r, offset); err != nil {
		return
	}

	// [end of central directory record]
	if offset, err = readEndOfCentralDirectoryRecord(r, offset); err != nil {
		return
	}

	return
}

// readDataDescriptorN possibly reads [data descriptor n] segment, where off must point at the beginning of
// the segment in a r stream. If the segment doesn't present in a stream at the location off, the returned value
// is the same as off.
func readDataDescriptorN(r io.ReaderAt, off int64) (int64, error) {
	sig := [4]byte{}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to possibly read data descriptor signature: %v", err)
	}

	switch sig {
	case dataDescriptorSignature:
		off += 4 // offset data descriptor signature
	case centralFileHeaderSignature: // non-empty zip/zip64, that doesn't have [data descriptor n]
		return off, nil
	case endOfCentralDirSignature: // empty zip/zip64, that doesn't have [data descriptor n]
		return off, nil
	default:
		// Data descriptor crc-32 or we have read 4 bytes of [archive decryption header] that doesn't have
		// [data descriptor n]
	}

	off += 4 + 4 + 4 // offset 12 bytes, which is the size of [data descriptor n]

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to possibly read [data descriptor n]: %v", err)
	}

	switch sig {
	case archiveExtraDataSignature:
		// [data descriptor n] doesn't present and we have offset 12 bytes of [archive decryption header], so
		// set offset back to the beginning
		off -= traditionalPKWAREEncryption

		return off, nil
	case centralFileHeaderSignature: // we have offset 12 bytes of [data description n]
		return off, nil
	default:
		// We have offset 12 bytes of zip64 [data description n], there are 8 bytes more left to offset, or
		// we have offset 12 bytes of [data description n] and [archive decryption header] is the next.
		//
		// Let's first assume that it is zip64 [data descriptor n]
		off += 8 // (4+4+4)+8 == 4+8+8
	}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to possibly read zip64 [data descriptor n]: %v", err)
	}

	switch sig {
	case centralFileHeaderSignature: // we have offset zip64 [data description n]
		return off, nil
	default:
		// We have offset zip64 [data description n] and [archive decryption header] is the next, or
		// we have read 8 bytes of [archive decryption header] and there are left 4 bytes more to reach
		// [archive extra data record].
		//
		// Let's offset 4 bytes to possibly reach [archive extra data record]
		off += 4
	}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to possibly read 4 bytes after zip64 [data descriptor n]: %v", err)
	}

	switch sig {
	case archiveExtraDataSignature:
		// We have reached [archive extra data record], so set offset back to the beginning of
		// [archive decryption header]
		off -= traditionalPKWAREEncryption
	default:
		// We have offset 4 bytes from the beginning of [archive decryption header], so set it back to the
		// beginning
		off -= 4
	}

	return off, nil
}

func TestEndOfCentralDirectory(t *testing.T) {
	zipFiles := map[string]int64{
		"testdata/zip64.zip":                            32768,
		"testdata/zip64-2.zip":                          32768,
		"testdata/subdir.zip":                           32768,
		"testdata/smartid1.bdoc":                        32768, // Smart-ID (https://www.smart-id.com/)
		"testdata/smartid2.bdoc":                        32768, // Smart-ID with trailing bytes
		"testdata/mid1.bdoc":                            32768, // Mobile-ID (https://www.id.ee/en/mobile-id/)
		"testdata/mid2.bdoc":                            32768, // Mobile-ID with trailing bytes
		"testdata/container1.asice":                     32768, // ID card (https://www.id.ee/en/rubriik/id-card-en/)
		"testdata/container2.asice":                     32768, // ID card with trailing bytes
		"testdata/container3_general_purpose_bit3.bdoc": 32768,
		"testdata/zipcrypto.zip":                        32768, // for some reason reading with archive/zip fails, but without this read works fine
		"testdata/unix.zip":                             32768,
		"testdata/utf8-7zip.zip":                        32768,
		"testdata/utf8-winrar.zip":                      32768,
		"testdata/deflate.zip":                          32768,
		"testdata/uncompressed.zip":                     32768,
		"testdata/deflate_trailing_bytes.zip":           32768,
		"testdata/uncompressed_trailing_bytes.zip":      32768,
		"testdata/empty.zip":                            32768,
		"testdata/aes256.zip":                           32768, // for some reason reading with archive/zip fails, but without this read works fine
		"testdata/symlink.zip":                          32768,
		"testdata/time-7zip.zip":                        32768,
		"testdata/time-osx.zip":                         32768,
		"testdata/time-win7.zip":                        32768,
		"testdata/time-winrar.zip":                      32768,
		"testdata/utf8-osx.zip":                         32768,
		"testdata/utf8-winzip.zip":                      32768,
		"testdata/deleted1.zip":                         32768,
		"testdata/42.zip":                               42838, // https://unforgettable.dk/. Fails with unsupported compression algorithm by archive/zip
		"testdata/zbsm.zip":                             42374, // https://www.bamsoftware.com/hacks/zipbomb/
	}

	// Tails are not declared in .ZIP file comment length section, so basically all these zips are malformed
	zipFilesWithTails := map[string]string{
		"testdata/container2.asice":                "This is trailing bytes",
		"testdata/mid2.bdoc":                       "Mobile-ID trailing bytes",
		"testdata/smartid2.bdoc":                   "Smart-ID trailing bytes",
		"testdata/deflate_trailing_bytes.zip":      "Unexpected trailing bytes on deflate compression",
		"testdata/uncompressed_trailing_bytes.zip": "Unexpected trailing bytes on uncompressed zip",
	}

	// NB! Never unzip these files!
	zipBombs := []string{
		"testdata/42.zip",
		"testdata/zbsm.zip",
	}

	unexpectedErrors := []string{
		"testdata/zipcrypto.zip",
		"testdata/aes256.zip",
	}

	for zipFile, zipFileSize := range zipFiles {
		b, err := os.ReadFile(zipFile)
		if err != nil {
			t.Fatal(err)
		}

		r := bytes.NewReader(b)
		zr, err := zip.NewReader(r, zipFileSize)
		if err != nil {
			t.Fatal(err)
		}

		opts := DecompressOptions{
			FilesLimit:    30,
			FileSizeLimit: 32768, // BDOC oriented value
			ZipSizeLimit:  zipFileSize,
		}

		offset, err := EndOfCentralDirectory(r, opts)
		if err != nil {
			var foundZipBomb bool

			for _, zipBomb := range zipBombs {
				if zipBomb == zipFile {
					foundZipBomb = true
				}
			}

			if foundZipBomb { // zip bombs are expected to fail
				continue
			}

			var foundUnexpectedErrors bool

			for _, unexpectedError := range unexpectedErrors {
				if unexpectedError == zipFile {
					foundUnexpectedErrors = true
				}
			}

			if foundUnexpectedErrors {
				continue
			}

			t.Fatal(err)
		}
		if offset <= 0 {
			t.Fatal("offset of even empty zip cannot be 0")
		}
		if offset != int64(len(b)) {
			tail, ok := zipFilesWithTails[zipFile]
			if !ok {
				t.Fatalf("zip content length mismatch, expected=%d, got=%d, and trailing text is: %s", offset, len(b), b[offset:])
			}

			if subtle.ConstantTimeCompare(b[offset:], []byte(tail)) != 1 {
				t.Fatalf("trailing bytes mismatch, expected=%s, got=%s", tail, b[offset:])
			}
		}

		// This is how caller may detect a malformed zip file
		//
		// 1. Find expected end of a zip file
		offset, err = EndOfCentralDirectory(r, opts)
		if err != nil {
			t.Fatal(err)
		}

		// 2. Try to read at least one byte more, and if succeeds, then there are appended bytes at the end of zip.
		// Only io.EOF is considered as a valid outcome
		p := [1]byte{}
		_, err = r.ReadAt(p[:], offset)
		if err != io.EOF {
			// Following 'if' block is only for testing purposes, caller must throw an error if non-EOF occurs
			if _, ok := zipFilesWithTails[zipFile]; ok {
				continue
			}

			t.Fatal("appended bytes at the end of zip")
		}

		// Compare that outcome of test is the same as real
		offset2, err := endOfCentralDirectory(r, zr.File)
		if err != nil {
			t.Fatal(err)
		}

		if offset != offset2 {
			t.Fatalf("EndOfCentralDirectory offset=%d, test-endOfCentralDirectory offset=%d", offset, offset2)
		}
	}
}

func BenchmarkEndOfCentralDirectory(b *testing.B) {
	f, err := os.ReadFile("testdata/smartid1.bdoc")
	if err != nil {
		b.Fatal(err)
	}

	r := bytes.NewReader(f)

	opts := DecompressOptions{
		FilesLimit:    30,
		FileSizeLimit: 32768,
		ZipSizeLimit:  32768,
	}

	var offset int64

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		offset, err = EndOfCentralDirectory(r, opts)
	}
	b.StopTimer()
	// 98358 ns/op	49350 B/op	68 allocs/op

	fmt.Fprint(io.Discard, offset, err)
}

func TestEndOfCentralDirectoryRegression(t *testing.T) {
	expected := 68.0

	f, err := os.ReadFile("testdata/smartid1.bdoc")
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(f)

	opts := DecompressOptions{
		FilesLimit:    30,
		FileSizeLimit: 32768,
		ZipSizeLimit:  32768,
	}

	var offset int64

	if allocs := testing.AllocsPerRun(10, func() {
		offset, err = EndOfCentralDirectory(r, opts)
	}); allocs > expected {
		t.Fatalf("EndOfCentralDirectory now requires %0.f heap allocations, while before required %0.f", allocs, expected)
	}

	fmt.Fprint(io.Discard, offset, err)
}

func BenchmarkClassicalZipBombEndOfCentralDirectoryRegression(b *testing.B) {
	f, err := os.ReadFile("testdata/42.zip")
	if err != nil {
		b.Fatal(err)
	}

	r := bytes.NewReader(f)

	opts := DecompressOptions{
		FilesLimit:    500,
		FileSizeLimit: 32768,
		ZipSizeLimit:  42838,
	}

	var offset int64

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		offset, err = EndOfCentralDirectory(r, opts)
	}
	b.StopTimer()
	// 7491 ns/op	19031 B/op	78 allocs/op

	fmt.Fprint(io.Discard, offset, err)
}

func TestClassicalZipBombEndOfCentralDirectoryRegression(t *testing.T) {
	expected := 78.0

	f, err := os.ReadFile("testdata/42.zip")
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(f)

	opts := DecompressOptions{
		FilesLimit:    500,
		FileSizeLimit: 32768,
		ZipSizeLimit:  42838,
	}

	var offset int64

	if allocs := testing.AllocsPerRun(10, func() {
		offset, err = EndOfCentralDirectory(r, opts)
	}); allocs > expected {
		t.Fatalf("EndOfCentralDirectory for 42.zip now requires %0.f heap allocations, while before required %0.f", allocs, expected)
	}

	fmt.Fprint(io.Discard, offset, err)
}

func BenchmarkBetterZipBombEndOfCentralDirectoryRegression(b *testing.B) {
	f, err := os.ReadFile("testdata/zbsm.zip")
	if err != nil {
		b.Fatal(err)
	}

	r := bytes.NewReader(f)

	opts := DecompressOptions{
		FilesLimit:    500,
		FileSizeLimit: 32768,
		ZipSizeLimit:  42374,
	}

	var offset int64

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		offset, err = EndOfCentralDirectory(r, opts)
	}
	b.StopTimer()
	// 60901 ns/op	73653 B/op	979 allocs/op

	fmt.Fprint(io.Discard, offset, err)
}

func TestBetterZipBombEndOfCentralDirectoryRegression(t *testing.T) {
	expected := 979.0

	f, err := os.ReadFile("testdata/zbsm.zip")
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewReader(f)

	opts := DecompressOptions{
		FilesLimit:    500,
		FileSizeLimit: 32768,
		ZipSizeLimit:  42374,
	}

	var offset int64

	if allocs := testing.AllocsPerRun(10, func() {
		offset, err = EndOfCentralDirectory(r, opts)
	}); allocs > expected {
		t.Fatalf("EndOfCentralDirectory for zbsm.zip now requires %0.f heap allocations, while before required %0.f", allocs, expected)
	}

	fmt.Fprint(io.Discard, offset, err)
}
