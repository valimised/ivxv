package ee.ivxv.common.service.bbox.impl;

import static java.util.stream.Collectors.groupingBy;
import static java.util.stream.Collectors.mapping;
import static java.util.stream.Collectors.toCollection;

import ee.ivxv.common.crypto.CryptoUtil.PublicKeyHolder;
import ee.ivxv.common.model.Ballot;
import ee.ivxv.common.model.BallotBox;
import ee.ivxv.common.model.VoterBallots;
import ee.ivxv.common.service.bbox.BboxHelper.BallotsChecked;
import ee.ivxv.common.service.bbox.BboxHelper.BboxLoader;
import ee.ivxv.common.service.bbox.BboxHelper.BboxLoaderResult;
import ee.ivxv.common.service.bbox.BboxHelper.IntegrityChecked;
import ee.ivxv.common.service.bbox.BboxHelper.RegDataLoaderResult;
import ee.ivxv.common.service.bbox.BboxHelper.RegDataRef;
import ee.ivxv.common.service.bbox.BboxHelper.Reporter;
import ee.ivxv.common.service.bbox.BboxHelper.VoterProvider;
import ee.ivxv.common.service.bbox.InvalidBboxException;
import ee.ivxv.common.service.bbox.Ref.BbRef;
import ee.ivxv.common.service.bbox.Ref.RegRef;
import ee.ivxv.common.service.bbox.Result;
import ee.ivxv.common.service.bbox.impl.FileName.RefProvider;
import ee.ivxv.common.service.bbox.impl.verify.TsVerifier;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.service.container.InvalidContainerException;
import eu.europa.esig.dss.DSSException;
import java.time.Instant;
import java.util.Collections;
import java.util.ConcurrentModificationException;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Optional;
import java.util.TreeSet;
import java.util.Vector;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeUnit;
import java.util.function.BiConsumer;
import java.util.function.Predicate;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

