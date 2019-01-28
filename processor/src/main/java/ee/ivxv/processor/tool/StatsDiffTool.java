package ee.ivxv.processor.tool;

import ee.ivxv.common.cli.Arg;
import ee.ivxv.common.cli.Args;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.util.I18nConsole;
import ee.ivxv.common.util.Json;
import ee.ivxv.processor.Msg;
import ee.ivxv.processor.ProcessorContext;
import ee.ivxv.processor.tool.StatsDiffTool.StatsDiffArgs;
import ee.ivxv.processor.util.Statistics;
import java.nio.file.Path;

public class StatsDiffTool implements Tool.Runner<StatsDiffArgs> {

    private final I18nConsole console;

    public StatsDiffTool(ProcessorContext ctx) {
        console = new I18nConsole(ctx.i.console, ctx.i.i18n);
    }

    @Override
    public boolean run(StatsDiffArgs args) throws Exception {
        Statistics compare = Json.read(args.compare.value(), Statistics.class);
        Statistics to = Json.read(args.to.value(), Statistics.class);

        Statistics diff = new Statistics(compare, to);
        diff.writeJSON(args.diff.value());
        console.println(Msg.m_stats_diff_saved, args.diff.value());

        return true;
    }

    public static class StatsDiffArgs extends Args {

        Arg<Path> compare = Arg.aPath(Msg.arg_compare, true, false);
        Arg<Path> to = Arg.aPath(Msg.arg_to, true, false);
        Arg<Path> diff = Arg.aPath(Msg.arg_diff, false, null);

        public StatsDiffArgs() {
            args.add(compare);
            args.add(to);
            args.add(diff);
        }

    }

}
