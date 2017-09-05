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

public class Sequence implements ee.ivxv.common.asn1.ASN1Encodable, ASN1Decodable {
    private ASN1Encodable[] fields;

    public Sequence() {}

    public Sequence(BigInteger... ints) {
        fields = new ASN1Integer[ints.length];
        for (int i = 0; i < ints.length; i++) {
            fields[i] = new ASN1Integer(ints[i]);
        }
    }

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

    public Sequence(String... strs) {
        fields = new DERGeneralString[strs.length];
        for (int i = 0; i < strs.length; i++) {
            fields[i] = new DERGeneralString(strs[i]);
        }
    }

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

    @Override
    public void readFromBytes(byte[] in) throws ASN1DecodingException {
        if (fields != null) {
            throw new ASN1DecodingException("Instance already initialized");
        }
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
