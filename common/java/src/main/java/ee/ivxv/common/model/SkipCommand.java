package ee.ivxv.common.model;

public class SkipCommand {

    private final String changeset;
    private final String election;
    private final String skip_voter_list;

    public SkipCommand(
        String changeset, String election, String skip_voter_list) {
        this.changeset = changeset;
        this.election = election;
        this.skip_voter_list = skip_voter_list;
    }

    public String getElection() {
        return election;
    }

    public String getChangeset() {
        return changeset;
    }

    public String getSkipVoterList() {
        return skip_voter_list;
    }

}
