package ee.ivxv.key.model;

import java.util.ArrayList;
import java.util.List;

public class Invalid {
    private final String election;
    private final List<Vote> invalid = new ArrayList<>();

    public Invalid(String election) {
        this.election = election;
    }

    public String getElection() {
        return election;
    }

    public List<Vote> getInvalid() {
        return invalid;
    }
}
