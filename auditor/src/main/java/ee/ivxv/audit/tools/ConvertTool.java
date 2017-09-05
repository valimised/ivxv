package ee.ivxv.audit.tools;

import ee.ivxv.audit.AuditContext;
import ee.ivxv.audit.Msg;
import ee.ivxv.audit.tools.ConvertTool.ConvertArgs;
import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.util.I18nConsole;
import java.nio.file.Path;

public class ConvertTool implements Tool.Runner<ConvertArgs> {
    private final AuditContext ctx;
    private final I18nConsole console;

    public static class ConvertArgs extends Args {
        Arg<Path> inputBbox = Arg.aPath(Msg.arg_input_bb);
        Arg<Path> outputBbox = Arg.aPath(Msg.arg_output_bb);
        Arg<Path> pubPath = Arg.aPath(Msg.arg_pub);
        Arg<Path> protPath = Arg.aPath(Msg.arg_protinfo, true, false);
        Arg<Path> proofPath = Arg.aPath(Msg.arg_proofdir, true, true);

        public ConvertArgs() {
            super();
            args.add(inputBbox);
            args.add(outputBbox);
            args.add(pubPath);
            args.add(protPath);
            args.add(proofPath);
        }
    }

    public ConvertTool(AuditContext ctx) {
        this.ctx = ctx;
        this.console = new I18nConsole(ctx.i.console, ctx.i.i18n);
    }

    @Override
    public boolean run(ConvertArgs args) throws Exception {
        return false;
    }
}
