package ee.ivxv.key.protocol;

import java.io.IOException;

public interface GenerationProtocol {
    byte[] generateKey() throws ProtocolException, IOException;
}
