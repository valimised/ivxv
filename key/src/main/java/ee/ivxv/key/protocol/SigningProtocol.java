package ee.ivxv.key.protocol;

import org.bouncycastle.operator.ContentSigner;

public interface SigningProtocol extends ContentSigner {
    byte[] sign(byte[] msg) throws ProtocolException;
}
