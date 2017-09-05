package ee.ivxv.key.protocol;

public class ThresholdParameters {
    private final int parties;
    private final int threshold;

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

    public int getParties() {
        return this.parties;
    }

    public int getThreshold() {
        return this.threshold;
    }
}
