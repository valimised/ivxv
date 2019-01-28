package ee.ivxv.audit;

import ee.ivxv.audit.tools.ConvertTool;
import ee.ivxv.audit.tools.ConvertTool.ConvertArgs;
import ee.ivxv.audit.tools.DecryptTool;
import ee.ivxv.audit.tools.DecryptTool.DecryptArgs;
import ee.ivxv.audit.tools.MixerTool;
import ee.ivxv.audit.tools.MixerTool.MixerArgs;
import ee.ivxv.common.cli.App;
import ee.ivxv.common.cli.CommonArgs;
import ee.ivxv.common.cli.InitialContext;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.conf.Conf;
import java.util.Arrays;
import java.util.List;

class AuditApp extends App<AuditContext> {

    AuditApp() {
        super(Msg.app_audit, createTools());
    }

    private static List<Tool<AuditContext, ?>> createTools() {
        return Arrays.asList( //
                new Tool<>(Msg.tool_convert, ConvertArgs::new, ConvertTool::new),
                new Tool<>(Msg.tool_mixer, MixerArgs::new, MixerTool::new),
                new Tool<>(Msg.tool_decrypt, DecryptArgs::new, DecryptTool::new));
    }

    @Override
    public AuditContext createContext(InitialContext ctx, Conf conf, CommonArgs args) {
        return new AuditContext(ctx, conf, args);
    }
}
