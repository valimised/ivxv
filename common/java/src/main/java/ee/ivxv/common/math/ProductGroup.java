package ee.ivxv.common.math;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Field;
import ee.ivxv.common.asn1.Sequence;
import ee.ivxv.common.crypto.Plaintext;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

/**
 * ProductGroup is a direct product of several groups.
 */
public class ProductGroup extends Group {
    private static final String NAME_MODP = "ModPGroup";
    private static final String NAME_EC = "EllipticCurve";

    private final Group[] groups;

    /**
     * Initialize using groups.
     * 
     * @param groups
     */
    public ProductGroup(Group... groups) {
        this.groups = groups;
    }

    /**
     * Initialize using a group with multiplicity.
     * 
     * @param group
     * @param multiplicity
     */
    public ProductGroup(Group group, int multiplicity) {
        this.groups = new Group[multiplicity];
        for (int i = 0; i < multiplicity; i++) {
            this.groups[i] = group;
        }
    }

    /**
     * Initialize from serialized value.
     * 
     * @param data
     * @throws IllegalArgumentException When can not parse.
     */
    public ProductGroup(byte[] data) throws IllegalArgumentException {
        Sequence s = new Sequence();
        try {
            s.readFromBytes(data);
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing data failed", e);
        }
        byte[][] encoded;
        try {
            encoded = s.getBytes();
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing values failed", e);
        }
        Group[] groups = new Group[encoded.length];
        for (int i = 0; i < encoded.length; i++) {
            Sequence ss = new Sequence();
            try {
                ss.readFromBytes(encoded[i]);
            } catch (ASN1DecodingException e) {
                throw new IllegalArgumentException("Parsing single value failed", e);
            }
            byte[][] value;
            try {
                value = ss.getBytes();
            } catch (ASN1DecodingException e) {
                throw new IllegalArgumentException("Invalid single value", e);
            }
            if (value.length != 2) {
                throw new IllegalArgumentException("Invalid value length");
            }
            Field f = new Field();
            try {
                f.readFromBytes(value[0]);
            } catch (ASN1DecodingException e) {
                throw new IllegalArgumentException("Can not read group type", e);
            }
            String type;
            try {
                type = f.getString();
            } catch (ASN1DecodingException e) {
                throw new IllegalArgumentException("Can not decode group type as string", e);
            }
            if (type.equals(NAME_MODP)) {
                groups[i] = new ModPGroup(value[1]);
            } else if (type.equals(NAME_EC)) {
                groups[i] = new ECGroup(value[1]);
            } else {
                throw new IllegalArgumentException("Unknown group type");
            }
        }
        this.groups = groups;
    }

    @Override
    public GroupElement getElement(byte[] data) throws IllegalArgumentException {
        return new ProductGroupElement(this, data);
    }

    @Override
    public BigInteger getOrder() {
        BigInteger[] orders = new BigInteger[groups.length];
        for (int i = 0; i < orders.length; i++) {
            orders[i] = groups[i].getOrder();
        }
        BigInteger order = MathUtil.lcm(orders);
        return order;
    }

    @Override
    public BigInteger getFieldOrder() {
        BigInteger[] orders = new BigInteger[groups.length];
        for (int i = 0; i < orders.length; i++) {
            orders[i] = groups[i].getFieldOrder();
        }
        BigInteger order = MathUtil.lcm(orders);
        return order;
    }

    @Override
    public GroupElement getIdentity() {
        GroupElement[] identities = new GroupElement[groups.length];
        for (int i = 0; i < identities.length; i++) {
            identities[i] = groups[i].getIdentity();
        }
        ProductGroupElement identity = new ProductGroupElement(this, identities);
        return identity;
    }

    @Override
    public GroupElement encode(Plaintext msg) throws MathException {
        // this method does not make sense
        throw new RuntimeException("Invalid use of ProductGroup");
    }

    @Override
    public Plaintext pad(Plaintext msg) {
        // this method does not make sense
        throw new RuntimeException("Invalid use of ProductGroup");
    }

    @Override
    public Decodable isDecodable(GroupElement el) {
        // this method does not make sense
        throw new RuntimeException("Invalid use of ProductGroup");
    }

    @Override
    public Plaintext decode(GroupElement msg) {
        // this method does not make sense
        throw new RuntimeException("Invalid use of ProductGroup");
    }

    @Override
    public boolean isGroupElement(GroupElement el) {
        return this.equals(el.getGroup());
    }

    /**
     * Serialize the value
     * <p>
     * Returns ASN1 SEQUENCE of serialized values of the underlying groups.
     * 
     * @return
     */
    @Override
    public byte[] getBytes() {
        byte[][] encoded = new byte[groups.length][];
        for (int i = 0; i < groups.length; i++) {
            byte[] groupBytes = groups[i].getBytes();
            Field f;
            if (groups[i] instanceof ModPGroup) {
                f = new Field(NAME_MODP);
            } else if (groups[i] instanceof ECGroup) {
                f = new Field(NAME_EC);
            } else {
                throw new RuntimeException("Invalid group");
            }
            Sequence s = new Sequence(f.encode(), groupBytes);
            encoded[i] = s.encode();
        }
        Sequence ret = new Sequence(encoded);
        return ret.encode();
    }

    @Override
    public boolean equals(Object other) {
        if (other == null || this.getClass() != other.getClass()) {
            return false;
        }
        if (other == this) {
            return true;
        }
        ProductGroup og = (ProductGroup) other;
        if (groups.length != og.groups.length) {
            return false;
        }
        for (int i = 0; i < groups.length; i++) {
            if (!groups[i].equals(og.groups[i])) {
                return false;
            }
        }
        return true;
    }

    @Override
    public String toString() {
        List<String> strings = new ArrayList<String>(groups.length);
        for (int i = 0; i < groups.length; i++) {
            strings.add(i, groups[i].toString());
        }
        String ret = String.format("ProductGroup(%s)", String.join(",", strings));
        return ret;
    }

    /**
     * Get the number of groups in the ProductGroup.
     * 
     * @return
     */
    public int getLength() {
        return groups.length;
    }

    /**
     * Get the groups used to construct the ProductGroup.
     * 
     * @return
     */
    public Group[] getGroups() {
        return this.groups;
    }
}
