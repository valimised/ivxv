package ee.ivxv.common.model;

import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

public class VoterList {

    public static final String TYPE_INITIAL = "algne";

    private final VoterList parent;
    private final String hash;
    private final String name;
    private final String versionNumber;
    private final String electionId;
    private final String type;
    private final Map<String, Voter> added;
    private final Map<String, Voter> removed;

    public VoterList(VoterList parent, String hash, String name, String versionNumber,
            String electionId, String type, List<Voter> voters) {
        this.parent = parent;
        this.hash = hash;
        this.name = name;
        this.versionNumber = versionNumber;
        this.electionId = electionId;
        this.type = type;
        Map<String, Voter> a = new LinkedHashMap<>();
        Map<String, Voter> r = new LinkedHashMap<>();
        voters.forEach(v -> (v.isAddition() ? a : r).put(v.getCode(), v));
        added = Collections.unmodifiableMap(a);
        removed = Collections.unmodifiableMap(r);
    }

    public VoterList getParent() {
        return parent;
    }

    public String getHash() {
        return hash;
    }

    public String getName() {
        return name;
    }

    public String getVersionNumber() {
        return versionNumber;
    }

    public String getElectionId() {
        return electionId;
    }

    public String getType() {
        return type;
    }

    public Map<String, Voter> getAdded() {
        return added;
    }

    public Map<String, Voter> getRemoved() {
        return removed;
    }

    public boolean isInitial() {
        return TYPE_INITIAL.equals(type);
    }

    /**
     * Tries to find the active voter from a voter list that is valid in the specified voter list.
     * 
     * @param voterId
     * @param vlHash the hash of the desired voter list
     * @return Returns the voter with the specified code that is active in the specified voter list
     *         or <tt>null</tt> if either such voter list does not exist of the voter is not active.
     */
    public Voter find(String voterId, String vlHash) {
        if (getHash().equals(vlHash)) {
            return find(voterId);
        }
        return Optional.ofNullable(parent).map(p -> p.find(voterId, vlHash)).orElse(null);
    }

    /**
     * Tries to find the voter from this voter list or it's parent.
     * 
     * @param voterId
     * @return The voter instance that is active in this voter list or <tt>null</tt>.
     */
    public Voter find(String voterId) {
        Voter v = added.get(voterId);

        if (v != null) {
            return v;
        }
        if (removed.containsKey(voterId)) {
            return null;
        }
        return Optional.ofNullable(parent).map(p -> p.find(voterId)).orElse(null);
    }

}
