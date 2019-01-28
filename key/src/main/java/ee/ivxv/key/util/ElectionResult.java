package ee.ivxv.key.util;

import ee.ivxv.common.model.CandidateList;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.model.Proof;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.service.report.Reporter;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Json;
import ee.ivxv.common.util.Util;
import ee.ivxv.key.model.Invalid;
import ee.ivxv.key.model.Tally;
import ee.ivxv.key.model.Vote;
import ee.ivxv.key.protocol.SigningProtocol;
import java.io.ByteArrayOutputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.Callable;
import java.util.concurrent.LinkedBlockingQueue;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Helper class for processing election result.
 */
public class ElectionResult {
    static final Logger log = LoggerFactory.getLogger(ElectionResult.class);

    private static final Path INVALID_VOTE_PATH = Paths.get("invalid");
    private static final Path PROOF_PATH = Paths.get("proof");
    private static final String TALLY_SUFFIX = ".tally";
    private static final String SIGNATURE_SUFFIX = TALLY_SUFFIX + ".signature";

    private final String electionName;
    private final CandidateList candidates;
    private final DistrictList districts;
    private final boolean withProof;
    private final BlockingQueue<Object> votes = new LinkedBlockingQueue<>();
    private final Map<String, Tally> tallySet;
    private final Proof proof;
    private final Invalid invalid;

    /**
     * Initialize using values.
     * 
     * @param electionName Election identifier
     * @param candidates Candidates list
     * @param districts Districts list
     * @param withProof Boolean indicating if encrypted ballots should be decrypted with proof of
     *        correct decryption.
     */
    public ElectionResult(String electionName, CandidateList candidates, DistrictList districts,
            boolean withProof) {
        this.electionName = electionName;
        this.candidates = candidates;
        this.districts = districts;
        this.withProof = withProof;
        this.tallySet = new HashMap<>();
        this.proof = withProof ? new Proof(electionName) : null;
        this.invalid = new Invalid(electionName);
    }

    /**
     * Get the worker for computing the tally.
     * <p>
     * The worker works in parallel to decryptions and continuously computes the tally. It separates votes which are invalid or are given for invalid candidates
     * 
     * @param voteCount
     * @param console
     * @param reporter
     * @return
     */
    public ResultWorker getResultWorker(int voteCount, I18nConsole console, Reporter reporter) {
        return new ResultWorker(voteCount, console, reporter);
    }

    /**
     * Set the end of incoming votes.
     * <p>
     * After the last vote has been added, set the end marker indicating the worker thread to stop
     * waiting for incoming votes.
     */
    public void setEot() {
        votes.add(Util.EOT);
    }

    /**
     * Add a vote for worker to process.
     * 
     * @param vote
     */
    public void addVote(Vote vote) {
        votes.add(vote);
    }

    /**
     * Output tally and the corresponding signature.
     * 
     * @param outDir Output directory to store the tally and signature.
     * @param signer Protocol for constructing the signature for the tally.
     * @throws Exception When writing or communication with card tokens fail.
     */
    public void outputTally(Path outDir, SigningProtocol signer) throws Exception {
        for (Map.Entry<String, Tally> tally : tallySet.entrySet()) {
            ByteArrayOutputStream in = new ByteArrayOutputStream();
            Json.write(tally.getValue(), in);
            byte[] signature = signer.sign(in.toByteArray());
            Files.write(outDir.resolve(Paths.get(tally.getKey() + TALLY_SUFFIX)), in.toByteArray());
            Files.write(outDir.resolve(Paths.get(tally.getKey() + SIGNATURE_SUFFIX)), signature);
        }
    }

    /**
     * Output proofs of correct decryptions.
     * 
     * @param outDir Output directory to store the proofs file.
     * @throws Exception When writing the proofs file fails.
     */
    public void outputProof(Path outDir) throws Exception {
        if (!withProof) {
            return;
        }
        Json.write(proof, outDir.resolve(PROOF_PATH));
    }

    /**
     * Output invalid votes.
     * 
     * @param outDir Output directory to store the invalid votes.
     * @throws Exception When writing the file fails.
     */
    public void outputInvalid(Path outDir) throws Exception {
        Json.write(invalid, outDir.resolve(INVALID_VOTE_PATH));
    }

    private class ResultWorker implements Callable<Void> {
        private final int voteCount;
        private I18nConsole console;
        private Reporter reporter;

        public ResultWorker(int voteCount, I18nConsole console, Reporter reporter) {
            this.voteCount = voteCount;
            this.console = console;
            this.reporter = reporter;
        }

        @Override
        public Void call() throws Exception {
            Progress progress = console.startProgress(voteCount);
            Object obj;
            while ((obj = votes.take()) != Util.EOT) {
                progress.increase(1);
                if (obj instanceof Vote) {
                    Vote vote = (Vote) obj;
                    String choice;
                    if (vote.getProof() != null) {
                        if (withProof) {
                            proof.addProof(vote.getProof());
                        }
                        choice = isValidChoice(vote) ? getCandidateNumber(vote)
                                : Tally.INVALID_VOTE_ID;
                    } else {
                        log.warn("Vote proof is missing - invalid vote");
                        choice = Tally.INVALID_VOTE_ID;
                    }
                    if (choice.equals(Tally.INVALID_VOTE_ID)) {
                        log.warn("Vote is invalid!");
                        invalid.getInvalid().add(vote);
                    }
                    addVoteToTally(vote, choice);
                } else {
                    throw new IllegalArgumentException(
                            "Unexpected decryption result type: " + obj.getClass());
                }
            }
            progress.finish();
            return null;
        }

        private String getCandidateNumber(Vote vote) {
            return vote.getProof().getDecrypted().getUTF8DecodedMessage()
                    .split(Util.UNIT_SEPARATOR)[0];
        }

        private boolean isValidChoice(Vote vote) {
            String voteStr = vote.getProof().decrypted.getUTF8DecodedMessage();
            String[] voteParts = voteStr.split(Util.UNIT_SEPARATOR, 3);
            if (voteParts.length != 3) {
                log.warn("isValidChoice() voteParts.length == {}, should be 3", voteParts.length);
                return false;
            }
            try {
                return candidates.getCandidates().get(vote.getDistrict()).get(voteParts[1])
                        .get(voteParts[0]).equals(voteParts[2]);
            } catch (NullPointerException ignored) {
                log.warn("isValidChoice() NullPointerException finding choice from candidate list");
                return false;
            }
        }

        private void addVoteToTally(Vote vote, String choice) {
            tallySet.computeIfAbsent(vote.getQuestion(),
                    q -> new Tally(electionName, candidates, districts)).getByStation()
                    .get(vote.getDistrict()).get(vote.getStation())
                    .compute(choice, (c, count) -> count + 1);
        }

    }
}
