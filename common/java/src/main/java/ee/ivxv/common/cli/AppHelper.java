package ee.ivxv.common.cli;

import ee.ivxv.common.service.i18n.Message;
import ee.ivxv.common.util.I18nConsole;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.function.BiConsumer;
import java.util.stream.Collectors;
import java.util.stream.Stream;

class AppHelper {

    private static final String LIST_EL = "-";
    private static final String LIST_SHIFT = " ";

    private final App<?> app;
    private final CommonArgs cargs;
    private final I18nConsole console;

    AppHelper(App<?> app, CommonArgs cargs, I18nConsole console) {
        this.app = app;
        this.cargs = cargs;
        this.console = console;
    }

    static void walk(Args args, int level, BiConsumer<Arg<?>, Integer> consumer) {
        for (Arg<?> arg : args.args) {
            consumer.accept(arg, level);
            if (arg instanceof Arg.Tree) {
                walk(((Arg.Tree) arg).value, level + 1, consumer);
            }
            if (arg instanceof Arg.TreeList) {
                ((Arg.TreeList<?>) arg).value.forEach(a -> walk(a, level + 1, consumer));
            }
        }
    }

    void showHelp() {
        console.println(Msg.app, app.name.getName(), app.name);
        console.println();
        showUsage();
        console.println();
        showTools();
        console.println();
        showArgs(cargs);
    }

    void showToolHelp(Tool<?, ?> tool) {
        console.println(Msg.app, app.name.getName(), app.name);
        console.println(Msg.tool, tool.name.getName(), tool.name);
        console.println();
        showUsage();
        console.println();
        showArgs(cargs);
        console.println();
        showParams(tool.createArgs());
    }

    private void showUsage() {
        console.println(Msg.usage);
        Object[] args = cargs.args.stream().filter(a -> a != cargs.help).map(this::formatUsageArg)
                .toArray();
        console.println(Msg.row_indent,
                Arrays.asList(app.name.getName(), Msg.tool_placeholder, args));

        Object helpArg = Optional.ofNullable(Option.formatShortName(cargs.help)).map(
                sn -> (Object) new Message(Msg.usage_arg_names, sn, Option.formatName(cargs.help)))
                .orElse(Option.formatName(cargs.help));

        console.println(Msg.row_indent,
                Arrays.asList(app.name.getName(), Msg.tool_placeholder, helpArg));
        console.println(Msg.row_indent, Arrays.asList(app.name.getName(), helpArg));
    }

    private Object formatUsageArg(Arg<?> arg) {
        boolean isFlag = arg.value != null && (arg.value() instanceof Boolean);
        Object a = isFlag ? Option.formatName(arg)
                : new Message(Msg.usage_arg_w_value, Option.formatName(arg), arg.name.getName());
        return arg.isRequired() ? a : new Message(Msg.usage_arg_optional, a);
    }

    private void showTools() {
        console.println(Msg.tools);
        for (Tool<?, ?> tool : app.tools.values()) {
            console.println(Msg.row_indent,
                    new Message(Msg.tool_row, tool.name.getName(), tool.name));
        }
    }

    private void showArgs(Args args) {
        console.println(Msg.args);
        walk(args, 0, (arg, level) -> {
            String names = Stream.of(Option.formatShortName(arg), Option.formatName(arg))
                    .filter(s -> !s.isEmpty()).collect(Collectors.joining(" "));
            Object name = required(Msg.param_name_required, names, arg.isRequired());
            Message indented = indent(level, Msg.arg_row, name, arg.name);
            console.println(indented);
            if (arg instanceof Arg.TreeList) {
                // Add a single child node for which to show arguments
                ((Arg.TreeList<?>) arg).addNew();
            }
        });
    }

    private void showParams(Args args) {
        console.println(Msg.params);
        AtomicBoolean isFirstTreeElem = new AtomicBoolean();
        // In case of YAML list, the additional '- ' is added to indentation. Keep track of those.
        List<Integer> levelShift = new ArrayList<>();

        walk(args, 0, (arg, level) -> {
            // Discard the unused tail of 'levelShift' and inherit the value from parent level
            levelShift.subList(level, levelShift.size()).clear();
            levelShift.add(level > 0 ? levelShift.get(level - 1) : 0);
            // Consider that a " " is added between array elements by the i18n service.
            String shift = String.join(" ", Collections.nCopies(levelShift.get(level), LIST_SHIFT));

            Object name = required(Msg.param_name_required, arg.name.getName(), arg.isRequired());
            Message paramRow = new Message(Msg.param_row, name, arg.name);
            Message indented;
            if (isFirstTreeElem.getAndSet(false)) {
                indented = indent(level, new Object[] {shift, LIST_EL, paramRow});
                // Other elements of the same level are shifted by LIST_EL
                levelShift.set(level - 1, levelShift.get(level) + 1);
            } else {
                indented = indent(level, new Object[] {shift, paramRow});
            }

            console.println(indented);

            if (arg instanceof Arg.MultiValueArg<?>) {
                // Show a single element placeholder
                console.println(
                        indent(level + 1, new Object[] {shift, LIST_EL, Msg.value_placeholder}));
            }
            if (arg instanceof Arg.TreeList) {
                // Add a single child node for which to show parameters
                ((Arg.TreeList<?>) arg).addNew();
                isFirstTreeElem.set(true);
            }
        });
    }

    private Message indent(int level, Enum<?> key, Object... args) {
        return indent(level, new Message(key, args));
    }

    private Message indent(int level, Object arg) {
        Object indentedArg = arg;
        for (int i = 0; i++ < level; indentedArg = new Message(Msg.row_indent, indentedArg));
        return new Message(Msg.row_indent, indentedArg);
    }

    private Object required(Enum<?> key, String name, boolean isRequired) {
        return isRequired ? new Message(key, name) : name;
    }

}
