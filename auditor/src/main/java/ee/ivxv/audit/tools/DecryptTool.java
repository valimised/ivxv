package ee.ivxv.audit.tools;

import ee.ivxv.audit.AuditContext;
import ee.ivxv.audit.Msg;
import ee.ivxv.audit.model.PlainBallotBox;
import ee.ivxv.audit.model.Tally;
import ee.ivxv.audit.tools.DecryptTool.DecryptArgs;
import ee.ivxv.audit.util.DiscardedBallot;
import ee.ivxv.audit.util.InvalidDecProofs;
import ee.ivxv.common.M;
import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.Plaintext;
import ee.ivxv.common.crypto.elgamal.ElGamalCiphertext;
import ee.ivxv.common.crypto.elgamal.ElGamalDecryptionProof;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.math.Group;
import ee.ivxv.common.math.GroupElement;
import ee.ivxv.common.math.MathException;
import ee.ivxv.common.model.AnonymousBallotBox;
import ee.ivxv.common.model.CandidateList;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.model.Proof;
import ee.ivxv.common.service.bbox.impl.BboxHelperImpl;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Json;
import ee.ivxv.common.util.ToolHelper;
import ee.ivxv.common.util.Util;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.apache.commons.collections4.Bag;
import org.apache.commons.collections4.bag.HashBag;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.*;
import java.util.concurrent.*;
import java.util.function.Consumer;
import java.util.stream.Collectors;

/**
 * Tool for verifying the correctness and consistency of ciphertext decryption.
 */
public class DecryptTool implements Tool.Runner<DecryptArgs> {
    private final Logger log = LoggerFactory.getLogger(DecryptTool.class);

    private final AuditContext ctx;
    private final I18nConsole console;
    private final ToolHelper tool;

    public DecryptTool(AuditContext ctx) {
        this.ctx = ctx;
        this.console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        tool = new ToolHelper(console, ctx.container, new BboxHelperImpl(ctx.conf, ctx.container));
    }

