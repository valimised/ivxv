package ee.ivxv.common.crypto.rnd;

import java.security.NoSuchAlgorithmException;
import java.security.NoSuchProviderException;
import java.security.SecureRandom;

/**
 * NativeRnd is a entropy source which reads the random bytes directly from system.
 */
public class NativeRnd implements Rnd {
    private SecureRandom src;

    /**
     * Initialize the source. Internally, different interfaces used depending on the operating
     * system. For Windows systems, {@link #initWin()} is called. For other systems (Linux,BSD,OS
     * X), {@link #initLinux()} is called.
     * 
     * @throws RuntimeException On exception during initializing the system source.
     */
    public NativeRnd() {
        String os_name = System.getProperty("os.name").toLowerCase();
        try {
            if (os_name.contains("windows")) {
                initWin();
            } else {
                initLinux();
            }
        } catch (NoSuchAlgorithmException | NoSuchProviderException ex) {
            throw new RuntimeException("Failed initializing native entropy source", ex);
        }
    }

    /**
     * Initialize {@code Windows-PRNG} instance. The instance uses {@code CryptGenRandom} from
     * Windows cryptographic API to obtain the random bytes.
     * 
     * @throws NoSuchAlgorithmException.
     * @throws NoSuchProviderException.
     */
    private void initWin() throws NoSuchAlgorithmException, NoSuchProviderException {
        src = SecureRandom.getInstance("Windows-PRNG");
    }

    /**
     * Initializes {@code NativePRNGNonBlocking} instance. The instance uses the non-blocking kernel
     * interface to obtain the random bytes.
     * 
     * @throws NoSuchAlgorithmException.
     * @throws NoSuchProviderException.
     */
    private void initLinux() throws NoSuchAlgorithmException, NoSuchProviderException {
        src = SecureRandom.getInstance("NativePRNGNonBlocking");
    }

    /**
     * Read the bytes from the system random provider.
     * 
     * @param buf The output buffer to write the bytes into.
     * @param offset The offset at the output buffer to start writing the random bytes into.
     * @param len The amount of bytes to be read.
     * @return The requested amount {@code len}.
     */
    @Override
    public int read(byte[] buf, int offset, int len) {
        byte[] tmp = new byte[len];
        src.nextBytes(tmp);
        System.arraycopy(tmp, 0, buf, offset, len);
        return len;
    }

    /**
     * Returns {@link read(buf, offset, len)}.
     * 
     * @param buf The output buffer to write the bytes into.
     * @param offset The offset at the output buffer to start writing the random bytes into.
     * @param len The amount of bytes to be read.
     * @return The requested amount {@code len}.
     */
    @Override
    public int mustRead(byte[] buf, int offset, int len) {
        return read(buf, offset, len);
    }

    /**
     * @return false.
     */
    @Override
    public boolean isFinite() {
        return false;
    }

    /**
     * A no-op.
     */
    @Override
    public void close() {}
}
