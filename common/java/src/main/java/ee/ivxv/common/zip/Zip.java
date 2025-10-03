package ee.ivxv.common.zip;

import java.io.IOException;
import java.io.InputStream;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;

/**
 * Zip implements .ZIP File Format Specification, version 6.3.10 as defined in
 * https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT.
 */
public final class Zip {

    // Size in bytes of a buffer to read a zip file content.
    private static final int zipFileBuffer = 8192; // 8 KB

    // 6.0 Traditional PKWARE Encryption.
    private static final int traditionalPKWAREEncryption = 12;

    // 4.4.4 general purpose bit flag: (2 bytes). Bit 0.
    private static final byte encryptedFlag = 0b00000001;

    // 4.4.4 general purpose bit flag: (2 bytes). Bit 3.
    private static final byte dataDescriptorFlag = 0b00001000;

    // APPENDIX E - AE-x encryption marker: https://www.winzip.com/en/support/aes-encryption/. Little-endian value
    // of 0x9901.
    private static final byte[] aexEncryption = {1, -103};

    // 4.4.5 compression method: (2 bytes). The file is stored (no compression).
    private static final byte[] noCompression = {0, 0};

    // 4.3.7 Local file header. Little-endian of 0x04034b50.
    private static final byte[] localFileHeaderSignature = {80, 75, 3, 4};

    // 8.5.5 The signature value 0x08074b50 is also used by some ZIP implementations as a marker for the
    // Data Descriptor record. Little-endian of 0x08074b50.
    private static final byte[] dataDescriptorSignature = {80, 75, 7, 8};

    // 4.3.11 Archive extra data record. Little-endian of 0x08064b50.
    private static final byte[] archiveExtraDataSignature = {80, 75, 6, 8};

    // 4.3.12 Central directory structure. Little-endian value of 0x02014b50.
    private static final byte[] centralFileHeaderSignature = {80, 75, 1, 2};

    // 4.3.13 Digital signature. Little-endian value of 0x05054b50.
    private static final byte[] headerSignature = {80, 75, 5, 5};

    // 4.3.14 Zip64 end of central directory record. Little-endian value of 0x06064b50.
    private static final byte[] zip64EndOfCentralDirSignature = {80, 75, 6, 6};

    // 4.3.15 Zip64 end of central directory locator. Little-endian value of 0x07064b50.
    private static final byte[] zip64EndOfCentralDirLocatorSignature = {80, 75, 6, 7};

    // 4.3.16 End of central directory record. Little-endian value of 0x06054b50.
    private static final byte[] endOfCentralDirSignature = {80, 75, 5, 6};

    /**
     * Segment is a zip segment.
     */
    public static final class Segment {
        // Offset of the segment from the beginning of a zip stream.
        private int off;

        // Signature of the segment.
        private byte[] signature;

        // Is true when next segment must be taken for parsing.
        private boolean hasNext;

        // This parameter is only relevant when the segment is [file data n], and is true
        // when this segment, having hasNext=true, is not yet counted as a zip file.
        private boolean isNotCountedFile;

        private Segment() {
        }
    }

    /**
     * File is a [file data n] segment of a zip.
     *
     * @param name
     * @param compressedSize
     * @param uncompressedSize
     */
    private record File(String name, long compressedSize, long uncompressedSize) {
    }

