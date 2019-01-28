package ee.ivxv.common.math;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Field;
import java.math.BigInteger;
import java.util.Arrays;
import org.bouncycastle.math.ec.ECFieldElement;
import org.bouncycastle.math.ec.ECPoint;

/**
 * ECGroupElement holds elliptic curve points.
 */
public class ECGroupElement extends GroupElement {
    private final ECGroup group;
    private final ECPoint point;

    /**
     * Initializie using a group and point.
     * 
     * @param group
     * @param point
     */
    public ECGroupElement(ECGroup group, ECPoint point) {
        this.group = group;
        this.point = point.normalize();
        checkPoint();
    }

    /**
     * Initialize a point at infinity.
     * 
     * @param group
     */
    public ECGroupElement(ECGroup group) {
        this(group, group.getInfinitePoint());
    }

    /**
     * Initialize a point field element integer representation as coordinates.
     * 
     * @param group
     * @param x
     * @param y
     * @throws IllegalArgumentException
     */
    public ECGroupElement(ECGroup group, BigInteger x, BigInteger y)
            throws IllegalArgumentException {
        this.group = group;
        this.point = this.group.getPoint(x, y).normalize();
        checkPoint();
    }

    /**
     * Initialize using field elements as coordinates.
     * 
     * @param group
     * @param X
     * @param Y
     * @throws IllegalArgumentException
     */
    public ECGroupElement(ECGroup group, ECFieldElement X, ECFieldElement Y)
            throws IllegalArgumentException {
        this(group, X.toBigInteger(), Y.toBigInteger());
    }

    /**
     * Initialize using serialized data.
     * 
     * @param group
     * @param data
     * @throws IllegalArgumentException
     */
    public ECGroupElement(ECGroup group, byte[] data) throws IllegalArgumentException {
        this.group = group;
        BigInteger[] coords = getANSICoords(data);
        if (coords == null) {
            this.point = this.group.getInfinitePoint();
        } else {
            this.point = this.group.getPoint(coords[0], coords[1]);
        }
        checkPoint();
    }

    private void checkPoint() throws IllegalArgumentException {
        if (!point.isValid()) {
            throw new IllegalArgumentException("Point coordinates are not valid");
        }
    }

    private static BigInteger[] getANSICoords(byte[] encoded) throws IllegalArgumentException {
        byte[] data;
        Field f = new Field();
        try {
            f.readFromBytes(encoded);
        } catch (ASN1DecodingException ex) {
            throw new IllegalArgumentException("Illegal point encoding");
        }
        try {
            data = f.getBytes();
        } catch (ASN1DecodingException ex) {
            throw new IllegalArgumentException("Illegal data in point encoding");
        }

        if (data.length == 1 && data[0] == 0x00) {
            return null;
        }
        if (data.length % 2 != 1) {
            throw new IllegalArgumentException(
                    "ANSI coordinates must consist of mode byte and coordinates");
        }
        if (data[0] != 0x04) {
            throw new IllegalArgumentException("Invalid point representation method");
        }
        int coordLen = (data.length - 1) / 2;
        BigInteger[] coords = new BigInteger[2];
        coords[0] = new BigInteger(1, Arrays.copyOfRange(data, 1, coordLen + 1));
        coords[1] = new BigInteger(1, Arrays.copyOfRange(data, coordLen + 1, 2 * coordLen + 1));
        return coords;
    }

    @Override
    public BigInteger getOrder() {
        return group.getOrder();
    }

    /**
     * Add another point to this point and return the result.
     * 
     * @param other
     * @return
     */
    @Override
    public GroupElement op(GroupElement other) throws MathException {
        if (!this.group.equals(other.getGroup())) {
            throw new MathException("Group elements from mismatching groups");
        }
        ECGroupElement e = (ECGroupElement) other;
        return new ECGroupElement(this.group, this.getPoint().add(e.getPoint()));
    }

    /**
     * Scale the point with factor.
     * 
     * @param factor
     * @return
     */
    @Override
    public GroupElement scale(BigInteger factor) {
        return new ECGroupElement(this.group, this.getPoint().multiply(factor));
    }

    /**
     * Return the inverse of the point.
     * 
     * @return
     */
    @Override
    public GroupElement inverse() {
        return new ECGroupElement(this.group, this.getPoint().negate());
    }

    /**
     * Serialize the point.
     * <p>
     * See X9-62 section 4.3.6 for EC point serialization algorithm. The point is then encoded as
     * ASN1 OCTETSTRING
     * 
     * @return
     */
    @Override
    public byte[] getBytes() {
        // ref: https://www.security-audit.com/files/x9-62-09-20-98.pdf section 4.3.6
        if (this.getPoint().isInfinity()) {
            return new byte[] {0x00};
        }
        byte[] res = this.getPoint().getEncoded(false);
        return new Field(res).encode();
    }

    @Override
    public String toString() {
        return String.format("ECGroupElement(%s, %s)",
                this.getPoint().getAffineXCoord().toBigInteger(),
                this.getPoint().getAffineYCoord().toBigInteger());
    }

    @Override
    public Group getGroup() {
        return this.group;
    }

    /**
     * Get the actual point of the element.
     * 
     * @return
     */
    public ECPoint getPoint() {
        return point;
    }

    @Override
    public boolean equals(Object other) {
        if (other == null || this.getClass() != other.getClass()) {
            return false;
        }
        if (other == this) {
            return true;
        }
        ECGroupElement el = (ECGroupElement) other;
        return this.group.equals(el.getGroup()) && this.getPoint().equals(el.getPoint());
    }

    @Override
    public int hashCode() {
        return this.getPoint().hashCode() ^ this.getGroup().hashCode();
    }
}
