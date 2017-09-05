package ee.ivxv.common.asn1;

import java.io.IOException;
import org.bouncycastle.asn1.ASN1Encodable;
import org.bouncycastle.asn1.ASN1InputStream;
import org.bouncycastle.asn1.ASN1ObjectIdentifier;
import org.bouncycastle.asn1.ASN1Primitive;
import org.bouncycastle.asn1.ASN1Sequence;
import org.bouncycastle.asn1.DERSequence;

public class Ciphertext implements ee.ivxv.common.asn1.ASN1Encodable, ASN1Decodable {
    private ASN1ObjectIdentifier oid;
    private byte[] ct;

    public Ciphertext() {

    }

    public Ciphertext(String oid, byte[] ct) {
        this.oid = new ASN1ObjectIdentifier(oid);
        this.ct = ct;
    }

    @Override
    public byte[] encode() {
        try {
            return new DERSequence(
                    new ASN1Encodable[] {new DERSequence(oid), ASN1Primitive.fromByteArray(ct)})
                            .getEncoded();
        } catch (IOException e) {
            // if exception is thrown here then this is a programming error.
            // these error must be fixed during development phase and thus they
            // are not thrown in production. so, silence IOExceptions
            return null;
        }
    }

    @Override
    public void readFromBytes(byte[] in) throws ASN1DecodingException {
        if (oid != null || ct != null) {
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
        ASN1Sequence s = (ASN1Sequence) p;
        if (s.size() != 2) {
            throw new ASN1DecodingException("Bytes do not correspond to Ciphertext structure");
        }
        if (!(s.getObjectAt(0) instanceof ASN1Sequence)) {
            throw new ASN1DecodingException("Bytes do not correspond to Ciphertext structure");
        }
        ASN1Sequence oidSeq = (ASN1Sequence) s.getObjectAt(0);
        if (oidSeq.size() != 1) {
            throw new ASN1DecodingException("Bytes do not correspond to Ciphertext structure");
        }
        oid = (ASN1ObjectIdentifier) oidSeq.getObjectAt(0);
        try {
            ct = ((ASN1Primitive) s.getObjectAt(1)).getEncoded("DER");
        } catch (IOException e) {
            throw new ASN1DecodingException("Bytes do not correspond to Ciphertext structure");
        }
    }

    public String getOID() throws ASN1DecodingException {
        if (oid == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        return oid.toString();
    }

    public byte[] getCiphertext() throws ASN1DecodingException {
        if (ct == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        return ct.clone();
    }
}
