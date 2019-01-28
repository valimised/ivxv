package ee.ivxv.key.protocol;

import ee.ivxv.common.crypto.CorrectnessUtil;
import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import java.io.IOException;

/**
 * DecryptionProtocol defines interface which the protocols for decrypting ciphertext must
 * implement.
 */
public interface DecryptionProtocol {
    /**
     * Take in a ciphertext as bytes and output a proof of correct decryption.
     * <p>
     * The returned {@link ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof} does not need to be
     * complete. If the protocol implementation does not allow for proving the correctness, then the
     * corresponding fields should be left null. Also, if the protocol can be initialized for not
     * providing proofs.
     * <p>
     * The fields for message, ciphertext and public key must be set.
     * 
     * @param msg
     * @return
     * @throws ProtocolException
     * @throws IOException
     */
    ElGamalDecryptionProof decryptMessage(byte[] msg) throws ProtocolException, IOException;

    /**
     * Check if the ciphertext could be decrypted using the protocol.
     * <p>
     * Check if the input ciphertext is correctly serialized and part of the ciphertext space. If
     * the result is valid, then a call to {@link #checkCorrectness(byte[])} decrypts correctly the
     * ciphertext.
     * <p>
     * This check does not actually decrypt the ciphertext. Thus, the decrypted message may not be
     * correctly padded or contain valuable information.
     * 
     * @param msg
     * @return
     * @throws ProtocolException
     */
    CorrectnessUtil.CiphertextCorrectness checkCorrectness(byte[] msg) throws ProtocolException;
}
