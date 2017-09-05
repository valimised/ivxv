package ee.ivxv.common.crypto;

import static javax.xml.bind.DatatypeConverter.printHexBinary;

import ee.ivxv.common.asn1.Field;
import ee.ivxv.common.util.Util;
import java.math.BigInteger;
import java.util.Arrays;

// Plaintext is an immutable plaintext
public class Plaintext {
    private final byte[] msg;
    private final boolean padded;

    public Plaintext(byte[] msg, boolean padded) {
        this.msg = msg;
        this.padded = padded;
    }

    public Plaintext(byte[] msg) {
        this(msg, false);
    }

    public Plaintext(String msg) {
        this(Util.toBytes(msg), false);
    }

    public Plaintext() {
        this(new byte[] {}, false);
    }

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

    public byte[] getBytes() {
        return new Field(this.msg).encode();
    }

    public byte[] getMessage() {
        return this.msg.clone();
    }

    // assumption is that the message is UTF-8. If not, then the output is
    // ill-defined (invalid characters are replaced with replacement
    // characters).
    public String getUTF8DecodedMessage() {
        return Util.toString(this.msg);
    }

    public Plaintext addPadding(int totalBytes) throws IllegalArgumentException {
        if (this.padded) {
            return this;
        }
        if (this.msg.length > totalBytes - 3) {
            throw new IllegalArgumentException("Padded message length too short");
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
