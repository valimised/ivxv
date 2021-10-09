package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import java.time.Instant;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;

/**
 * Ballot represents the choices on a single ballot.
 */
public class Ballot {

    private final String id;
    private final Instant time;
    private final String version;
    private final Voter voter;
    private final Map<String, byte[]> votes;
    private boolean isInvalid;

    @JsonCreator
    public Ballot( //
            @JsonProperty("id") String id, //
            @JsonProperty("time") String time, //
            @JsonProperty("version") String version, //
            @JsonProperty("name") String name, //
            @JsonProperty("districtId") String districtId, //
            @JsonProperty("parish") String parishId, //
            @JsonProperty("votes") Map<String, byte[]> votes) {
        this(id, Instant.parse(time), version,
                new Voter(null, name, null, parishId, new LName(districtId)), votes);
    }

    public Ballot(String id, Instant time, String version, Voter voter, Map<String, byte[]> votes) {
        this.id = id;
        this.time = time;
        this.version = version;
        this.voter = voter;
        this.votes = Collections.unmodifiableMap(new LinkedHashMap<>(votes));
    }

    public String getId() {
        return id;
    }

    public Instant getTime() {
        return time;
    }

    public String getVersion() {
        return version;
    }

    public String getName() {
        return voter.getName();
    }

    @JsonIgnore
    public LName getDistrict() {
        return voter.getDistrict();
    }

    public String getDistrictId() {
        return voter.getDistrict().getId();
    }

    public String getParish() {
        return voter.getParish();
    }

    /**
     * @return Returns the map from question id to the encrypted vote
     */
    public Map<String, byte[]> getVotes() {
        return votes;
    }

    boolean isInvalid() {
        return isInvalid;
    }

    /**
     * Sets the invalid status of the ballot if the status is not already the same as the parameter.
     *
     * @param isInvalid
     * @return Returns whether the operation was successful.
     */
    boolean setInvalidState(boolean isInvalid) {
        if (this.isInvalid == isInvalid) {
            return false;
        }
        this.isInvalid = isInvalid;
        return true;
    }


}
