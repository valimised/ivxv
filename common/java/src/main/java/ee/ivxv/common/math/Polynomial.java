package ee.ivxv.common.math;

import ee.ivxv.common.crypto.rnd.Rnd;
import java.io.IOException;
import java.math.BigInteger;

public class Polynomial {
    private final BigInteger[] coefs;
    private final BigInteger modulus;

    public Polynomial(BigInteger modulus, BigInteger[] coefs) throws IllegalArgumentException {
        verifyModulus(modulus);
        this.modulus = modulus;
        this.coefs = coefs;
    }

    public Polynomial(int degree, BigInteger modulus, Rnd rnd)
            throws IllegalArgumentException, IOException {
        if (degree < 0) {
            throw new IllegalArgumentException("Degree of a polynomial must be non-negative");
        }
        verifyModulus(modulus);
        this.modulus = modulus;
        this.coefs = new BigInteger[degree + 1];
        for (int i = 0; i <= degree; i++) {
            this.coefs[i] = IntegerConstructor.construct(rnd, modulus);
        }
    }

    public Polynomial(int degree, BigInteger modulus, BigInteger free, Rnd rnd)
            throws IllegalArgumentException, IOException {
        if (degree < 0) {
            throw new IllegalArgumentException("Degree of a polynomial must be non-negative");
        }
        verifyModulus(modulus);
        this.modulus = modulus;
        this.coefs = new BigInteger[degree + 1];
        this.coefs[0] = free;
        for (int i = 1; i <= degree; i++) {
            this.coefs[i] = IntegerConstructor.construct(rnd, modulus);
        }
    }

    // throw an exception if the modulus is not permitted
    static void verifyModulus(BigInteger modulus) throws IllegalArgumentException {
        if (modulus.compareTo(BigInteger.ZERO) <= 0) {
            throw new IllegalArgumentException("Modulus must be positive");
        }
    }

    public BigInteger evaluate(BigInteger arg) {
        BigInteger res = BigInteger.ZERO;
        for (int i = 0; i < this.coefs.length; i++) {
            res = res.add(evaluateMonomial(arg, i)).mod(this.modulus);
        }
        return res;
    }

    public BigInteger evaluate(int arg) {
        return evaluate(BigInteger.valueOf(arg));
    }

    private BigInteger evaluateMonomial(BigInteger arg, int degree) {
        if (degree >= this.coefs.length) {
            return BigInteger.ZERO;
        }
        return arg.modPow(BigInteger.valueOf(degree), this.modulus).multiply(this.coefs[degree])
                .mod(this.modulus);
    }
}
