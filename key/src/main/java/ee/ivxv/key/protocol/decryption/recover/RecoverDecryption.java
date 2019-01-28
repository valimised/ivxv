package ee.ivxv.key.protocol.decryption.recover;

import ee.ivxv.common.crypto.CorrectnessUtil;
import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.crypto.elgamal.ElGamalCiphertext;
import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import ee.ivxv.common.crypto.elgamal.ElGamalParameters;
import ee.ivxv.common.crypto.elgamal.ElGamalPrivateKey;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.math.LagrangeInterpolation;
import ee.ivxv.common.math.MathException;
import ee.ivxv.common.service.smartcard.IndexedBlob;
import ee.ivxv.key.protocol.DecryptionProtocol;
import ee.ivxv.key.protocol.ProtocolException;
import ee.ivxv.key.protocol.ThresholdParameters;
import java.io.IOException;
import java.math.BigInteger;
import java.util.Set;

/**
 * RecoverDecryption is a protocol for decrypting the ciphertext by reconstructing the private key
 * from private key shares.
 */
public class RecoverDecryption implements DecryptionProtocol {
    private final Set<IndexedBlob> blobs;
    private final ThresholdParameters tparams;
    private final boolean withProof;
    private ElGamalPrivateKey sk;

    /**
     * Initialize the protocol from values.
     * 
     * @param blobs The set of blobs that contain the private key shares.
     * @param tparams Parameters for threshold decryption.
     * @throws ProtocolException When the number of cards is less than is required for decryption.
     */
    public RecoverDecryption(Set<IndexedBlob> blobs, ThresholdParameters tparams)
            throws ProtocolException {
        this(blobs, tparams, true);
    }

    /**
     * Initialize the protocol from values.
     * 
     * @param blobs The set of blobs that contain the private key shares.
     * @param tparams Parameters for threshold decryption.
     * @param withProof Boolean indicating if decrypting without providing proofs of correct
     *        decryption.
     * @throws ProtocolException When the number of cards is less than is required for decryption.
     */
    public RecoverDecryption(Set<IndexedBlob> blobs, ThresholdParameters tparams, boolean withProof)
            throws ProtocolException {
        if (blobs.size() < tparams.getThreshold()) {
            throw new ProtocolException("Fewer cards available than threshold");
        }
        this.blobs = blobs;
        this.tparams = tparams;
        this.withProof = withProof;
        recoverKey();
    }

    private void recoverKey() throws ProtocolException {
        if (this.sk == null) {
            this.sk = forceKeyRecover();
        }
    }

    private ElGamalPrivateKey forceKeyRecover() throws ProtocolException {
        ElGamalPrivateKey[] parsedKeys = parseAllBlobs(blobs);
        if (!filterKeys(parsedKeys)) {
            throw new ProtocolException("Key share parameters mismatch");
        }
        ElGamalPrivateKey secretKey = combineKeys(parsedKeys);
        return secretKey;
    }

    private ElGamalPrivateKey[] parseAllBlobs(Set<IndexedBlob> blobs) throws ProtocolException {
        ElGamalPrivateKey[] parsed = new ElGamalPrivateKey[tparams.getParties()];
        for (IndexedBlob blob : blobs) {

            int i = blob.getIndex();
            try {
                parsed[i - 1] = new ElGamalPrivateKey(blob.getBlob());
            } catch (IllegalArgumentException e) {
                throw new ProtocolException(
                        "Exception while parsing secret share: " + e.toString());
            }

        }
        return parsed;
    }

    private boolean filterKeys(ElGamalPrivateKey[] keys) throws ProtocolException {
        // this method checks that the parameters for all the keys are the
        // same. If not, then we do not perform any recovery.
        if (keys.length == 0) {
            throw new ProtocolException("No key shares ready for filtering");
        }
        ElGamalParameters params = null;
        for (ElGamalPrivateKey key : keys) {
            if (key == null) {
                continue;
            }
            if (params == null) {
                params = key.getParameters();
            } else if (!params.equals(key.getParameters())) {
                return false;
            }
        }
        return true;
    }

    private ElGamalPrivateKey combineKeys(ElGamalPrivateKey[] keys) throws ProtocolException {
        if (keys.length == 0) {
            throw new ProtocolException("No key shares ready for filtering");
        }
        if (keys.length < tparams.getThreshold()) {
            throw new ProtocolException("Fewer keys available than is the threshold");
        }
        ElGamalParameters params = null;
        BigInteger k = BigInteger.ZERO;
        for (int i = 0; i < keys.length; i++) {
            if (keys[i] == null) {
                continue;
            }
            params = keys[i].getParameters();
            BigInteger p = keys[i].getSecretPart();
            p = p.multiply(LagrangeInterpolation.basisPolynomial(params.getGeneratorOrder(), keys,
                    BigInteger.valueOf(i + 1)));
            k = k.add(p).mod(params.getGeneratorOrder());
        }
        return new ElGamalPrivateKey(params, k);
    }

    /**
     * Decrypt the message using private key reconstructed in memory.
     * <p>
     * If the protocol was initialized to decrypt without proofs, then the corresponding values in
     * the returned value are null.
     * 
     * @param msg Message to be decrypted. Must be serialized instance of
     *        {@link ee.ivxv.common.crypto.elgamal.ElGamalCiphertext}
     * @return Decrypted message
     * @throws ProtocolException If the secret key has not bee decrypted or computation exception
     *         occurs.
     * @throws IllegalArgumentException If invalid input.
     */
    @Override
    public ElGamalDecryptionProof decryptMessage(byte[] msg)
            throws ProtocolException, IllegalArgumentException, IOException {
        if (this.sk == null) {
            throw new ProtocolException("Secret key not reconstructed");
        }
        ElGamalCiphertext ct = new ElGamalCiphertext(this.sk.getParameters(), msg);
        try {
            if (withProof) {
                return this.sk.provableDecrypt(ct);
            } else {
                Plaintext pt = this.sk.decrypt(ct, true);
                return new ElGamalDecryptionProof(ct, pt, this.sk.getPublicKey());
            }
        } catch (MathException e) {
            throw new ProtocolException("Arithmetic error: " + e.toString());
        }
    }

    /**
     * Check the ciphertext correctness.
     * <p>
     * Calls {@link ee.ivxv.common.crypto.CorrectnessUtil#isValidCiphertext(ElGamalPublicKey, byte[])}.
     * 
     * @see ee.ivxv.common.crypto.CorrectnessUtil
     * 
     * @param msg
     * @throws ProtocolException When the protocol is not fully initialized
     */
    @Override
    public CorrectnessUtil.CiphertextCorrectness checkCorrectness(byte[] msg)
            throws ProtocolException {
        if (this.sk == null) {
            throw new ProtocolException("Secret key not reconstructed");
        }
        ElGamalPublicKey pk = sk.getPublicKey();
        return CorrectnessUtil.isValidCiphertext(pk, msg);
    }
}
