package ee.ivxv.common.crypto.elgamal;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.PKCS8PrivateKey;
import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.crypto.rnd.NativeRnd;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.IntegerConstructor;
import ee.ivxv.common.math.MathException;
import ee.ivxv.common.math.Group.Decodable;
import ee.ivxv.common.util.Util;
import java.io.IOException;
import java.math.BigInteger;
import java.nio.file.Path;

/**
 * ElGamalPrivateKey is the private key for decrypting ElGamal ciphertexts.
 */
public class ElGamalPrivateKey {
    private final ElGamalParameters parameters;
    private final BigInteger key;
    private ElGamalPublicKey pub;
    private Rnd rnd = new NativeRnd();

    /**
     * Initialize using ElGamal parameters and the private key.
     * 
     * @param parameters
     * @param key
     * @throws IllegalArgumentException
     */
    public ElGamalPrivateKey(ElGamalParameters parameters, BigInteger key)
            throws IllegalArgumentException {
        this.parameters = parameters;
        this.key = key;
    }

    /**
     * Initialize using serialized private key.
     * 
     * @see #getBytes()
     * 
     * @param packed
     * @throws IllegalArgumentException When parsing fails
     */
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

    /**
     * Initialize using armored private key file.
     * <p>
     * The input key file should be MIME-encoded private key file with private key headers.
     * 
     * @param p
     * @throws IllegalArgumentException When decoding or parsing fails.
     * @throws IOException When reading fails
     */
    public ElGamalPrivateKey(Path p) throws IllegalArgumentException, IOException {
        this(Util.decodePrivateKey(p));
    }

    /**
     * Compute the public key which corresponds to the private key.
     * 
     * @return
     */
    public ElGamalPublicKey getPublicKey() {
        if (pub == null) {
            pub = new ElGamalPublicKey(getParameters(), computePublicPart());
        }
        return pub;
    }

    private GroupElement computePublicPart() {
        return getParameters().getGenerator().scale(key);
    }

    /**
     * Get the key parameters.
     * 
     * @return
     */
    public ElGamalParameters getParameters() {
        return this.parameters;
    }

    /**
     * Get the key value.
     * 
     * @return
     */
    public BigInteger getSecretPart() {
        return this.key;
    }

    /**
     * Serialize the private as PKCS8 private key
     * <p>
     * The key is represented as ASN1 INTEGER and parameters as ElGamalParameters.
     * 
     * @see ee.ivxv.common.asn1.PKCS8PrivateKey
     * @see ee.ivxv.common.crypto.elgamal.ElGamalParameters#getBytes()
     * 
     * @return
     */
    public byte[] getBytes() {
        PKCS8PrivateKey der = new PKCS8PrivateKey(getParameters().getOID(), key.toByteArray(),
                getParameters().getBytes());
        return der.encode();
    }

    @Override
    public String toString() {
        return String.format("ElGamalPrivateKey(%s, %s)", parameters, key);
    }

    /**
     * Decrypt a ciphertext using the private key.
     * <p>
     * This method performs checks to ensure that the ciphertext is correct.
     * 
     * @param ct Ciphertext.
     * @return Decrypted message
     * @throws MathException When decryption fails.
     */
    public Plaintext decrypt(ElGamalCiphertext ct) throws MathException {
        return decrypt(ct, false);
    }

    /**
     * Decrypt a ciphertext using the private key, possibly omitting checks.
     * <p>
     * If the user is sure that the ciphertext is correct, then it is possible to omit correctness
     * checking, gaining a 3x speedup during decryption. If the ciphertext is identified falsely as
     * correct, then the outcome is undefined.
     * 
     * @param ct Ciphertext.
     * @param assumeDecodable If true, omit decodability checks.
     * @return Decrypted message
     * @throws MathException When decryption fails.
     */
    public Plaintext decrypt(ElGamalCiphertext ct, boolean assumeDecodable) throws MathException {
        if (!assumeDecodable
                && getParameters().getGroup().isDecodable(ct.getBlind()) != Decodable.VALID) {
            throw new MathException("Blind is not decodable");
        }
        GroupElement msg = ct.getBlindedMessage().op(ct.getBlind().scale(key).inverse());
        Plaintext decoded;
        if (assumeDecodable || getParameters().getGroup().isDecodable(msg) == Decodable.VALID) {
            decoded = getParameters().getGroup().decode(msg);
        } else {
            throw new MathException("Message is not decodable");
        }
        Plaintext unpadded = decoded.stripPadding();
        return unpadded;
    }

    /**
     * Decrypt a ciphertext and provide a proof of correct decryption.
     * <p>
     * See the documentation of decryption proof for the explanation of the values of the proof.
     * This method does not check that the message is actually decodable.
     * 
     * @see ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof
     * 
     * @param ct Ciphertext
     * @return Proof of correct decryption (includes decrypted message)
     * @throws MathException When computation fails.
     * @throws IOException When can not read entropy for constructing commitment.
     */
    public ElGamalDecryptionProof provableDecrypt(ElGamalCiphertext ct)
            throws MathException, IOException {
        // provable decryption must perform all checks - but currently we still
        // assume that the ciphertext is decodable
        Plaintext plaintext = decrypt(ct, true);
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
