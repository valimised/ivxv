package ee.ivxv.common.math;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Field;
import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.util.Util;
import java.math.BigInteger;

public class ModPGroup extends Group {
    private final BigInteger p;
    private BigInteger q;
    private final ModPGroupElement one;

    public ModPGroup(BigInteger p) {
        // we assume p=2*q+1 with q prime
        this(p, false);
    }

    public ModPGroup(BigInteger p, boolean verify) {
        if (verify) {
            if (!p.isProbablePrime(80)) {
                throw new IllegalArgumentException(
                        "Multiplicative order is not p=2*1+1 with q prime");
            }
        }
        this.p = p;
        this.one = new ModPGroupElement(this, BigInteger.ONE);
    }

    public ModPGroup(byte[] data) {
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

    @Override
    public GroupElement getElement(byte[] data) throws IllegalArgumentException {
        return new ModPGroupElement(this, data);
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
            this.q = Util.safePrimeOrder(this.p);
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
        return msg.addPadding(msgBytes());
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

    // encode message as quadratic residue so we wouldn't leak any bits
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
    public Plaintext decode(GroupElement el) throws IllegalArgumentException {
        return decode(el, false);
    }

    public Plaintext decode(GroupElement el, boolean nocheck) throws IllegalArgumentException {
        if (!isGroupElement(el)) {
            throw new IllegalArgumentException("Group element is not from this group");
        }
        BigInteger e = ((ModPGroupElement) el).getValue();
        if (!nocheck) {
            if (e.compareTo(BigInteger.ZERO) <= 0) {
                throw new IllegalArgumentException("Can not decode non-positive integer");
            }
            if (e.compareTo(getOrder()) > 0) {
                throw new IllegalArgumentException(
                        "Can not decode integer larger than group order");
            }
            if (legendre(e) != 1) {
                throw new IllegalArgumentException("Decodable integer is not a quadratic residue");
            }
        }
        BigInteger decoded =
                e.compareTo(getMultiplicativeGroupOrder()) > 0 ? getOrder().subtract(e) : e;
        return new Plaintext(decoded, msgBits(), true);
    }

    @Override
    public boolean isGroupElement(GroupElement el) {
        return this.equals(el.getGroup());
    }

    @Override
    public byte[] getBytes() {
        return new Field(getOrder()).encode();
    }

    @Override
    public String toString() {
        return String.format("ModPGroup(%s)", getOrder());
    }
}
