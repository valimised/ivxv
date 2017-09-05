package ee.ivxv.common.math;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

public class LagrangeInterpolation {
    /*
     * interpolation is not required in the current context, so we omit the complexity of computing
     * the coeficcients public static Polynomial interpolate() { }
     */

    public static BigInteger basisPolynomial(BigInteger modulus, Object[] points, BigInteger base) {
        List<BigInteger> realPoints = new ArrayList<>();
        for (int i = 0; i < points.length; i++) {
            if (points[i] != null) {
                realPoints.add(BigInteger.valueOf(i + 1));
            }
        }
        return genericBasisPolynomial(modulus,
                realPoints.toArray(new BigInteger[realPoints.size()]), base);
    }

    private static BigInteger genericBasisPolynomial(BigInteger modulus, BigInteger[] points,
            BigInteger base) throws IllegalArgumentException {
        BigInteger nominator = new BigInteger("1");
        BigInteger denominator = new BigInteger("1");
        for (int i = 0; i < points.length; i++) {
            if (base.compareTo(points[i]) != 0) {
                nominator = nominator.multiply(points[i]).mod(modulus);
                denominator = denominator.multiply(base.subtract(points[i])).mod(modulus);
            }
        }
        if (points.length % 2 == 0) {
            nominator = nominator.negate().mod(modulus);
        }
        return denominator.modInverse(modulus).multiply(nominator).mod(modulus);
    }

    public static BigInteger basisInverselessPolynomial(Object[] points, BigInteger base) {
        List<BigInteger> realPoints = new ArrayList<>();
        for (int i = 0; i < points.length; i++) {
            if (points[i] != null) {
                realPoints.add(BigInteger.valueOf(i + 1));
            }
        }
        return genericBasisInverselessPolynomial(points.length,
                realPoints.toArray(new BigInteger[realPoints.size()]), base);
    }

    private static BigInteger genericBasisInverselessPolynomial(int pointLimit, BigInteger[] points,
            BigInteger base) throws IllegalArgumentException {
        BigInteger nominator = new BigInteger("1");
        BigInteger denominator = new BigInteger("1");
        for (int i = 0; i < points.length; i++) {
            if (base.compareTo(points[i]) != 0) {
                nominator = nominator.multiply(points[i]);
                denominator = denominator.multiply(base.subtract(points[i]));
            }
        }
        if (points.length % 2 == 0) {
            nominator = nominator.negate();
        }
        // instead of finding inverse (which may not even exist), we multiply
        // the nominator with factorial(pointLimit). The evaluation points are
        // 1, ..., pointLimit, then as
        // - |x_i - x_j| < pointLimit
        // - |x_i - x_j| are distinct for j=1, ..., n
        // then we can do integer division without losing information.
        nominator = nominator.multiply(MathUtil.factorial(pointLimit));
        return nominator.divide(denominator);
    }
}
