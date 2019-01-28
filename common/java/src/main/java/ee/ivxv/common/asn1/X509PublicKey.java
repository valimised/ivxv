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

/**
 * X509PublicKey is a class for encoding and decoding X509 SubjectPublicKeyInfo structure.
 *
 */
public class X509PublicKey implements ee.ivxv.common.asn1.ASN1Encodable, ASN1Decodable {
    private ASN1ObjectIdentifier oid;
    public byte[] key;
    public byte[] params;

    /**
     * Initialize empty X509PublicKey for decoding from encoded data.
     */
    public X509PublicKey() {}

    /**
     * Initialize X509PublicKey using encryption scheme OID, raw key bytes and encoded parameters.
     * If the parameters are not set, then it may be null.
     * 
     * @param oid OID of the encryption scheme, given as a String.
     * @param key Raw public key bytes.
     * @param params ASN1 DER-encoded key parameters.
     */
    public X509PublicKey(String oid, byte[] key, byte[] params) {
        this.oid = new ASN1ObjectIdentifier(oid);
        this.key = new Sequence(key).encode();
        this.params = params;
    }

    /**
     * Encode the X509PublicKey into following ASN1 structure:
     * 
     * <pre>
     * {@code
     *  SubjectPublicKeyInfo ::= SEQUENCE {
     *    algorithm   AlgorithmIdentifier,
     *    publicKey   BIT STRING
     *    }
     *    
     *  AlgorithmIdentifier  ::=  SEQUENCE {
     *    algorithm   OBJECT IDENTIFIER,
     *    parameters  ANY DEFINED BY algorithm OPTIONAL
     *    }
     * }
     * </pre>
     */
    @Override
    public byte[] encode() {
        try {
            if (params != null) {
                return new SubjectPublicKeyInfo(
                        new AlgorithmIdentifier(oid, ASN1Primitive.fromByteArray(params)), key)
                                .getEncoded("DER");
            } else {
                return new SubjectPublicKeyInfo(new AlgorithmIdentifier(oid), key)
                        .getEncoded("DER");
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

    /**
     * Decode input into uninitialized X509PublicKey instance. The input must be formatted according
     * to the structure defined in {@link #encode()}.
     * 
     * @param in ASN1 DER-encoded byte array to be decoded.
     * @throws ASN1DecodingException When instance is already initialized or input does not
     *         correspond to the expected structure.
     */
    @Override
    public void readFromBytes(byte[] in) throws ASN1DecodingException {
        if (oid != null || key != null || params != null) {
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

    /**
     * Get the OID of the encryption scheme.
     * 
     * @return OID
     * @throws ASN1DecodingException When OID value is not set.
     */
    public String getOID() throws ASN1DecodingException {
        if (oid == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        return oid.toString();
    }

    /**
     * Get the raw key bytes.
     * 
     * @return Key as raw bytes.
     * @throws ASN1DecodingException When key is not set..
     */
    public byte[] getKey() throws ASN1DecodingException {
        if (key == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        Sequence keys = new Sequence();
        try {
            keys.readFromBytes(this.key);
        } catch (ASN1DecodingException ex) {
            // suppress this exception as the key is constructed during
            // instance initialization
        }
        try {
            return keys.getBytes()[0];
        } catch (ASN1DecodingException ex) {
            // suppress this exception as the key is constructed during
            // instance initialization
            return null;
        }
    }

    /**
     * Get the key parameters. Key parameters may be null if parameters are not needed or set.
     * 
     * @return Encoded key parameters.
     */
    public byte[] getParameters() {
        // params can be null
        return params;
    }
}
