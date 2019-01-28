package ee.ivxv.common.crypto.elgamal;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.X509PublicKey;
import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.crypto.rnd.NativeRnd;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.IntegerConstructor;
import ee.ivxv.common.math.MathException;
import ee.ivxv.common.math.ProductGroup;
import ee.ivxv.common.math.ProductGroupElement;
import ee.ivxv.common.util.Util;
import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Path;
import java.security.KeyException;

/**
 * ElGamalPublicKey allows encrypting messages.
 */
public class ElGamalPublicKey {
    private final ElGamalParameters parameters;
    private final GroupElement key;

    /**
     * Initialize using parameters and a public key instance.
     * 
     * @param parameters
     * @param key A group element which represents the ElGamal public key.
     * @throws IllegalArgumentException
     */
    public ElGamalPublicKey(ElGamalParameters parameters, GroupElement key)
            throws IllegalArgumentException {
        this.parameters = parameters;
        this.key = key;
    }

    /**
     * Initialize using serialized value.
     * 
     * @see #getBytes()
     * 
     * @param packed
     * @throws IllegalArgumentException
     */
    public ElGamalPublicKey(byte[] packed) throws IllegalArgumentException {
        X509PublicKey p = new X509PublicKey();
        try {
            p.readFromBytes(packed);
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Unpacking ElGamalPublicKey failed: " + e);
        }
        byte[] paramsb = p.getParameters();
        if (paramsb == null) {
            throw new IllegalArgumentException("Key parameters not defined");
        }
        String oids = "";
        try {
            oids = p.getOID();
        } catch (ASN1DecodingException e) {
            // does not happen but silence compilator warning
        }
        byte[] keyb = null;
        try {
            keyb = p.getKey();
        } catch (ASN1DecodingException e) {
            // does not happen but silence compilator warning
        }
        parameters = new ElGamalParameters(oids, paramsb);
        key = parameters.getGroup().getElement(keyb);
    }

    /**
     * Initialize using armored public key file.
     * <p>
     * The input key file should be MIME-encoded public key file with public key headers.
     * 
     * @param p
     * @throws IllegalArgumentException When decoding or parsing fails.
     * @throws IOException When reading fails
     */
    public ElGamalPublicKey(Path p) throws IllegalArgumentException, IOException {
        this(Util.decodePublicKey(p));
    }

    /**
     * Get the parameters.
     * 
     * @return
     */
    public ElGamalParameters getParameters() {
        return parameters;
    }

    /**
     * Get the public key instance.
     * 
     * @return
     */
    public GroupElement getKey() {
        return key;
    }

    /**
     * Serialize as X509 public key.
     * 
     * @see ee.ixvx.common.asn1.X509PublicKey
     * @see ee.ivxv.common.math.GroupElement#getBytes()
     * @see ee.ivxv.common.crypto.elgamal.ElGamalParameters#getBytes()
     * 
     * @return
     */
    public byte[] getBytes() {
        X509PublicKey der = new X509PublicKey(getParameters().getOID(), key.getBytes(),
                getParameters().getBytes());
        return der.encode();
    }

    /**
     * Represent the instance as a product group element.
     * <p>
     * Returns the value {@literal (g, y)}, where {@literal g} is the generator of the parameters
     * and {@literal y} is the value of the public key.
     * 
     * @return
     */
    public ProductGroupElement getAsProductGroupElement() {
        ProductGroup pgroup = new ProductGroup(getKey().getGroup(), 2);
        ProductGroupElement ret =
                new ProductGroupElement(pgroup, getParameters().getGenerator(), getKey());
        return ret;
    }

    @Override
    public String toString() {
        return String.format("ElGamalPublicKey(%s, %s)", parameters, key);
    }

    /**
     * Encrypt the message using system random source.
     * 
     * @param msg
     * @return
     * @throws KeyException When encryption fails.
     * @throws IOException When reading randomness fails.
     */
    public ElGamalCiphertext encrypt(Plaintext msg) throws KeyException, IOException {
        return encrypt(msg, new NativeRnd());
    }

    /**
     * Encrypt the message using specified random source.
     * 
     * @param msg
     * @param rnd
     * @return
     * @throws KeyException When encryption fails.
     * @throws IOException When reading randomness fails.
     */
    public ElGamalCiphertext encrypt(Plaintext msg, Rnd rnd) throws KeyException, IOException {
        Plaintext padded = getParameters().getGroup().pad(msg);
        GroupElement el;
        try {
            el = getParameters().getGroup().encode(padded);
        } catch (MathException e) {
            throw new KeyException("Encoding for key parameters failed: " + e);
        }
        BigInteger r = IntegerConstructor.construct(rnd, getParameters().getGeneratorOrder());
        try {
            return new ElGamalCiphertext(getParameters().getGenerator().scale(r),
                    getKey().scale(r).op(el), getParameters().getOID());
        } catch (MathException e) {
            throw new KeyException("Key initialization error");
        }
    }
}
