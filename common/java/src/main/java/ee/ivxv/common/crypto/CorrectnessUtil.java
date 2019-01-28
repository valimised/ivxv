package ee.ivxv.common.crypto;

import ee.ivxv.common.crypto.elgamal.ElGamalCiphertext;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.math.GroupElement;

/**
 * CorrectnessUtil is a utility class for checking the correctness of ElGamal ciphertexts.
 */
public class CorrectnessUtil {
    /**
     * Ciphertext correctness results.
     */
    public enum CiphertextCorrectness {
        // if adding enum values, update
        // ee.ivxv.processor.util.ReportHelper#translate(CiphertextCorrectness) method.
        /**
         * Valid ciphertext
         */
        VALID,
        /**
         * Byte array not decodable as ElGamalCiphertext
         */
        INVALID_BYTES,
        /**
         * Value is not element of expected group
         */
        INVALID_GROUP,
        /**
         * Value is out of range.
         */
        INVALID_RANGE,
        /**
         * Value is not quadratic residue.
         */
        INVALID_QR,
        /**
         * Value is not a point on the curve.
         */
        INVALID_POINT,
        /**
         * Invalid ciphertext.
         */
        INVALID;
    }

    /**
     * Verify the correctness of the serialized ciphertext.
     * <p>
     * Verify that the ciphertext is correct and safe to use.
     * 
     * @param pk The expected public key which was used to encode the ciphertext
     * @param ctb Serialized ciphertext to check for correctness.
     * @return Correctness check value.
     */
    public static CiphertextCorrectness isValidCiphertext(ElGamalPublicKey pk, byte[] ctb) {
        ElGamalCiphertext ciphertext;
        try {
            ciphertext = new ElGamalCiphertext(pk.getParameters(), ctb);
        } catch (IllegalArgumentException e) {
            return CiphertextCorrectness.INVALID_BYTES;
        }
        for (GroupElement el : new GroupElement[] {ciphertext.getBlind(),
                ciphertext.getBlindedMessage()}) {
            switch (pk.getParameters().getGroup().isDecodable(el)) {
                case VALID:
                    continue;
                case INVALID_GROUP:
                    return CiphertextCorrectness.INVALID_GROUP;
                case INVALID_RANGE:
                    return CiphertextCorrectness.INVALID_RANGE;
                case INVALID_QR:
                    return CiphertextCorrectness.INVALID_QR;
                case INVALID_POINT:
                    return CiphertextCorrectness.INVALID_POINT;
                default:
                    return CiphertextCorrectness.INVALID;
            }
        }
        return CiphertextCorrectness.VALID;
    }
}
