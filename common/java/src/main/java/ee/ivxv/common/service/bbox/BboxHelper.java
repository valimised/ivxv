package ee.ivxv.common.service.bbox;

import ee.ivxv.common.crypto.CryptoUtil.PublicKeyHolder;
import ee.ivxv.common.model.BallotBox;
import ee.ivxv.common.model.Voter;
import ee.ivxv.common.service.console.Progress;
import java.nio.file.Path;
import java.time.Instant;
import java.util.Map;
import java.util.Optional;
import java.util.function.BiConsumer;
import java.util.function.Consumer;

public interface BboxHelper {

    Loader<?> getLoader(Path path, Progress.Factory pf, int nThreads);

    interface Loader<U> {
        BboxLoader<U> getBboxLoader(Path path, Reporter<Ref.BbRef> r) throws InvalidBboxException;

        RegDataLoader<U> getRegDataLoader(Path path, Reporter<Ref.RegRef> r)
                throws InvalidBboxException;
    }

    byte[] getChecksum(Path path) throws Exception;

    boolean compareChecksum(byte[] sum1, byte[] sum2);

    @FunctionalInterface
    interface VoterProvider {
        Voter find(String voterId, String version);
    }

    @FunctionalInterface
    interface Reporter<T extends Ref> {
        void report(T ref, Result res, Object... args);
    }

    interface BboxLoader<U> extends InitialStage {
        /**
         * Checks the integrity of the ballot box, i.e. all required data for each record must be
         * present.
         * 
         * @return
         */
        IntegrityChecked<U> checkIntegrity();
    }

    /**
     * All records in ballot box are complete, i.e. all required data for each record is present.
     */
    interface IntegrityChecked<U> extends Stage {
        /**
         * Checks the consistency of the ballot box, e.g. the digital signature of all ballots.
         * 
         * @param vp Function to find active voters.
         * @param tsKey The key that was used to sign the timestamp requests.
         * @param elStart The election start time.
         * @return
         */
        BallotsChecked<U> checkBallots(VoterProvider vp, PublicKeyHolder tsKey, Instant elStart);

        /**
         * Exports the ballots of the current ballot box to the appointed destination.
         * 
         * @param voterId The voter ID to export, may be missing.
         * @param exporter An exporter callback.
         */
        void export(Optional<String> voterId, BiConsumer<Ref.BbRef, byte[]> exporter);

        /**
         * Lists the voters in the current ballot box. A voter is reported per ballot, so if they
         * have voted multiple times, then they will be listed the same number of times (but not
         * necessarily sequentially).
         * <p>
         * <tt>start</tt> and <tt>end</tt> can be used to limit the output to voters after a start
         * time and/or before an end time (both inclusive).
         * 
         * @param start The period start time, may be <tt>null</tt>.
         * @param end The period end time, may be <tt>null</tt>.
         * @param vp Function to find active voters.
         * @param consumer A callback that consumes the list.
         */
        void listVoters(Instant start, Instant end, VoterProvider vp, Consumer<Voter> consumer);
    }

    /**
     * All records in ballot box are consistent, e.g. digital signature is correct.
     */
    interface BallotsChecked<U> extends Stage {
        /**
         * Compares the ballot box to the registration data and retains only the intersection of
         * records that are consistent.
         * 
         * @param regData
         * @return
         */
        BboxLoaderResult checkRegData(RegDataLoaderResult<U> regData);
    }

    /**
     * Ballot box is fully loaded and checked. All records are correct and consistent.
     */
    interface BboxLoaderResult extends Stage {
        /**
         * Constructs ballot box instance with the loaded information.
         * 
         * @param electionId
         * @return
         */
        BallotBox getBallotBox(String electionId);
    }

    interface RegDataLoader<U> extends InitialStage {
        /**
         * Checks the integrity of the registration data container, i.e. all required data for each
         * record must be present.
         * 
         * @return
         */
        RegDataIntegrityChecked<U> checkIntegrity();
    }

    /**
     * All records in the registration data container are complete, i.e all required data for each
     * record is present.
     */
    interface RegDataIntegrityChecked<U> extends Stage {
        /**
         * Constructs and returns the registration data loader result object.
         * 
         * @return
         */
        RegDataLoaderResult<U> getRegData();
    }

    /**
     * Registration data is fully loaded and checked. All records are correct and consistent.
     */
    interface RegDataLoaderResult<U> extends Stage, Reporter<Ref.RegRef> {
        /**
         * @return Returns unmodifiable map with results. The map key is abstract correlation key
         *         between ballot box and registration data that depends on the {@code Loader}
         *         implementation.
         */
        Map<Object, RegDataRef<U>> getRegData();
    }

    class RegDataRef<U> {
        public final Ref.RegRef ref;
        public final U data;

        public RegDataRef(Ref.RegRef ref, U data) {
            this.ref = ref;
            this.data = data;
        }
    }

    interface InitialStage extends Stage {
        @Override
        default int getNumberOfInvalidBallots() {
            return 0;
        }
    }

    /**
     * Stage represents generic stage of the process of loading ballot box or registration data.
     */
    interface Stage {
        int getNumberOfValidBallots();

        /**
         * @return Returns the number of invalid ballots compared to the previous stage.
         */
        int getNumberOfInvalidBallots();
    }

}
