package ee.ivxv.common.model;

import ee.ivxv.common.service.bbox.Result;
import ee.ivxv.common.service.bbox.impl.ResultException;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

public class VoterList {

    public static final String TYPE_INITIAL = "0";

    private final VoterList parent;
    private final String name;
    private final String versionNumber;
    private final String electionId;
    private final String changeset;
    private final Map<String, Voter> added;
    private final Map<String, Voter> removed;

    public VoterList(VoterList parent, String name, String versionNumber,
            String electionId, String changeset, List<Voter> voters) {
        this.parent = parent;
        this.name = name;
        this.versionNumber = versionNumber;
        this.electionId = electionId;
        this.changeset = changeset;
        Map<String, Voter> a = new LinkedHashMap<>();
        Map<String, Voter> r = new LinkedHashMap<>();
        voters.forEach(v -> {
            if (a.containsKey(v.getCode()) && r.containsKey(v.getCode())) {
                a.remove(v.getCode());
                r.remove(v.getCode());
            }

            if (v.isAddition()) {
                a.put(v.getCode(), v);
            } else {
                r.put(v.getCode(), v);
            }
        });
        added = Collections.unmodifiableMap(a);
        removed = Collections.unmodifiableMap(r);
    }

    public VoterList getParent() {
        return parent;
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

    public String getChangeset() {
        return changeset;
    }

    public Map<String, Voter> getAdded() {
        return added;
    }

    public Map<String, Voter> getRemoved() {
        return removed;
    }

    public boolean isInitial() {
        return TYPE_INITIAL.equals(changeset);
    }

    /**
     * Tries to find the active voter from a voter list that is valid in the specified voter list.
     *
     * @param voterId
     * @param changeSet the change number of the desired voter list
     * @return Returns the voter with the specified code that is active in the specified voter list
     * 	or <tt>null</tt> if either such voter list does not exist of the voter is not active.
     */
    public Voter find(String voterId, String changeSet) throws ResultException {
        if (getChangeset().equals(changeSet)) {
            return find(voterId);
        }
        if (parent == null) {
            throw new ResultException(Result.VOTERLIST_NOT_FOUND, changeSet);
        }
        return parent.find(voterId, changeSet);
    }

    /**
     * Tries to find the voter from this voter list or it's parent.
     *
     * @param voterId
     * @return The voter instance that is active in this voter list or <tt>null</tt>.
     */
    public Voter find(String voterId) {

        Voter prev = Optional.ofNullable(parent).map(p -> p.find(voterId)).orElse(null);
        Voter current_add = added.get(voterId);
        Voter current_del = removed.get(voterId);

        // no change
        if (current_add == null && current_del == null) {
            return prev;
        }

        // prev exists - change station
        // otherwise nop
        if (current_add != null && current_del != null) {
            if (prev == null) {
                return null;
            } else {
                return current_add;
            }
        }

        if (current_add != null) {
            return current_add;
        }

        return null;
    }

}
