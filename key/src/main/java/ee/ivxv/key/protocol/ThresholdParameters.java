package ee.ivxv.key.protocol;

/**
 * Parameters for operations with threshold.
 */
public class ThresholdParameters {
    private final int parties;
    private final int threshold;

    /**
     * Initialize the threshold parameters.
     * 
     * @param parties The number of parties of the protocol.
     * @param threshold The number of parties required for reconstruction.
     * @throws IllegalArgumentException If parties < 2*threshold-1
     */
    public ThresholdParameters(int parties, int threshold) throws IllegalArgumentException {
        if (parties <= 0) {
            throw new IllegalArgumentException("Number of parties must be positive");
        }
        if (parties < (2 * threshold - 1)) {
            throw new IllegalArgumentException(
                    "Number of parties must be higher than 2*threshold - 1");
        }
        this.parties = parties;
        this.threshold = threshold;
    }

    /**
     * Get the number of parties.
     * 
     * @return
     */
    public int getParties() {
        return this.parties;
    }

    /**
     * Get the threshold for reconstructing the output.
     * 
     * @return
     */
    public int getThreshold() {
        return this.threshold;
    }
}
