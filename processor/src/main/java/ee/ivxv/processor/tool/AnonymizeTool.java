package ee.ivxv.processor.tool;

import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.CorrectnessUtil;
import ee.ivxv.common.crypto.CorrectnessUtil.CiphertextCorrectness;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.model.AnonymousBallotBox;
import ee.ivxv.common.model.Ballot;
import ee.ivxv.common.model.BallotBox;
import ee.ivxv.common.model.BallotBox.VoteFilter;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.ToolHelper;
import ee.ivxv.common.util.Util;
import ee.ivxv.processor.Msg;
import ee.ivxv.processor.ProcessorContext;
import ee.ivxv.processor.tool.AnonymizeTool.AnonymizeArgs;
import ee.ivxv.processor.util.ReportHelper;
import java.io.Closeable;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class AnonymizeTool implements Tool.Runner<AnonymizeArgs> {

    private static final String OUT_BB = "bb-4.json";

    private static final Map<String, Object> EMPTY = new HashMap<>();

    final I18nConsole console;
    final ReportHelper reporter;
    private final ToolHelper tool;

    final Map<String, Map<String, Object>> excluded = new ConcurrentHashMap<>();

    public AnonymizeTool(ProcessorContext ctx) {
        console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        reporter = new ReportHelper(ctx, console);
        tool = new ToolHelper(console, ctx.container, ctx.bbox);
    }

    @Override
    public boolean run(AnonymizeArgs args) throws Exception {
        tool.checkBbChecksum(args.bb.value(), args.bbChecksum.value());

        BallotBox bb = tool.readJsonBb(args.bb.value(), BallotBox.Type.DOUBLE_VOTERS_REMOVED);
        AnonymousBallotBox abb = anonymize(bb, loadElGamalPublicKey(args.encKey.value()));

        reporter.writeBbErrors(args.out.value());

        reporter.writeLog2(args.out.value(), bb,
                (voterId, qid) -> excluded.getOrDefault(voterId, EMPTY).containsKey(qid));
        reporter.writeLog3(args.out.value(), bb,
                (voterId, qid) -> !excluded.getOrDefault(voterId, EMPTY).containsKey(qid));

        tool.writeJsonBb(abb, args.out.value().resolve(OUT_BB));

        return true;
    }

    private ElGamalPublicKey loadElGamalPublicKey(Path path) throws Exception {
        String keyString = Util.toString(Files.readAllBytes(path));

        return new ElGamalPublicKey(Util.decodePublicKey(keyString));
    }

    private AnonymousBallotBox anonymize(BallotBox bb, ElGamalPublicKey pk) {
        console.println();
        console.println(Msg.m_anonymizing_ballot_box);

        try (CiphertextFilter filter = new CiphertextFilter(pk, bb.getNumberOfBallots())) {
            return bb.anonymize(filter::accept);
        }
    }

    private class CiphertextFilter implements VoteFilter, Closeable {

        final ElGamalPublicKey pk;
        final Progress p;

        public CiphertextFilter(ElGamalPublicKey pk, int total) {
            this.pk = pk;
            this.p = console.startProgress(total);
        }

        @Override
        public boolean accept(String voterId, Ballot b, String qid, byte[] vote) {
            CiphertextCorrectness res = CorrectnessUtil.isValidCiphertext(pk, vote);
            boolean isValid = res == CiphertextCorrectness.VALID;

            p.increase(1);

            if (!isValid) {
                excluded.computeIfAbsent(voterId, x -> new ConcurrentHashMap<>()).put(qid, true);

                reporter.reportAbbError(voterId, b, qid, res);
            }

            return isValid;
        }

        @Override
        public void close() {
            p.finish();
        }

    } // class CiphertextFilter

    public static class AnonymizeArgs extends Args {

        Arg<Path> bb = Arg.aPath(Msg.arg_ballotbox, true, false);
        Arg<Path> bbChecksum = Arg.aPath(Msg.arg_ballotbox_checksum, true, false);
        Arg<Path> encKey = Arg.aPath(Msg.arg_enckey, true, false);

        Arg<Path> out = Arg.aPath(Msg.arg_out, false, null);

        public AnonymizeArgs() {
            args.add(bb);
            args.add(bbChecksum);
            args.add(encKey);
            args.add(out);
        }

    }

}