class IvxvBboxLoader<T extends Record<?>, U extends Record<?>, RT extends Keyable, RU extends Keyable>
        extends AbstractStage implements BboxLoader<RU> {

    static final Logger log = LoggerFactory.getLogger(IvxvBboxLoader.class);

    static final int MAX_NUMBER_OF_RETRIES = 5;

    final Profile<T, U, RT, RU> profile;
    final LoaderHelper<BbRef> helper;
    private final int nThreads;

    public IvxvBboxLoader(Profile<T, U, RT, RU> profile, FileSource source, Progress.Factory pf,
            Reporter<BbRef> reporter, int nThreads) throws InvalidBboxException {
        this(profile, new LoaderHelper<>(source, new BbRefProvider(), pf, reporter), nThreads);
    }

    IvxvBboxLoader(Profile<T, U, RT, RU> profile, LoaderHelper<BbRef> helper, int nThreads)
            throws InvalidBboxException {
        super(helper.getAllRefs().size());
        this.profile = profile;
        this.helper = helper;
        this.nThreads = nThreads;
        log.info("ZipBboxLoader instantiated with thread count {}", nThreads);
    }

    @Override
    public IntegrityChecked<RU> checkIntegrity() {
        int n = getNumberOfValidBallots();
        return new IntegrityCheckedImpl(helper.checkIntegrity(profile::createBbRecord, n), n);
    }

    ExecutorService createExecutorService() {
        if (nThreads <= 0) {
            return Executors.newCachedThreadPool();
        }
        return Executors.newFixedThreadPool(nThreads);
    }

    static class BbRefProvider implements RefProvider<BbRef> {

        private static final char DIR_SEP = '/';
        private static final String EXPECTED = "<voter-id>" + DIR_SEP + "<ballot-id>";

        @Override
        public BbRef get(String s) {
            int i = s.lastIndexOf(DIR_SEP);
            if (i < 0) {
                throw new FileName.InvalidNameException(s, EXPECTED);
            }
            int j = s.lastIndexOf(DIR_SEP, i - 1);
            String voter = s.substring(j + 1, i);
            String ballot = s.substring(i + 1);

            return new BbRef(voter, ballot);
        }
    }

    class IntegrityCheckedImpl extends AbstractStage implements IntegrityChecked<RU> {

        private final Map<BbRef, T> records;

        IntegrityCheckedImpl(Map<BbRef, T> records, int oldValid) {
            super(records.size(), oldValid);
            this.records = records;
        }

        @Override
        public BallotsChecked<RU> checkBallots(VoterProvider vp, PublicKeyHolder tsKey,
                Instant elStart) {
            Map<BbRef, BallotResponse> ballots = Collections.synchronizedMap(new LinkedHashMap<>());
            ExecutorService executor = createExecutorService();
            TsVerifier tsv = new TsVerifier(tsKey);
            Progress pb = helper.getProgress(getNumberOfValidBallots());
            Predicate<FileName<BbRef>> filter = name -> records.containsKey(name.ref);

            helper.processRecords(filter, profile::createBbRecord, (name, record) -> {
                // Ensure stable order of votes
                ballots.putIfAbsent(name.ref, null);

                executor.submit(() -> {
                    try {
                        BallotResponse br = createBallotResponse(name, record, vp, tsv, elStart);
                        if (br != null) {
                            ballots.put(name.ref, br);
                        }
                    } finally {
                        pb.increase(1);
                    }
                });
            });

            executor.shutdown();

            try {
                executor.awaitTermination(1, TimeUnit.DAYS);
            } catch (InterruptedException e) {
                throw new RuntimeException(e);
            }

            // Remove nulls added before
            ballots.values().removeIf(v -> v == null);

            removeRecurrentResponses(ballots);

            pb.finish();

            return new BallotsCheckedImpl(ballots, getNumberOfValidBallots());
        }

        @Override
        public void export(Optional<String> voterId, BiConsumer<BbRef, byte[]> exporter) {
            Optional<Progress> pb = voterId.isPresent() ? Optional.empty()
                    : Optional.of(helper.getProgress(getNumberOfValidBallots()));
            Predicate<FileName<BbRef>> filter = name -> records.containsKey(name.ref)
                    && voterId.map(vid -> name.ref.voter.equals(vid)).orElse(true);

            helper.processRecords(filter, profile::createBbRecord, (name, record) -> {
                try {
                    exporter.accept(name.ref, profile.combineBallotContainer(record));
                } catch (Exception e) {
                    helper.handleTechnicalError(name, e);
                } finally {
                    pb.ifPresent(p -> p.increase(1));
                }
            });

            pb.ifPresent(p -> p.finish());
        }

        private BallotResponse createBallotResponse(FileName<BbRef> name, T record,
                VoterProvider vp, TsVerifier tsv, Instant elStart) {
            int i = 0;
            retry: try {
                RT response = profile.getResponse(record);
                log.info("BALLOT-NONCE ballot: {}/{} nonce: {}", name.ref.voter, name.ref.ballot,
                        response.getKey());

                Ballot b = profile.createBallot(name, record, vp, tsv);
                if (b.getTime().isBefore(elStart)) {
                    throw new ResultException(Result.TIME_BEFORE_START, b.getTime().toString(),
                            elStart.toString());
                }

                return new BallotResponse(b, response);
            } catch (ResultException e) {
                log.error("ResultException was thrown processing file {}: ", name.path, e);
                helper.report(name.ref, e.result, e.args);
            } catch (InvalidContainerException e) {
                log.error("Invalid container '{}': {}", e.path, e.getMessage(), e);
                helper.report(name.ref, Result.INVALID_BALLOT_SIGNATURE, e);
            } catch (DSSException e) {
                /*
                 * In very rare cases a concurrent modification exception is re-thrown from a
                 * DSS-(?) library. Just try again for MAX times.
                 */
                log.error("DSSException occurred while processing {}", name.path, e);
                if (e.getCause() instanceof ConcurrentModificationException) {
                    log.error("!!! Caused by ConcurrentModificationException !!!");
                    if (++i < MAX_NUMBER_OF_RETRIES) {
                        break retry;
                    }
                }
                // Not concurrent modification exception - process as normally
                helper.handleTechnicalError(name, e);
            } catch (Exception e) {
                helper.handleTechnicalError(name, e);
            }
            return null;
        }

        /**
         * Ensure the uniqueness of all responses (response signatures/keys). In case of recurrence
         * only use the earliest one, remove others and report.
         * 
         * @param ballots
         */
        private void removeRecurrentResponses(Map<BbRef, BallotResponse> ballots) {
            ballots.entrySet().stream()
                    .collect(groupingBy(e -> e.getValue().response.getKey(), // Group by resp key
                            mapping(e -> e, toCollection(() -> new TreeSet<>(this::compare)))))
                    .values().stream().filter(e -> e.size() > 1) // Having only colliding entries
                    .flatMap(es -> es.stream().skip(1)) // Skip the first in sorted set as earliest
                    .forEach(e -> { // Process the others as the recurring ones
                        helper.report(e.getKey(), Result.REG_RESP_NOT_UNIQUE);
                        ballots.remove(e.getKey());
                    });
        }

        private int compare(Entry<BbRef, BallotResponse> e1, Entry<BbRef, BallotResponse> e2) {
            int cmp1 = e1.getValue().ballot.getTime().compareTo(e2.getValue().ballot.getTime());
            // Ensure equality is consistent with {@code equals} as required by {@code TreeSet}
            if (cmp1 != 0) {
                return cmp1;
            }
            return e1.getValue().ballot.getId().compareTo(e2.getValue().ballot.getId());
        }

    } // class IntegrityCheckedImpl

    class BallotResponse {
        final Ballot ballot;
        final RT response;

        public BallotResponse(Ballot ballot, RT response) {
            this.ballot = ballot;
            this.response = response;
        }
    }

    class BallotsCheckedImpl extends AbstractStage implements BallotsChecked<RU> {

        private final Map<BbRef, BallotResponse> ballots;

        BallotsCheckedImpl(Map<BbRef, BallotResponse> ballots, int oldValid) {
            super(ballots.size(), oldValid);
            this.ballots = ballots;
        }

        @Override
        public BboxLoaderResult checkRegData(RegDataLoaderResult<RU> regData) {
            Map<String, List<Ballot>> voters = Collections.synchronizedMap(new LinkedHashMap<>());
            Map<Object, RegRef> regFiles = new LinkedHashMap<>();
            ExecutorService executor = createExecutorService();
            Progress pb = helper.getProgress(getNumberOfValidBallots());

            regData.getRegData().forEach((key, rr) -> regFiles.put(key, rr.ref));

            ballots.forEach((ref, br) -> {
                RegDataRef<RU> rr = regData.getRegData().get(br.response.getKey());
                regFiles.remove(br.response.getKey());

                voters.computeIfAbsent(ref.voter, s -> new Vector<>());

                executor.submit(() -> {
                    try {
                        if (rr == null) {
                            // Request not found - report and continue
                            helper.report(ref, Result.BALLOT_WITHOUT_REG_REQ);
                        } else {
                            // Check registration request and response
                            Result regResult = profile.checkRegistration(br.response, rr.data);
                            if (regResult != Result.OK) {
                                // Report and break
                                helper.report(ref, regResult, rr.ref);
                                return;
                            }
                        }

                        voters.get(ref.voter).add(br.ballot);
                    } catch (Exception e) {
                        helper.handleTechnicalError(ref, e);
                    } finally {
                        pb.increase(1);
                    }
                });
            });

            executor.shutdown();

            try {
                executor.awaitTermination(1, TimeUnit.DAYS);
            } catch (InterruptedException e) {
                throw new RuntimeException(e);
            }

            // Report registration data without corresponding ballots using regData's reporting
            regFiles.forEach((key, ref) -> regData.report(ref, Result.REG_REQ_WITHOUT_BALLOT));

            // Remove entries for voters that never got a valid vote
            voters.values().removeIf(ballotList -> ballotList.isEmpty());

            pb.finish();

            return new BboxLoaderResultImpl(voters, getNumberOfValidBallots());
        }

    } // class BallotsCheckedImpl

    class BboxLoaderResultImpl extends AbstractStage implements BboxLoaderResult {

        private final Map<String, List<Ballot>> voters;

        BboxLoaderResultImpl(Map<String, List<Ballot>> voters, int oldValid) {
            super(voters.values().stream().mapToInt(m -> m.size()).sum(), oldValid);
            this.voters = voters;
        }

        @Override
        public BallotBox getBallotBox(String electionId) {
            Map<String, VoterBallots> ballots = new LinkedHashMap<>();

            voters.forEach((k, v) -> ballots.put(k, new VoterBallots(k, v)));

            // Detect ballots with the same timestamp as the latest ballot - report only
            ballots.values().forEach(vb -> Optional.ofNullable(vb.getLatest()).ifPresent(l -> {
                vb.getBallots().stream().filter(b -> b != l && b.getTime().equals(l.getTime()))
                        .forEach(b -> helper.report(new BbRef(vb.getVoterId(), b.getId()),
                                Result.SAME_TIME_AS_LATEST, l.getTime().toString(), l.getId()));
            }));

            return new BallotBox(electionId, ballots);
        }

    }

}
