package ee.ivxv.common.asn1;

import ee.ivxv.common.asn1.ASN1Decodable;
import ee.ivxv.common.asn1.ASN1DecodingException;
import org.bouncycastle.asn1.ASN1ObjectIdentifier;
import org.bouncycastle.asn1.x509.AlgorithmIdentifier;
import org.bouncycastle.asn1.pkcs.PrivateKeyInfo;
import org.bouncycastle.asn1.DEROctetString;
import org.bouncycastle.asn1.ASN1OctetString;
import org.bouncycastle.asn1.ASN1Primitive;
import org.bouncycastle.asn1.ASN1Encodable;
import org.bouncycastle.asn1.ASN1InputStream;
import java.io.IOException;

/**
 * PKCS8PrivateKey is class for encoding and decoding private keys in PKCS8 structure.
 *
 */
public class PKCS8PrivateKey implements ee.ivxv.common.asn1.ASN1Encodable, ASN1Decodable {
    private ASN1ObjectIdentifier oid;
    private byte[] key;
    private byte[] params;

    /**
     * Create uninitialized PKCS8PrivateKey for decoding.
     */
    public PKCS8PrivateKey() {}

    /**
     * Initialize PKCS8PrivateKey using an encryption scheme OID, raw key bytes and ASN1-encoded key
     * parameters. If parameters are not set, it may be null.
     * 
     * @param oid Encryption scheme OID
     * @param key Raw key bytes
     * @param params ASN1-encoded key parameters, may be null.
     */
    public PKCS8PrivateKey(String oid, byte[] key, byte[] params) {
        this.oid = new ASN1ObjectIdentifier(oid);
        this.key = key.clone();
        this.params = params;
    }

    /**
     * Encode the key in PKCS8 structure. The structure is defined as:
     * 
     * <pre>
     * {@code
     * PrivateKeyInfo ::= SEQUENCE {
     *   version Version,
     *   privateKeyAlgorithm PrivateKeyAlgorithmIdentifier,
     *   privateKey PrivateKey,
     *   attributes [0] IMPLICIT Attributes OPTIONAL
     *   }
     * 
     * Version ::= INTEGER
     * PrivateKeyAlgorithmIdentifier ::= AlgorithmIdentifier
     * PrivateKey ::= OCTET STRING
     * Attributes ::= SET OF Attribute
     * 
     * AlgorithmIdentifier ::= SEQUENCE {
     *   algorithm OBJECT IDENTIFIER,
     *   parameters ANY DEFINED BY algorithm OPTIONAL
     *   }
     * }
     * </pre>
     */
    @Override
    public byte[] encode() {
        try {
            if (params != null) {
                return new PrivateKeyInfo(
                        new AlgorithmIdentifier(oid, ASN1Primitive.fromByteArray(params)),
                        new DEROctetString(key)).getEncoded("DER");
            } else {
                return new PrivateKeyInfo(new AlgorithmIdentifier(oid), new DEROctetString(key))
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
     * Use ASN1 DER-encoded bytes to set the fields of PKCS8PrivateKey. The expected input structure
     * is defined in {@link #encode()}.
     * 
     * @param in Input ASN1 DER-encoded bytes
     * @throw ASN1DecodingException When instance is initialized or input bytes do not correspond to
     *        the PKCS8PrivateKey
     *
     * @see #encode()
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
        PrivateKeyInfo pki;
        pki = PrivateKeyInfo.getInstance(p);
        oid = pki.getPrivateKeyAlgorithm().getAlgorithm();
        ASN1Encodable k;
        try {
            k = pki.parsePrivateKey();
        } catch (IOException e) {
            throw new ASN1DecodingException("Input not PKCS8PrivateKey");
        }
        if (!(k instanceof ASN1OctetString)) {
            throw new ASN1DecodingException("Input not PKCS8PrivateKey");
        }
        key = ((ASN1OctetString) k).getOctets();
        k = pki.getPrivateKeyAlgorithm().getParameters();
        if (k != null) {
            try {
                params = ((ASN1Primitive) k).getEncoded("DER");
            } catch (IOException e) {
                throw new ASN1DecodingException("Input not PKCS8PrivateKey");
            }
        }
    }

    /**
     * Get the encryption scheme OID
     * 
     * @return Encryption scheme OID as a string.
     * @throws ASN1DecodingException When not set.
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
     * @return Raw key bytes.
     * @throws ASN1DecodingException When not set.
     */
    public byte[] getKey() throws ASN1DecodingException {
        if (key == null) {
            throw new ASN1DecodingException("Instance not initialized");
        }
        return key;
    }

    /**
     * Get the key parameters. May be null.
     * 
     * @return ASN1 DER-encoded key parameters. May be null.
     */
    public byte[] getParameters() {
        // params can be null
        return params;
    }
}
