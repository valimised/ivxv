package ee.ivxv.common.math;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Field;
import ee.ivxv.common.crypto.Plaintext;
import java.math.BigInteger;
import org.bouncycastle.math.ec.ECCurve;
import org.bouncycastle.math.ec.ECFieldElement;
import org.bouncycastle.math.ec.ECPoint;
import org.bouncycastle.math.ec.custom.sec.SecP384R1Curve;

public class ECGroup extends Group {
    /** The curve name constant for P-384 curve. */
    public final static String P384 = "P-384";

    // XXX: there is no objective reasoning for this value. It seems to give
    // reasonable success probability (i.e. failure rate 1/2^1024)
    // the probability of encoding failure is 2^-(2^ENCODING_SUCCESS)
    private final static int ENCODING_SUCCESS = 10;
    private final String name;
    private final ECCurve curve;
    private final ECGroupElement inf;
    private final ECGroupElement base;

    // initialize a P-384 group
    public ECGroup() {
        this(P384);
    }

    public ECGroup(byte[] data) {
        this(getASN1CurveName(data));
    }

    public ECGroup(String curvename) {
        name = curvename;
        switch (curvename) {
            case P384:
                curve = new SecP384R1Curve();
                BigInteger x = new BigInteger("aa87ca22be8b05378eb1c71ef320ad746e1d3b628ba79b9859f741e082542a385502f25dbf55296c3a545e3872760ab7", 16);
                BigInteger y = new BigInteger("3617de4a96262c6f5d9e98bf9292dc29f8f41dbd289a147ce9da3113b5f0b8c00a60b1ce1d7e819d7a431d7c90ea0e5f", 16);
                base = new ECGroupElement(this, getPoint(x, y));
                break;
            default:
                throw new IllegalArgumentException("Unknown curve: " + curvename);
        }
        inf = new ECGroupElement(this);
    }

    private static String getASN1CurveName(byte[] data) throws IllegalArgumentException {
        Field f = new Field();
        try {
            f.readFromBytes(data);
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing byte array failed: " + e.toString());
        }
        String curvename;
        try {
            curvename = f.getString();
        } catch (ASN1DecodingException e) {
            throw new IllegalArgumentException("Parsing string failed: " + e.toString());
        }
        return curvename;
    }

    @Override
    public GroupElement getElement(byte[] data) throws IllegalArgumentException {
        return new ECGroupElement(this, data);
    }

    @Override
    public BigInteger getOrder() {
        return curve.getOrder();
    }

    @Override
    public BigInteger getFieldOrder() {
        return curve.getField().getCharacteristic()
                .multiply(BigInteger.valueOf(curve.getField().getDimension()));
    }

    @Override
    public GroupElement getIdentity() {
        return inf;
    }

    @Override
    public GroupElement encode(Plaintext msg) {
        int limit = paddedMessageLengthBits();
        BigInteger x = msg.toBigInteger();
        // if the msg is larger than the point size + padding size, then the message is too large
        if (x.bitLength() > limit) {
            return null;
        }
        // we allocate the room for second padding
        x = x.shiftLeft(ENCODING_SUCCESS);

        // weierstrass curve equation is y^2 = x^3 + ax + b.
        // thus, we compute f(x) = x^3 - 3x + b
        ECFieldElement X, Y, tmp;
        // we take x as x = m || 0, then we check if f(x) is quadratic residue,
        // allowing us to take the square root of f(x). the square root of f(x)
        // will then become the y-coordinate of the point. as the probability that
        // a random element is quadratic residue is 1/2, then the probability that
        // k consequent elements are NOT quadratic residues is 2^-k. if we want
        // failure probability to be less than 2^-k, then we need to reserve the
        // padding for log_2(k) bits.
        do {
            X = curve.fromBigInteger(x);
            tmp = X.square();
            tmp = tmp.add(curve.getA());
            tmp = tmp.multiply(X);
            tmp = tmp.add(curve.getB());
            Y = tmp.sqrt();
            x = x.add(BigInteger.ONE);
        } while (Y == null);
        return new ECGroupElement(this, X, Y);
    }

    @Override
    public Plaintext decode(GroupElement msg) {
        if (!isGroupElement(msg)) {
            throw new IllegalArgumentException("Message is not group element");
        }
        ECGroupElement m = (ECGroupElement) msg;
        ECPoint M = m.getPoint();
        BigInteger MM = M.getAffineXCoord().toBigInteger().shiftRight(ENCODING_SUCCESS);
        return new Plaintext(MM, paddedMessageLengthBits(), true);
    }

    @Override
    public Plaintext pad(Plaintext msg) {
        return msg.addPadding(paddedMessageLengthBytes());
    }

    private int paddedMessageLengthBits() {
        int bitlength = getFieldOrder().bitLength();
        return bitlength - ENCODING_SUCCESS;
    }

    private int paddedMessageLengthBytes() {
        return (paddedMessageLengthBits() + 7) / 8;
    }

    @Override
    public boolean isGroupElement(GroupElement el) {
        return this.equals(el.getGroup());
    }

    @Override
    public byte[] getBytes() {
        return new Field(name).encode();
    }

    public String getCurveName() {
        return name;
    }

    ECPoint getPoint(BigInteger x, BigInteger y) {
        return curve.createPoint(x, y);
    }

    ECPoint getInfinitePoint() {
        return curve.getInfinity();
    }

    ECCurve getCurve() {
        return this.curve;
    }

    public ECGroupElement getBasePoint() {
        return base;
    }

    @Override
    public boolean equals(Object other) {
        if (other == null || this.getClass() != other.getClass()) {
            return false;
        }
        if (other == this) {
            return true;
        }
        ECGroup o = (ECGroup) other;
        return this.getCurve().equals(o.getCurve());
    }

    @Override
    public int hashCode() {
        return this.curve.hashCode() ^ 0x0002;
    }
}
