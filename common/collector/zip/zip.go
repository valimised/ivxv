// Package zip implements .ZIP File Format Specification, version 6.3.10 as defined in
// https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT.
package zip

import (
	"archive/zip"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	// 6.0 Traditional PKWARE Encryption.
	traditionalPKWAREEncryption = 12

	// 4.4.4 general purpose bit flag: (2 bytes). Bit 0.
	encryptedFlag = 0b00000001

	// 4.4.4 general purpose bit flag: (2 bytes). Bit 3.
	dataDescriptorFlag = 0b00001000

	// _maxFilesPrealloc is a limit for files count in a zip. If there are more files in a zip, then
	// runtime allocation happens for each subsequent file in a list.
	_maxFilesPrealloc = 10

	// _zipFileBuffer is a size in bytes of a buffer to read a zip file content.
	_zipFileBuffer = 8192 // 8 KiB
)

var (
	// APPENDIX E - AE-x encryption marker: https://www.winzip.com/en/support/aes-encryption/. Little-endian value
	// of 0x9901.
	aexEncryption = [2]byte{1, 153}

	// 4.4.5 compression method: (2 bytes). The file is stored (no compression).
	noCompression = [2]byte{0, 0}

	// 4.3.7 Local file header. Little-endian of 0x04034b50.
	localFileHeaderSignature = [4]byte{80, 75, 3, 4}

	// 8.5.5 The signature value 0x08074b50 is also used by some ZIP implementations as a marker for the
	// Data Descriptor record. Little-endian of 0x08074b50.
	dataDescriptorSignature = [4]byte{80, 75, 7, 8}

	// 4.3.11 Archive extra data record. Little-endian of 0x08064b50.
	archiveExtraDataSignature = [4]byte{80, 75, 6, 8}

	// 4.3.12 Central directory structure. Little-endian value of 0x02014b50.
	centralFileHeaderSignature = [4]byte{80, 75, 1, 2}

	// 4.3.13 Digital signature. Little-endian value of 0x05054b50.
	headerSignature = [4]byte{80, 75, 5, 5}

	// 4.3.14 Zip64 end of central directory record. Little-endian value of 0x06064b50.
	zip64EndOfCentralDirSignature = [4]byte{80, 75, 6, 6}

	// 4.3.15 Zip64 end of central directory locator. Little-endian value of 0x07064b50.
	zip64EndOfCentralDirLocatorSignature = [4]byte{80, 75, 6, 7}

	// 4.3.16 End of central directory record. Little-endian value of 0x06054b50.
	endOfCentralDirSignature = [4]byte{80, 75, 5, 6}
)

var (
	ErrNextSegment      = errors.New("attempting to read bytes of a next segment")
	ErrFileCountLimit   = errors.New("there are more files in a zip than expected")
	ErrUncompressedSize = errors.New("zip file uncompressed size exceeds a limit")
	ErrFileSizeLimit    = errors.New("zip file content size exceeds a limit")
)

// file is a [file data n] segment of a zip.
type file struct {
	// name is a filename.
	name string

	// compressedSize is a compressed size of a file in a zip.
	compressedSize uint64

	// uncompressedSize is an uncompressed size of a file in a zip.
	uncompressedSize uint64
}

// bufferedDecompress decompresses a file f, using a buffer of a size buf.
func bufferedDecompress(f *zip.File, buf []byte, fileSizeLimit int) (file, error) {
	rc, err := f.Open()
	if err != nil {
		return file{}, fmt.Errorf("failed to read a zip file: %v", err)
	}
	defer rc.Close()

	// Standard, but unreliable approach to catch a file that exceeds a size limit
	if f.UncompressedSize64 > uint64(fileSizeLimit) { //nolint:gosec
		return file{}, ErrUncompressedSize
	}

	var fileSize int

loop:
	// https://wiki.sei.cmu.edu/confluence/display/java/IDS04-J.+Safely+extract+files+from+ZipInputStream
	for { // safe endless loop, since we don't let to read more than limit bytes
		n, err := rc.Read(buf)
		switch err {
		case io.EOF:
			// end of file
			break loop
		case nil:
			// no errors
		default:
			return file{}, fmt.Errorf("unexpected error during zip file reading: %v", err)
		}

		fileSize += n

		// Reliable way to catch a file that exceeds a size limit
		if fileSize > fileSizeLimit {
			return file{}, ErrFileSizeLimit
		}
	}

	return file{
		name:             f.Name,
		compressedSize:   f.CompressedSize64,
		uncompressedSize: f.UncompressedSize64,
	}, nil
}

