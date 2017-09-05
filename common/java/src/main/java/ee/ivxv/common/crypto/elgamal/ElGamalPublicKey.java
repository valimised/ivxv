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
import java.io.IOException;
import java.math.BigInteger;
import java.security.KeyException;

public class ElGamalPublicKey {
    private final ElGamalParameters parameters;
    private final GroupElement key;

    public ElGamalPublicKey(ElGamalParameters parameters, GroupElement key)
            throws IllegalArgumentException {
        this.parameters = parameters;
        this.key = key;
    }

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

    public ElGamalParameters getParameters() {
        return parameters;
    }

    public GroupElement getKey() {
        return key;
    }

    public byte[] getBytes() {
        X509PublicKey der = new X509PublicKey(getParameters().getOID(), key.getBytes(),
                getParameters().getBytes());
        return der.encode();
    }

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

    public ElGamalCiphertext encrypt(Plaintext msg) throws KeyException, IOException {
        return encrypt(msg, new NativeRnd());
    }

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
