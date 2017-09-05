package ee.ivxv.processor;

import ee.ivxv.common.cli.App;
import ee.ivxv.common.cli.CommonArgs;
import ee.ivxv.common.cli.InitialContext;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.cli.VerifyTool;
import ee.ivxv.common.cli.VerifyTool.VerifyArgs;
import ee.ivxv.common.conf.Conf;
import ee.ivxv.processor.tool.AnonymizeTool;
import ee.ivxv.processor.tool.AnonymizeTool.AnonymizeArgs;
import ee.ivxv.processor.tool.CheckTool;
import ee.ivxv.processor.tool.CheckTool.CheckArgs;
import ee.ivxv.processor.tool.ExportTool;
import ee.ivxv.processor.tool.ExportTool.ExportArgs;
import ee.ivxv.processor.tool.RevokeTool;
import ee.ivxv.processor.tool.RevokeTool.RevokeArgs;
import ee.ivxv.processor.tool.SquashTool;
import ee.ivxv.processor.tool.SquashTool.SquashArgs;
import java.util.Arrays;
import java.util.List;

class ProcessorApp extends App<ProcessorContext> {

    ProcessorApp() {
        super(Msg.app_processor, createTools());
    }

    private static List<Tool<ProcessorContext, ?>> createTools() {
        return Arrays.asList( //
                new Tool<>(Msg.tool_check, CheckArgs::new, CheckTool::new),
                new Tool<>(Msg.tool_squash, SquashArgs::new, SquashTool::new),
                new Tool<>(Msg.tool_revoke, RevokeArgs::new, RevokeTool::new),
                new Tool<>(Msg.tool_anonymize, AnonymizeArgs::new, AnonymizeTool::new),
                new Tool<>(Msg.tool_export, ExportArgs::new, ExportTool::new),
                new Tool<>(ee.ivxv.common.cli.Msg.tool_verify, VerifyArgs::new, VerifyTool::new));
    }

    @Override
    public ProcessorContext createContext(InitialContext ctx, Conf conf, CommonArgs args) {
        return new ProcessorContext(ctx, conf, args);
    }
}