// DecompressOptions contain parameters for decompressing a zip file.
type DecompressOptions struct {
	// FilesLimit represents max files count in a zip.
	FilesLimit int

	// FileSizeLimit represents max size of a single zipped file in bytes.
	FileSizeLimit int

	// ZipSizeLimit represents max size of a zip in bytes.
	ZipSizeLimit int64
}

// EndOfCentralDirectory locates the end of [end of central directory record] segment in a zip stream r, which is
// assumed to have the given size in bytes. For zip files decompression DecompressOptions are used.
func EndOfCentralDirectory(r io.ReaderAt, opts DecompressOptions) (offset int64, err error) {
	// Validate r according to zip specification as much as archive/zip allows
	stream, err := zip.NewReader(r, opts.ZipSizeLimit)
	if err != nil {
		return offset, fmt.Errorf("failed to read zip stream: %v", err)
	}

	// We have to locate all zip files beforehand, using archive/zip, in order to obtain compressed/uncompressed
	// sizes of each file. The edge case that we are delegating to archive/zip is the situation where general
	// purpose flag bit 3 is set, and therefore compressed/uncompressed sizes may not appear in
	// [local file header n], but rather in [data descriptor n], which, unfortunately, appears after
	// [local file header n] and therefore makes it hard to parse these values in a sequential stream
	files := make([]file, 0, _maxFilesPrealloc) // no realloc if enough cap

	// Read uncompressed zip file content step-by-step by a buffer size
	buf := make([]byte, _zipFileBuffer)

	for i := range stream.File {
		if i > opts.FilesLimit {
			return offset, ErrFileCountLimit
		}

		f, err := bufferedDecompress(stream.File[i], buf, opts.FileSizeLimit)
		if err != nil {
			return offset, err
		}

		files = append(files, f)
	}

	var fileOffest int64 // offset from [local file header n] ... [file data n]
	var countFiles int

loop2:
	for { // infinite loop is dangerous, but we have validated zip with NewReader and the size, so it is safe here
		countFiles++
		fileOffest, err = readFile(r, fileOffest, files)

		if offset < fileOffest { // filter the largest offset
			offset = fileOffest
		}

		switch err {
		case nil:
			// More [local file header n] ... [data descriptor n] segments left
		case ErrNextSegment:
			// No [local file header n] ... [data descriptor n] segments left
			break loop2
		default:
			return offset, fmt.Errorf("failed to offset [local file header n] ... [data descriptor n]: %v", err) //nolint:lll
		}
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
	if offset, err = readCentralDirectoryHeader(r, countFiles, offset); err != nil {
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

// readFile returns the offset of [local file header n] ... [data descriptor n]. Files are [file data n]
// segments of a stream r, which must be parsed beforehand.
func readFile(r io.ReaderAt, off int64, files []file) (int64, error) { //nolint:gocyclo
	sig := [4]byte{}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to read local file header signature: %v", err)
	}

	switch sig {
	case localFileHeaderSignature:
		// There is a file in a zip
		off += 4 // offset local file header signature
	case endOfCentralDirSignature:
		// Empty zip
		return off, ErrNextSegment
	default:
		return off, errors.New("expected [local file header n]")
	}

	off += 2 // offset version needed to extract

	if _, err := r.ReadAt(sig[:2], off); err != nil {
		return off, fmt.Errorf("failed to read general purpose bit flag: %v", err)
	}
	off += 2 // offset general purpose bit flag

	// 4.4.4 general purpose bit flag: (2 bytes). Bit 0: If set, indicates that the file is encrypted.
	isEncrypted := sig[0]&encryptedFlag == encryptedFlag

	// 4.3.9.1 This descriptor MUST exist if bit 3 of the general purpose bit flag is set.
	withDataDescriptor := sig[0]&dataDescriptorFlag == dataDescriptorFlag

	if _, err := r.ReadAt(sig[:2], off); err != nil {
		return off, fmt.Errorf("failed to read compression method: %v", err)
	}
	off += 2 // offset compression method

	isCompressed := subtle.ConstantTimeCompare(sig[:2], noCompression[:]) != 1

	if isCompressed {
		off += 8 // offset last mod file time ... crc-32
	} else {
		off += 12 // offset last mod file time ... compressed size
	}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to read compression size: %v", err)
	}

	size := int64(binary.LittleEndian.Uint32(sig[:])) // either compressed or uncompressed size

	if isCompressed {
		off += 8 // offset compressed size ... uncompressed size
	} else {
		off += 4 // offset uncompressed size
	}

	if _, err := r.ReadAt(sig[:2], off); err != nil {
		return off, fmt.Errorf("failed to read file name length: %v", err)
	}
	off += 2 // offset file name length

	name := int64(binary.LittleEndian.Uint16(sig[:2]))

	if _, err := r.ReadAt(sig[:2], off); err != nil {
		return off, fmt.Errorf("failed to read extra field length: %v", err)
	}
	off += 2 // offset extra field length

	extra := int64(binary.LittleEndian.Uint16(sig[:2]))

	fileName := make([]byte, name)
	if _, err := r.ReadAt(fileName, off); err != nil {
		return off, fmt.Errorf("failed to read file name (variable size): %v", err)
	}
	off += name // offset file name (variable size)

	// Only use provided files if size doesn't present in [local file header n], which actually means that
	// size is stored in [data descriptor n], which we cannot be parsed yet (sequential stream)
	if size <= 0 {
		entryID := -1

		for i := range files {
			if files[i].name == string(fileName) {
				entryID = i
			}
		}

		if entryID == -1 {
			return off, fmt.Errorf("%s not found in a zip stream", fileName)
		}

		size = int64(files[entryID].uncompressedSize) //nolint:gosec
		if isCompressed {
			size = int64(files[entryID].compressedSize) //nolint:gosec
		}
	}

	extraField := make([]byte, extra)
	if _, err := r.ReadAt(extraField, off); err != nil {
		return off, fmt.Errorf("failed to read extra field (variable size): %v", err)
	}

	var isPKWAREEncryption bool // traditional PKWARE Encryption

	// Gather encryption info from extra field
	if isEncrypted {
		// Header ID - 2 bytes
		switch {
		case subtle.ConstantTimeCompare(extraField[:2], aexEncryption[:]) == 1:
			// AE-x (e.g. AES) encryption marker, encryption info is stored within extra field,
			// not within [encryption header n]
		default:
			// Encryption info is stored withing [encryption header n]
			isPKWAREEncryption = true
		}
	}

	off += extra // offset extra field (variable size)

	// If the file is encrypted, the encryption header for the file SHOULD be placed after the local header and
	// before the file data (relevant for traditional PKWARE Encryption only)
	if isPKWAREEncryption {
		off += traditionalPKWAREEncryption // offset [encryption header n]
	}

	// After possibly [encryption header n] there must be [file data n]
	off += size // offset [file data n]

	if withDataDescriptor {
		if _, err := r.ReadAt(sig[:], off); err != nil {
			return off, fmt.Errorf("failed to read possibly data descriptor signature: %v", err)
		}

		if subtle.ConstantTimeCompare(sig[:], dataDescriptorSignature[:]) == 1 { // signature is optional
			off += 4 // offset data descriptor signature
		}

		off += 12 // offset data descriptor (4+4+4)

		if _, err := r.ReadAt(sig[:], off); err != nil {
			return off, fmt.Errorf("failed to read 4 bytes after [data descriptor n]: %v", err)
		}

		switch sig {
		case localFileHeaderSignature:
			// [data descriptor 1] [local file header 2], next file to be parsed
			return off, nil
		case centralFileHeaderSignature:
			// [data descriptor n] [central directory header 1], all files are parsed, start parsing
			// central dir
			return off, ErrNextSegment
		default:
			// [data descriptor n] is either zip64, or we have encountered [archive decryption header]
			// segment.
			//
			// Assume it is zip64, so possibly offset zip64 data descriptor
			off += 8
		}

		if _, err := r.ReadAt(sig[:], off); err != nil {
			return off, fmt.Errorf("failed to read 4 bytes after zip64 [data descriptor n]: %v", err)
		}

		switch sig {
		case localFileHeaderSignature:
			// zip64 [data descriptor 1] [local file header 2], so next file to be parsed
			return off, nil
		case centralFileHeaderSignature:
			// zip64 [data descriptor n] [central directory header 1], all files are parsed,
			// start parsing central dir
			return off, ErrNextSegment
		default:
			// Either zip64 [data descriptor n] [archive decryption header] or we have already read
			// 8 bytes of [archive decryption header].
			//
			// Assume we have read 8 bytes of [archive decryption header], so read 4 more to reach
			// [archive extra data record]
			off += 4
		}

		if _, err := r.ReadAt(sig[:], off); err != nil {
			return 0, fmt.Errorf("failed to read 8 bytes after zip64 [data descriptor n]: %v", err)
		}

		switch sig {
		case archiveExtraDataSignature:
			// Reached [archive extra data record], so unread [archive decryption header] to reset stream
			// offset at the beginning of [archive decryption header]. There are no files left in a stream
			off -= traditionalPKWAREEncryption

			return off, ErrNextSegment
		default:
			// We have read 4 bytes of [archive decryption header], so unread it
			off -= 4
		}
	}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return 0, fmt.Errorf("failed to read 4 bytes after possibly [data descriptor n]: %v", err)
	}

	if subtle.ConstantTimeCompare(sig[:], centralFileHeaderSignature[:]) == 1 {
		// No more zip files
		return off, ErrNextSegment
	}

	return off, nil // read next zip file
}

