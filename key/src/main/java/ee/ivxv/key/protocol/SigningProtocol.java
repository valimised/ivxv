package ee.ivxv.key.protocol;

import org.bouncycastle.operator.ContentSigner;

/**
 * Interface for signing messages.
 */
public interface SigningProtocol extends ContentSigner {
    /**
     * Sign the message using the signing protocol and return serialized signature.
     * 
     * @param msg Message to be signed.
     * @return Serialized signature.
     * @throws ProtocolException When exception occurs during protocol run.
     */
    byte[] sign(byte[] msg) throws ProtocolException;
}
