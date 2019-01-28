package ee.ivxv.key.protocol;

import java.io.IOException;

/**
 * Interface for constructing a key pair usable for signing or encryption.
 */
public interface GenerationProtocol {
    /**
     * Generate the key and output serialized public key.
     * <p>
     * The serialized public key depends on the protocol and underlying crypto system.
     * 
     * @return
     * @throws ProtocolException
     * @throws IOException
     */
    byte[] generateKey() throws ProtocolException, IOException;
}