// readArchiveDecryptionHeader possibly reads [archive decryption header] segment, where off must point at the
// beginning of the segment in a r stream. If the segment doesn't present in a stream at the location off, the
// returned value is the same as off. Only supports traditional PKWARE encryption header.
func readArchiveDecryptionHeader(r io.ReaderAt, off int64) (int64, error) {
	sig := [4]byte{}

	// Even though [archive decryption header] doesn't have a signature, we try to find out what these
	// next 4 bytes are
	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to possibly read 4 bytes of [arhive decryption header]: %v", err)
	}

	switch sig {
	case centralFileHeaderSignature:
		// Non-empty zip file without [archive decryption header]
		return off, nil
	case endOfCentralDirSignature:
		// Empty zip file without [archive decryption header]
		return off, nil
	default:
		// 6.1 Traditional PKWARE Encryption
		off += traditionalPKWAREEncryption
	}

	return off, nil
}

// readArchiveExtraDataRecord possibly reads [archive extra data record] segment, where off must point at the
// beginning of the segment in a r stream. If the segment doesn't present in a stream at the location off, the
// returned value is the same as off.
func readArchiveExtraDataRecord(r io.ReaderAt, off int64) (int64, error) {
	sig := [4]byte{}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to possibly read arhive extra data record signature: %v", err)
	}

	switch sig {
	case archiveExtraDataSignature:
		// [archive decryption header] is present
		off += 4 // offset archive extra data signature
	case centralFileHeaderSignature:
		// Non-empty zip file without [archive decryption header]
		return off, nil
	case endOfCentralDirSignature:
		// Empty zip file without [archive decryption header]
		return off, nil
	default:
		return off, errors.New("expected [archive extra data record]")
	}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to read extra field length: %v", err)
	}
	off += 4 // offset extra field length

	off += int64(binary.LittleEndian.Uint32(sig[:])) // offset extra field data (variable size)

	return off, nil
}

