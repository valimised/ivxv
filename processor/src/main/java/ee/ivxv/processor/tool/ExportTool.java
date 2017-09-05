package ee.ivxv.processor.tool;

import ee.ivxv.common.M;
import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.model.BallotBox;
import ee.ivxv.common.service.bbox.BboxHelper;
import ee.ivxv.common.service.bbox.InvalidBboxException;
import ee.ivxv.common.service.bbox.Ref;
import ee.ivxv.common.service.i18n.MessageException;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.ToolHelper;
import ee.ivxv.processor.Msg;
import ee.ivxv.processor.ProcessorContext;
import ee.ivxv.processor.tool.ExportTool.ExportArgs;
import ee.ivxv.processor.util.ReportHelper;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Optional;

public class ExportTool implements Tool.Runner<ExportArgs> {

    private static final String OUT_EXP = "export";

    private final ProcessorContext ctx;
    private final I18nConsole console;
    private final ReportHelper reporter;
    private final ToolHelper tool;

    public ExportTool(ProcessorContext ctx) {
        this.ctx = ctx;
        console = new I18nConsole(ctx.i.console, ctx.i.i18n);
        reporter = new ReportHelper(ctx, console);
        tool = new ToolHelper(console, ctx.container, ctx.bbox);
    }

    @Override
    public boolean run(ExportArgs args) throws Exception {
        export(args);
        reporter.writeBbErrors(args.out.value());

        return true;
    }

    private void export(ExportArgs args) throws Exception {
        try {
            tool.checkBbChecksum(args.bb.value(), args.bbChecksum.value());

            int tc = ctx.args.threads.value();
            BboxHelper.Loader<?> loader =
                    ctx.bbox.getLoader(args.bb.value(), console::startProgress, tc);

            export(args.bb.value(), args.voter.value(), loader, args.out.value().resolve(OUT_EXP));
        } catch (InvalidBboxException e) {
            throw new MessageException(e, Msg.e_bb_read_error, e.path, e);
        }
    }

    private <T> void export(Path path, String voterId, BboxHelper.Loader<T> l, Path out) {
        console.println();
        console.println(M.m_bb_loading, path);
        BboxHelper.BboxLoader<T> loader = l.getBboxLoader(path, reporter::reportBbError);
        console.println(M.m_bb_loaded);
        console.println(M.m_bb_checking_type);
        // If no error has occurred so far, the file structure must be correct and type UNORGANIZED
        console.println(M.m_bb_type, BallotBox.Type.UNORGANIZED);

        console.println(M.m_bb_checking_integrity);
        BboxHelper.IntegrityChecked<T> ic = loader.checkIntegrity();
        console.println(M.m_bb_data_is_integrous);
        console.println(M.m_bb_numof_ballots, ic.getNumberOfValidBallots());

        console.println();
        if (voterId != null) {
            console.println(M.m_bb_exporting_voter, voterId, out);
        } else {
            console.println(M.m_bb_exporting, out);
        }
        ic.export(Optional.ofNullable(voterId), (ref, bytes) -> writeContainer(ref, bytes, out));
        console.println(M.m_bb_exported);
    }

    private void writeContainer(Ref.BbRef ref, byte[] bytes, Path out) {
        try {
            String fileName = String.format("%s.%s", ref.ballot, ctx.container.getFileExtension());
            Path ballotPath = out.resolve(ref.voter).resolve(fileName);

            Files.createDirectories(ballotPath.getParent());
            Files.write(ballotPath, bytes);
        } catch (Exception e) {
            // Let the BboxLoader take care of error reporting
            throw new RuntimeException(e);
        }
    }

    public static class ExportArgs extends Args {

        Arg<Path> bb = Arg.aPath(Msg.arg_ballotbox, true, false);
        Arg<Path> bbChecksum = Arg.aPath(Msg.arg_ballotbox_checksum, true, false);
        Arg<String> voter = Arg.aString(Msg.arg_voter_id).setOptional();

        Arg<Path> out = Arg.aPath(Msg.arg_out, false, null);

        public ExportArgs() {
            args.add(bb);
            args.add(bbChecksum);
            args.add(voter);
            args.add(out);
        }

    }

}
