package ee.ivxv.key;

import ee.ivxv.common.cli.App;
import ee.ivxv.common.cli.CommonArgs;
import ee.ivxv.common.cli.InitialContext;
import ee.ivxv.common.cli.Tool;
import ee.ivxv.common.conf.Conf;
import ee.ivxv.key.tool.DecryptTool;
import ee.ivxv.key.tool.DecryptTool.DecryptArgs;
import ee.ivxv.key.tool.GroupGenTool;
import ee.ivxv.key.tool.GroupGenTool.GroupGenArgs;
import ee.ivxv.key.tool.InitTool;
import ee.ivxv.key.tool.InitTool.InitArgs;
import ee.ivxv.key.tool.UtilTool;
import ee.ivxv.key.tool.UtilTool.UtilArgs;
import java.util.Arrays;
import java.util.List;

public class KeyApp extends App<KeyContext> {

    KeyApp() {
        super(Msg.app_key, createTools());
    }

    private static List<Tool<KeyContext, ?>> createTools() {
        return Arrays.asList( //
                new Tool<>(Msg.tool_decrypt, DecryptArgs::new, DecryptTool::new),
                new Tool<>(Msg.tool_groupgen, GroupGenArgs::new, GroupGenTool::new),
                new Tool<>(Msg.tool_init, InitArgs::new, InitTool::new),
                new Tool<>(Msg.tool_util, UtilArgs::new, UtilTool::new));
    }

    @Override
    public KeyContext createContext(InitialContext ctx, Conf conf, CommonArgs args) {
        return new KeyContext(ctx, conf, args);
    }
}
