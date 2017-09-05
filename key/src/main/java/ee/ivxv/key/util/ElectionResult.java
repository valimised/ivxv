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
import javax.xml.bind.DatatypeConverter;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

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

    public final Map<String, List<Reporter.Record>> log4;
    public final Map<String, List<Reporter.Record>> log5;

    public ElectionResult(String electionName, CandidateList candidates, DistrictList districts,
            boolean withProof) {
        this.electionName = electionName;
        this.candidates = candidates;
        this.districts = districts;
        this.withProof = withProof;
        this.tallySet = new HashMap<>();
        this.proof = withProof ? new Proof(electionName) : null;
        this.invalid = new Invalid(electionName);
        this.log4 = new HashMap<>();
        this.log5 = new HashMap<>();
    }

    public ResultWorker getResultWorker(int voteCount, I18nConsole console, Reporter reporter) {
        return new ResultWorker(voteCount, console, reporter);
    }

    public void setEot() {
        votes.add(Util.EOT);
    }

    public void addVote(Vote vote) {
        votes.add(vote);
    }

    public void outputTally(Path outDir, SigningProtocol signer) throws Exception {
        for (Map.Entry<String, Tally> tally : tallySet.entrySet()) {
            ByteArrayOutputStream in = new ByteArrayOutputStream();
            Json.write(tally.getValue(), in);
            byte[] signature = signer.sign(in.toByteArray());

            Files.write(outDir.resolve(Paths.get(tally.getKey() + TALLY_SUFFIX)), in.toByteArray());

            Files.write(outDir.resolve(Paths.get(tally.getKey() + SIGNATURE_SUFFIX)),
                    DatatypeConverter.printHexBinary(signature).getBytes());
        }
    }

    public void outputProof(Path outDir) throws Exception {
        if (!withProof) {
            return;
        }
        Json.write(proof, outDir.resolve(PROOF_PATH));
    }

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
                        addVoteToLog4(vote);
                    } else {
                        addVoteToLog5(vote);
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
            String[] voteParts = voteStr.split(Util.UNIT_SEPARATOR);
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

        private void addVoteToLog4(Vote vote) {
            addVoteToLog(vote, log4);
        }

        private void addVoteToLog5(Vote vote) {
            addVoteToLog(vote, log5);
        }

        private void addVoteToLog(Vote vote, Map<String, List<Reporter.Record>> log) {
            log.computeIfAbsent(vote.getQuestion(), q -> new ArrayList<>())
                    .add(reporter.newLog45Record(vote.getDistrict(), vote.getVote()));
        }
    }
}
