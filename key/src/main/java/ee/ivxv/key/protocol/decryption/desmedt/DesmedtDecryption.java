package ee.ivxv.key.protocol.decryption.desmedt;

import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import ee.ivxv.key.protocol.DecryptionProtocol;

// this method combines the decryption shares into a decryption.
public class DesmedtDecryption implements DecryptionProtocol {
    public byte[] getDecryptionShare(int party, byte[] msg) {
        return null;
    }

    @Override
    public ElGamalDecryptionProof decryptMessage(byte[] msg) {
        return null;
    }
}
