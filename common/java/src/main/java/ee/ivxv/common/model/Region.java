package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * Region represents a municipal region.
 * 
 * <p>
 * Example: <code>
 *     {
 *       "county": "Viljandi maakond", 
 *       "parish": "Abja vald", 
 *       "state": "Eesti Vabariik"
 *     }
 * </code>
 */
public class Region {

    private final String county;
    private final String parish;
    private final String state;

    @JsonCreator
    public Region( //
            @JsonProperty("county") String county, //
            @JsonProperty("parish") String parish, //
            @JsonProperty("state") String state) {
        this.county = county;
        this.parish = parish;
        this.state = state;
    }

    public String getCounty() {
        return county;
    }

    public String getParish() {
        return parish;
    }

    public String getState() {
        return state;
    }

}
