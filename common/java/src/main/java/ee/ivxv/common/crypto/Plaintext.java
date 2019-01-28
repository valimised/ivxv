package ee.ivxv.common.crypto;

import static javax.xml.bind.DatatypeConverter.printHexBinary;

import ee.ivxv.common.asn1.Field;
import ee.ivxv.common.util.Util;
import java.math.BigInteger;
import java.util.Arrays;

/**
 * Plaintext instance represents an immutable plaintext suitable for encryption. It includes
 * additional methods for padding, encoding and decoding the messages.
 *
 */
public class Plaintext {
    private final byte[] msg;
    private final boolean padded;

    /**
     * Initialize the Plaintext instance from byte array.
     * 
     * @param msg Byte array to set message.
     * @param padded A boolean indicating if the message is already padded.
     */
    public Plaintext(byte[] msg, boolean padded) {
        this.msg = msg;
        this.padded = padded;
    }

    /**
     * Initialize Plaintext from unpadded byte array.
     * 
     * @param msg Byte array to set message.
     */
    public Plaintext(byte[] msg) {
        this(msg, false);
    }

    /**
     * Initialize Plaintext from unpadded UTF-8 encoded String.
     * 
     * @param msg Unpadded UTF-8 encoded message.
     */
    public Plaintext(String msg) {
        this(Util.toBytes(msg), false);
    }

    /**
     * Initialize empty Plaintext.
     */
    public Plaintext() {
        this(new byte[] {}, false);
    }

    /**
     * Initialize Plaintext from BigInteger from its byte representation. Truncating or padding with
     * zeros the byte representation if necessary.
     * 
     * @param msg BigInteger to initialize from.
     * @param totalBits The number of bits to use. Lower bits are truncated if necessary.
     * @param padded Indicate if the BigInteger denotes a padded message.
     */
    // the user must be sure that the msg fits into this number of bits. we
    // truncate the lower bits if they do not fit
    public Plaintext(BigInteger msg, int totalBits, boolean padded) {
        byte[] bib = msg.toByteArray();
        this.msg = new byte[(totalBits + 7) / 8];
        System.arraycopy(bib, 0, this.msg,
                this.msg.length - bib.length >= 0 ? this.msg.length - bib.length : 0, bib.length);
        this.padded = padded;
    }

    @Override
    public String toString() {
        return String.format("Plaintext(%s)", printHexBinary(msg));
    }

    /**
     * Get byte-encoded representation of the Plaintext, embedded in a ASN1 OCTETSTRING.
     * 
     * @return ASN1 encoded message in bytes.
     */
    public byte[] getBytes() {
        return new Field(this.msg).encode();
    }

    /**
     * Get byte-encoded representation of the Plaintext.
     * 
     * @return Message in bytes.
     */
    public byte[] getMessage() {
        return this.msg.clone();
    }

    /**
     * Get UTF-8 decoded representation of the message. If the message is not valid UTF-8, then the
     * result is undefined.
     * 
     * @return
     */
    public String getUTF8DecodedMessage() {
        return Util.toString(this.msg);
    }

    /**
     * Return a Plaintext with additional padding.
     * 
     * @param totalBytes The number of bytes to pad to.
     * @return Plaintext with padded message.
     * @throws IllegalArgumentException If the message does not fit into given number of bytes.
     */
    public Plaintext addPadding(int totalBytes) throws IllegalArgumentException {
        if (this.padded) {
            return this;
        }
        if (this.msg.length > totalBytes - 3) {
            throw new IllegalArgumentException("Padded message length too long");
        }
        byte[] padded = new byte[totalBytes];
        int i;
        padded[0] = 0x00;
        padded[1] = 0x01;
        padded[totalBytes - this.msg.length - 1] = 0x00;
        for (i = 2; i < totalBytes - this.msg.length - 1; i++) {
            padded[i] = (byte) 0xff;
        }
        System.arraycopy(this.msg, 0, padded, i + 1, this.msg.length);
        return new Plaintext(padded, true);
    }

    /**
     * Strip padding. No-op if message is already unpadded.
     * 
     * @return Unpadded Plaintext.
     * @throws IllegalArgumentException If padding is invalid.
     */
    public Plaintext stripPadding() throws IllegalArgumentException {
        if (!this.padded) {
            return this;
        }
        if (this.msg.length < 3) {
            throw new IllegalArgumentException("Source message can not contain padding");
        }
        if (this.msg[0] != 0x00 || this.msg[1] != 0x01) {
            throw new IllegalArgumentException("Incorrect padding head");
        }
        for (int i = 2; i < this.msg.length; i++) {
            switch (this.msg[i]) {
                case 0:
                    // found padding end
                    return new Plaintext(Arrays.copyOfRange(this.msg, i + 1, this.msg.length),
                            false);
                case (byte) 0xff:
                    continue;
                default:
                    // incorrect padding byte
                    throw new IllegalArgumentException("Incorrect padding byte");
            }
        }
        throw new IllegalArgumentException("Padding unexpected");
    }

    /**
     * Return a BigInteger representation of the message.
     * 
     * @return BigInteger representation of the message.
     */
    public BigInteger toBigInteger() {
        return new BigInteger(1, this.msg);
    }

    @Override
    public boolean equals(Object o) {
        if (this == o)
            return true;
        if (o == null || getClass() != o.getClass())
            return false;

        Plaintext plaintext = (Plaintext) o;

        return Arrays.equals(this.stripPadding().msg, plaintext.stripPadding().msg);

    }

    @Override
    public int hashCode() {
        return Arrays.hashCode(msg);
    }
}
