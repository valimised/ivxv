package ee.ivxv.audit.shuffle;

import ee.ivxv.common.util.Util;
import java.io.IOException;
import java.io.OutputStream;
import java.security.MessageDigest;

/**
 * Implements Random Oracle as defined in Verificatum manual for implementing independent verifier.
 */
public class RO extends OutputStream {
    private String hashname;
    private MessageDigest hash;
    private byte[] seed;
    private int amount;

    /**
     * Initialize RO using a hashname and a seed.
     * 
     * @param hashname Hash function to use.
     * @param seed Seed bytes.
     */
    public RO(String hashname, byte[] seed) {
        this.hashname = hashname;
        this.flush();
        this.seed = seed.clone();
    }

    /**
     * Initialize RO instance using hashname.
     * 
     * In this mode, the user has to provide the amount to be read and the seed to seed this
     * instance separately.
     * 
     * @param hashname Hash function to use.
     */
    public RO(String hashname) {
        this.hashname = hashname;
        this.flush();
    }

    /**
     * Reset the RO instance
     * 
     */
    public void flush() {
        this.hash = PRNG.init_hash(hashname, null);
        this.amount = 0;
    }

    /**
     * Set the expected amount to be read.
     * 
     * As the amount is written to the digest instance, then it must be set before updating the RO
     * instance.
     * 
     * @param amount
     */
    public void setAmount(int amount) {
        this.amount = amount;
        hash.update(Util.toBytes(amount));
    }



    private void readOut(byte[] out) {
        if ((this.amount + 7) / 8 != out.length) {
            throw new IllegalArgumentException(
                    "Output buffer length does not correspond to requested read amount");
        }
        PRNG p = new PRNG(hashname, hash.digest());
        p.read(out);
        if (amount % 8 != 0) {
            out[0] &= (1 << (amount % 8)) - 1;
        }
    }

    /**
     * Fill the output buffer with bytes from the RO.
     * 
     * If the amount field is 0, then it is assumed that it has been set by a call to
     * {@link #setAmount(int)}. If the seed was given during the initialization, then it is added to
     * the input.
     * 
     * @param out Output buffer to fill.
     * @param amount Number of bits to fill
     */
    public void read(byte[] out, int amount) {
        if (this.amount == 0) {
            setAmount(amount);
        }
        if (this.seed != null) {
            try {
                write(seed);
            } catch (IOException e) {
                // checked exception
            }
        }
        readOut(out);
    }

    /**
     * Update the RO with some input.
     * 
     * The expected amount to be read must be set.
     * 
     * @param b The data
     * @throws IOException When amount to be read is not set.
     */
    @Override
    public void write(int b) throws IOException {
        write(new byte[] {(byte) b});
    }

    /**
     * Update the RO with some input.
     * 
     * The expected amount to be read must be set.
     * 
     * @param b The data
     * @param off Offset of the data
     * @param len Length of the data
     * @throws IOException when amount to be read is not set.
     */
    @Override
    public void write(byte[] b, int off, int len) throws IOException {
        if (amount == 0) {
            throw new IllegalArgumentException("Amount not set");
        }
        hash.update(b, off, len);
    }

    /**
     * Update the RO with some input.
     * 
     * The expected amount to be read must be set.
     * 
     * @param in A byte array to seed into the RO instance.
     * @throws IOException when amount to be read is not set.
     */
    @Override
    public void write(byte[] b) throws IOException {
        write(b, 0, b.length);
    }
}
