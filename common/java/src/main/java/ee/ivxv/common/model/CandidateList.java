package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.Collections;
import java.util.Map;

public class CandidateList {
    private final String election;
    private final Map<String, Map<String, Map<String, String>>> candidates;

    @JsonCreator
    public CandidateList( //
            @JsonProperty("election") String election, //
            @JsonProperty("choices") Map<String, Map<String, Map<String, String>>> districts) {
        this.election = election;
        this.candidates = Collections.unmodifiableMap(districts);
    }

    public String getElection() {
        return election;
    }

    /**
     * @return Returns a map from district id to a map from party name to a map from candidate id to
     *         candidate name.
     */
    public Map<String, Map<String, Map<String, String>>> getCandidates() {
        return candidates;
    }

    @JsonIgnore
    public int getCount() {
        return candidates.values().stream()
                .mapToInt(d -> d.values().stream().mapToInt(p -> p.size()).sum()).sum();
    }
}
