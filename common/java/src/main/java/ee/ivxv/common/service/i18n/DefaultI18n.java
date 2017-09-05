package ee.ivxv.common.service.i18n;

import ee.ivxv.common.M;
import java.time.Instant;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.util.Collection;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;
import java.util.stream.Stream;

/**
 * Default implementation of i18n service that does not depend on any actual translation provider.
 */
public abstract class DefaultI18n implements I18n {

    /** The parsed and re-used date-time formatters. */
    private final Map<String, DateTimeFormatter> formats = new ConcurrentHashMap<>();

    /**
     * Returns {@code getInternal} with the specified key and resolved arguments.
     */
    @Override
    public final String get(Enum<?> key, Object... args) {
        return getInternal(key, resolveArgs(args));
    }

    /**
     * The implementation of translation.
     * 
     * @param key
     * @param args Resolved arguments.
     * @return
     */
    protected abstract String getInternal(Enum<?> key, Object... args);

    @Override
    public String format(Instant i) {
        String p = get(M.m_datetime_pattern);
        DateTimeFormatter fmt = formats.computeIfAbsent(p, key -> DateTimeFormatter.ofPattern(p));
        return fmt.format(i.atZone(ZoneId.systemDefault()));
    }

    /**
     * Returns a new array of arguments applying {@code resolveArg(arg)} to each element.
     * 
     * @param args
     * @return
     */
    public Object[] resolveArgs(Object[] args) {
        if (args == null) {
            return new Object[0];
        }
        return Stream.of(args).map(a -> resolveArg(a)).toArray();
    }

    /**
     * Performs the following replacements for {@code arg}:
     * <ul>
     * <li>{@code Object[]} or {@code Collection<?>} {@literal -> } {@code resolveArg} applied to
     * elements, result is converted to space-separated string.
     * <li>{@code Translatable} {@literal -> } {@code get((Translatable) arg)},
     * <li>{@code Throwable} {@literal -> } {@code ((Throwable) arg).getMessage()},
     * <li>{@code Instant} {@literal -> } {@code format((Instant) arg)},
     * </ul>
     * 
     * @param arg
     * @return
     */
    public Object resolveArg(Object arg) {
        if (arg instanceof Object[]) {
            return Stream.of((Object[]) arg).map(a -> String.valueOf(resolveArg(a)))
                    .filter(s -> !s.isEmpty()).collect(Collectors.joining(" "));
        }
        if (arg instanceof Collection<?>) {
            return ((Collection<?>) arg).stream().map(a -> String.valueOf(resolveArg(a)))
                    .filter(s -> !s.isEmpty()).collect(Collectors.joining(" "));
        }
        if (arg instanceof Translatable) {
            return get((Translatable) arg);
        }
        if (arg instanceof Throwable) {
            String message = ((Throwable) arg).getMessage();
            return message != null ? message : arg;
        }
        if (arg instanceof Instant) {
            return format((Instant) arg);
        }
        return arg;
    }

}
