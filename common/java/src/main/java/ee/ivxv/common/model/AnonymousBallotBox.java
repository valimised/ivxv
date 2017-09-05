package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class AnonymousBallotBox implements IBallotBox {

    private final String election;
    private final Map<String, Map<String, Map<String, List<byte[]>>>> districts;

    @JsonCreator
    public AnonymousBallotBox( //
            @JsonProperty("election") String election, //
            @JsonProperty("districts") Map<String, Map<String, Map<String, List<byte[]>>>> districts) {
        this.election = election;
        this.districts = Collections.unmodifiableMap(districts);
    }

    @Override
    public String getElection() {
        return election;
    }

    @Override
    @JsonIgnore
    public Type getType() {
        return Type.ANONYMIZED;
    }

    /**
     * @return Returns the number of votes, because ballots are unknown quantity here.
     */
    @Override
    @JsonIgnore
    public int getNumberOfBallots() {
        return districts.values().stream().mapToInt( //
                s -> s.values().stream()
                        .mapToInt(q -> q.values().stream().mapToInt(b -> b.size()).sum()).sum())
                .sum();
    }

    @JsonIgnore
    public int getNumberOfQuestions() {
        Map<String, Boolean> res = new HashMap<>();
        districts.forEach((d, sMap) -> sMap
                .forEach((s, qMap) -> qMap.keySet().forEach((q) -> res.putIfAbsent(q, true))));
        return res.size();
    }

    /**
     * @return Returns a map from district id to a map from station id to a map from question id to
     *         list of encrypted votes.
     */
    public Map<String, Map<String, Map<String, List<byte[]>>>> getDistricts() {
        return districts;
    }

}