    @Override
    public boolean run(DecryptArgs args) throws Exception {
        // Verify whether all provided files are found on the filesystem.
        boolean allFilesExist = verifyInputFiles(args);
        if (!allFilesExist) {
            console.println();
            return false;
        }

        boolean abortEarly = args.abortEarly.value();
        boolean verifyInvalid = Objects.nonNull(args.invalidInputPath.value());
        boolean allChecksPassed = true;

        DistrictList districts = tool.readJsonDistricts(args.districts.value());
        CandidateList candidates = tool.readJsonCandidates(args.candidates.value(), districts);

        console.println();

        console.println(Msg.m_anon_loading, args.anonBoxPath.value());
        // The anonymous ballot box will subsequently be mutated!
        AnonymousBallotBox anonBb = Json.read(args.anonBoxPath.value(), AnonymousBallotBox.class);
        console.println(Msg.m_anon_loaded);

        console.println();

        console.println(Msg.m_pub_loading, args.pubPath.value());
        ElGamalPublicKey pub = new ElGamalPublicKey(args.pubPath.value());
        console.println(Msg.m_pub_loaded);

        // Use bags since it might be that there are multiple identical ciphertexts.
        // While these occurrences are suspicious and should be flagged, the flagging is done by
        // the integrity tool, not the decrypt tool.
        Bag<String> vProofCiphers = new HashBag<>(); // from proofs of ballot validity
        Bag<String> iProofCiphers = new HashBag<>(); // from proofs of ballot invalidity

        // For privacy reasons, these proofs might not exist: their generation is optional.
        if (verifyInvalid) {
            Proof proofs = tool.readJsonProofs(args.invalidInputPath.value());
            // Cryptographically verify proofs and get the base64 encoded invalid ballots.
            boolean allCorrect = processVerification(proofs, pub, ctx.args.threads.value(),
                    args.outputPath.value(), false, iProofCiphers);

            if (!allCorrect && abortEarly) {
                printFailure();
                return true;
            } else allChecksPassed = allCorrect;

            // Verify invalidity: plaintexts are indeed invalid.
            // Verify consistency: all ciphertexts in the proof file have a counterpart in the anonymous BB.
            // If this step becomes a performance bottleneck, it could be combined with the previous check.
            // Currently, the checks are separate for implementation clarity.
            // Note! Mutates the anonymous ballot box: removes ciphertexts of invalid votes from it.
            boolean areConsistent = verifyPlaintexts(pub, anonBb, proofs, candidates, false);
            if (!areConsistent && abortEarly) {
                printFailure();
                return true;
            } else allChecksPassed = allChecksPassed && areConsistent;
        }

        Proof proofs = tool.readJsonProofs(args.inputPath.value());
        // Cryptographically verify proofs and get the base64 encoded valid ballots.
        boolean allCorrect = processVerification(proofs, pub, ctx.args.threads.value(),
                args.outputPath.value(), true, vProofCiphers);

        if (!allCorrect && abortEarly) {
            printFailure();
            return true;
        } else allChecksPassed = allChecksPassed && allCorrect;

        // Verify validity: plaintexts are indeed valid.
        // Verify consistency: all ciphertexts in the proof file have a counterpart in the anonymous BB.
        // If this step becomes a performance bottleneck, it could be combined with the previous check.
        // Currently, the checks are separate for implementation clarity.
        // Note! Mutates the anonymous ballot box: removes ciphertexts of valid votes from it.
        boolean areConsistent = verifyPlaintexts(pub, anonBb, proofs, candidates, true);
        if (!areConsistent && abortEarly) {
            printFailure();
            return true;
        } else allChecksPassed = allChecksPassed && areConsistent;

        // Get the ciphertexts deemed invalid.
        // The file must always be provided, even if there are none: the file then contains none.
        Bag<String> discardedCiphers = getDiscardedCiphers(args.discardedInputPath.value());

        if (verifyInvalid) {
            // Verify whether the invalidity proofs are consistent with the ciphertexts declared invalid:
            // - all ciphertexts in the proofs file must have a counterpart in the invalids file
            // - all ciphertexts in the invalids file must have a counterpart in the proofs file
            boolean consistentInvalids = verifyInvalidConsistency(discardedCiphers, iProofCiphers);
            console.println(Msg.m_decrypt_consistent_invalids, getYesNoMessage(consistentInvalids));
            if (!consistentInvalids && abortEarly) {
                printFailure();
                return true;
            } else allChecksPassed = allChecksPassed && consistentInvalids;
        } else {
            // If the invalidity proofs are not provided, verify whether all ciphertexts declared invalid
            // have a counterpart in the anonymous BB. This is already done if the proofs are provided.
            // Note! Mutates the anonymous ballot box: removes ciphertexts of invalid votes from it.
            boolean invalidsInBb = verifyCiphersInBb(anonBb, discardedCiphers);
            console.println(Msg.m_decrypt_bb_has_invalids, getYesNoMessage(invalidsInBb));
            if (!invalidsInBb && abortEarly) {
                printFailure();
                return true;
            } else allChecksPassed = allChecksPassed && invalidsInBb;
        }

        // Verify that no ciphertext is declared both valid and invalid.
        console.println();
        boolean disjointOutputs = Collections.disjoint(vProofCiphers, discardedCiphers);
        console.println(Msg.m_decrypt_one_per_file, getYesNoMessage(disjointOutputs));
        if (!disjointOutputs && abortEarly) {
            printFailure();
            return true;
        } else allChecksPassed = allChecksPassed && disjointOutputs;

        // The anonymous BB should now be empty if it is consistent with the proof files:
        // all ciphertexts in the anonymous BB then have a counterpart in the proof files.
        // If only validity proofs were provided, the counterparts were checked through the invalid
        // ciphertexts file.
        // The converse consistency has already been verified, i.e. that all ciphertexts have a
        // counterpart in the anonymous BB.
        // In summary, if the following check passes, then: anon bb = valid ciphertexts + invalid ciphertexts
        boolean abbIsEmpty = isAnonBBEmpty(anonBb);
        console.println(Msg.m_decrypt_dec_consistent, getYesNoMessage(abbIsEmpty));
        if (!abbIsEmpty && abortEarly) {
            printFailure();
            return true;
        } else allChecksPassed = allChecksPassed && abbIsEmpty;

        // Get the plaintext ballot box.
        console.println();
        console.println(Msg.m_plain_loading, args.plainBoxPath.value());
        PlainBallotBox plainBb = Json.read(args.plainBoxPath.value(), PlainBallotBox.class);
        console.println(Msg.m_plain_loaded);

        // Verify whether the plaintext ballot box is consistent with the plaintexts in the proofs file:
        // - all plaintexts in the proofs file must have a counterpart in the plaintext ballot box
        // - all entries in the plaintext ballot box must have a counterpart in the proofs file
        // - the number of invalid votes should be consistent
        // If this step becomes a performance bottleneck, it should be somehow combined with the
        // validity proof verification stage.
        boolean consistentValids = verifyValidConsistency(pub, proofs, plainBb, discardedCiphers.size());
        console.println(Msg.m_decrypt_consistent_valids, getYesNoMessage(consistentValids));
        if (!consistentValids && abortEarly) {
            printFailure();
            return true;
        } else allChecksPassed = allChecksPassed && consistentValids;

        console.println();

        // Verify whether the plaintext ballot box matches with the computed tally.
        boolean tallyValid = verifyTally(plainBb, args.tallyPath.value());
        console.println(Msg.m_tally_match, getYesNoMessage(tallyValid));
        if (!tallyValid && abortEarly) {
            printFailure();
            return true;
        } else allChecksPassed = allChecksPassed && tallyValid;

        console.println();
        if (allChecksPassed) console.println(Msg.m_decrypt_success);
        else console.println(Msg.m_decrypt_failure);
        console.println();

        return true;
    }