    /**
     * endOfCentralDirectoryRecord locates the end of [end of central directory record] segment in a zip stream.
     *
     * @param zipStream
     * @return
     * @throws IOException occurs when reading from a zip stream fails
     */
    public static int endOfCentralDirectoryRecord(InputStream zipStream, int filesCountLimit, int fileSizeLimit, int zipSizeLimit) throws IOException {
        // We have to buffer a stream, so after reading files from a stream we could reset the position pointer back
        // to the start of a stream
        zipStream.mark(zipSizeLimit);

        // Validate zip stream according to zip specification as much as java.util.zip.ZipInputStream allows
        ZipInputStream zis = new ZipInputStream(zipStream);

        // We have to locate all zip files beforehand, using java.util.zip.ZipInputStream, in order to obtain
        // compressed/uncompressed sizes of each file. The edge case that we are delegating to
        // java.util.zip.ZipInputStream is the situation where general purpose flag bit 3 is set, and therefore
        // compressed/uncompressed sizes may not appear in [local file header n], but rather in [data descriptor n],
        // which, unfortunately, appears after [local file header n] and therefore makes it hard to parse these values
        // in a sequential stream
        List<File> files = new ArrayList<>();

        // Read uncompressed zip file content step-by-step by a buffer size
        byte[] buf = new byte[zipFileBuffer];

        int filesCount = 0;

        for (ZipEntry entry; (entry = zis.getNextEntry()) != null; ) {
            filesCount++;

            if (filesCount > filesCountLimit) {
                throw new IOException("there are more files in a zip than expected");
            }

            // Standard, but unreliable approach to catch a file that exceeds a size limit
            if (entry.getSize() > fileSizeLimit) {
                throw new IOException("zip file uncompressed size exceeds a limit");
            }

            int fileSize = 0;

            // https://wiki.sei.cmu.edu/confluence/display/java/IDS04-J.+Safely+extract+files+from+ZipInputStream
            while (true) { // safe endless loop, since we don't let to read more that limit bytes
                int n = zis.read(buf);
                if (n == -1) {
                    // End of file
                    break;
                }

                fileSize += n;

                // Reliable way to catch a file that exceeds a size limit
                if (fileSize > fileSizeLimit) {
                    throw new IOException("zip file content size exceeds a limit");
                }
            }

            String name = entry.getName();
            long compressedSize = entry.getCompressedSize();
            long uncompressedSize = entry.getSize();

            files.add(new File(name, compressedSize, uncompressedSize));
        }

        // Reset stream pointer back to the beginning of a stream, so we could read it again
        zipStream.reset();

        // Mutable zip stream segment. Each method that parses a segment will mutate this instance
        Segment segment = new Segment();

        int countFiles = 0;

        // Infinite loop is dangerous, but we have validated zip with ZipInputStream, so it is safe here
        for (; ; countFiles++) {
            readFile(zipStream, files, segment);

            if (segment.hasNext) { // no [local file header n] ... [data descriptor n] segments left
                if (segment.isNotCountedFile) { // if the file is not the same as before, then count it
                    countFiles++;
                }

                break;
            }

            // More [local file header n] ... [data descriptor n] segments left
        }

        // [archive decryption header] 
        readArchiveDecryptionHeader(zipStream, segment);

        // [archive extra data record]
        readArchiveExtraDataRecord(zipStream, segment);

        // [central directory header 1] ... [central directory header n]
        readCentralDirectoryHeader(zipStream, countFiles, segment);

        // [zip64 end of central directory record]
        readZip64EndOfCentralDirectoryRecord(zipStream, segment);

        // [zip64 end of central directory locator]
        readZip64EndOfCentralDirectoryLocator(zipStream, segment);

        // [end of central directory record]
        readEndOfCentralDirectoryRecord(zipStream, segment);

        return segment.off;
    }

