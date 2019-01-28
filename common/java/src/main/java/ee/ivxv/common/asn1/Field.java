package ee.ivxv.common.asn1;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.ASN1Encodable;
import ee.ivxv.common.asn1.ASN1Decodable;
import org.bouncycastle.asn1.ASN1Primitive;
import org.bouncycastle.asn1.ASN1Integer;
import org.bouncycastle.asn1.DEROctetString;
import org.bouncycastle.asn1.DERGeneralString;
import java.math.BigInteger;
import java.io.IOException;

/**
 * Field class is used to encode and decode arbitrary types.
 *
 */
public class Field implements ASN1Encodable, ASN1Decodable {
    private ASN1Primitive f;

    /**
     * Create uninitialized instance for decoding.
     */
    public Field() {}

    /**
     * Initialize using {@link java.math.BigInteger}.
     * 
     * @param i
     */
    public Field(BigInteger i) {
        f = new ASN1Integer(i);
    }

    /**
     * Initialize using string.
     * 
     * @param s
     */
    public Field(String s) {
        f = new DERGeneralString(s);
    }

    /**
     * Initialize using raw byte array.
     * 
     * @param b
     */
    public Field(byte[] b) {
        f = new DEROctetString(b);
    }

    /**
     * Return ASN1 DER-encoded byte arrays representing the value the instance was initialized with.
     * 
     * @return ASN1 DER-encoded byte array.
     */
    @Override
    public byte[] encode() {
        try {
            return f.getEncoded("DER");
        } catch (IOException e) {
            // if exception is thrown here then this is a programming error.
            // these error must be fixed during development phase and thus they
            // are not thrown in production. so, silence IOExceptions
            return null;
        }
    }

    /**
     * Decode the value from ASN1 DER-encoded input.
     * 
     * @param b ASN1 DER-encoded input.
     * @throws ASN1DecodingException When instance is already initialized or decoding from byte
     *         array fails.
     */
    @Override
    public void readFromBytes(byte[] b) throws ASN1DecodingException {
        if (f != null) {
            throw new ASN1DecodingException("Instance already initialized");
        }
        try {
            f = ASN1Primitive.fromByteArray(b);
        } catch (IOException e) {
            throw new ASN1DecodingException("Reading from byte array failed: " + e.toString());
        }
    }

    /**
     * Get the field value as {@link java.math.BigInteger}.
     * 
     * @return
     * @throws ASN1DecodingException When not decodable as an integer.
     */
    public BigInteger getInteger() throws ASN1DecodingException {
        if (f == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        if (!(f instanceof ASN1Integer)) {
            throw new ASN1DecodingException("Field is not an integer");
        }
        return ((ASN1Integer) f).getValue();
    }

    /**
     * Get the field values as string.
     * 
     * @return
     * @throws ASN1DecodingException When not decodable as a string.
     */
    public String getString() throws ASN1DecodingException {
        if (f == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        if (!(f instanceof DERGeneralString)) {
            throw new ASN1DecodingException("Field is not a string");
        }
        return ((DERGeneralString) f).getString();
    }

    /**
     * Get the field value as raw bytes.
     * 
     * @return
     * @throws ASN1DecodingException When not decodable as raw bytes.
     */
    public byte[] getBytes() throws ASN1DecodingException {
        if (f == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        if (!(f instanceof DEROctetString)) {
            throw new ASN1DecodingException("Field is not a byte array");
        }
        return ((DEROctetString) f).getOctets();
    }
}