    private Msg getYesNoMessage(boolean success) {
        if (success) return Msg.m_yes;
        return Msg.m_no;
    }

    private void printFailure() {
        console.println();
        console.println(Msg.m_decrypt_failure);
        console.println();
    }

    private boolean isAnonBBEmpty(AnonymousBallotBox abb) {
        boolean isEmpty = true;
        for (Map<String, Map<String, List<byte[]>>> smap : abb.getDistricts().values()) {
            for (Map<String, List<byte[]>> qmap : smap.values()) {
                for (List<byte[]> clist : qmap.values()) {
                    if (clist.isEmpty()) continue;
                    isEmpty = false;
                    for (byte[] c : clist) {
                        String b64ct = Base64.getEncoder().encodeToString(c);
                        log.warn("Output files missing ballot: {}", b64ct);
                    }
                }
            }
        }
        return isEmpty;
    }

    /**
     * Parses the file of discarded ciphertexts and extracts them.
     *
     * @param path the path of the file containing discarded ciphertexts
     * @return a bag of base64-encoded ciphertexts
     * @throws Exception If the file cannot be read or parsed
     */
    private Bag<String> getDiscardedCiphers(Path path) throws Exception {
        Bag<String> discardedCTs = new HashBag<>();

        console.println();
        console.println(Msg.m_discarded_loading, path);
        DiscardedBallot discarded = Json.read(path, DiscardedBallot.class);
        console.println(Msg.m_discarded_loaded);

        console.println(Msg.m_discarded_count, discarded.getCount());

        discarded.getDiscarded().forEach(vote ->
                discardedCTs.add(Base64.getEncoder().encodeToString(vote.getVote())));

        return discardedCTs;
    }

    /**
     * Verifies the proofs of correct decryption.
     * <p>
     * Additionally, verifies whether the decrypted results indeed are valid or invalid,
     * thus confirming whether proofs are valid proofs of validity or invalidity.
     * Collects also the base64-encoded ciphertexts for subsequent audit operations.
     *
     * @param proofs         the decryption proofs
     * @param pub            the election public key
     * @param threadCount    the number of threads to use for processing
     * @param outPath        the directory where to output verification failures
     * @param validityProofs whether the proofs prove validity or invalidity of ballots
     * @param b64CTs         the bag where to store base64-encoded ciphertexts
     * @return true if all proofs verify, false otherwise
     * @throws Exception If the verification process cannot complete
     */
    private boolean processVerification(Proof proofs, ElGamalPublicKey pub, int threadCount,
                                        Path outPath, boolean validityProofs,
                                        Bag<String> b64CTs) throws Exception {
        console.println();
        console.println(Msg.m_verify_start);
        InvalidDecProofs invalid = verifyDecryption(proofs, pub, threadCount, b64CTs);
        console.println(Msg.m_verify_finish);

        console.println(Msg.m_failurecount, invalid.getCount());
        if (invalid.getCount() == 0) return true;

        console.println(M.m_out_start, outPath);
        if (!Files.exists(outPath)) Files.createDirectory(outPath);

        if (validityProofs)
            invalid.outputFailedProofsOfValidity(outPath);
        else
            invalid.outputFailedProofsOfInvalidity(outPath);

        console.println(M.m_out_done, outPath);

        return false;
    }

