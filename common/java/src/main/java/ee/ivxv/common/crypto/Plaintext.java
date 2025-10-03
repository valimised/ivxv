package ee.ivxv.common.crypto;

import ee.ivxv.common.asn1.Field;
import ee.ivxv.common.math.Group;
import ee.ivxv.common.util.Util;

import java.math.BigInteger;
import java.util.Arrays;
import java.util.HexFormat;
import java.util.Optional;
import java.util.function.Function;

/**
 * Plaintext instance represents an immutable plaintext suitable for encryption. It includes
 * additional methods for padding, encoding and decoding the messages.
 *
 */
public class Plaintext {
    private final byte[] msg;
    public final boolean padded;

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
        return String.format("Plaintext(%s)", HexFormat.of().formatHex(msg).toUpperCase());
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

        return Arrays.equals(this.msg, plaintext.msg);

    }

    @Override
    public int hashCode() {
        return Arrays.hashCode(msg);
    }
}
