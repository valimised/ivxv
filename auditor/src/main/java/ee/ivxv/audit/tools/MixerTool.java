package ee.ivxv.audit.tools;

import ee.ivxv.audit.AuditContext;
import ee.ivxv.audit.Msg;
import ee.ivxv.audit.shuffle.ShuffleException;
import ee.ivxv.audit.shuffle.ShuffleProof;
import ee.ivxv.audit.shuffle.Verifier;
import ee.ivxv.audit.tools.MixerTool.MixerArgs;
import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.util.I18nConsole;
import java.nio.file.Path;

public class MixerTool implements Tool.Runner<MixerArgs> {
    private final AuditContext ctx;
    private final I18nConsole console;

    public static class MixerArgs extends Args {
        Arg<Path> protPath = Arg.aPath(Msg.arg_protinfo, true, false);
        Arg<Path> proofPath = Arg.aPath(Msg.arg_proofdir, true, true);

        public MixerArgs() {
            super();
            args.add(protPath);
            args.add(proofPath);
        }
    }

    public MixerTool(AuditContext ctx) {
        this.ctx = ctx;
        this.console = new I18nConsole(ctx.i.console, ctx.i.i18n);
    }

    @Override
    public boolean run(MixerArgs args) throws Exception {
        console.println();
        console.println(Msg.m_shuffle_proof_loading, args.protPath.value(), args.proofPath.value());
        ShuffleProof proof = new ShuffleProof(args.protPath.value(), args.proofPath.value());
        Verifier ver = new Verifier(proof);
        boolean res = false;
        try {
            res = ver.verify_all();
        } catch (ShuffleException e) {
            console.println(Msg.m_shuffle_proof_failed_reason, e);
            res = false;
        }
        if (res) {
            console.println(Msg.m_shuffle_proof_succeeded);
        } else {
            console.println(Msg.m_shuffle_proof_failed);
        }
        return res;
    }
}