    /**
     * Verifies whether plaintexts in a proof file really are valid/invalid.
     * <p>
     * Mutates the anonymous ballot box: removes ciphertexts for which it has
     * checked the plaintext. This speeds up subsequent verifications and keeps
     * track of the verification status.
     *
     * @param pub           the election public key
     * @param abb           the anonymous ballot box
     * @param proofs        the proof file
     * @param candidates    the list of accepted candidates
     * @param expectedValid whether the plaintexts should be valid or invalid
     * @return the result of the checks
     */
    private boolean verifyPlaintexts(ElGamalPublicKey pub, AnonymousBallotBox abb, Proof proofs, CandidateList candidates,
                                     boolean expectedValid) {
        // Using a set is fine here even if there are repetitions since we have already verified the proofs:
        // inconsistent repetitions should no longer be possible.
        Map<String, GroupElement> votes = new HashMap<>();
        // However, we still need to keep track of how many duplications there are when we compare with the
        // ballot box later.
        Bag<String> multiples = new HashBag<>();

        // Get all ciphertext-message pairs.
        for (Proof.ProofJson proof : proofs.getProofs()) {
            String b64 = Base64.getEncoder().encodeToString(proof.getCiphertext());
            GroupElement old = votes.put(b64, pub.getParameters().getGroup().getElement(proof.getMessage()));

            if (Objects.isNull(old)) continue; // not a duplicate

            if (!multiples.contains(b64)) multiples.add(b64, 2); // add two the first time
            else multiples.add(b64);
        }

        boolean allCorrect = true;

        for (Map.Entry<String, Map<String, Map<String, List<byte[]>>>> districtMap : abb.getDistricts().entrySet()) {
            for (Map<String, List<byte[]>> sMap : districtMap.getValue().values()) {
                for (List<byte[]> cList : sMap.values()) {
                    Iterator<byte[]> i = cList.iterator();
                    while (i.hasNext()) {
                        byte[] c = i.next();
                        String b64 = Base64.getEncoder().encodeToString(c);
                        GroupElement decrypted = votes.remove(b64);
                        multiples.remove(b64);
                        if (Objects.isNull(decrypted)) continue;
                        Group group = pub.getParameters().getGroup();
                        Plaintext decoded = group.decode(decrypted);
                        Plaintext unpadded = group.unpad(decoded);

                        boolean isValid = isValidChoice(unpadded, districtMap.getKey(), candidates);
                        if (isValid != expectedValid) {
                            log.warn("Plaintext is declared {} but is not: {}",
                                    expectedValid ? "valid" : "invalid", unpadded.getUTF8DecodedMessage());
                        }
                        allCorrect = allCorrect && (isValid == expectedValid);
                        i.remove();
                    }
                }
            }
        }

        // Verify whether all proof file entries were indeed in the anonymous ballot box.
        boolean allInBb = votes.isEmpty() && multiples.isEmpty();
        if (!allInBb) {
            log.warn("There were {} ballots missing from the input ballot box:", votes.size() + multiples.size());
            logMissingBallots(votes.keySet());
            logMissingBallots(multiples);
        }

        if (expectedValid) console.println(Msg.m_decrypt_verified_valid, getYesNoMessage(allCorrect));
        else console.println(Msg.m_decrypt_verified_invalid, getYesNoMessage(allCorrect));

        console.println(Msg.m_decrypt_bb_has_proofs, getYesNoMessage(allInBb));
        return allCorrect && allInBb;
    }

