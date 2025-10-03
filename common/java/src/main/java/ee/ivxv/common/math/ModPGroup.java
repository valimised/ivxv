package ee.ivxv.common.math;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Field;
import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.crypto.rnd.Rnd;
import org.bouncycastle.math.raw.Mod;
import org.bouncycastle.pqc.math.linearalgebra.IntegerFunctions;

import java.io.IOException;
import java.math.BigInteger;
import java.util.Arrays;

/**
 * Group of integers modulo a safe prime.
 */
public class ModPGroup extends Group {
    private final BigInteger p;
    private BigInteger q;
    private final ModPGroupElement one;

    /**
     * Initialize group using a safe prime modulus.
     * <p>
     * This constructor does not verify that modulus is a safe prime.
     * 
     * @param p
     */
    public ModPGroup(BigInteger p) {
        // we assume p=2*q+1 with q prime
        this(p, false);
    }

    /**
     * Initialize group using a safe prime modulus and verify that the modulus is a safe prime.
     * 
     * @param p
     * @param verify
     * @throws IllegalArgumentException When modulus is not safe prime.
     */
    public ModPGroup(BigInteger p, boolean verify) throws IllegalArgumentException {
        if (verify) {
            if (!p.isProbablePrime(80)) {
                throw new IllegalArgumentException(
                        "Multiplicative order is not p=2*1+1 with q prime");
            }
        }
        this.p = p;
        this.one = new ModPGroupElement(this, BigInteger.ONE);
    }

    /**
     * Initialize group using serialized group.
     * 
     * @see #getBytes()
     * 
     * @param data
     * @throws IllegalArgumentException When can not parse
     */
    public ModPGroup(byte[] data) throws IllegalArgumentException {
        Field f = new Field();
        try {
            f.readFromBytes(data);
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing byte array failed: " + e);
        }
        try {
            this.p = f.getInteger();
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing integer failed: " + e);
        }
        this.one = new ModPGroupElement(this, BigInteger.ONE);
    }

    /**
     * Generate a safe prime and initialize the group.
     * <p>
     * Use the random source for obtaining modulus candidates. The tries argument defines the number
     * of tries to repeat. When tries is set to 0, then test candidate values until a suitable
     * modulus is found.
     * 
     * @param len
     * @param rnd
     * @param tries
     * @throws IllegalArgumentException When len is not a positive value.
     * @throws IOException When exception during a read from random source.
     * @throws MathException When couldn't find a safe prime within tries
     */
    public ModPGroup(int len, Rnd rnd, int tries)
            throws IllegalArgumentException, IOException, MathException {
        int sglen = len - 1;
        BigInteger bigTwo = new BigInteger("2");
        BigInteger genp, genq;

        // For performance reasons, we first test primes lightly.
        // If both pass, test them thoroughly.
        for (int i = 0; tries == 0 || i < tries; i++) {
            genq = IntegerConstructor.construct(rnd, bigTwo.pow(sglen).subtract(BigInteger.ONE));
            if (!genq.isProbablePrime(2)) {
                continue;
            }
            genp = genq.multiply(bigTwo).add(BigInteger.ONE);
            if (genp.bitLength() != len || !genp.isProbablePrime(2)) {
                continue;
            }
            if (genp.isProbablePrime(80) && genq.isProbablePrime(80)) {
                p = genp;
                q = genq;
                one = new ModPGroupElement(this, BigInteger.ONE);
                return;
            }
        }
        throw new MathException("Could not generate group parameters during tries");
    }

    @Override
    public String getName() {
        return "ModSafePrime";
    }

    @Override
    public GroupElement getElement(byte[] data) throws IllegalArgumentException {
        return new ModPGroupElement(this, data);
    }

    /**
     * Sample a random element from the group.
     * 
     * @param rnd
     * @return
     * @throws IOException
     */
    public ModPGroupElement getRandomElement(Rnd rnd) throws IOException {
        BigInteger template = IntegerConstructor.construct(rnd, getOrder());
        // square the value to make sure it is quadratic residue
        template = template.modPow(BigInteger.valueOf(2), getOrder());
        return new ModPGroupElement(this, template);
    }

    @Override
    public BigInteger getOrder() {
        return this.p;
    }

    @Override
    public BigInteger getFieldOrder() {
        return this.p;
    }

    @Override
    public GroupElement getIdentity() {
        return this.one;
    }

    private BigInteger getMultiplicativeGroupOrder() {
        if (this.q == null) {
            this.q = MathUtil.safePrimeOrder(this.p);
        }
        return this.q;
    }

    private int msgBits() {
        return getMultiplicativeGroupOrder().bitLength();
    }

    private int msgBytes() {
        return (msgBits() + 7) / 8;
    }

    @Override
    public Plaintext pad(Plaintext msg) {
        if (msg.padded) {
            return msg;
        }

        byte[] padded = msg.getMessage();
        int totalBytes = msgBytes();

        if (padded.length > totalBytes - 3) {
            throw new IllegalArgumentException("Padded message length too long");
        }
        byte[] padded2 = new byte[totalBytes];
        int i;
        padded2[0] = 0x00;
        padded2[1] = 0x01;
        padded2[totalBytes - padded.length - 1] = 0x00;
        for (i = 2; i < totalBytes - padded.length - 1; i++) {
            padded2[i] = (byte) 0xff;
        }
        System.arraycopy(padded, 0, padded2, i + 1, padded.length);
        return new Plaintext(padded2, true);
    }

