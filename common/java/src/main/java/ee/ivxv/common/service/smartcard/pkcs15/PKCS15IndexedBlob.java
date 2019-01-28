package ee.ivxv.common.service.smartcard.pkcs15;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Field;
import ee.ivxv.common.asn1.Sequence;
import ee.ivxv.common.service.smartcard.IndexedBlob;
import ee.ivxv.common.util.Util;

/**
 * PKCS15IndexedBlob is a blob with an index.
 */
public class PKCS15IndexedBlob extends IndexedBlob {
    /**
     * Initialize using a file data and the index.
     * 
     * @param index
     * @param blob
     */
    public PKCS15IndexedBlob(int index, byte[] blob) {
        super(index, blob);
    }

    /**
     * Parse a serialized blob.
     * 
     * @param file
     * @return
     * @throws PKCS15Exception
     */
    public static PKCS15IndexedBlob create(byte[] file) throws PKCS15Exception {
        Sequence s = new Sequence();
        try {
            s.readFromBytes(file);
        } catch (ASN1DecodingException e) {
            throw new PKCS15Exception("Error reading indexed blob: " + e.toString());
        }
        byte[][] fields;
        try {
            fields = s.getBytes();
        } catch (ASN1DecodingException e) {
            throw new PKCS15Exception("Error decoding blob: " + e.toString());
        }
        if (fields.length != 2) {
            throw new PKCS15Exception("Invalid format for indexed blob");
        }
        Field indexField = new Field();
        Field blobField = new Field();
        try {
            indexField.readFromBytes(fields[0]);
            blobField.readFromBytes(fields[1]);
        } catch (ASN1DecodingException e) {
            throw new PKCS15Exception("Invalid field: " + e.toString());
        }
        int index;
        try {
            index = Util.toInt(indexField.getBytes());
        } catch (ASN1DecodingException e) {
            throw new PKCS15Exception("Error while decoding index field: " + e.toString());
        }
        byte[] blob;
        try {
            blob = blobField.getBytes();
        } catch (ASN1DecodingException e) {
            throw new PKCS15Exception("Error while decoding blob field: " + e.toString());
        }
        return new PKCS15IndexedBlob(index, blob);
    }

    /**
     * Serialize and indexed blob.
     * <p>
     * The index is encoded as a 4-byte big-endian value
     * {@code
     * SEQUENCE (
     *    index OCTETSTRING
     *    blob  OCTETSTRING
     *    )
     * }
     * 
     * @return
     */
    public byte[] encode() {
        return new Sequence(new Field(Util.toBytes(index)).encode(), new Field(blob).encode())
                .encode();
    }
}
