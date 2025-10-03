package ee.ivxv.common.math;

import ee.ivxv.common.asn1.ASN1DecodingException;
import ee.ivxv.common.asn1.Field;
import ee.ivxv.common.crypto.Plaintext;

import java.math.BigInteger;
import java.util.*;
import java.util.function.Supplier;
import java.util.stream.IntStream;

import org.bouncycastle.math.ec.ECCurve;
import org.bouncycastle.math.ec.ECFieldElement;
import org.bouncycastle.math.ec.ECPoint;
import org.bouncycastle.math.ec.custom.sec.SecP224R1Curve;
import org.bouncycastle.math.ec.custom.sec.SecP384R1Curve;

/**
 * Elliptic curve group
 */
public class ECGroup extends Group {
    public final static String P224 = "P-224";
    public final static String P384 = "P-384";

    // there is no objective reasoning for this value. It seems to give
    // reasonable success probability (i.e. failure rate 1/2^1024)
    // the probability of encoding failure is 2^-(2^ENCODING_SUCCESS)
    private final static int ENCODING_SUCCESS = 10;
    private final String name;
    private final ECCurve curve;
    private final ECGroupElement inf;
    private final ECGroupElement base;

    // initialize a P-384 group
    /**
     * Initialize an elliptic curve P-384
     */
    public ECGroup() {
        this(P384);
    }

    /**
     * Initialize an elliptic curve using a serialized value.
     *
     * @see #getBytes()
     * @param data
     * @throws IllegalArgumentException When parsing fails
     */
    public ECGroup(byte[] data) throws IllegalArgumentException {
        this(getASN1CurveName(data));
    }

    /**
     * Initialize an elliptic curve using a standard name for the curve.
     *
     * @param curvename
     * @throws IllegalArgumentException When unknown curve name is used.
     */
    public ECGroup(String curvename) throws IllegalArgumentException {
        name = curvename;
        switch (curvename) {
            case P224:
                {
                curve = new SecP224R1Curve();
                BigInteger x = new BigInteger(
                        "b70e0cbd6bb4bf7f321390b94a03c1d356c21122343280d6115c1d21",
                        16);
                BigInteger y = new BigInteger(
                        "bd376388b5f723fb4c22dfe6cd4375a05a07476444d5819985007e34",
                        16);
                base = new ECGroupElement(this, getPoint(x, y));
                }
                break;
            case P384:
                {
                curve = new SecP384R1Curve();
                BigInteger x = new BigInteger(
                        "aa87ca22be8b05378eb1c71ef320ad746e1d3b628ba79b9859f741e082542a385502f25dbf55296c3a545e3872760ab7",
                        16);
                BigInteger y = new BigInteger(
                        "3617de4a96262c6f5d9e98bf9292dc29f8f41dbd289a147ce9da3113b5f0b8c00a60b1ce1d7e819d7a431d7c90ea0e5f",
                        16);
                base = new ECGroupElement(this, getPoint(x, y));
                }
                break;
            default:
                throw new IllegalArgumentException("Unknown curve: " + curvename);
        }
        inf = new ECGroupElement(this);
    }

    /**
     * Initialize an elliptic curve over a field with specified bitlength.
     *
     * @param len
     */
    public ECGroup(int len) {
        this(getIntCurveName(len));
    }

