package ee.ivxv.common.math;

import ee.ivxv.common.crypto.rnd.Rnd;
import java.io.IOException;
import java.math.BigInteger;

public class IntegerConstructor {
    public static BigInteger construct(Rnd rnd, BigInteger upper)
            throws IllegalArgumentException, IOException {
        if (upper.signum() != 1) {
            throw new IllegalArgumentException("Nonpositive limit");
        }
        int noBytes = (upper.bitLength() + 7) / 8;
        int maskLen = upper.bitLength() % 8;
        maskLen = maskLen == 0 ? 8 : maskLen;

        byte[] bytes = new byte[noBytes];
        BigInteger n;
        while (true) {
            rnd.read(bytes, 0, bytes.length);
            bytes[0] &= (byte) ((1 << maskLen) - 1);
            n = new BigInteger(1, bytes);
            if (n.compareTo(upper) < 0) {
                return n;
            }
        }
    }

    public static BigInteger constructPrime(Rnd rnd, BigInteger upper)
            throws IllegalArgumentException, IOException {
        BigInteger res;
        while (true) {
            while (true) {
                res = construct(rnd, upper);
                if (res.isProbablePrime(20)) {
                    // quickcheck
                    break;
                }
            }
            if (res.isProbablePrime(100)) {
                // slow check
                return res;
            }
        }
    }
}
