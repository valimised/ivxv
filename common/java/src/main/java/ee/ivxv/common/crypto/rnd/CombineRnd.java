package ee.ivxv.common.crypto.rnd;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import org.bouncycastle.crypto.digests.SHAKEDigest;

/**
 * CombineRnd reads bytes from the source and mixes them securely. The output is as predictable as
 * is the least predictable source.
 */
public class CombineRnd implements Rnd {
    private List<Rnd> sources;
    private SHAKEDigest combiner;
    private byte[] buf = new byte[FileRnd.BUF_SIZE];
    private int cwritten;
    private final int RATE;

    /**
     * Initialize the instance. The instance is not yet seeded. To add sources and seed the
     * instance, call {@link #addSource(Rnd)} method with actual entropy sources.
     */
    public CombineRnd() {
        sources = new ArrayList<Rnd>();
        combiner = new SHAKEDigest(256);
        cwritten = 0;
        RATE = combiner.getByteLength();
    }

    /**
     * @return false.
     */
    @Override
    public boolean isFinite() {
        return false;
    }

    /**
     * Add a entropy source into the pool. During reading, every source is asked for entropy.
     * 
     * @param rnd The configured entropy source. The {@link CombineRnd} instances can be chained.
     * @throws IOException On IOException while reading from finite source.
     */
    public void addSource(Rnd rnd) throws IOException {
        if (rnd.isFinite()) {
            addFiniteSource(rnd);
        } else {
            addInfiniteSource(rnd);
        }
    }

    /**
     * Add infinite source to the sources. The source is read only when invoking
     * {@link #read(byte[], int, int)}.
     * 
     * @param rnd The infinite entropy source.
     */
    private void addInfiniteSource(Rnd rnd) {
        sources.add(rnd);
    }

    /**
     * Add finite source. The whole source is read and mixed into the entropy pool. It is not read
     * during invoking {@link #read(byte[], int, int)}.
     * 
     * @param rnd The finite entropy source.
     * @throws IOException On exception during reading the source.
     */
    private void addFiniteSource(Rnd rnd) throws IOException {
        int p;
        while ((p = rnd.read(buf, 0, buf.length)) > 0) {
            write(buf, 0, p);
        }
    }

    /**
     * Fill the buffer with random bytes. Before every read, available bytes from every infinite
     * entropy source is read and mixed into the pool.
     * 
     * @param out The output buffer to store the random bytes in.
     * @param offset The offset at which to start storing the bytes in output buffer.
     * @param len The amount of bytes to store in the output buffer.
     * @return The requested amount {@code len}.
     * @throws IOException On exception during reading the entropy source.
     */
    @Override
    public int read(byte[] out, int offset, int len) throws IOException {
        return read(out, offset, len, false);
    }

    /**
     * Fill the buffer with random bytes. Before every read, exactly requested number of bytes from
     * every infinite entropy source is read and mixed into the pool.
     * 
     * @param out The output buffer to store the random bytes in.
     * @param offset The offset at which to start storing the bytes in output buffer.
     * @param len The amount of bytes to store in the output buffer.
     * @return The requested amount {@code len}.
     * @throws IOException On exception during reading the entropy source.
     */
    @Override
    public int mustRead(byte[] out, int offset, int len) throws IOException {
        return read(out, offset, len, true);
    }

    /**
     * Read from sources and mix into the pool. The number of bytes read from every soure depends on
     * the {@code must} parameter.
     * 
     * @param out The output buffer to store the random bytes in.
     * @param offset The offset at which to start storing the bytes in output buffer.
     * @param len The amount of bytes to store in the output buffer.
     * @param must Denote if must read exactly the amount of bytes from every source.
     * @return The requested amount {@code len}.
     * @throws IOException On exception during reading the entropy source.
     */
    private int read(byte[] out, int offset, int len, boolean must) throws IOException {
        updateFromAllSources(len, must);
        computeDigest(out, offset, len);
        padAndFill();
        writeZeros(len);
        return len;
    }

    /**
     * Write bytes into the entropy pool.
     * 
     * @param buf The source buffer to read bytes from.
     * @param offset The source offset in the source buffer to start reading the bytes.
     * @param len The length of bytes to read from the source buffer.
     */
    public void write(byte[] buf, int offset, int len) {
        combiner.update(buf, offset, len);
        cwritten = (cwritten + len - offset) % RATE;
    }

    /**
     * Convenience method for {@link #write(buf, 0, buf.length)}.
     * 
     * @param buf The source buffer to read bytes from.
     */
    public void write(byte[] buf) {
        write(buf, 0, buf.length);
    }

    /**
     * Read from the entropy source and mix the read bytes into the entropy pool.
     * 
     * @param source The entropy source to read from.
     * @param len The maximum amount of bytes to read. If bytes are not available, then actual
     *        number of read bytes may be smaller.
     * @param must If must read until obtain the {@code len} amount of bytes.
     * @throws IOException On exception while reading from the source.
     */
    private void updateFromSource(Rnd source, int len, boolean must) throws IOException {
        int read = 0;
        int p;
        do {
            if (must) {
                p = source.mustRead(buf, 0, len - read < buf.length ? len - read : buf.length);

            } else {
                p = source.read(buf, 0, len - read < buf.length ? len - read : buf.length);
            }
            read += p;
            write(buf, 0, p);
        } while (p > 0 && read < len);
    }

    /**
     * Read from all sources.
     * 
     * @param len The maximum amount of bytes to read from every source. If bytes are not available,
     *        then actual number of read bytes may be smaller.
     * @param must If must read until obtain the {@code len} amount of bytes.
     * @throws IOException On exception while reading from the source.
     */
    private void updateFromAllSources(int len, boolean must) throws IOException {
        for (Rnd source : sources) {
            updateFromSource(source, len, must);
        }
    }

    /**
     * Add the SHAKE-256 padding to the entropy pool and fill the rest of the block with zeros. The
     * padding separates the SHAKE-256 absorbing and squeezing states.
     */
    private void padAndFill() {
        Arrays.fill(buf, (byte) 0);
        buf[cwritten] = 0x1f;
        buf[RATE - 1] |= (byte) 0x80;
        combiner.update(buf, cwritten, RATE - cwritten);
        cwritten = 0;
    }

    /**
     * Write zeros to entropy pool.
     * 
     * @param len The amount of zeros to write.
     */
    private void writeZeros(int len) {
        // we write as much zeros to vanilla copy as we are reading
        int written = 0;
        int to_write;
        Arrays.fill(buf, (byte) 0);
        while (written < len) {
            to_write = len - written < buf.length ? len - written : buf.length;
            combiner.update(buf, 0, to_write);
            written += to_write;
            cwritten = (cwritten + to_write) % RATE;
        }
    }

    /**
     * Copy the entropy pool and use the copy to compute the digest.
     * 
     * @param out The output buffer to write the digest to.
     * @param offset The start location in the output buffer to start writing the digest.
     * @param len The length of requested digest.
     */
    private void computeDigest(byte[] out, int offset, int len) {
        // copy the vanilla digest and use the copy to compute the output
        SHAKEDigest tmpd = new SHAKEDigest(combiner);
        tmpd.doOutput(out, offset, len);
    }

    /**
     * Close all underlying sources.
     */
    @Override
    public void close() {
        sources.forEach(source -> source.close());
    }
}
