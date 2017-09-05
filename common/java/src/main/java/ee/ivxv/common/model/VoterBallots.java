package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.function.BiConsumer;

public class VoterBallots {

    private final String voterId;
    private final List<Ballot> ballots;
    private final Ballot latest;

    @JsonCreator
    public VoterBallots( //
            @JsonProperty("voterId") String voterId, //
            @JsonProperty("ballots") List<Ballot> ballots) {
        this.voterId = voterId;
        this.ballots = createSortedList(ballots);
        latest = this.ballots.isEmpty() ? null : this.ballots.get(this.ballots.size() - 1);
    }

    private static List<Ballot> createSortedList(List<Ballot> ballots) {
        List<Ballot> result = new ArrayList<>(ballots);
        Collections.sort(result, (b1, b2) -> b1.getTime().compareTo(b2.getTime()));
        return result;
    }

    public String getVoterId() {
        return voterId;
    }

    /**
     * @return Returns chronologically ordered unmodifiable list of ballots.
     */
    public List<Ballot> getBallots() {
        return Collections.unmodifiableList(ballots);
    }

    @JsonIgnore
    public Ballot getLatest() {
        return latest;
    }

    void removeOldBallots(BiConsumer<String, Ballot> cb) {
        while (ballots.size() > 1) {
            cb.accept(voterId, ballots.remove(0));
        }
    }

}
