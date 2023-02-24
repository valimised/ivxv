package ee.ivxv.processor.tool;

import ee.ivxv.common.M;
import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.crypto.CorrectnessUtil;
import ee.ivxv.common.crypto.elgamal.ElGamalPublicKey;
import ee.ivxv.common.model.Ballot;
import ee.ivxv.common.model.BallotBox;
import ee.ivxv.common.model.DistrictList;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.service.report.Reporter;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.ToolHelper;
import ee.ivxv.common.util.Util;
import ee.ivxv.processor.Msg;
import ee.ivxv.processor.ProcessorContext;
import ee.ivxv.processor.tool.SquashTool.SquashArgs;
import ee.ivxv.processor.util.ReportHelper;

import java.io.Closeable;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

public class SquashTool implements Tool.Runner<SquashArgs> {

    private static final String OUT_BB_TMPL = "bb-2.json";
    private static final String OUT_IVLJSON_TMPL = "ivoterlist.json";
    private static final String OUT_IVLPDF_TMPL = "ivoterlist.pdf";
    private static final String OUT_RR_TMPL = "revocation-report.csv";

    private final ProcessorContext ctx;
    private final I18nConsole console;
    private final ReportHelper reporter;
    private final ToolHelper tool;

    private final List<Reporter.Record> revocationRecords = new ArrayList<>();
    private final List<Reporter.LogNRecord> log2Records = new ArrayList<>();

    public SquashTool(ProcessorContext ctx) {
        this.ctx = ctx;
        console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        reporter = new ReportHelper(ctx, console);
        tool = new ToolHelper(console, ctx.container, ctx.bbox);
    }

    @Override
    public boolean run(SquashArgs args) throws Exception {
        tool.checkBbChecksum(args.bb.value(), args.bbChecksum.value());

        BallotBox bb = tool.readJsonBb(args.bb.value(), BallotBox.Type.INTEGRITY_CONTROLLED);
        DistrictList dl = tool.readJsonDistricts(args.districts.value());
        Path out = args.out.value();
        ElGamalPublicKey pub = new ElGamalPublicKey(args.encKey.value());

        removeRecurrentVotes(bb);
        removeInvalidCiphertexts(bb, pub);

        Path OUT_IVLJSON = Util.prefixedPath(bb.getElection(), OUT_IVLJSON_TMPL);
        Path OUT_IVLPDF = Util.prefixedPath(bb.getElection(), OUT_IVLPDF_TMPL);
        Path OUT_RR = Util.prefixedPath(bb.getElection(), OUT_RR_TMPL);
        Path OUT_RR_ANONYMOUS = Util.prefixedPath(bb.getElection(), OUT_RR_TMPL + ".anonymous");
        Path OUT_BB = Util.prefixedPath(bb.getElection(), OUT_BB_TMPL);

        reporter.writeIVoterList(out.resolve(OUT_IVLJSON), out.resolve(OUT_IVLPDF), bb, dl);
        reporter.writeRevocationReport(out.resolve(OUT_RR), bb.getElection(), revocationRecords,
                Reporter.AnonymousFormatter.NOT_ANONYMOUS);
        reporter.writeRevocationReport(out.resolve(OUT_RR_ANONYMOUS), bb.getElection(), revocationRecords,
                Reporter.AnonymousFormatter.REVOCATION_REPORT_CSV);
        reporter.writeLog2(out, bb.getElection(), log2Records);

        tool.writeJsonBb(bb, out.resolve(OUT_BB));

        return true;
    }

    private void removeRecurrentVotes(BallotBox bb) {
        console.println();
        console.println(Msg.m_removing_recurrent_votes);

        bb.removeRecurrentVotes(this::collect);

        console.println();
        console.println(M.m_bb_type, bb.getType());
        console.println(M.m_bb_numof_ballots, bb.getNumberOfBallots());
    }

    private void removeInvalidCiphertexts(BallotBox bb, ElGamalPublicKey pk) {
        console.println();
        console.println(Msg.m_removing_invalid_ciphertexts);

        try (CiphertextFilter filter = new CiphertextFilter(pk, bb.getNumberOfBallots())) {
            bb.removeInvalidCiphertexts(filter::accept);
        }
        console.println();
        console.println(M.m_bb_type, bb.getType());
        console.println(M.m_bb_numof_ballots, bb.getNumberOfBallots());
    }

    private void collect(String vid, Ballot b) {
        revocationRecords.add(ctx.reporter.newRevocationRecordForRecurrentVote(vid, b));
        log2Records.add(ctx.reporter.newLog123Record(vid, b));
    }

    private void collectinvalid(String vid, Ballot b) {
        revocationRecords.add(ctx.reporter.newRevocationRecordForInvalidVote(vid, b));
        log2Records.add(ctx.reporter.newLog123Record(vid, b));
    }

    private class CiphertextFilter implements BallotBox.VoteFilter, Closeable {

        final ElGamalPublicKey pk;
        final Progress p;

        public CiphertextFilter(ElGamalPublicKey pk, int total) {
            this.pk = pk;
            this.p = console.startProgress(total);
        }

        @Override
        public boolean accept(String voterId, Ballot b, String qid, byte[] vote) {
            CorrectnessUtil.CiphertextCorrectness res = CorrectnessUtil.isValidCiphertext(pk, vote);
            boolean isValid = res == CorrectnessUtil.CiphertextCorrectness.VALID;

            p.increase(1);

            if (!isValid) {
                collectinvalid(voterId, b);
            }

            return isValid;
        }

        @Override
        public void close() {
            p.finish();
        }

    } // class CiphertextFilter

    public static class SquashArgs extends Args {

        Arg<Path> bb = Arg.aPath(Msg.arg_ballotbox, true, false);
        Arg<Path> bbChecksum = Arg.aPath(Msg.arg_ballotbox_checksum, true, false);
        Arg<Path> districts = Arg.aPath(Msg.arg_districts, true, false);
        Arg<Path> encKey = Arg.aPath(Msg.arg_enckey, true, false);

        Arg<Path> out = Arg.aPath(Msg.arg_out, false, null);

        public SquashArgs() {
            args.add(bb);
            args.add(bbChecksum);
            args.add(districts);
            args.add(encKey);
            args.add(out);
        }

    }

}
