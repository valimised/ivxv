package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.ArrayList;
import java.util.List;

public class RevocationList {

    static final String TYPE_REVOKE = "revoke";

    private final String election;
    private final String type;
    private final List<String> persons;

    @JsonCreator
    public RevocationList(//
            @JsonProperty("election") String election, //
            @JsonProperty("type") String type, //
            @JsonProperty("persons") List<String> persons) {
        this.election = election;
        this.type = type;
        this.persons = new ArrayList<>(persons);
    }

    public String getElection() {
        return election;
    }

    public String getType() {
        return type;
    }

    public List<String> getPersons() {
        return persons;
    }

    public boolean isRevoke() {
        return TYPE_REVOKE.equals(type);
    }

}
