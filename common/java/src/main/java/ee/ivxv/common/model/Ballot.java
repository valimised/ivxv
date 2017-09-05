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
    private boolean isRevoked;

    @JsonCreator
    public Ballot( //
            @JsonProperty("id") String id, //
            @JsonProperty("time") String time, //
            @JsonProperty("version") String version, //
            @JsonProperty("name") String name, //
            @JsonProperty("districtId") String districtId, //
            @JsonProperty("stationId") String stationId, //
            @JsonProperty("rowNumber") Long rowNumber, //
            @JsonProperty("votes") Map<String, byte[]> votes) {
        this(id, Instant.parse(time), version,
                new Voter(null, name, null, districtId, stationId, rowNumber), votes);
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

    @JsonIgnore
    public LName getStation() {
        return voter.getStation();
    }

    public String getStationId() {
        return voter.getStation().getId();
    }

    public Long getRowNumber() {
        return voter.getRowNumber();
    }

    /**
     * @return Returns the map from question id to the encrypted vote
     */
    public Map<String, byte[]> getVotes() {
        return votes;
    }

    boolean isRevoked() {
        return isRevoked;
    }

    /**
     * Sets the revoked status of the ballot if the status is not already the same as the parameter.
     * 
     * @param isRevoked
     * @return Returns whether the operation was successful.
     */
    boolean setRevokedState(boolean isRevoked) {
        if (this.isRevoked == isRevoked) {
            return false;
        }
        this.isRevoked = isRevoked;
        return true;
    }

}