    /**
     * readFile returns the offset of [local file header n] ... [data descriptor n]. Files are [file data n]
     * segments of a zip stream, which must be parsed beforehand.
     *
     * @param zipStream
     * @return
     * @throws IOException
     */
    private static void readFile(InputStream zipStream, List<File> files, Segment s) throws IOException {
        byte[] signature = new byte[4];

        zipStream.read(signature); // offset local file header signature
        s.off += 4;

        if (MessageDigest.isEqual(signature, localFileHeaderSignature)) {
            // There is a file in a zip
        } else if (MessageDigest.isEqual(signature, centralFileHeaderSignature)) {
            // All zip files are read already
            s.hasNext = true;
            s.signature = centralFileHeaderSignature;

            return;
        } else if (MessageDigest.isEqual(signature, endOfCentralDirSignature)) {
            // Empty zip
            s.hasNext = true;
            s.signature = endOfCentralDirSignature;

            return;
        } else {
            throw new IOException("expected [local file header n]");
        }

        zipStream.skip(2); // offset version needed to extract
        s.off += 2;

        byte[] field = new byte[2];

        zipStream.read(field); // offset general purpose bit flag
        s.off += 2;

        // 4.4.4 general purpose bit flag: (2 bytes). Bit 0: If set, indicates that the file is encrypted.
        boolean isEncrypted = (field[0] & encryptedFlag) == encryptedFlag;

        // 4.3.9.1 This descriptor MUST exist if bit 3 of the general purpose bit flag is set.
        boolean withDataDescriptor = (field[0] & dataDescriptorFlag) == dataDescriptorFlag;

        zipStream.read(field); // offset compression method
        s.off += 2;

        boolean isCompressed = !MessageDigest.isEqual(field, noCompression);

        if (isCompressed) {
            zipStream.skip(8); // offset last mod file time ... crc-32
            s.off += 8;
        } else {
            zipStream.skip(12); // offset last mod file time ... compressed size
            s.off += 12;
        }

        zipStream.read(signature); // offset either compressed or uncompressed size
        s.off += 4;

        ByteBuffer buf = ByteBuffer.wrap(signature);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        int size = buf.getInt(); // either compressed or uncompressed size

        if (isCompressed) {
            zipStream.skip(4); // offset uncompressed size
            s.off += 4;
        }

        zipStream.read(field); // offset file name length
        s.off += 2;

        buf = ByteBuffer.wrap(field);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        short name = buf.getShort();

        zipStream.read(field); // offset extra field length
        s.off += 2;

        buf = ByteBuffer.wrap(field);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        short extra = buf.getShort();

        byte[] nameBytes = new byte[name];
        zipStream.read(nameBytes); // offset file name (variable size)
        s.off += name;
        String fileName = new String(nameBytes, StandardCharsets.UTF_8);

        // Only use provided files if size doesn't present in [local file header n], which actually means that
        // size is stored in [data descriptor n], which we cannot be parsed yet (sequential stream)
        if (size <= 0) {
            Optional<File> entry = files
                    .stream()
                    .filter(x -> fileName.equals(x.name))
                    .findFirst();
            if (entry.isEmpty()) {
                throw new IOException("no matching file (" + fileName + ") found in a zip stream");
            }

            size = (int) entry.get().uncompressedSize;
            if (isCompressed) {
                size = (int) entry.get().compressedSize;
            }
        }

        byte[] extraField = new byte[extra];
        byte[] headerID = new byte[2]; // Header ID - 2 bytes

        zipStream.read(extraField); // offset extra field (variable size)
        s.off += extra;

        if (extraField.length > 1) {
            System.arraycopy(extraField, 0, headerID, 0, 2);
        }

        boolean isPKWAREEncryption = false; // traditional PKWARE Encryption

        // Gather encryption info from extra field
        if (isEncrypted) {
            if (MessageDigest.isEqual(headerID, aexEncryption)) {
                // AE-x (e.g. AES) encryption marker, encryption info is stored within extra field, not within
                // [encryption header n]
            } else {
                // Encryption info is stored withing [encryption header n]
                isPKWAREEncryption = true;
            }
        }

        // If the file is encrypted, the encryption header for the file SHOULD be placed after the local header and
        // before the file data (relevant for traditional PKWARE Encryption only)
        if (isPKWAREEncryption) {
            zipStream.skip(traditionalPKWAREEncryption);
            s.off += traditionalPKWAREEncryption; // offset [encryption header n]
        }

        // After possibly [encryption header n] there must be [file data n]
        zipStream.skip(size); // offset file data
        s.off += size;

        if (withDataDescriptor) {
            zipStream.read(signature); // read possibly data descriptor signature, or crc-32
            s.off += 4;

            if (!MessageDigest.isEqual(signature, dataDescriptorSignature)) {
                zipStream.skip(8); // offset data descriptor without signature (4+4), crc-32 already read
                s.off += 8;
            } else {
                zipStream.skip(12); // offset data descriptor with signature (4+4+4), signature already read
                s.off += 12;
            }

            zipStream.read(signature); // offset next 4 bytes
            s.off += 4;

            if (MessageDigest.isEqual(signature, localFileHeaderSignature)) {
                // [data descriptor 1] [local file header 2], next file to be parsed
                s.hasNext = false;
                s.signature = localFileHeaderSignature;

                return;
            } else if (MessageDigest.isEqual(signature, centralFileHeaderSignature)) {
                // [data descriptor n] [central directory header 1], all files are parsed, start parsing central dir
                s.hasNext = true;
                s.signature = centralFileHeaderSignature;
                s.isNotCountedFile = true;

                return;
            } else {
                // [data descriptor n] is either zip64, or we have encountered [archive decryption header] segment.
                //
                // Assume it is zip64, so possibly offset zip64 data descriptor
                zipStream.skip(4);
                s.off += 4;
            }

            zipStream.read(signature); // offset next 4 bytes
            s.off += 4;

            if (MessageDigest.isEqual(signature, localFileHeaderSignature)) {
                // zip64 [data descriptor 1] [local file header 2], so next file to be parsed
                s.hasNext = false;
                s.signature = localFileHeaderSignature;

                return;
            } else if (MessageDigest.isEqual(signature, centralFileHeaderSignature)) {
                // zip64 [data descriptor n] [central directory header 1], all files are parsed, start parsing central
                // dir
                s.hasNext = true;
                s.signature = centralFileHeaderSignature;
                s.isNotCountedFile = true;

                return;
            } else {
                // Either zip64 [data descriptor n] [archive decryption header] or we have already read 8 bytes of
                // [archive decryption header].
                //
                // Assume we have read 8 bytes of [archive decryption header], so read 4 more to reach
                // [archive extra data record]
            }

            zipStream.read(signature); // offset 4 bytes to possibly reach [archive extra data record]
            s.off += 4;

            if (MessageDigest.isEqual(signature, archiveExtraDataSignature)) {
                // Reached [archive extra data record]. There are no files left in a stream
                s.hasNext = false;
                s.signature = archiveExtraDataSignature;

                return;
            } else {
                // We have read 4 bytes of [archive decryption header], now read 8 more to reach
                // [archive extra data record]
                zipStream.skip(traditionalPKWAREEncryption - 4);
                s.off += traditionalPKWAREEncryption - 4;
            }
        }

        // Read next zip file
        s.hasNext = false;
        s.signature = null;
    }