// readCentralDirectoryHeader possibly reads [central directory header 1] ... [central directory header n]
// segment, where off must point at the beginning of the segment in a r stream. If the segment doesn't present
// in a stream at the location off, the returned value is the same as off.
func readCentralDirectoryHeader(r io.ReaderAt, filesCount int, off int64) (int64, error) {
	sig := [4]byte{}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to possibly read central file header signature: %v", err)
	}

	switch sig {
	case centralFileHeaderSignature:
		// Do nothing as we offset central file header signature in a 'for range files' anyway
	case endOfCentralDirSignature:
		// Empty zip
		return off, nil
	default:
		return off, errors.New("expected [central directory header n]")
	}

	// All files within [local file header 1] ... [local file header n] must be present in
	// [central directory header 1] ... [central directory header n]
	for range filesCount {
		off += 4 + 24 // offset central file header signature ... uncompressed size

		if _, err := r.ReadAt(sig[:2], off); err != nil {
			return off, fmt.Errorf("failed to read file name length: %v", err)
		}
		off += 2 // offset file name length

		name := binary.LittleEndian.Uint16(sig[:2])

		if _, err := r.ReadAt(sig[:2], off); err != nil {
			return off, fmt.Errorf("failed to read extra field length: %v", err)
		}
		off += 2 // offset extra field length

		extra := binary.LittleEndian.Uint16(sig[:2])

		if _, err := r.ReadAt(sig[:2], off); err != nil {
			return off, fmt.Errorf("failed to read file comment length: %v", err)
		}
		off += 2 // offset file comment length

		com := binary.LittleEndian.Uint16(sig[:2])

		off += 12 // offset disk number start ... relative offset of local header

		off += int64(name + extra + com) // offset file name (variable size) ... file comment (variable size)
	}

	// 4.3.13 Digital signature, optional
	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to possibly read central directory digital signature: %v", err)
	}

	switch sig {
	case headerSignature:
		// Digital signature present
		off += 4 // offset header signature
	case zip64EndOfCentralDirSignature:
		// Non-empty zip file without a digital signature
		return off, nil
	case endOfCentralDirSignature:
		// Empty zip file without a digital signature
		return off, nil
	default:
		return off, errors.New("expected central directory header digital signature")
	}

	if _, err := r.ReadAt(sig[:2], off); err != nil {
		return off, fmt.Errorf("failed to read size of data: %v", err)
	}
	off += 2 // offset size of data

	off += int64(binary.LittleEndian.Uint16(sig[:2])) // offset signature data

	return off, nil
}

