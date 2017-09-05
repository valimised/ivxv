package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;

public class DistrictList {

    private final String election;
    private final Map<String, District> districts;
    private final Map<String, Region> regions;

    @JsonCreator
    public DistrictList( //
            @JsonProperty("election") String election, //
            @JsonProperty("districts") Map<String, District> districts, //
            @JsonProperty("regions") Map<String, Region> regions) {
        this.election = election;
        this.districts = Collections.unmodifiableMap(new LinkedHashMap<>(districts));
        this.regions = Collections.unmodifiableMap(new LinkedHashMap<>(regions));
    }

    public String getElection() {
        return election;
    }

    public Map<String, District> getDistricts() {
        return districts;
    }

    public Map<String, Region> getRegions() {
        return regions;
    }

    @JsonIgnore
    public int getCount() {
        return districts.size();
    }

}
