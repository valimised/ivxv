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

public class RandomSourceArg extends Args {
    static enum RndType {
        stream(p -> newFileRnd(p, false)), file(p -> newFileRnd(p, true)), //
        DPRNG(RndType::newDPRNG), system(p -> new NativeRnd()), //
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

    public static class RndListEntry extends Args {
        Arg<RndType> type = Arg.aChoice(Msg.arg_random_source_type, RndType.values());
        // Must be optional, because of NATIVE random type that does not require path
        Arg<Path> path = Arg.aPath(Msg.arg_random_source_path, true, false).setOptional();

        RndListEntry() {
            args.add(type);
            args.add(path);
        }
    }

    public static Arg<List<RndListEntry>> getArgument() {
        return new TreeList<>(Msg.arg_random_source, RndListEntry::new);
    }

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
