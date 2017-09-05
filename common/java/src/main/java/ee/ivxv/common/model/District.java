package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.Collections;
import java.util.List;

public class District {

    private final String name;
    private final List<String> stations;

    @JsonCreator
    public District( //
            @JsonProperty("name") String name, //
            @JsonProperty("stations") List<String> stations) {
        this.name = name;
        this.stations = Collections.unmodifiableList(stations);
    }

    public String getName() {
        return name;
    }

    public List<String> getStations() {
        return stations;
    }

}
