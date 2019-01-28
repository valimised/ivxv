package ee.ivxv.common.crypto.rnd;

import java.io.IOException;
import java.nio.ByteBuffer;
import java.nio.channels.FileChannel;
import java.nio.file.Path;
import java.nio.file.StandardOpenOption;
import java.security.MessageDigest;

/**
 * DPRNG is a deterministic pseudorandom number generator. It is an entropy source which extends the
 * input seed.
 * <p>
 * It is intended to be used together with {@link CombineRnd}, adding this as a source to it.
 */
public class DPRNG implements Rnd {
    private MessageDigest dgst;
    private int dgstLen;
    private byte[] seed;
    private long counter;
    private byte[] buffer;
    private int bufferptr;

    /**
     * Initialize the DPRNG from a seed.
     * 
     * @param seed The seed.
     */
    public DPRNG(byte[] seed) {
        seedInit(seed);
    }

    /**
     * Initialize the DPRNG from the seed read from a file. A digest of the file is computed and the
     * digest is used as seed.
     * 
     * @param rndSource The path for the seed file.
     * @throws IOException On exception while reading the file.
     */
    public DPRNG(Path rndSource) throws IOException {
        byte[] seed = computeFileDigest(rndSource);
        seedInit(seed);
    }

    /**
     * @return false
     */
    @Override
    public boolean isFinite() {
        return false;
    }

    /**
     * Store the seed and set internals.
     * 
     * @param seed The seed value.
     */
    private void seedInit(byte[] seed) {
        this.dgst = getDigestInstance();
        this.dgstLen = dgst.getDigestLength();
        this.counter = 1;
        this.seed = seed.clone();
        this.buffer = new byte[dgstLen];
        this.bufferptr = Integer.MAX_VALUE;
    }

    /**
     * Compute digest of a file.
     * 
     * @param path Path leading to file.
     * @return The digest value.
     * @throws IOException On exception during reading the file.
     */
    private byte[] computeFileDigest(Path path) throws IOException {
        MessageDigest d = getDigestInstance();
        ByteBuffer b = ByteBuffer.allocate(1024);
        FileChannel f = FileChannel.open(path, StandardOpenOption.READ);
        while (f.read(b) >= 0) {
            b.flip();
            d.update(b);
            b.clear();
        }
        return d.digest();
    }

    /**
     * Read bytes from DPRNG. Given the {@code seed}, the output is obtained sequentially from the
     * infinite byte array <code>H(1||seed) || H(2||seed) || ..</code>.
     * <p>
     * The hash function <code>H</code> used is given in {@link #getDigestInstance()}.
     * <p>
     * This method is not thread-safe and synchronization must be handled by the caller.
     * 
     * @param buf The output buffer to store the value in.
     * @param off The offset in output buffer for storing the value.
     * @param len The amount of bytes to store in the output buffer.
     * @return The requested amount {@code len}.
     */
    @Override
    public int read(byte[] buf, int off, int len) {
        int read = 0;
        int to_read;
        while (read < len) {
            refill();
            to_read = Math.min(len - read, dgstLen - bufferptr);
            System.arraycopy(buffer, bufferptr, buf, off + read, to_read);
            bufferptr = bufferptr + to_read;
            read += to_read;
        }
        return len;
    }

    /**
     * Returns {@link #read(byte[], int, int)}
     * <p>
     * This method is not thread-safe and synchronization must be handled by the caller.
     * 
     * @param buf The output buffer to store the value in.
     * @param off The offset in output buffer for storing the value.
     * @param len The amount of bytes to store in the output buffer.
     * @return The requested amount {@code len}.
     */
    @Override
    public int mustRead(byte[] buf, int off, int len) {
        return read(buf, off, len);
    }

    /**
     * @return SHA-256 digest instance.
     */
    private MessageDigest getDigestInstance() {
        try {
            return MessageDigest.getInstance("SHA-256");
        } catch (java.security.NoSuchAlgorithmException e) {
            // this should not happen, SHA-256 is widespread
            // we throw a generic exception so that we wouldn't need to add
            // exception checking mess everywhere
            throw new RuntimeException("SHA-256 digest algorithm not found");
        }
    }

    private void refill() {
        if (bufferptr < dgstLen) {
            return;
        }
        byte[] rndSeed = getRoundInput();
        buffer = dgst.digest(rndSeed);
        bufferptr = 0;
        counter += 1;
    }

    private byte[] getRoundCounter() {
        byte[] ret = new byte[8];
        ret[0] = (byte) ((counter >>> 56) & 0xff);
        ret[1] = (byte) ((counter >>> 48) & 0xff);
        ret[2] = (byte) ((counter >>> 40) & 0xff);
        ret[3] = (byte) ((counter >>> 32) & 0xff);
        ret[4] = (byte) ((counter >>> 24) & 0xff);
        ret[5] = (byte) ((counter >>> 16) & 0xff);
        ret[6] = (byte) ((counter >>> 8) & 0xff);
        ret[7] = (byte) ((counter >>> 0) & 0xff);
        return ret;
    }

    private byte[] getRoundInput() {
        byte[] packedCounter = getRoundCounter();
        byte[] ret = new byte[packedCounter.length + seed.length];
        System.arraycopy(packedCounter, 0, ret, 0, packedCounter.length);
        System.arraycopy(seed, 0, ret, packedCounter.length, seed.length);
        return ret;
    }

    /**
     * Close random source.
     */
    @Override
    public void close() {
        // no-op
    }
}
