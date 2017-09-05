package ee.ivxv.key.protocol;

import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import java.io.IOException;

public interface DecryptionProtocol {
    ElGamalDecryptionProof decryptMessage(byte[] msg) throws ProtocolException, IOException;
}
