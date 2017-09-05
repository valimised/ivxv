package ee.ivxv.common.cli;

import java.util.List;

/**
 * An instance of <code>Option</code> represents a single command line option name that should
 * correspond to a valid command line argument (instance of {@link Arg}). The class provides helper
 * methods to match the option with an argument.
 * 
 * Option may represent sub-arguments. In that case the option is either in the form
 * '--parent-child' (long argument names) or '-p-c' (short argument names).
 */
class Option {

    private static final String SEP = "-";
    private static final String SHORT_PREFIX = SEP;
    private static final String LONG_PREFIX = SEP + SEP;

    private final String fullName;
    private final boolean isShort;
    private final String[] names;
    private final int i;
    private final List<String> value;

    Option(String name, List<String> value) {
        this(name, isShortName(name), name.substring(name.lastIndexOf(SEP, 1) + 1).split(SEP), 0,
                value);
    }

    private Option(String fullName, boolean isShort, String[] names, int i, List<String> value) {
        this.fullName = fullName;
        this.isShort = isShort;
        this.names = names;
        this.i = i;
        this.value = value;
    }

    static boolean isShortName(String name) {
        return !name.startsWith(LONG_PREFIX);
    }

    static boolean isOption(String s) {
        return s != null && (s.startsWith(SHORT_PREFIX) || s.startsWith(LONG_PREFIX));
    }

    static String formatName(Arg<?> arg) {
        return arg.name.getName() != null ? LONG_PREFIX + arg.name.getName() : "";
    }

    static String formatShortName(Arg<?> arg) {
        return arg.name.getShortName() != null ? SHORT_PREFIX + arg.name.getShortName() : "";
    }

    public String getName() {
        return fullName;
    }

    String getLevelName() {
        return names[i];
    }

    List<String> getValue() {
        return value;
    }

    boolean hasChild() {
        return i < names.length - 1;
    }

    Option child() {
        return new Option(fullName, isShort, names, i + 1, value);
    }

    boolean matches(Arg<?> arg) {
        return getLevelName().equals(isShort ? arg.name.getShortName() : arg.name.getName());
    }

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder();

        sb.append(fullName);
        sb.append(" = ");
        sb.append(value);

        return sb.toString();
    }
}
