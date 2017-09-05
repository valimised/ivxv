package ee.ivxv.audit.shuffle;

import ee.ivxv.common.util.Util;
import java.security.MessageDigest;

/**
 * Class PRNG implements pseudo-random number generator as defined in Verificatum independent
 * verifier implementation description.
 */
public class PRNG {
    private MessageDigest cleanhash;
    private byte[] buf;
    private int it;
    private int bufp;
    private int digestLen;

    /**
     * Initialize PRNG using a hash function and a seed.
     * 
     * @param hashname Defined hash function.
     * @param seed Seed to initialize PRNG with.
     */
    PRNG(String hashname, byte[] seed) {
        this.cleanhash = init_hash(hashname, seed);
        this.digestLen = cleanhash.getDigestLength();
        this.buf = new byte[digestLen];
        this.it = 0;
        this.bufp = digestLen;
    }

    /**
     * Fill the output buffer with bytes from the PRNG.
     * 
     * @param out Output buffer to fill.
     */
    public void read(byte[] out) {
        int read = 0;
        int to_read;
        int len = out.length;
        while (read < len) {
            refill();
            to_read = Math.min(len - read, digestLen - bufp);
            System.arraycopy(buf, bufp, out, read, to_read);
            bufp = bufp + to_read;
            read += to_read;
        }
    }

    private void refill() {
        if (bufp < digestLen) {
            return;
        }
        MessageDigest h;
        try {
            h = (MessageDigest) cleanhash.clone();
        } catch (CloneNotSupportedException e) {
            // already checked
            return;
        }
        buf = h.digest(Util.toBytes(it));
        bufp = 0;
        it += 1;
    }

    static MessageDigest init_hash(String hashname, byte[] seed) {
        MessageDigest md = DataParser.getHash(hashname);
        try {
            md.clone();
        } catch (CloneNotSupportedException e) {
            throw new IllegalArgumentException(
                    String.format("Hash function %s support is incomplete", hashname));
        }
        if (seed != null) {
            md.update(seed);
        }
        return md;
    }

}
