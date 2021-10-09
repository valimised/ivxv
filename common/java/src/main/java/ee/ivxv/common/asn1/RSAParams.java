package ee.ivxv.common.asn1;

import java.io.IOException;
import java.math.BigInteger;
import org.bouncycastle.asn1.ASN1Encodable;
import org.bouncycastle.asn1.ASN1InputStream;
import org.bouncycastle.asn1.ASN1Integer;
import org.bouncycastle.asn1.ASN1Primitive;
import org.bouncycastle.asn1.ASN1Sequence;
import org.bouncycastle.asn1.DERSequence;

/**
 * RSAParams is a class for serializing RSA parameters (public and private exponents and modulus) in
 * ASN.1 structure.
 */
public class RSAParams implements ee.ivxv.common.asn1.ASN1Encodable, ASN1Decodable {

    private BigInteger publicExponent;
    private BigInteger privateExponent;
    private BigInteger modulus;

    /**
     * Constructor for uninitialized instance for later decoding.
     */
    public RSAParams() {}

    /**
     * Creates a new instance with the specified parameters.
     * 
     * @param publicExponent the public exponent, not {@code null}.
     * @param privateExponent the private exponent, not{@code null}.
     * @param modulus the modulus, not {@code null}.
     */
    public RSAParams(BigInteger publicExponent, BigInteger privateExponent, BigInteger modulus) {
        this.publicExponent = publicExponent;
        this.privateExponent = privateExponent;
        this.modulus = modulus;
    }

    /**
     * Creates a new partially evaluated instance of RSAParams with private exponent and modulus.
     * 
     * @param d the private exponent, not {@code null}.
     * @param n the modulus, not {@code null}.
     * @return the new RSAParams.
     */
    public static RSAParams createPrivateKeyParams(BigInteger d, BigInteger n) {
        return new RSAParams(BigInteger.ZERO, d, n);
    }

    public BigInteger getPublicExponent() {
        return publicExponent;
    }

    public BigInteger getPrivateExponent() {
        return privateExponent;
    }

    public BigInteger getModulus() {
        return modulus;
    }

    /**
     * Encode the fields in ASN1 SEQUENCE.
     * 
     * @returns ASN1 DER-encoded SEQUENCE of fields.
     */
    @Override
    public byte[] encode() {
        try {
            ASN1Encodable[] fields = {new ASN1Integer(publicExponent),
                    new ASN1Integer(privateExponent), new ASN1Integer(modulus)};
            return new DERSequence(fields).getEncoded();
        } catch (IOException ex) {
            // if exception is thrown here then this is a programming error.
            // these error must be fixed during development phase and thus they
            // are not thrown in production. so, silence IOExceptions
            return null;
        }
    }

    /**
     * Decode the ASN1 DER-encode byte array as an RSAKeyPair object.
     * 
     * @param in Input ASN1 DER-encoded byte array.
     * @throws ASN1DecodingException when input is not ASN1 SEQUENCE containing 3 ASN1 INTEGERs.
     */
    @Override
    public void readFromBytes(byte[] in) throws ASN1DecodingException {
        if (publicExponent != null || privateExponent != null || modulus != null) {
            throw new ASN1DecodingException("Instance already initialized");
        }
        @SuppressWarnings("resource")
        ASN1InputStream a = new ASN1InputStream(in);
        ASN1Primitive p;
        try {
            p = a.readObject();
        } catch (IOException ex) {
            throw new ASN1DecodingException("Invalid ASN1");
        }
        if (!(p instanceof ASN1Sequence)) {
            throw new ASN1DecodingException("Input not ASN1 SEQUENCE");
        }
        ASN1Encodable[] fields = ((ASN1Sequence) p).toArray();

        if (fields.length != 3) {
            throw new ASN1DecodingException("Expected 3 elements, got " + fields.length);
        }
        for (int i = 0; i < fields.length; i++) {
            if (!(fields[i] instanceof ASN1Integer)) {
                throw new ASN1DecodingException("Sequence field not an integer");
            }
        }
        publicExponent = ((ASN1Integer) fields[0]).getValue();
        privateExponent = ((ASN1Integer) fields[1]).getValue();
        modulus = ((ASN1Integer) fields[2]).getValue();
    }

}