// readZip64EndOfCentralDirectoryRecord possibly reads [zip64 end of central directory record] segment, where
// off must point at the beginning of the segment in a r stream. If the segment doesn't present in a stream at
// the location off, the returned value is the same as off.
func readZip64EndOfCentralDirectoryRecord(r io.ReaderAt, off int64) (int64, error) {
	sig := [8]byte{} // allocate space suitable for size of zip64 end of central directory record

	if _, err := r.ReadAt(sig[:4], off); err != nil {
		return off, fmt.Errorf("failed to possibly read zip64 end of central directory signature: %v", err)
	}

	switch {
	case subtle.ConstantTimeCompare(sig[:4], zip64EndOfCentralDirSignature[:]) == 1:
		// [zip64 end of central directory record] is present
		off += 4 // offset zip64 end of central dir signature
	case subtle.ConstantTimeCompare(sig[:4], endOfCentralDirSignature[:]) == 1:
		// Non-empty/empty zip
		return off, nil
	default:
		return off, errors.New("expected [zip64 end of central directory record]")
	}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to read size of zip64 end of central directory record: %v", err)
	}
	off += 8 // offset size of zip64 end of central directory record

	// 4.3.14.1 The value stored into the "size of zip64 end of central directory record" SHOULD be the size of
	// the remaining record and SHOULD NOT include the leading 12 bytes.
	off += int64(binary.LittleEndian.Uint64(sig[:])) //nolint:gosec

	return off, nil
}

// readZip64EndOfCentralDirectoryLocator possibly reads [zip64 end of central directory locator] segment, where
// off must point at the beginning of the segment in a r stream. If the segment doesn't present in a stream at
// the location off, the returned value is the same as off.
func readZip64EndOfCentralDirectoryLocator(r io.ReaderAt, off int64) (int64, error) {
	sig := [4]byte{}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to read zip64 end of central directory locator signature: %v", err)
	}

	switch sig {
	case zip64EndOfCentralDirLocatorSignature:
		// [zip64 end of central directory record] [zip64 end of central directory locator] are present
		off += 4 // offset zip64 end of central dir locator signature
	case endOfCentralDirSignature:
		// Non-empty/empty zip
		return off, nil
	default:
		return off, errors.New("expected [zip64 end of central directory locator]")
	}

	off += 16 // offset [zip64 end of central directory locator]

	return off, nil
}

// readEndOfCentralDirectoryRecord possibly reads [end of central directory record] segment, where off must point
// at the beginning of the segment in a r stream. If the segment doesn't present in a stream at the location off,
// the returned value is the same as off.
func readEndOfCentralDirectoryRecord(r io.ReaderAt, off int64) (int64, error) {
	sig := [4]byte{}

	if _, err := r.ReadAt(sig[:], off); err != nil {
		return off, fmt.Errorf("failed to read end of central dir signature: %v", err)
	}

	if subtle.ConstantTimeCompare(sig[:], endOfCentralDirSignature[:]) != 1 {
		return off, errors.New("expected [end of central directory record]")
	}
	off += 4 // offset end of central dir signature

	off += 16 // offset number of this disk ... starting disk number

	if _, err := r.ReadAt(sig[:2], off); err != nil {
		return off, fmt.Errorf("failed to read .ZIP file comment length")
	}
	off += 2 // offset .ZIP file comment length

	off += int64(binary.LittleEndian.Uint16(sig[:2])) // offset .ZIP file comment (variable size)

	return off, nil
}
