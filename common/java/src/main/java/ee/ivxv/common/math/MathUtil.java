package ee.ivxv.common.math;

import ee.ivxv.common.util.Util;
import java.math.BigInteger;

public class MathUtil {
    public static BigInteger factorial(int f) {
        return factorial(BigInteger.valueOf(f));
    }

    public static BigInteger factorial(BigInteger f) {
        if (f.compareTo(BigInteger.ONE) <= 0) {
            return BigInteger.ONE;
        }
        BigInteger res = BigInteger.ONE;
        while (f.compareTo(BigInteger.ONE) > 0) {
            res = res.multiply(f);
            f = f.subtract(BigInteger.ONE);
        }
        return res;
    }
    
    public static BigInteger gcd(BigInteger... v) {
        if (v.length == 0) {
            return BigInteger.ZERO;
        } else if (v.length == 1) {
            return v[0];
        }
        BigInteger res = v[0];
        for (int i = 1; i < v.length; i++) {
            res = res.gcd(v[i]);
        }
        return res;
    }
    
    public static BigInteger lcm(BigInteger... v) {
        if (v.length == 0) {
            return BigInteger.ZERO;
        } else if (v.length == 1) {
            return v[0];
        }
        BigInteger p = v[0];
        for (int i = 1; i < v.length; i++) {
            p = p.multiply(v[i]);
        }
        BigInteger res = p.divide(gcd(v));
        return res;
    }

    public static BigInteger[] extendedEuclidean(BigInteger a, BigInteger b) {
        /*-
         * https://en.wikipedia.org/wiki/Extended_Euclidean_algorithm#Pseudocode
         * function extended_gcd(a, b)
         * s := 0;    old_s := 1
         * t := 1;    old_t := 0
         * r := b;    old_r := a
         * while r ≠ 0
         *    quotient := old_r div r
         *    (old_r, r) := (r, old_r - quotient * r)
         *    (old_s, s) := (s, old_s - quotient * s)
         *    (old_t, t) := (t, old_t - quotient * t)
         * output "Bézout coefficients:", (old_s, old_t)
         * output "greatest common divisor:", old_r
         * output "quotients by the gcd:", (t, s)
         */
        BigInteger q, s, t, r, old_s, old_r, old_t, temp;
        s = BigInteger.ZERO;
        t = BigInteger.ONE;
        r = b;
        old_s = BigInteger.ONE;
        old_t = BigInteger.ZERO;
        old_r = a;
        while (r.compareTo(BigInteger.ZERO) != 0) {
            q = old_r.divide(r);

            temp = r;
            r = old_r.subtract(q.multiply(r));
            old_r = temp;

            temp = s;
            s = old_s.subtract(q.multiply(s));
            old_s = temp;

            temp = t;
            t = old_t.subtract(q.multiply(t));
            old_t = temp;
        }
        return new BigInteger[] {old_s, old_t};
    }

    public static BigInteger phiSemiprime(BigInteger p, BigInteger q) {
        // compute Euler's phi for p*q. We assume that p, q prime.
        return p.subtract(BigInteger.ONE).multiply(q.subtract(BigInteger.ONE));
    }

    public static int legendre(BigInteger e, BigInteger p) {
        BigInteger q = Util.safePrimeOrder(p);
        BigInteger big = e.modPow(q, p);
        if (big.compareTo(BigInteger.ONE) == 0) {
            return 1;
        } else if (big.compareTo(BigInteger.ZERO) == 0) {
            return 0;
        } else if (big.compareTo(p.subtract(BigInteger.ONE)) == 0) {
            return -1;
        }

        // should not reach here, but return something to silence compiler
        // warning
        return 1;
    }

    public BigInteger weierstrass(BigInteger a, BigInteger b, BigInteger x, BigInteger p) {
        BigInteger res = x.modPow(BigInteger.valueOf(2), p).add(a).multiply(x).mod(p).add(b).mod(p);
        return res;
    }
}
