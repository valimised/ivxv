package ee.ivxv.key.protocol.decryption.desmedt;

import ee.ivxv.common.crypto.CorrectnessUtil.CiphertextCorrectness;
import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import ee.ivxv.key.protocol.DecryptionProtocol;

/**
 * DesmedtDecryption combines the decryption shares into a decrypted message without recovering the
 * key. Currently this protocol is not implemented.
 */
public class DesmedtDecryption implements DecryptionProtocol {
    /**
     * Get the corresponding decryption share.
     * <p>
     * Currently, this method is not implemented and returns null.
     * 
     * @param party
     * @param msg
     * @return
     */
    public byte[] getDecryptionShare(int party, byte[] msg) {
        return null;
    }

    /**
     * Get the decrypted message.
     * <p>
     * Decryption using DesmedDecryption is not implemented and the method returns null.
     * 
     * @param msg
     */
    @Override
    public ElGamalDecryptionProof decryptMessage(byte[] msg) {
        return null;
    }

    /**
     * Check the correctness of the message.
     * <p>
     * Checking correctness using DesmedtDecryption is not implemented and the method returns null.
     * 
     * @param msg
     */
    @Override
    public CiphertextCorrectness checkCorrectness(byte[] msg) {
        return null;
    }
}