    private boolean verifyTally(PlainBallotBox pbb, Path tallyPath) throws Exception {
        console.println(Msg.m_tally_loading, tallyPath);
        Tally tally = Json.read(tallyPath, Tally.class);
        console.println(Msg.m_tally_loaded);

        boolean electionMatches = tally.getElection().equals(pbb.election());
        if (!electionMatches) {
            log.warn("The tally is for election '{}' while the plaintext ballot box is for election '{}'",
                    tally.getElection(), pbb.election());
            return false;
        }

        boolean parishesMatch = tally.getByParish().keySet().equals(pbb.byParish().keySet());
        if (!parishesMatch) {
            log.warn("Parishes do not match between the tally and the plaintext ballot box");
            return false;
        }

        // This iteration is comprehensive since the parish sets must now be equal.
        for (Map.Entry<String, Map<String, List<String>>> parishMap : pbb.byParish().entrySet()) {
            String parishId = parishMap.getKey();
            Map<String, List<String>> districtVotes = parishMap.getValue();

            Map<String, Map<String, Integer>> parishDistricts = tally.getByParish().get(parishId);
            boolean districtsMatch = parishDistricts.keySet().equals(districtVotes.keySet());
            if (!districtsMatch) {
                log.warn("Districts do not match between the tally and the plaintext ballot box");
                return false;
            }

            // This iteration is comprehensive since the district sets must now be equal.
            for (Map.Entry<String, List<String>> votesMap : districtVotes.entrySet()) {
                // The plaintext votes for that district.
                List<String> votes = votesMap.getValue();

                // Counter the (choice, count) pairs from the plaintext BB entry.
                Map<String, Integer> counts = votes.stream()
                        .collect(Collectors.groupingBy(c -> c, Collectors.summingInt(c -> 1)));
                // The tallied (choice, count) pairs. Create a copy to avoid mutation.
                Map<String, Integer> tallyCounts = new LinkedHashMap<>(parishDistricts.get(votesMap.getKey()));

                for (Map.Entry<String, Integer> count : counts.entrySet()) {
                    // The tally counts must match the count obtained from the plaintext BB.
                    if (!tallyCounts.remove(count.getKey()).equals(count.getValue())) {
                        log.warn("Plaintext ballot box and tally mismatch for the choice '{}'", count.getKey());
                        return false;
                    }
                }

                // Whatever choice is present in the tally map should now have 0 votes.
                // The plaintext BB choices have already been checked for validity.
                for (Integer val : tallyCounts.values()) {
                    if (val != 0) {
                        log.warn("Vote count > 0 for a choice not present in the plaintext ballot box");
                        return false;
                    }
                }
            }
        }

        // We have checked the byParish tally. The byDistrict tally must also be checked.
        // Since byParish is valid, simply check whether the byDistrict tally obtained
        // from the byParish tally matches with the byDistrict tally read form the tally.
        return tally.getByDistrict().equals(tally.computeByDistrict());
    }

    private boolean isValidChoice(Plaintext pt, String district, CandidateList candidates) {
        String candidateCode;
        try {
            candidateCode = pt.getUTF8DecodedMessage();
        } catch (IllegalArgumentException ignored) {
            return false;
        }

        // Get choices list
        Map<String, Map<String, Map<String, String>>> ds = candidates.getCandidates();
        // Ensure that vote district exists in a choices list
        if (!ds.containsKey(district)) {
            return false;
        }

        // Get choices for a specific district (district that vote belongs to)
        Map<String, Map<String, String>> ps = ds.get(district);
        // Ensure that there are choices per that district
        if (ps.isEmpty()) {
            return false;
        }

        // Loop over all choices per that district
        for (Map<String, String> parties : ps.values()) {
            // Is there a candidate code that matches candidateCode
            if (parties.containsKey(candidateCode)) {
                return true;
            }
        }

        // Choice is not valid
        return false;
    }

    private InvalidDecProofs verifyDecryption(Proof input, ElGamalPublicKey pub,
                                              int threadCount, Bag<String> b64CTs) throws Exception {
        InvalidDecProofs idp = new InvalidDecProofs(input.getElection());
        ExecutorService ioExecutor = Executors.newFixedThreadPool(2);
        CompletionService<Void> CompService = new ExecutorCompletionService<>(ioExecutor);

        ExecutorService verifyExecutor;
        threadCount = threadCount > 0 ? threadCount : 1;
        verifyExecutor = new ThreadPoolExecutor(threadCount, threadCount, 0L, TimeUnit.MILLISECONDS,
                new ArrayBlockingQueue<>(threadCount * 2));

        WorkManager manager =
                new WorkManager(input, getVerifyConsumer(pub, idp, b64CTs), verifyExecutor, idp);
        CompService.submit(manager);
        CompService.submit(idp.getResultWorker());

        try {
            for (int done = 0; done < 2; done++) {
                CompService.take().get();
            }
        } finally {
            ioExecutor.shutdown();
            verifyExecutor.shutdown();
        }
        return idp;
    }

