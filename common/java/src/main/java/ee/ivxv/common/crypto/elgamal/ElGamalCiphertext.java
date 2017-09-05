package ee.ivxv.common.crypto.elgamal;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Ciphertext;
import ee.ivxv.common.asn1.Sequence;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.ProductGroup;
import ee.ivxv.common.math.ProductGroupElement;

public class ElGamalCiphertext {
    // given a message m and public key h and randomness r and generator g, the
    // ElGamal ciphertext is a tuple c = (c1, c2) = (g^r, m * h^r).
    //
    // we call the variable c1 as 'blind', denoting that it is just that
    //
    // we also call the variable c2 as 'blindedMessage', given that it includes
    // the message, which is 'blinded' by some other value
    //
    // we also need the oid of the cryptosystem (per specification)
    private final GroupElement blind;
    private final GroupElement blindedMessage;
    private final String oid;

    public ElGamalCiphertext(GroupElement blind, GroupElement blindedMessage, String oid) {
        this.blind = blind;
        this.blindedMessage = blindedMessage;
        this.oid = oid;
    }

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

    public GroupElement getBlind() {
        return blind;
    }

    public GroupElement getBlindedMessage() {
        return blindedMessage;
    }

    public String getOID() {
        return oid;
    }

    public byte[] getBytes() {
        return new Ciphertext(getOID(),
                new Sequence(getBlind().getBytes(), getBlindedMessage().getBytes()).encode())
                        .encode();
    }

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
