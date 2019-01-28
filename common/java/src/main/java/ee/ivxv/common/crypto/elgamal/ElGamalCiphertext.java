package ee.ivxv.common.crypto.elgamal;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Ciphertext;
import ee.ivxv.common.asn1.Sequence;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.ProductGroup;
import ee.ivxv.common.math.ProductGroupElement;

/**
 * ElGamalCiphertext holds ElGamal ciphertext information.
 * <p>
 * Given the group {@code G}, the message {@code m}, public key {@code h}, randomness {@code r} and
 * generator {@code g}, the ElGamal ciphertext is a tuple {@code c = (c1, c2) = (g^r, m * h^r)}.
 * <p>
 * We call the part {@code c1} of the ciphertext as {@literal blind} as it is using a blinding
 * factor to hide the message.
 * <p>
 * We call the variable {@code c2} of the ciphertext as {@literal blinded message} as it is a
 * message which is blinded by the blinding factor.
 *
 */
public class ElGamalCiphertext {
    private final GroupElement blind;
    private final GroupElement blindedMessage;
    private final String oid;

    /**
     * Create the ElGamalCiphertext using blind, blinded message and OID of the encryption scheme.
     * 
     * @param blind
     * @param blindedMessage
     * @param oid
     */
    public ElGamalCiphertext(GroupElement blind, GroupElement blindedMessage, String oid) {
        this.blind = blind;
        this.blindedMessage = blindedMessage;
        this.oid = oid;
    }

    /**
     * Initialize using blind and blinded message encoded as a
     * {@link ee.ivxv.common.math.ProductGroupElement} and OID of the encryption scheme.
     * 
     * @param el
     * @param oid
     * @throws IllegalArgumentException
     */
    public ElGamalCiphertext(ProductGroupElement el, String oid) throws IllegalArgumentException {
        if (el.getElements().length != 2) {
            throw new IllegalArgumentException("Invalid dimension of group element");
        }
        this.blind = el.getElements()[0];
        this.blindedMessage = el.getElements()[1];
        this.oid = oid;
    }

    /**
     * Create the ElGamalCiphertext using {@link ElGamalParameters} and ASN1 DER-encoded ciphertext.
     * 
     * @param params
     * @param data
     * @throws IllegalArgumentException When can not decode ciphertext from input.
     */
    public ElGamalCiphertext(ElGamalParameters params, byte[] data)
            throws IllegalArgumentException {
        Ciphertext p = new Ciphertext();
        try {
            p.readFromBytes(data);
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Unpacking ElGamalCiphertext failed: " + e);
        }
        String poid = "";
        try {
            poid = p.getOID();
        } catch (ASN1DecodingException e) {
            // does not happen but silence compilator warning
        } finally {
            oid = poid;
        }
        if (!params.getOID().equals(oid)) {
            throw new IllegalArgumentException("Invalid key parameters.");
        }
        byte[] ct = null;
        try {
            ct = p.getCiphertext();
        } catch (ASN1DecodingException e) {
            // does not happen but silence compilator warning
        }
        Sequence s = new Sequence();
        try {
            s.readFromBytes(ct);
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Unpacking ElGamalCiphertext failed: " + e);
        }
        byte[][] parts;
        try {
            parts = s.getBytes();
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Unpacking ElGamalCiphertext failed: " + e);
        }
        if (parts.length != 2) {
            throw new IllegalArgumentException(
                    "Unpacking ElGamalCiphertext failed: invalid ciphertext length: "
                            + parts.length);
        }
        blind = params.getGroup().getElement(parts[0]);
        blindedMessage = params.getGroup().getElement(parts[1]);
    }

    /**
     * Get the blind of the ciphertext.
     * 
     * @return Blind of the ciphertext.
     */
    public GroupElement getBlind() {
        return blind;
    }

    /**
     * Get the blinded message of the ciphertext.
     * 
     * @return Blinded message of the ciphertext.
     */
    public GroupElement getBlindedMessage() {
        return blindedMessage;
    }

    /**
     * Get the OID of encryption scheme.
     * 
     * @return OID of encryption scheme.
     */
    public String getOID() {
        return oid;
    }

    /**
     * Encode the ciphertext as ASN1 DER-encoded structure.
     * <p>
     * Algorithm-specific structure:
     * 
     * {@code
     * SEQUENCE (
     *   blind          GroupElement
     *   blindedMessage GroupElement
     *   )
     * }
     * 
     * @see ee.ivxv.common.asn1.Ciphertext
     * 
     * @return ASN1 DER-encoded bytes representing ciphertext.
     */
    public byte[] getBytes() {
        return new Ciphertext(getOID(),
                new Sequence(getBlind().getBytes(), getBlindedMessage().getBytes()).encode())
                        .encode();
    }

    /**
     * Represent the abstract ElGamalCiphertext structure as an element {@code (c1, c2)} in a
     * {@link ee.ivxv.common.math.ProductGroup}.
     * 
     * @return {@link ee.ivxv.common.math.ProductGroup} element representing ciphertext.
     */
    public ProductGroupElement getAsProductGroupElement() {
        ProductGroup pgroup =
                new ProductGroup(getBlind().getGroup(), getBlindedMessage().getGroup());
        ProductGroupElement ret = new ProductGroupElement(pgroup, getBlind(), getBlindedMessage());
        return ret;
    }

    @Override
    public String toString() {
        return String.format("ElGamalCipherText(%s,%s)", blind, blindedMessage);
    }
}
