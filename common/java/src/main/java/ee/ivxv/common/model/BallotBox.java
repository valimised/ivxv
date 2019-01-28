package ee.ivxv.common.model;

import com.fasterxml.jackson.annotation.JsonCreator;
import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.annotation.JsonProperty;
import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.function.BiConsumer;
import java.util.function.Supplier;
import java.util.stream.Collectors;
import java.util.stream.Stream;

/**
 * 
 */
public class BallotBox implements IBallotBox {

    private final String election;
    private Type type;
    private final Map<String, VoterBallots> ballots = new LinkedHashMap<>();

    public BallotBox(String election, Map<String, VoterBallots> ballots) {
        this(election, Type.INTEGRITY_CONTROLLED, ballots);
    }

    @JsonCreator
    public BallotBox( //
            @JsonProperty("election") String election, //
            @JsonProperty("type") Type type, //
            @JsonProperty("ballots") Map<String, VoterBallots> ballots) {
        this.election = election;
        this.type = type;
        this.ballots.putAll(ballots);
    }

    @Override
    public String getElection() {
        return election;
    }

    @Override
    public Type getType() {
        return type;
    }

    /**
     * @return Returns unmodifiable map of ballots.
     */
    public Map<String, VoterBallots> getBallots() {
        return Collections.unmodifiableMap(ballots);
    }

    @Override
    @JsonIgnore
    public int getNumberOfBallots() {
        return ballots.values().stream().mapToInt(vb -> vb.getBallots().size()).sum();
    }

    /**
     * Removes recurrent ballots, i.e. retains only the latest ballot for each voter.
     * 
     * @param cb Callback to be called on every removal of recurrent ballot.
     */
    public void removeRecurrentVotes(BiConsumer<String, Ballot> cb) {
        requireType(Type.INTEGRITY_CONTROLLED);
        ballots.forEach((vid, vb) -> vb.removeOldBallots(cb));
        type = Type.RECURRENT_VOTES_REMOVED;
    }

    /**
     * Revokes double ballots according to the revocation lists. The provided callback is called on
     * every revocation activity (revoke or restore). Revocation activity can be either successful
     * or unsuccessful. The latter case happens when voter or ballot does not exist or the ballot
     * has unexpected state - revoking a revoked ballot or restoring a non-revoked ballot.
     * 
     * @param rls Revocation lists
     * @param cb Callback to be called on every revocation activity.
     */
    public void revokeDoubleVotes(Stream<Supplier<RevocationList>> rls, RevokeCallback cb) {
        // Check type
        requireType(Type.RECURRENT_VOTES_REMOVED);
        // Mark voters' latest ballot revoked/restored according to revocation lists and report
        rls.forEach(rlSupplier -> {
            RevocationList rl = rlSupplier.get();
            rl.getPersons().forEach(vid -> {
                Ballot ballot = ballots.containsKey(vid) ? ballots.get(vid).getLatest() : null;
                boolean success = ballot != null && ballot.setRevokedState(rl.isRevoke());
                cb.call(vid, ballot, rl.isRevoke(), success);
            });
        });
        // Remove voters with revoked latest ballot
        ballots.values().removeIf(vb -> vb.getLatest().isRevoked());
        // Change type
        type = Type.DOUBLE_VOTERS_REMOVED;
    }

    /**
     * @param filter The filter to apply to all votes using parallel processing.
     * @return Returns an anonymous ballot box with the latest votes of this ballot box.
     */
    public AnonymousBallotBox anonymize(VoteFilter filter) {
        requireType(Type.DOUBLE_VOTERS_REMOVED);

        Map<String, Map<String, Map<String, List<byte[]>>>> anonymous = new LinkedHashMap<>();

        ballots.entrySet().parallelStream()
                .flatMap(ve -> Vote.streamOf(ve.getKey(), ve.getValue().getLatest())) // All votes
                .filter(v -> filter.accept(v.voterId, v.ballot, v.questionId, v.vote)) // Filter
                .collect(Collectors.toList()).stream() // Join threads, restore initial order
                .forEach(v -> anonymous
                        .computeIfAbsent(v.ballot.getDistrictId(), d -> new LinkedHashMap<>())
                        .computeIfAbsent(v.ballot.getStationId(), s -> new LinkedHashMap<>())
                        .computeIfAbsent(v.questionId, q -> new ArrayList<>()) //
                        .add(v.vote));

        return new AnonymousBallotBox(election, anonymous);
    }

    public interface RevokeCallback {
        void call(String voterId, Ballot b, boolean revoke, boolean success);
    }

    public interface VoteFilter {
        boolean accept(String voterId, Ballot b, String qid, byte[] vote);
    }

    private static class Vote {
        final String voterId;
        final Ballot ballot;
        final String questionId;
        final byte[] vote;

        Vote(String voterId, Ballot ballot, String questionId, byte[] vote) {
            this.voterId = voterId;
            this.ballot = ballot;
            this.questionId = questionId;
            this.vote = vote;
        }

        static Stream<Vote> streamOf(String voterId, Ballot ballot) {
            return ballot.getVotes().entrySet().stream()
                    .map(e -> new Vote(voterId, ballot, e.getKey(), e.getValue()));
        }
    }

}
