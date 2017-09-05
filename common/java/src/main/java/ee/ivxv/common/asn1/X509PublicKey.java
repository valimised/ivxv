package ee.ivxv.common.asn1;

import ee.ivxv.common.asn1.ASN1Decodable;
import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Sequence;
import org.bouncycastle.asn1.x509.SubjectPublicKeyInfo;
import org.bouncycastle.asn1.ASN1ObjectIdentifier;
import org.bouncycastle.asn1.ASN1Primitive;
import org.bouncycastle.asn1.x509.AlgorithmIdentifier;
import org.bouncycastle.asn1.ASN1InputStream;
import org.bouncycastle.asn1.ASN1Encodable;
import java.io.IOException;

public class X509PublicKey implements ee.ivxv.common.asn1.ASN1Encodable, ASN1Decodable {
    /*

    SubjectPublicKeyInfo ::= SEQUENCE {
        algorithm AlgorithmIdentifier,
        publicKey BIT STRING }

    AlgorithmIdentifier  ::=  SEQUENCE  {
        algorithm               OBJECT IDENTIFIER,
        parameters              ANY DEFINED BY algorithm OPTIONAL  }
    */
    private ASN1ObjectIdentifier oid;
    public byte[] key;
    public byte[] params;

    public X509PublicKey() {
    }

    public X509PublicKey(String oid, byte[] key, byte[] params) {
        this.oid = new ASN1ObjectIdentifier(oid);
        this.key = new Sequence(key).encode();
        this.params = params;
    }

    @Override
    public byte[] encode() {
        try {
            if (params != null) {
                return new SubjectPublicKeyInfo(
                        new AlgorithmIdentifier(
                            oid, 
                            ASN1Primitive.fromByteArray(params)),
                        key
                ).getEncoded("DER");
            } else {
                return new SubjectPublicKeyInfo(
                        new AlgorithmIdentifier(oid),
                        key
                ).getEncoded("DER");
            }
        } catch (IOException e) {
            // if exception is thrown here then this is a programming error.
            // these error must be fixed during development phase and thus they
            // are not thrown in production. so, silence IOExceptions
            // the exception thrown here from are:
            // 1. ASN1Primitive.fromByteArray(params)),
            // 2. return new PrivateKeyInfo(
            // 3. ).getEncoded("DER");
            return null;
        }
    }

    @Override
    public void readFromBytes(byte[] in) throws ASN1DecodingException {
        if (oid != null || key != null || params != null) {
            throw new ASN1DecodingException("Instance already initialized");
        }
        ASN1InputStream a = new ASN1InputStream(in);
        ASN1Primitive p;
        try {
            p = a.readObject();
        } catch (IOException e) {
            throw new ASN1DecodingException("Invalid ASN1");
        }
        SubjectPublicKeyInfo spki;
        spki = SubjectPublicKeyInfo.getInstance(p);
        oid = spki.getAlgorithm().getAlgorithm();
        key = spki.getPublicKeyData().getBytes();
        ASN1Encodable k;
        k = spki.getAlgorithm().getParameters();
        if (k != null) {
            try {
                params = ((ASN1Primitive) k).getEncoded("DER");
            } catch (IOException e) {
                throw new ASN1DecodingException("Input not X509PrivateKey");
            }
        }
    }

    public String getOID() throws ASN1DecodingException {
        if (oid == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        return oid.toString();
    }

    public byte[] getKey() throws ASN1DecodingException {
        if (key == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        Sequence keys = new Sequence();
        try {
            keys.readFromBytes(this.key);
        } catch (ASN1DecodingException ex) {
            // suppress this exception as the key is constructed during
            // instance initalization
        }
        try {
            return keys.getBytes()[0];
        } catch (ASN1DecodingException ex) {
            // suppress this exception as the key is constructed during
            // instance initalization
            return null;
        }
    }

    public byte[] getParameters() {
        // params can be null
        return params;
    }
}
