package ee.ivxv.key.protocol;

import ee.ivxv.common.math.Polynomial;
import java.math.BigInteger;

/**
 * Utility functions for threshold protocols.
 */
public class ProtocolUtil {
    /**
     * Evaluate the polynomial at points to generate shares for Lagrange interpolation.
     * 
     * @param pol Polynomial to secret share
     * @param amount The number of evaluation points to generate.
     * @return List of polynomial evaluations.
     */
    public static BigInteger[] generateShares(Polynomial pol, int amount) {
        BigInteger[] shares = new BigInteger[amount];
        for (int i = 0; i < amount; i++) {
            shares[i] = pol.evaluate(i + 1);
        }
        return shares;
    }
}
