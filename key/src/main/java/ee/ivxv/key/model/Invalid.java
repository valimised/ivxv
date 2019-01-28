package ee.ivxv.key.model;

import java.util.ArrayList;
import java.util.List;

/**
 * JSON serializable list of invalid votes.
 */
public class Invalid {
    private final String election;
    private final List<Vote> invalid = new ArrayList<>();

    /**
     * Initialize using election identifier.
     * 
     * @param election
     */
    public Invalid(String election) {
        this.election = election;
    }

    /**
     * Get the election identifier.
     * 
     * @return
     */
    public String getElection() {
        return election;
    }

    /**
     * Get the list of invalid votes.
     * 
     * @return
     */
    public List<Vote> getInvalid() {
        return invalid;
    }
}