    private static String getIntCurveName(int len) {
        switch (len) {
            case 224:
                return P224;
            case 384:
                return P384;
            default:
                throw new IllegalArgumentException("Invalid curve length");
        }
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
    public String getName() {
        return this.getCurveName();
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

    /**
     * Return point at infinity.
     *
     * @return Point at infinity
     */
    @Override
    public GroupElement getIdentity() {
        return inf;
    }

    /**
     * Encode a plaintext as an elliptic curve point.
     * <p>
     * We perform deterministic hashing. The algorithm for encoding the message is: {@code
     * * if the msg is larger than the point size + padding size, then the message is too large
     * 1. we allocate the room for second padding
     * 2. compute f(x) = x^3 - 3x + b
     * 3. set i = 0
     * 4. we take x as x = m || i,
     * 5. then we check if f(x) is quadratic residue, allowing us to take the square root of f(x).
     * 5.1 if not, let i += 1 and got to 4.
     * 5. the square root of f(x) will then become the y-coordinate of the point.
     * }
     *
     * @param msg
     * @return
     */
    @Override
    public GroupElement encode(Plaintext msg) {
        int limit = getFieldOrder().bitLength() - ENCODING_SUCCESS;

        BigInteger X = msg.toBigInteger();

        if (X.bitLength() > limit) {
            throw new IllegalArgumentException("Message is too large");
        }

        // We allocate room for encoding msg into a curve point
        X = X.shiftLeft(ENCODING_SUCCESS);

        return weierstrassQuadraticResidue(X);
    }

    /**
     * Returns result of Weierstrass equation y^2 = x^3 + ax + b
     * as modSqrt(y^2). Resulting group element is guaranteed to be
     * a quadratic residue modulo EC field order with a probability
     * of 99,9(9)%.
     *
     * @param x X coordinate of a curve
     * @return probably EC group element that is quadratic residue modulo EC field order
     */
    public ECGroupElement weierstrassQuadraticResidue(final BigInteger x) {
        int i = 0;
        int limit = BigInteger.valueOf(2).pow(ENCODING_SUCCESS).intValue();

        BigInteger X = x;
        BigInteger Y;

        // Loop until Y is a quadratic residue or limit is exceeded
        do {
            Y = weierstrassEquation(X);
            X = Y == null ? X.add(BigInteger.ONE) : X;
            i++;
        } while (Y == null && i < limit);

        Objects.requireNonNull(Y, "Y coordinate is null");

        // Y is guaranteed to be a quadratic residue
        return new ECGroupElement(this, X, Y);
    }

    /**
     * Returns smallest solution of Weierstrass equation y^2 = x^3 + ax + b, e.g. if Weierstrass solved
     * to (3, 5) and (3, 2) then since 5 > 2, 2 is returned.
     * Returns null if y^2 is not a perfect root.
     * @param x X coordinate of an elliptic curve point
     * @return smallest solution of Weierstrass equation or null if y^2 is not a perfect root
     */
    private BigInteger weierstrassEquation(final BigInteger x) {
        ECFieldElement XCoord = this.curve.fromBigInteger(x);
        ECFieldElement YCoord = XCoord
                    .square() // x^2
                    .add(this.curve.getA()) // (x^2 + a)
                    .multiply(XCoord) // x * (x^2 + a) == x^3 + ax
                    .add(this.curve.getB()) // x^3 + ax + b
                    .sqrt(); // SQRT(x^3 + ax + b) mod P

        // null means not a perfect square. Point may lie on a curve, but definitely not suitable for encoding
        // plaintexts, as only quadratic residues should be used
        if (YCoord == null)
            return null;

        // First solution of Weierstrass equation (X,Y)
        BigInteger Y = YCoord.toBigInteger();

        // Second solution of Weierstrass equation (X,-Y)
        BigInteger minusY = YCoord.negate().toBigInteger();

        // Take smallest solution
        // https://github.com/verificatum/verificatum-vcr/blob/97974cfc4ebbb323e49396222823e226cae2bebe/src/java/com/verificatum/arithm/ECqPGroup.magic#L485
        if (minusY.compareTo(Y) < 0) {
            return minusY;
        } else {
            return Y;
        }
    }

    @Override
    public Decodable isDecodable(GroupElement el) {
        // Is valid group element
        Decodable isValid = isGroupElement(el);
        if (isValid != Decodable.VALID) {
            return Decodable.INVALID_GROUP;
        }

        // Padded plaintext should be exactly of (group field bits length - 1),
        // "-1" because Java strips leading 0 bit in BigInteger
        if(decode((ECGroupElement) el) == null)
            return Decodable.INVALID_RANGE;

        return Decodable.VALID;
    }

    /**
     * Returns X coordinate of an elliptic curve if it is of correct bytes length, otherwise null.
     * @param msg
     * @return
     */
    private BigInteger decode(ECGroupElement msg) {
        ECPoint XY = msg.getPoint();

        BigInteger X = XY.getAffineXCoord().toBigInteger();

        if ((X.bitLength()) != this.getFieldOrder().bitLength() - 1) {
            return null;
        }

        return X;
    }

    @Override
    public Plaintext decode(GroupElement msg) {
        BigInteger X = decode((ECGroupElement) msg);
        if (X == null) {
            throw new IllegalArgumentException("Element is not decodable");
        }

        X = X.shiftRight(ENCODING_SUCCESS);

        return new Plaintext(X.toByteArray(), true);
    }

    @Override
    public Plaintext pad(Plaintext msg) {
        if (msg.padded) {
            return msg;
        }

        int fieldOrderBitLen = this.getFieldOrder().bitLength();

        // Padded plaintext looks like 0b01111111 0b11111111 ... 0b1...0PLAINTEXT_BITS00 0b00000000,
        // "-2" are first 0,1 bits and "-1" is last "0" that separates 0b1... and PLAINTEXT_BITS,
        // last 00 0b00000000 are encoding bits reserved to encode plaintext into EC point
        int maxPlaintextBitsLen = fieldOrderBitLen - 2 - 1 - ENCODING_SUCCESS;

        int plaintextBitLen = msg.getMessage().length * 8;

        // right now reserve 1 extra byte
        if (plaintextBitLen + 8> maxPlaintextBitsLen) {
            throw new IllegalArgumentException("Message is too large");
        }

        // Padding is 0b01111111 0b11111111 ... 0b1...0, see comment above
        int paddingBitLen = this.getFieldOrder().bitLength() - 1 - plaintextBitLen - ENCODING_SUCCESS;

        BigInteger padding = BigInteger.ONE;

        // 0b10000000 ... 0b00000000
        padding = padding.shiftLeft(paddingBitLen);

        // 0b01111111 ... 0b11111110
        padding = padding.subtract(BigInteger.TWO);

        byte[] paddingB = padding.toByteArray();

        byte[] paddedPlaintext = new byte[msg.getMessage().length + paddingB.length];

        System.arraycopy(paddingB, 0, paddedPlaintext, 0, paddingB.length);
        System.arraycopy(msg.getMessage(), 0, paddedPlaintext, paddingB.length, msg.getMessage().length);

        return new Plaintext(paddedPlaintext, true);
    }

    @Override
    public Plaintext unpad(Plaintext msg) {
        if (!msg.padded) {
            return msg;
        }

        byte[] XBytes = msg.getMessage();

        int mostSignificantByte = XBytes[0] >> 1;

        int zerosCount = Byte.SIZE - (Integer.SIZE - Integer.numberOfLeadingZeros(mostSignificantByte));
        int onesCount = Integer.bitCount(mostSignificantByte);

        // After leading zeros there should only be continuous stream of ones up until
        // padding end bit (0) is met, e.g. 0b00000111 0b11111111 ... 0b1110
        //
        // Therefore we assume that inside a single byte (8 bits), once we subtract all 1 bits,
        // there would only be leading 0 bits, so such combinations like 0b00000101 are not possible
        // since 1 bits count is indeed 2, leading 0 bits are 5 and therefore 8 - 2 = 6 which makes it
        // 6 != 5
        if (8 - onesCount != zerosCount) {
            throw new IllegalArgumentException("Message not padded correctly by inspecting first byte");
        }

        // Strip padding bytes, starting from second byte as first byte is already checked
        for (int i = 1; i < XBytes.length; i++) {
            switch (XBytes[i]) {
                case (byte) 0xFF: // padding byte
                    break;
                case (byte) 0xFE: // end of padding
                    // encoded plaintext contains only padding bytes, so decoded plaintext is ""
                    if (i + 1 == XBytes.length) {
                        return new Plaintext("");
                    }
                    byte[] unpaddedPlaintextB = Arrays.copyOfRange(XBytes, i + 1, XBytes.length);
                    BigInteger unpaddedPlaintext = new BigInteger(unpaddedPlaintextB);
                    return new Plaintext(unpaddedPlaintext, unpaddedPlaintext.bitLength(), false);
                default:
                    throw new IllegalArgumentException("Message has incorrect padding byte");
            }
        }

        throw new IllegalArgumentException("Unexpected padding");
    }

    @Override
    public Decodable isGroupElement(GroupElement el) {
        // Element belongs to group
        if (!this.equals(el.getGroup())) {
            return Decodable.INVALID_GROUP;
        }

        // Element is valid EC point
        ECGroupElement m = (ECGroupElement) el;
        ECPoint M = m.getPoint();
        if (!M.isValid()) {
            return Decodable.INVALID_POINT;
        }

        return Decodable.VALID;
    }

    /**
     * Serialize the group.
     * <p>
     * Returns GENERALSTRING of the curve name.
     *
     * @return
     */
    @Override
    public byte[] getBytes() {
        return new Field(name).encode();
    }

    /**
     * Get the curve name.
     *
     * @return
     */
    public String getCurveName() {
        return name;
    }

    ECPoint getPoint(BigInteger x, BigInteger y) {
        if (x == null || y == null || x.intValue() == -1 && y.intValue() == -1) {
            return curve.getInfinity();
        }
        return curve.createPoint(x, y);
    }

    ECPoint getInfinitePoint() {
        return curve.getInfinity();
    }

    ECCurve getCurve() {
        return this.curve;
    }

    /**
     * Get the base point of the group.
     *
     * @return
     */
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

    @Override
    public String toString() {
        return String.format("ECGroup(\"%s\")", getCurveName());
    }

    /**
     * Returns exactly "count" amount of EC group elements, which are not guaranteed to be quadratic residues.
     * EC group elements are derived using curve base point and random X coordinates.
     *
     * @param count amount of EC group elements to create
     * @param rng random X coordinates generator
     * @return array of EC group elements, which are not guaranteed to be quadratic residues
     */
    public ECGroupElement[] elementsOfRandomScalars(final int count, final Supplier<BigInteger> rng) {
        return IntStream
                .range(0, count)
                .mapToObj(i -> (ECGroupElement) getBasePoint().scale(rng.get()))
                .toArray(ECGroupElement[]::new);
    }

    /**
     * Returns exactly "count" amount of EC group elements, which are all guaranteed to be quadratic residues.
     * X coordinates are derived using deterministic pseudo random bytes generator.
     *
     * @param count amount of EC group elements to create
     * @param dprng deterministic pseudo random bytes generator
     * @return array of EC group elements which are guaranteed to be quadratic residues
     */
    public ECGroupElement[] pseudoRandomElements(final int count, final Supplier<byte[]> dprng) {
        List<ECGroupElement> ec = new ArrayList<>();

        // Loop until exactly count elements are generated
        while (ec.size() < count) {
            // Get X coordinate from DPRNG and apply modulo operation on returned value
            BigInteger X = new BigInteger(1, dprng.get()).mod(this.getFieldOrder());

            // Solve Weierstrass
            BigInteger Y = weierstrassEquation(X);

            if (Objects.isNull(Y)) {
                // don't collect Y coordinates that are not quadratic residues
                continue;
            }

            // Y is guaranteed to be quadratic residue
            ec.add(new ECGroupElement(this, X, Y));
        }

        // From List to array
        return ec.toArray(ECGroupElement[]::new);
    }
}
