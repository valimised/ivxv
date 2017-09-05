package ee.ivxv.common.crypto.hash;

import java.io.IOException;
import java.io.InputStream;

public interface HashFunction {
    byte[] digest(byte[] input);

    byte[] digest(InputStream input) throws IOException;
}
