package ee.ivxv.common.crypto.elgamal;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.PKCS8PrivateKey;
import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.crypto.rnd.NativeRnd;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.IntegerConstructor;
import ee.ivxv.common.math.MathException;
import java.io.IOException;
import java.math.BigInteger;

public class ElGamalPrivateKey {
    private final ElGamalParameters parameters;
    private final BigInteger key;
    private ElGamalPublicKey pub;
    private Rnd rnd = new NativeRnd();

    public ElGamalPrivateKey(ElGamalParameters parameters, BigInteger key)
            throws IllegalArgumentException {
        this.parameters = parameters;
        this.key = key;
    }

    public ElGamalPrivateKey(byte[] packed) throws IllegalArgumentException {
        PKCS8PrivateKey p = new PKCS8PrivateKey();
        try {
            p.readFromBytes(packed);
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Unpacking ElGamalPrivateKey failed: " + e);
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
        key = new BigInteger(keyb);
    }

    public ElGamalPublicKey getPublicKey() {
        if (pub == null) {
            pub = new ElGamalPublicKey(getParameters(), computePublicPart());
        }
        return pub;
    }

    private GroupElement computePublicPart() {
        return getParameters().getGenerator().scale(key);
    }

    public ElGamalParameters getParameters() {
        return this.parameters;
    }

    public BigInteger getSecretPart() {
        return this.key;
    }

    public byte[] getBytes() {
        PKCS8PrivateKey der = new PKCS8PrivateKey(getParameters().getOID(), key.toByteArray(),
                getParameters().getBytes());
        return der.encode();
    }

    @Override
    public String toString() {
        return String.format("ElGamalPrivateKey(%s, %s)", parameters, key);
    }

    public Plaintext decrypt(ElGamalCiphertext ct) throws MathException {
        GroupElement msg = ct.getBlindedMessage().op(ct.getBlind().scale(key).inverse());
        Plaintext decoded = getParameters().getGroup().decode(msg);
        Plaintext unpadded = decoded.stripPadding();
        return unpadded;
    }

    public ElGamalDecryptionProof provableDecrypt(ElGamalCiphertext ct)
            throws MathException, IOException {
        Plaintext plaintext = decrypt(ct);
        ElGamalDecryptionProof proof = new ElGamalDecryptionProof(ct, plaintext, getPublicKey());
        BigInteger r = IntegerConstructor.construct(rnd, getParameters().getGeneratorOrder());
        GroupElement a = ct.getBlind().scale(r);
        GroupElement b = getParameters().getGenerator().scale(r);
        proof.setMessageCommitment(a);
        proof.setKeyCommitment(b);
        BigInteger k = proof.computeChallenge();
        BigInteger s = k.multiply(key).add(r).mod(getParameters().getGeneratorOrder());
        proof.setResponse(s);
        return proof;
    }
}
