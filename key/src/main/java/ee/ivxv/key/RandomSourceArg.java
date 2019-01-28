package ee.ivxv.key;

import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Arg.TreeList;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.crypto.rnd.CombineRnd;
import ee.ivxv.common.crypto.rnd.DPRNG;
import ee.ivxv.common.crypto.rnd.FileRnd;
import ee.ivxv.common.crypto.rnd.NativeRnd;
import ee.ivxv.common.crypto.rnd.Rnd;
import ee.ivxv.common.crypto.rnd.UserRnd;
import ee.ivxv.common.service.i18n.MessageException;
import java.io.IOException;
import java.nio.file.Path;
import java.util.List;
import java.util.function.Function;

/**
 * Helper class for parsing random source arguments.
 */
public class RandomSourceArg extends Args {
    /**
     * Randomness source type.
     */
    static enum RndType {
        /**
         * Stream random source reads infinite random source.
         * <p>
         * Stream file source corresponds to {@link ee.ivxv.common.crypto.rnd.FileRnd}, initializing
         * it with finite argument true.
         * <p>
         * The command line takes a single argument denoting the location of the stream source.
         */
        stream(p -> newFileRnd(p, false)),
        /**
         * File random source reads finite random source.
         * <p>
         * Stream file source corresponds to {@link ee.ivxv.common.crypto.rnd.FileRnd}, initializing
         * it with finite argument false.
         * <p>
         * The command line takes a single argument denoting the location of the file.
         */
        file(p -> newFileRnd(p, true)),
        /**
         * DPRNG random source reads a seed from the file and initializes a deterministic pseudo
         * random number generator.
         * <p>
         * It uses {@link ee.ivxv.common.crypto.rnd.DPRNG}.
         * <p>
         * The command line takes a single argument denoting the location of the seed file.
         */
        DPRNG(RndType::newDPRNG),
        /**
         * System random source uses the system random source.
         * <p>
         * Internally, it uses {@link ee.ivxv.common.crypto.rnd.NativeRnd}.
         * <p>
         * Does not process given arguments.
         */
        system(p -> new NativeRnd()),
        /**
         * User random source uses external program to obtain randomness.
         * <p>
         * This method uses {@link ee.ivxv.common.crypto.rnd.UserRnd}.
         * <p>
         * Takes as a location the external program which is run for obtaining entropy.
         */
        user(p -> newUserRnd(p));

        private final Function<Path, Rnd> supplier;

        RndType(Function<Path, Rnd> supplier) {
            this.supplier = supplier;
        }

        static Rnd newFileRnd(Path path, boolean finite) {
            try {
                requirePath(path);
                return new FileRnd(path, finite);
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }

        static Rnd newDPRNG(Path path) {
            try {
                requirePath(path);
                return new DPRNG(path);
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }

        static Rnd newUserRnd(Path path) {
            try {
                requirePath(path);
                // port 22062 is chosen randomly from high range
                // the current implementation is continuous
                return new UserRnd(path, true, 22062);
            } catch (IOException e) {
                throw new RuntimeException(e);
            }
        }

        static void requirePath(Path path) {
            if (path == null) {
                throw new MessageException(ee.ivxv.common.cli.Msg.e_invalid_path_not_exists, path);
            }
        }

        public Rnd getRnd(Path path) {
            return supplier.apply(path);
        }
    }

    /**
     * Single random source argument.
     */
    public static class RndListEntry extends Args {
        Arg<RndType> type = Arg.aChoice(Msg.arg_random_source_type, RndType.values());
        // Must be optional, because of NATIVE random type that does not require path
        Arg<Path> path = Arg.aPath(Msg.arg_random_source_path, true, false).setOptional();

        RndListEntry() {
            args.add(type);
            args.add(path);
        }
    }

    /**
     * Get argument for setting multiple random source values.
     * 
     * @return
     */
    public static Arg<List<RndListEntry>> getArgument() {
        return new TreeList<>(Msg.arg_random_source, RndListEntry::new);
    }

    /**
     * Construct a random source combining given argument values.
     * <p>
     * If the argument is not set, then returns null. Otherwise, initialize all random sources and
     * combine it using {@link ee.ivxv.common.crypto.rnd.CombineRnd} and return the instance.
     * 
     * @param argument Argument values
     * @return Combined random source or null if arguments not given.
     * @throws IOException When exception occurs during initialization of single random source.
     */
    public static CombineRnd combineFromArgument(Arg<List<RndListEntry>> argument)
            throws IOException {
        if (!argument.isSet()) {
            return null;
        }
        CombineRnd cr = new CombineRnd();
        for (RndListEntry rle : argument.value()) {
            cr.addSource(rle.type.value().getRnd(rle.path.value()));
        }
        return cr;
    }
}
