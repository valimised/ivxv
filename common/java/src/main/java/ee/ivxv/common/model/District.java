package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.Collections;
import java.util.List;

public class District {

    private final String name;
    private final List<String> parish;

    @JsonCreator
    public District( //
            @JsonProperty("name") String name, //
            @JsonProperty("parish") List<String> parish) {
        this.name = name;
        this.parish = Collections.unmodifiableList(parish);
    }

    public String getName() {
        return name;
    }

    public List<String> getParish() {
        return parish;
    }

}
