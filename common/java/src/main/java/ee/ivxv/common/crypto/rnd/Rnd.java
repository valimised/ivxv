package ee.ivxv.common.crypto.rnd;

import java.io.IOException;

/**
 * Define the methods required from any random bytes source.
 */
public interface Rnd {
    /**
     * Write the random bytes into output buffer.
     * 
     * @param buf The output buffer.
     * @param offset The offset to start writing at.
     * @param len The amount of random bytes requested.
     * @return The number of actual bytes written.
     * @throws IOException
     */
    int read(byte[] buf, int offset, int len) throws IOException;

    /**
     * Write the random bytes into output buffer, blocking if the requested amount is not available
     * in full.
     * 
     * @param buf The output buffer.
     * @param offset The offset to start writing at.
     * @param len The amount of random bytes requested.
     * @return The number of bytes requested.
     * @throws IOException
     */
    int mustRead(byte[] buf, int offset, int len) throws IOException;

    /**
     * @return Boolean indicating if the total number of bytes the instance can return is bounded.
     */
    boolean isFinite();

    /**
     * Close the random source.
     */
    void close();
}
