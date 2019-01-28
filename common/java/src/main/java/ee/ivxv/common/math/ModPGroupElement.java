package ee.ivxv.common.math;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Field;
import java.math.BigInteger;

/**
 * ModPGroupElement is an integer modulo a safe prime.
 */
public class ModPGroupElement extends GroupElement {
    private final ModPGroup group;
    private final BigInteger value;
    private BigInteger order;

    /**
     * Initialize using group and integer value.
     * 
     * @param group
     * @param value
     * @throws IllegalArgumentException
     */
    public ModPGroupElement(ModPGroup group, BigInteger value) throws IllegalArgumentException {
        this.group = group;
        this.value = value;
    }

    /**
     * Initialize using group and serialized value.
     * 
     * @see #getBytes()
     * 
     * @param group
     * @param data
     * @throws IllegalArgumentException
     */
    public ModPGroupElement(ModPGroup group, byte[] data) throws IllegalArgumentException {
        Field f = new Field();
        try {
            f.readFromBytes(data);
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing byte array failed: " + e);
        }
        this.group = group;
        try {
            this.value = f.getInteger();
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing integer failed: " + e);
        }
    }

    @Override
    public BigInteger getOrder() {
        if (order == null) {
            try {
                order = computeOrder();
            } catch (MathException e) {
                throw new IllegalArgumentException("Computing multiplicative order failed: " + e);
            }
        }
        return this.order;
    }

    private BigInteger computeOrder() throws MathException {
        // we assume a group with p = 2*q + 1. The order divides p - 1, thus it
        // is either 1, 2, q or p-1.
        BigInteger[] possible = new BigInteger[] {BigInteger.ONE, BigInteger.valueOf(2),
                group.getOrder().subtract(BigInteger.ONE).divide(BigInteger.valueOf(2)),
                group.getOrder().subtract(BigInteger.ONE)};
        for (BigInteger p : possible) {
            if (value.modPow(p, group.getOrder()).equals(BigInteger.ONE)) {
                return p;
            }
        }
        throw new MathException("Invalid group parameters");
    }

    /**
     * Multiply this value with other.
     * 
     * @param other
     * @return
     * @throws MathException When other element is from different group.
     */
    @Override
    public GroupElement op(GroupElement other) throws MathException {
        if (!this.group.equals(other.getGroup())) {
            throw new MathException("Group elements from mismatching groups");
        }
        ModPGroupElement o = (ModPGroupElement) other;
        return new ModPGroupElement(this.group,
                this.getValue().multiply(o.getValue()).mod(this.group.getOrder()));
    }

    /**
     * Exponentiate the value.
     * 
     * @param factor
     * @return
     */
    @Override
    public GroupElement scale(BigInteger factor) {
        try {
            return new ModPGroupElement(this.group,
                    this.getValue().modPow(factor, this.group.getOrder()));
        } catch (IllegalArgumentException e) {
            // this can not happen as the exception is only thrown when
            // computing element order fails. this happens only if group
            // parameters are incorrect. however, the parameters are
            // instantiated from this element's parameters.
        }
        return null;
    }

    /**
     * Find the modular inverse of the value.
     * 
     * @return
     */
    @Override
    public GroupElement inverse() {
        try {
            return new ModPGroupElement(this.group,
                    this.getValue().modInverse(this.group.getOrder()));
        } catch (IllegalArgumentException e) {
            // this can not happen as the exception is only thrown when
            // computing element order fails. this happens only if group
            // parameters are incorrect. however, the parameters are
            // instantiated from this element's parameters.
        }
        return null;
    }

    @Override
    public boolean equals(Object other) {
        if (other == null || this.getClass() != other.getClass()) {
            return false;
        }
        if (other == this) {
            return true;
        }
        ModPGroupElement o = (ModPGroupElement) other;
        return this.group.equals(o.group) && this.getValue().equals(o.getValue());
    }

    @Override
    public int hashCode() {
        return this.group.hashCode() ^ this.getValue().hashCode();
    }

    /**
     * Get the integer value of the element.
     * 
     * @return
     */
    public BigInteger getValue() {
        return this.value;
    }

    /**
     * Serialize the value
     * <p>
     * Returns the element as ASN1 INTEGER
     *
     * @return
     */
    @Override
    public byte[] getBytes() {
        return new Field(value).encode();
    }

    @Override
    public String toString() {
        return String.format("ModPGroupElement(%s)", value);
    }

    @Override
    public Group getGroup() {
        return this.group;
    }
}