    private Consumer<Proof.ProofJson> getVerifyConsumer(ElGamalPublicKey pub,
                                                        InvalidDecProofs out,
                                                        Bag<String> b64CTs) {
        return (proofJson) -> {
            GroupElement decrypted = pub.getParameters().getGroup().getElement(proofJson.getMessage());
            ElGamalCiphertext ct =
                    new ElGamalCiphertext(pub.getParameters(), proofJson.getCiphertext());
            ElGamalDecryptionProof proof =
                    new ElGamalDecryptionProof(ct, decrypted, pub, proofJson.getProof());

            // Get the base64-encoded ballot for later comparison with the ballot box.
            b64CTs.add(Base64.getEncoder().encodeToString(proofJson.getCiphertext()));

            try {
                boolean res = proof.verifyProof();
                if (!res) {
                    log.warn("Proof verification failed: {}", proof);
                    out.addInvalidProof(proof);
                }
            } catch (MathException e) {
                log.warn("Proof verification exception: {}, {}", proof, e);
                out.addInvalidProof(proof);
            }
        };
    }

    /**
     * Verifies whether ciphertexts are part of the anonymous ballot box.
     *
     * @param abb    the anonymous ballot box
     * @param b64CTs the base64-encoded ciphertexts to check inclusion for
     * @return true if all the ciphertexts are included, false otherwise
     */
    private boolean verifyCiphersInBb(AnonymousBallotBox abb, Bag<String> b64CTs) {
        // Do not mutate the ciphers.
        Bag<String> b64copy = new HashBag<>(b64CTs);

        for (Map<String, Map<String, List<byte[]>>> sMap : abb.getDistricts().values()) {
            for (Map<String, List<byte[]>> qMap : sMap.values()) {
                for (List<byte[]> cList : qMap.values()) {
                    Iterator<byte[]> i = cList.iterator();
                    while (i.hasNext()) {
                        byte[] c = i.next();
                        String b64ct = Base64.getEncoder().encodeToString(c);
                        if (!b64copy.remove(b64ct, 1)) continue;
                        i.remove();
                    }
                }
            }
        }

        // Verify whether all entries were indeed in the anonymous ballot box.
        boolean allInBb = b64copy.isEmpty();
        if (!allInBb) {
            log.warn("There were {} ballots missing from the input ballot box:", b64copy.size());
            logMissingBallots(b64copy);
        }
        return allInBb;
    }

    private boolean verifyInvalidConsistency(Bag<String> discardedCiphers, Bag<String> iProofCiphers) {
        if (iProofCiphers.equals(discardedCiphers)) return true;

        // Create copies to avoid mutating inputs.
        Bag<String> iProofCopy = new HashBag<>(iProofCiphers);
        Bag<String> discardedCopy = new HashBag<>(discardedCiphers);

        // Obtain the differences and log them.
        iProofCopy.removeAll(discardedCiphers);
        discardedCopy.removeAll(iProofCiphers);

        log.warn("There were {} ballots missing from the discarded votes file:", iProofCopy.size());
        logMissingBallots(iProofCopy);

        log.warn("There were {} ballots missing from the invalid proofs:", discardedCopy.size());
        logMissingBallots(discardedCopy);

        return false;
    }

    private void logMissingBallots(Collection<String> b64CTs) {
        for (String b64ct : b64CTs) log.warn("Missing ballot: {}", b64ct);
    }

    /**
     * Verifies plaintext ballot box consistency with the validity proofs.
     * <p>
     * Carries out the verification solely based on the choice identifier, and not
     * on the complete plaintext vote. The full validation of plaintexts in the proofs
     * file must be carried out beforehand for stronger consistency guarantees.
     *
     * @param pub          the election public key
     * @param proofs       the validity proofs
     * @param pbb          the plaintext ballot box
     * @param invalidCount the number of expected invalid votes
     * @return true if the verification succeeds, false otherwise
     */
    private boolean verifyValidConsistency(ElGamalPublicKey pub, Proof proofs, PlainBallotBox pbb, int invalidCount) {
        Bag<String> plainVotes = new HashBag<>(pbb.byDistrict().values().stream()
                .flatMap(Collection::stream).toList());
        console.println(Msg.m_plain_count, plainVotes.size());

        console.println();
        console.println(Msg.m_decrypt_consistent_valids_begin);

        boolean noExtra = true;
        try {
            for (Proof.ProofJson proof : proofs.getProofs()) {
                GroupElement decrypted = pub.getParameters().getGroup().getElement(proof.getMessage());
                Group group = decrypted.getGroup();
                Plaintext decoded = group.decode(decrypted);
                // Note! Here we do not validate the full vote, meaning that the validation
                // should be done prior to calling this function.
                String voteCode = group.unpad(decoded).getUTF8DecodedMessage();
                if (plainVotes.remove(voteCode, 1)) continue;
                noExtra = false;
                log.warn("A vote for {} is missing from the plaintext ballot box", voteCode);
            }
        } catch (Exception e) {
            // Barring software errors, this can happen only if
            // the proofs file is manipulated/corrupted.
            log.error(e.toString());
            return false;
        }

        // There were proofs for votes not present in the plaintext BB.
        if (!noExtra) {
            log.warn("There were proofs for votes not present in the plaintext ballot box");
            return false;
        }

        // We have checked the byDistrict votes. The byParish votes must also be checked.
        // Since byDistrict is valid, simply check whether the byDistrict votes obtained
        // from the byParish votes matches with the byDistrict votes read form the plaintext BB.
        boolean tallyConsistent = pbb.byDistrict().equals(pbb.computeByDistrict());
        if (!tallyConsistent) {
            log.warn("byparish inconsistent with bydistrict in the plaintext ballot box");
            return false;
        }

        return plainVotes.size() == invalidCount;
    }