    /**
     * readArchiveDecryptionHeader possibly reads [archive decryption header] segment, where off must point at the
     * beginning of the segment in a zip stream. If the segment doesn't present in a stream at the location off, the
     * returned value is the same as off. Only supports traditional PKWARE encryption header.
     *
     * @param zipStream
     * @param s
     * @throws IOException
     */
    private static void readArchiveDecryptionHeader(InputStream zipStream, Segment s) throws IOException {
        if (MessageDigest.isEqual(s.signature, centralFileHeaderSignature)) {
            // Non-empty zip file without [archive decryption header]
            return;
        } else if (MessageDigest.isEqual(s.signature, endOfCentralDirSignature)) {
            // Empty zip file without [archive decryption header]
            return;
        }

        zipStream.skip(traditionalPKWAREEncryption); // 6.1 Traditional PKWARE Encryption
        s.off += traditionalPKWAREEncryption;
    }

    /**
     * readArchiveExtraDataRecord possibly reads [archive extra data record] segment, where off must point at the
     * beginning of the segment in a zip stream. If the segment doesn't present in a stream at the location off, the
     * returned value is the same as off.
     *
     * @param zipStream
     * @param s
     * @throws IOException
     */
    private static void readArchiveExtraDataRecord(InputStream zipStream, Segment s) throws IOException {
        byte[] signature = new byte[4];

        if (MessageDigest.isEqual(s.signature, archiveExtraDataSignature)) {
            // [archive decryption header] is present
            zipStream.skip(4); // offset archive extra data signature
            s.off += 4;
        } else if (MessageDigest.isEqual(s.signature, centralFileHeaderSignature)) {
            // Non-empty zip file without [archive decryption header]
            return;
        } else if (MessageDigest.isEqual(s.signature, endOfCentralDirSignature)) {
            // Empty zip file without [archive decryption header]
            return;
        } else {
            throw new IOException("expected [archive extra data record]");
        }

        zipStream.read(signature); // offset extra field length
        s.off += 4;

        ByteBuffer buf = ByteBuffer.wrap(signature);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        int extra = buf.getInt();

        zipStream.skip(extra); // offset extra field data (variable size)
        s.off += extra;

        s.signature = centralFileHeaderSignature; // next segment must be [central directory header 1]
    }

