package ee.ivxv.common.asn1;

import ee.ivxv.common.asn1.ASN1Decodable;
import ee.ivxv.common.asn1.ASN1DecodingException;
import java.math.BigInteger;
import java.io.IOException;
import org.bouncycastle.asn1.ASN1Encodable;
import org.bouncycastle.asn1.ASN1Primitive;
import org.bouncycastle.asn1.ASN1Integer;
import org.bouncycastle.asn1.DERSequence;
import org.bouncycastle.asn1.ASN1InputStream;
import org.bouncycastle.asn1.ASN1Sequence;
import org.bouncycastle.asn1.DERGeneralString;

/**
 * Sequence can be used to encode and decode a series of values.
 *
 */
public class Sequence implements ee.ivxv.common.asn1.ASN1Encodable, ASN1Decodable {
    private ASN1Encodable[] fields;

    /**
     * Constructor for uninitialized instance for later decoding.
     */
    public Sequence() {}

    /**
     * Constructor for a sequence of integers.
     * 
     * @param ints
     */
    public Sequence(BigInteger... ints) {
        fields = new ASN1Integer[ints.length];
        for (int i = 0; i < ints.length; i++) {
            fields[i] = new ASN1Integer(ints[i]);
        }
    }

    /**
     * Constructor for a sequence of ASN1-encoded byte arrays.
     * 
     * @param encoded
     */
    public Sequence(byte[]... encoded) {
        fields = new ASN1Primitive[encoded.length];
        for (int i = 0; i < encoded.length; i++) {
            try {
                fields[i] = ASN1Primitive.fromByteArray(encoded[i]);
            } catch (IOException e) {
                // if exception is thrown here then this is a programming error.
                // these error must be fixed during development phase and thus they
                // are not thrown in production. so, silence IOExceptions
            }
        }
    }

    /**
     * Constructor for a sequence of strings. Every string is encoded as GeneralString.
     * 
     * @param strs
     */
    public Sequence(String... strs) {
        fields = new DERGeneralString[strs.length];
        for (int i = 0; i < strs.length; i++) {
            fields[i] = new DERGeneralString(strs[i]);
        }
    }

    /**
     * Encode the fields in ASN1 SEQUENCE.
     * 
     * @returns ASN1 DER-encoded SEQUENCE of fields.
     */
    @Override
    public byte[] encode() {
        try {
            return new DERSequence(fields).getEncoded();
        } catch (IOException e) {
            // if exception is thrown here then this is a programming error.
            // these error must be fixed during development phase and thus they
            // are not thrown in production. so, silence IOExceptions
            return null;
        }
    }

    /**
     * Decode the ASN1 DER-encode byte array as a Sequence.
     * 
     * @param in Input ASN1 DER-encoded byte array.
     * @throws ASN1DecodingException when input is not ASN1 SEQUENCE.
     */
    @Override
    public void readFromBytes(byte[] in) throws ASN1DecodingException {
        if (fields != null) {
            throw new ASN1DecodingException("Instance already initialized");
        }
        @SuppressWarnings("resource")
        ASN1InputStream a = new ASN1InputStream(in);
        ASN1Primitive p;
        try {
            p = a.readObject();
        } catch (IOException e) {
            throw new ASN1DecodingException("Invalid ASN1");
        }
        if (!(p instanceof ASN1Sequence)) {
            throw new ASN1DecodingException("Input not ASN1 SEQUENCE");
        }
        fields = ((ASN1Sequence) p).toArray();
    }

    /**
     * Get the sequence elements as byte arrays.
     * 
     * @return Sequence elements as byte arrays.
     * @throws ASN1DecodingException When sequence elements are not byte arrays.
     */
    public byte[][] getBytes() throws ASN1DecodingException {
        if (fields == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        byte[][] ret = new byte[fields.length][];
        try {
            for (int i = 0; i < fields.length; i++) {
                ret[i] = ((ASN1Primitive) fields[i]).getEncoded("DER");
            }
        } catch (IOException e) {
            throw new ASN1DecodingException("Sequence field incorrectly encoded");
        }
        return ret;
    }

    /**
     * Get the sequence elements as integers.
     * 
     * @return Sequence elements as integers.
     * @throws ASN1DecodingException When sequence elements are not integers.
     */
    public BigInteger[] getIntegers() throws ASN1DecodingException {
        if (fields == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        BigInteger[] ret = new BigInteger[fields.length];
        for (int i = 0; i < fields.length; i++) {
            if (!(fields[i] instanceof ASN1Integer)) {
                throw new ASN1DecodingException("Sequence field not an integer");
            }
            ret[i] = ((ASN1Integer) fields[i]).getValue();
        }
        return ret;
    }

    /**
     * Get the sequence elements as strings.
     * 
     * @return Sequence elements as strings.
     * @throws ASN1DecodingException When sequence elements are not strings.
     */
    public String[] getStrings() throws ASN1DecodingException {
        if (fields == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        String[] ret = new String[fields.length];
        for (int i = 0; i < fields.length; i++) {
            if (!(fields[i] instanceof DERGeneralString)) {
                throw new ASN1DecodingException("Sequence field not a string");
            }
            ret[i] = ((DERGeneralString) fields[i]).getString();
        }
        return ret;
    }
}
