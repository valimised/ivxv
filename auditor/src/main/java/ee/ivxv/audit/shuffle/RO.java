package ee.ivxv.audit.shuffle;

import ee.ivxv.common.util.Util;
import java.security.MessageDigest;

/**
 * Implements Random Oracle as defined in Verificatum manual for implementing independent verifier.
 */
public class RO {
    private String hashname;
    private MessageDigest cleanhash;
    private byte[] seed;

    /**
     * Initialize RO using a hashname and a seed.
     * 
     * @param hashname Hash function to use.
     * @param seed Seed bytes.
     */
    RO(String hashname, byte[] seed) {
        this.hashname = hashname;
        this.seed = seed.clone();
        this.cleanhash = PRNG.init_hash(hashname, null);
    }

    private byte[] seed(int amount) {
        MessageDigest h;
        try {
            h = (MessageDigest) cleanhash.clone();
        } catch (CloneNotSupportedException e) {
            // already checked
            return null;
        }
        h.update(Util.toBytes(amount));
        h.update(seed);
        return h.digest();
    }

    /**
     * Fill the output buffer with bytes from the RO.
     * 
     * @param out Output buffer to fill.
     */
    public void read(byte[] out, int amount) {
        if ((amount + 7) / 8 != out.length) {
            throw new IllegalArgumentException(
                    "Output buffer length does not correspond to requested read amount");
        }
        byte[] s = seed(amount);
        PRNG p = new PRNG(hashname, s);
        p.read(out);
        if (amount % 8 != 0) {
            out[0] &= (1 << (amount % 8)) - 1;
        }
    }
}