    /**
     * readCentralDirectoryHeader possibly reads [central directory header 1] ... [central directory header n]
     * segment, where off must point at the beginning of the segment in a zip stream. If the segment doesn't present
     * in a stream at the location off, the returned value is the same as off.
     *
     * @param zipStream
     * @param filesCount
     * @param s
     * @throws IOException
     */
    private static void readCentralDirectoryHeader(InputStream zipStream, int filesCount, Segment s) throws IOException {
        byte[] signature = new byte[4];

        if (MessageDigest.isEqual(s.signature, centralFileHeaderSignature)) {
            // Do nothing as we offset central file header signature in a 'for loop' anyway
        } else if (MessageDigest.isEqual(s.signature, endOfCentralDirSignature)) {
            // Empty zip
            return;
        } else {
            throw new IOException("expected [central directory header n]");
        }

        byte[] field = new byte[2];

        // All files within [local file header 1] ... [local file header n] must be present in
        // [central directory header 1] ... [central directory header n]
        for (int i = 0; i < filesCount; i++) {
            // Offset central file header signature ... uncompressed size
            if (i > 0) {
                zipStream.skip(4);
                s.off += 4;
            }
            zipStream.skip(24);
            s.off += 24;

            zipStream.read(field); // offset file name length
            s.off += 2;

            ByteBuffer buf = ByteBuffer.wrap(field);
            buf.order(ByteOrder.LITTLE_ENDIAN);
            short name = buf.getShort();

            zipStream.read(field); // offset extra field length
            s.off += 2;

            buf = ByteBuffer.wrap(field);
            buf.order(ByteOrder.LITTLE_ENDIAN);
            short extra = buf.getShort();

            zipStream.read(field); // offset file comment length
            s.off += 2;

            buf = ByteBuffer.wrap(field);
            buf.order(ByteOrder.LITTLE_ENDIAN);
            short com = buf.getShort();

            zipStream.skip(12); // offset disk number start ... relative offset of local header
            s.off += 12;

            zipStream.skip(name + extra + com); // offset file name (variable size) ... file comment (variable size)
            s.off += name + extra + com;
        }

        // 4.3.13 Digital signature, optional
        zipStream.read(signature);
        s.off += 4; // offset possibly header signature

        if (MessageDigest.isEqual(signature, headerSignature)) {
            // Digital signature present
        } else if (MessageDigest.isEqual(signature, zip64EndOfCentralDirSignature)) {
            // Non-empty zip64 file without a digital signature
            s.signature = zip64EndOfCentralDirSignature;

            return;
        } else if (MessageDigest.isEqual(signature, endOfCentralDirSignature)) {
            // Empty zip file without a digital signature
            s.signature = endOfCentralDirSignature;

            return;
        } else {
            throw new IOException("expected central directory header digital signature");
        }

        zipStream.read(field);  // offset size of data
        s.off += 2;

        ByteBuffer buf = ByteBuffer.wrap(field);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        short data = buf.getShort();

        zipStream.skip(data); // offset signature data
        s.off += data;

        zipStream.read(signature); // offset extra 4 bytes to determine which segment comes next
        s.off += 4;

        if (MessageDigest.isEqual(signature, zip64EndOfCentralDirSignature)) {
            // Non-empty zip64 file with a digital signature
            s.signature = zip64EndOfCentralDirSignature;

            return;
        } else if (MessageDigest.isEqual(signature, endOfCentralDirSignature)) {
            // Empty zip file with a digital signature
            s.signature = endOfCentralDirSignature;

            return;
        }

        throw new IOException("expected [zip64 end of central directory record] or [end of central directory record]");

    }