    private boolean verifyInputFiles(DecryptArgs args) {
        List<Path> paths = new ArrayList<>(List.of(
                args.inputPath.value(),
                args.pubPath.value(),
                args.discardedInputPath.value(),
                args.anonBoxPath.value(),
                args.plainBoxPath.value(),
                args.tallyPath.value(),
                args.districts.value(),
                args.candidates.value()));

        if (Objects.nonNull(args.invalidInputPath.value()))
            paths.add(args.invalidInputPath.value());

        for (Path path : paths) {
            if (Files.notExists(path)) {
                console.println();
                console.println(Msg.e_file_missing, path);
                return false;
            }
        }

        return true;
    }

    public static class DecryptArgs extends Args {
        Arg<Path> inputPath = Arg.aPath(Msg.arg_proofs);
        Arg<Path> invalidInputPath = Arg.aPath(Msg.arg_invalidity_proofs).setOptional();
        Arg<Path> pubPath = Arg.aPath(Msg.arg_pub);
        Arg<Path> anonBoxPath = Arg.aPath(Msg.arg_anon_bb);
        Arg<Path> plainBoxPath = Arg.aPath(Msg.arg_plain_bb);
        Arg<Path> tallyPath = Arg.aPath(Msg.arg_tally);
        Arg<Path> discardedInputPath = Arg.aPath(Msg.arg_discarded);
        Arg<Path> districts = Arg.aPath(Msg.arg_districts);
        Arg<Path> candidates = Arg.aPath(Msg.arg_candidates);
        Arg<Path> outputPath = Arg.aPath(Msg.arg_out, false, true);
        Arg<Boolean> abortEarly = Arg.aFlag(Msg.arg_abort_early).setDefault(true);

        public DecryptArgs() {
            super();
            args.add(inputPath);
            args.add(pubPath);
            args.add(discardedInputPath);
            args.add(anonBoxPath);
            args.add(plainBoxPath);
            args.add(tallyPath);
            args.add(districts);
            args.add(candidates);
            args.add(outputPath);
            args.add(invalidInputPath);
            args.add(abortEarly);
        }
    }

    private class WorkManager implements Callable<Void> {
        private final Proof in;
        private final Consumer<Proof.ProofJson> consumer;
        private final ExecutorService verifyExecutor;
        private final InvalidDecProofs idp;

        WorkManager(Proof in, Consumer<Proof.ProofJson> consumer, ExecutorService verifyExecutor,
                    InvalidDecProofs idp) {
            this.in = in;
            this.consumer = consumer;
            this.verifyExecutor = verifyExecutor;
            this.idp = idp;
        }

        @Override
        public Void call() throws Exception {
            Progress progress = console.startProgress(in.getCount());
            in.getProofs().forEach(proof -> {
                boolean taskAdded = false;
                do {
                    try {
                        verifyExecutor.execute(() -> consumer.accept(proof));
                        taskAdded = true;
                    } catch (RejectedExecutionException e) {
                        try {
                            Thread.sleep(20);
                        } catch (InterruptedException e1) {
                            log.warn("Unexpected interruption", e1);
                        }
                    }
                } while (!taskAdded);
                progress.increase(1);
            });
            verifyExecutor.shutdown();
            verifyExecutor.awaitTermination(1, TimeUnit.DAYS);
            idp.setEot();
            progress.finish();
            return null;
        }
    }
}
