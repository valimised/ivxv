package ee.ivxv.common.math;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Sequence;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

/**
 * ProductGroupElement is a direct product of group elements
 */
public class ProductGroupElement extends GroupElement {
    private final ProductGroup group;
    private GroupElement[] elements;

    /**
     * Initialize using a ProductGroup and corresponding group elements.
     * 
     * @param group
     * @param elements
     */
    public ProductGroupElement(ProductGroup group, GroupElement... elements) {
        this.group = group;
        this.elements = elements;
    }

    /**
     * Initialize using a ProductGroup and uninitalized group elements.
     * 
     * @param group
     */
    public ProductGroupElement(ProductGroup group) {
        this.group = group;
        this.elements = new GroupElement[group.getGroups().length];
    }

    /**
     * Initialize using a ProductGroup and serialized data.
     * 
     * @param group
     * @param data
     * @throws IllegalArgumentException When parsing fails
     */
    public ProductGroupElement(ProductGroup group, byte[] data) throws IllegalArgumentException {
        this.group = group;
        Sequence s = new Sequence();
        try {
            s.readFromBytes(data);
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing byte array failed", e);
        }
        byte[][] bs;
        try {
            bs = s.getBytes();
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Invalid input byte array", e);
        }
        if (bs.length != group.getLength()) {
            throw new IllegalArgumentException("Input data does not correspond to groups");
        }
        elements = new GroupElement[bs.length];
        for (int i = 0; i < bs.length; i++) {
            elements[i] = group.getGroups()[i].getElement(bs[i]);
        }
    }

    @Override
    public BigInteger getOrder() {
        BigInteger[] orders = new BigInteger[elements.length];
        for (int i = 0; i < orders.length; i++) {
            orders[i] = elements[i].getOrder();
        }
        BigInteger order = MathUtil.lcm(orders);
        return order;
    }

    /**
     * Perform pointwise operation with other element and return the result.
     * 
     * @param other
     * @return
     * @throws MatchException When group elements are from different groups.
     */
    @Override
    public GroupElement op(GroupElement other) throws MathException {
        if (!this.group.equals(other.getGroup())) {
            throw new MathException("Group elements from mismatching groups");
        }
        ProductGroupElement el = (ProductGroupElement) other;
        GroupElement[] reselements = new GroupElement[elements.length];
        for (int i = 0; i < reselements.length; i++) {
            reselements[i] = elements[i].op(el.elements[i]);
        }
        ProductGroupElement res = new ProductGroupElement(this.group, reselements);
        return res;
    }

    /**
     * Scale all direct product elements with a scale factor and return the result.
     * 
     * @param factor
     * @return
     */
    @Override
    public GroupElement scale(BigInteger factor) {
        GroupElement[] reselements = new GroupElement[elements.length];
        for (int i = 0; i < reselements.length; i++) {
            reselements[i] = elements[i].scale(factor);
        }
        ProductGroupElement res = new ProductGroupElement(this.group, reselements);
        return res;
    }

    /**
     * Scale all direct product elements with corresponding scale factor and return the result.
     * 
     * @param factors
     * @return
     */
    public ProductGroupElement scale(BigInteger[] factors) {
        GroupElement[] reselements = new GroupElement[elements.length];
        for (int i = 0; i < reselements.length; i++) {
            reselements[i] = elements[i].scale(factors[i]);
        }
        ProductGroupElement res = new ProductGroupElement(this.group, reselements);
        return res;
    }

    /**
     * Inverse all direct product elements and return the result.
     * 
     * @return
     */
    @Override
    public GroupElement inverse() {
        GroupElement[] inverses = new GroupElement[elements.length];
        for (int i = 0; i < inverses.length; i++) {
            inverses[i] = elements[i].inverse();
        }
        ProductGroupElement ret = new ProductGroupElement(this.group, inverses);
        return ret;
    }

    @Override
    public boolean equals(Object other) {
        if (other == null || this.getClass() != other.getClass()) {
            return false;
        }
        if (other == this) {
            return true;
        }
        ProductGroupElement o = (ProductGroupElement) other;
        if (getElements().length != o.getElements().length) {
            return false;
        }
        for (int i = 0; i < getElements().length; i++) {
            if (!getElements()[i].equals(o.getElements()[i]))
                return false;
        }
        return true;
    }

    @Override
    public Group getGroup() {
        return this.group;
    }

    /**
     * Serialize element
     * <p>
     * Serializes all direct product elements and returns ASN1 SEQUENCE of them.
     * 
     * @return
     */
    @Override
    public byte[] getBytes() {
        byte[][] encoded = new byte[elements.length][];
        for (int i = 0; i < encoded.length; i++) {
            encoded[i] = elements[i].getBytes();
        }
        Sequence s = new Sequence(encoded);
        byte[] ret = s.encode();
        return ret;
    }

    @Override
    public String toString() {
        List<String> strings = new ArrayList<String>(elements.length);
        for (int i = 0; i < elements.length; i++) {
            strings.add(i, elements[i].toString());
        }
        String ret = String.format("ProductGroupElement(%s)", String.join(",", strings));
        return ret;
    }

    /**
     * Get the corresponding direct product elements.
     * 
     * @return
     */
    public GroupElement[] getElements() {
        return this.elements;
    }
}
