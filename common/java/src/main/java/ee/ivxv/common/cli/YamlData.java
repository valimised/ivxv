package ee.ivxv.common.cli;

import ee.ivxv.common.cli.Arg.Resolver;
import java.io.InputStream;
import java.nio.file.Path;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.Date;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.yaml.snakeyaml.Yaml;

public class YamlData {

    static final Logger log = LoggerFactory.getLogger(YamlData.class);

    private final Resolver resolver;
    private final Map<?, ?> map;

    private YamlData(Path path, Map<?, ?> map) throws ParseException {
        this.map = map;
        resolver = new Resolver(path::resolveSibling);
    }

    public static YamlData parse(Path path, InputStream in) {
        try {
            Yaml yaml = new Yaml();
            Object o = yaml.load(in);

            return new YamlData(path, expectMap(o, ""));
        } catch (Exception e) {
            throw new ParseException(e, Msg.e_yaml_invalid_file, e.getMessage());
        }
    }

    private static Map<?, ?> expectMap(Object o, String path) {
        if (o == null) {
            return new HashMap<>();
        }
        if (!(o instanceof Map<?, ?>)) {
            throw new ParseException(Msg.e_yaml_map_expected, path, o);
        }

        return (Map<?, ?>) o;
    }

    public void set(String toolName, Args args) throws ParseException {
        Map<?, ?> toolMap = expectMap(map.get(toolName), toolName);
        set(args, toolMap, toolName, ""); // Default for optionalValuePrefix is ""
    }

    private void set(Args args, Map<?, ?> m, String argPath, String optionalValuePrefix) {
        for (Map.Entry<?, ?> entry : m.entrySet()) {
            String name = entry.getKey().toString();
            String subPath = getPath(argPath, name);
            Object value = entry.getValue();

            if (YamlDataExtension.isVoterListsDirTag(name)) {
                YamlDataExtension.setVoterListsDirName(value);
                continue;
            }

            if (!optionalValuePrefix.equals("")) {
                value = optionalValuePrefix + (String) value;
            }

            log.debug("Path: {}, name: {}, value: {} ({})", argPath, name, value,
                    value != null ? value.getClass() : null);

            Arg<?> arg = args.find(name);

            if (arg == null) {
                throw new ParseException(Msg.e_yaml_invalid_key, argPath, name);
            }

            if (arg instanceof Arg.TreeList) {
                if (value == null) {
                    continue;
                }

                if (((List<?>) value).isEmpty()) {
                    throw new ParseException(Msg.e_yaml_list_expected, subPath, value);
                }
                if (!(value instanceof List)) {
                    throw new ParseException(Msg.e_yaml_list_expected, subPath, value);
                }

                Arg.TreeList<?> tl = (Arg.TreeList<?>) arg;

                ((List<?>) value).forEach(o -> set(tl.addNew(), expectMap(o, argPath), argPath, YamlDataExtension.getVoterlistsDirName()));

                continue;
            }

            if (arg instanceof Arg.Tree) {
                Arg.Tree tree = (Arg.Tree) arg;
                // Default for optionalValuePrefix is ""
                set(tree.value(), expectMap(value, subPath), subPath, "");
                continue;
            } else if (value instanceof Map) {
                throw new ParseException(Msg.e_yaml_scalar_expected, subPath, value);
            }

            List<String> values = new ArrayList<>();

            if (value instanceof List) {
                ((List<?>) value).forEach(o -> values.add(format(o)));
            } else if (value != null) {
                values.add(format(value));
            }

            arg.parse(values, resolver);
        }
    }

    private String format(Object o) {
        if (o instanceof Date) {
            return DateTimeFormatter.ISO_INSTANT.format(((Date) o).toInstant());
        }
        return o.toString();
    }

    private String getPath(String parent, String child) {
        return String.format("%s.%s", parent, child);
    }

}
