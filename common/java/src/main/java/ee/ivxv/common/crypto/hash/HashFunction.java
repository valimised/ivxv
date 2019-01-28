package ee.ivxv.common.crypto.hash;

import java.io.IOException;
import java.io.InputStream;

/**
 * Interface for hash functions.
 *
 */
public interface HashFunction {
    /**
     * Construct digest from byte array.
     * 
     * @param input
     * @return Digest of the input.
     */
    byte[] digest(byte[] input);

    /**
     * Construct digest from input stream.
     * 
     * @param input
     * @return Digest of the input.
     * @throws IOException When stream read fails.
     */
    byte[] digest(InputStream input) throws IOException;
}
