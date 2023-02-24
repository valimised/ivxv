package ee.ivxv.common.cli;

import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.Set;

/**
 * CommandLine is the logical data model of command line arguments Command line consists of
 * <i>commands</i> and <i>options</i>. Commands are non-option arguments starting from the beginning
 * of the command line. Options are command line arguments that start with <tt>-</tt>. Options can
 * have a value, that is the list of non-option arguments that immediately follow the option. Both
 * commands and options are ordered.
 */
public class CommandLine {

    /** List of non-option arguments starting from the beginning of the command line arguments. */
    private final List<String> commands;
    /** List of options that have not bee matched with any argument so far. */
    private final List<Option> options;

    private CommandLine(List<String> commands, List<Option> options) {
        this.commands = Collections.unmodifiableList(commands);
        this.options = options;
    }

    /**
     * Parses an instance of <tt>CommandLine</tt> out of the provided command line arguments.
     *
     * @param args the command line arguments
     * @return Returns a <tt>CommandLine</tt> instance that corresponds to the provided arguments.
     */
    public static CommandLine parse(String[] args) {
        List<String> commands = new ArrayList<>();
        List<Option> options = new ArrayList<>();

        for (String arg : args) {
            if (Option.isOption(arg)) {
                break;
            }
            commands.add(arg);
        }

        for (int i = commands.size(), size = args.length; i < size;) {
            // The loop always starts with an option name
            String name = args[i];
            List<String> value = new ArrayList<>();

            // Leap over and process the option arguments
            for (i++; i < size && !Option.isOption(args[i]); i++) {
                value.add(args[i]);
            }

            Option o = new Option(name, value);

            options.add(o);
        }

        return new CommandLine(commands, options);
    }

    public void set(Args args) throws ParseException {
        Map<Arg<?>, Arg.Tree> branches = new HashMap<>();
        Set<Arg<?>> processedArgs = new HashSet<>();

        for (Iterator<Option> i = options.iterator(); i.hasNext();) {
            Option o = i.next();
            Arg<?> arg = find(o, args, branches);

            if (arg == null) {
                continue;
            }

            i.remove();

            if (!processedArgs.add(arg)) {
                throw new ParseException(Msg.e_multiple_assignments, o.getName());
            }

            if (arg instanceof Arg.Tree) {
                Arg.Tree tree = (Arg.Tree) arg;

                if (!tree.isExclusive()) {
                    throw new ParseException(Msg.e_value_not_allowed, o.getName(), o.getValue());
                }
                if (o.getValue().size() != 1) {
                    throw new ParseException(Msg.e_requires_single_value, o.getName(),
                            o.getValue());
                }

                String first = o.getValue().get(0);
                Arg<?> branch = tree.value.find(first);

                // If the found branch is not a tree, it does not make any sense
                if (branch == null || !(branch instanceof Arg.Tree)) {
                    throw new ParseException(Msg.e_branch_expected, o.getName(), first);
                }
                branches.put(tree, (Arg.Tree) branch);
            } else {
                arg.parse(o.getValue());
            }
        }
    }

    /**
     * @return Returns an unmodifiable list of commands.
     */
    public List<String> getCommands() {
        return commands;
    }

    /**
     * @return Returns the copy of the list of options that are unused so far.
     */
    public List<Option> getUnusedOptions() {
        return new ArrayList<>(options);
    }

    private Arg<?> find(Option o, Args args, Map<Arg<?>, Arg.Tree> branches) {
        for (Arg<?> arg : args.args) {
            if (o.matches(arg)) {
                if (!o.hasChild()) {
                    return arg;
                }
                if (arg instanceof Arg.Tree) {
                    return find(o.child(), branches.getOrDefault(arg, (Arg.Tree) arg).value,
                            branches);
                }
                // Referring to sub-argument of a non-tree argument
                return null;
            }
        }
        return null;
    }
}