    /**
     * readZip64EndOfCentralDirectoryRecord possibly reads [zip64 end of central directory record] segment, where
     * off must point at the beginning of the segment in a zip stream. If the segment doesn't present in a stream at
     * the location off, the returned value is the same as off.
     *
     * @param zipStream
     * @param s
     * @throws IOException
     */
    private static void readZip64EndOfCentralDirectoryRecord(InputStream zipStream, Segment s) throws IOException {
        byte[] size = new byte[8];

        if (MessageDigest.isEqual(s.signature, zip64EndOfCentralDirSignature)) {
            // [zip64 end of central directory record] is present
        } else if (MessageDigest.isEqual(s.signature, endOfCentralDirSignature)) {
            // Non-empty/empty zip
            return;
        } else {
            throw new IOException("expected [zip64 end of central directory record]");
        }

        zipStream.read(size); // offset size of zip64 end of central directory record
        s.off += 8;

        // 4.3.14.1 The value stored into the "size of zip64 end of central directory record" SHOULD be the size of
        // the remaining record and SHOULD NOT include the leading 12 bytes.
        ByteBuffer buf = ByteBuffer.wrap(size);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        int extra = buf.getInt();

        zipStream.skip(extra); // offset extra field data (variable size)
        s.off += extra;

        // Next segment must be [zip64 end of central directory locator]
        s.signature = zip64EndOfCentralDirLocatorSignature;
    }

    /**
     * readZip64EndOfCentralDirectoryLocator possibly reads [zip64 end of central directory locator] segment, where
     * off must point at the beginning of the segment in a zip stream. If the segment doesn't present in a stream at
     * the location off, the returned value is the same as off.
     *
     * @param zipStream
     * @param s
     * @throws IOException
     */
    private static void readZip64EndOfCentralDirectoryLocator(InputStream zipStream, Segment s) throws IOException {
        byte[] signature = new byte[4];

        if (MessageDigest.isEqual(s.signature, zip64EndOfCentralDirLocatorSignature)) {
            // [zip64 end of central directory record] [zip64 end of central directory locator] are present
            zipStream.skip(4); // offset zip64 end of central dir locator signature
            s.off += 4;
        } else if (MessageDigest.isEqual(s.signature, endOfCentralDirSignature)) {
            // Non-empty/empty zip
            return;
        } else {
            throw new IOException("expected [zip64 end of central directory locator]");
        }

        zipStream.skip(16); // offset [zip64 end of central directory locator]
        s.off += 16;

        zipStream.read(signature); // offset next 4 bytes, which must be end of central dir signature
        s.off += 4;

        if (!MessageDigest.isEqual(signature, endOfCentralDirSignature)) {
            throw new IOException("expected end of central dir signature");
        }

        s.signature = endOfCentralDirSignature;
    }

    /**
     * readEndOfCentralDirectoryRecord possibly reads [end of central directory record] segment, where off must point
     * at the beginning of the segment in a zip stream. If the segment doesn't present in a stream at the location off,
     * the returned value is the same as off.
     *
     * @param zipStream
     * @param s
     * @throws IOException
     */
    private static void readEndOfCentralDirectoryRecord(InputStream zipStream, Segment s) throws IOException {
        if (!MessageDigest.isEqual(s.signature, endOfCentralDirSignature)) {
            throw new IOException("expected [end of central directory record]");
        }

        zipStream.skip(16); // offset number of this disk ... starting disk number
        s.off += 16;

        byte[] field = new byte[2];

        zipStream.read(field); // offset .ZIP file comment length
        s.off += 2;

        ByteBuffer buf = ByteBuffer.wrap(field);
        buf.order(ByteOrder.LITTLE_ENDIAN);
        short com = buf.getShort();

        zipStream.skip(com); // offset .ZIP file comment (variable size)
        s.off += com;
    }
}