package ee.ivxv.common.cli;

import ee.ivxv.common.conf.Conf;
import ee.ivxv.common.util.NameHolder;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Command line application. An application has one or more tools that define applications
 * functionality.
 */
public abstract class App<T extends AppContext<?>> {

    public final NameHolder name;
    public final Map<String, Tool<T, ?>> tools;

    public App(NameHolder name, List<Tool<T, ?>> tools) {
        this.name = name;
        this.tools = createToolsMap(tools);
    }

    public abstract T createContext(InitialContext ctx, Conf conf, CommonArgs args);

    private Map<String, Tool<T, ?>> createToolsMap(List<Tool<T, ?>> toolList) {
        Map<String, Tool<T, ?>> tmpTools = new LinkedHashMap<>();

        for (Tool<T, ?> tool : toolList) {
            tmpTools.put(tool.name.getName(), tool);
        }

        return Collections.unmodifiableMap(tmpTools);
    }

}