    @Override
    public Plaintext unpad(Plaintext msg) {
        if (!msg.padded) {
            return msg;
        }

        byte[] padded = msg.getMessage();

        if (padded.length < 3) {
            throw new IllegalArgumentException("Source message can not contain padding");
        }
        if (padded[0] != 0x00 || padded[1] != 0x01) {
            throw new IllegalArgumentException("Incorrect padding head");
        }
        for (int i = 2; i < padded.length; i++) {
            switch (padded[i]) {
                case 0:
                    // found padding end
                    return new Plaintext(Arrays.copyOfRange(padded, i + 1, padded.length),
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

    @Override
    public boolean equals(Object other) {
        if (other == null || this.getClass() != other.getClass()) {
            return false;
        }
        if (other == this) {
            return true;
        }
        ModPGroup o = (ModPGroup) other;

        return this.getOrder().equals(o.getOrder());
    }

    @Override
    public int hashCode() {
        return this.getOrder().hashCode() ^ 0x0002;
    }

    private int legendre(BigInteger e) {
        return MathUtil.legendre(e, getOrder());
    }

    /**
     * Encode the message as a quadratic residue.
     * <p>
     * If the value is quadratic residue, then return the message encoded as integer. Otherwise,
     * return modulus-msg encoded as integer.
     * 
     * @param msg
     * @return
     * @throws IllegalArgumentException is value is not less than half the order size or greater
     *         than 0
     * @throws MathException When computation fails.
     */
    @Override
    public GroupElement encode(Plaintext msg) throws IllegalArgumentException, MathException {
        BigInteger m = msg.toBigInteger();
        if (m.compareTo(BigInteger.ZERO) <= 0) {
            throw new IllegalArgumentException("Value to be encoded is less than 0");
        }
        if (m.compareTo(getMultiplicativeGroupOrder()) > 0) {
            throw new IllegalArgumentException("Value to be encoded is larger equal than order/2");
        }
        switch (legendre(m)) {
            case -1:
                return new ModPGroupElement(this, getOrder().subtract(m));
            case 1:
                return new ModPGroupElement(this, m);
            default:
                // Legendre's function only takes values -1,0,1
                throw new IllegalArgumentException("Can not encode 0 as quadratic residue");
        }
    }

    @Override
    public Decodable isDecodable(GroupElement el) {
        // Is valid group element
        Decodable isValid = isGroupElement(el);
        if (isValid != Decodable.VALID)
            return isValid;

        // Padded plaintext should be exactly of a group generator order bytes length - 1
        if (decode((ModPGroupElement) el) == null) {
            return Decodable.INVALID_RANGE;
        }

        return Decodable.VALID;
    }

    /**
     * Decodes group element into BigInteger equivalent, or null if group element is not decodable.
     * @param encoded
     * @return
     */
    private BigInteger decode(ModPGroupElement encoded) {
        BigInteger e = encoded.getValue();

        BigInteger decoded =
                e.compareTo(getMultiplicativeGroupOrder()) > 0 ? getOrder().subtract(e) : e;

        // Java strips leading 0x00 padding byte that's why "-1"
        if (MathUtil.toBytesLen(decoded.bitLength()) != MathUtil.toBytesLen(this.getMultiplicativeGroupOrder().bitLength()) - 1) {
            return null;
        }
        return decoded;
    }
    /**
     * Decode group element as a message.
     * <p>
     * If the value is larger than order/2, then return order-el.
     * 
     * @param el
     * @return
     */
    @Override
    public Plaintext decode(GroupElement el) {
        BigInteger decoded = decode((ModPGroupElement) el);
        if (decoded == null) {
            throw new IllegalArgumentException("Element is not decodable");
        }

        Plaintext padded = new Plaintext(decoded, msgBits(), true);
        if ((padded.getMessage().length) != MathUtil.toBytesLen(this.getFieldOrder().bitLength())) {
            throw new IllegalArgumentException("Message not padded to correct length");
        }
        return padded;

    }

    @Override
    public Decodable isGroupElement(GroupElement el) {
        // Element belongs to group
        if (!this.equals(el.getGroup())) {
            return Decodable.INVALID_GROUP;
        }

        // Element value > 0
        BigInteger e = ((ModPGroupElement) el).getValue();
        if (e.compareTo(BigInteger.ZERO) <= 0) {
            return Decodable.INVALID_RANGE;
        }

        // Element value < group field order
        if (e.compareTo(getOrder()) > 0) {
            return Decodable.INVALID_RANGE;
        }

        // Is quadratic residue
        if (IntegerFunctions.jacobi(((ModPGroupElement)el).getValue(), this.getFieldOrder()) != 1) {
            return Decodable.INVALID_QR;
        }

        return Decodable.VALID;
    }

    /**
     * Serialize group.
     * <p>
     * Returns group modulus as ASN1 INTEGER
     * 
     * @return
     */
    @Override
    public byte[] getBytes() {
        return new Field(getOrder()).encode();
    }

    @Override
    public String toString() {
        return String.format("ModPGroup(%s)", getOrder());
    }
}
