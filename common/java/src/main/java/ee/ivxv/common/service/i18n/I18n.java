package ee.ivxv.common.service.i18n;

import java.time.Instant;

/**
 * I18n service provides translated messages for the locale the service was initialized with.
 */
public interface I18n {

    /**
     * Returns the translated message for the locale of the I18n service.
     * 
     * The following modifications are applied to the arguments before composing the translation:
     * <ul>
     * <li>{@code Translatable} {@literal -> } {@code get((Translatable) arg)},
     * <li>{@code Throwable} {@literal -> } {@code ((Throwable) arg).getMessage()},
     * <li>{@code Instant} {@literal -> } {@code format((Instant) arg)},
     * </ul>
     * 
     * @param key The message key.
     * @param args The message arguments.
     * @return The internationalized message.
     */
    String get(Enum<?> key, Object... args);

    /**
     * @param msg The message description.
     * @return The result of <code>get(msg.key, msg.args)</code> by default.
     * @throws NullPointerException If the argument is <code>null</code>.
     */
    default String get(Translatable msg) throws NullPointerException {
        return get(msg.getKey(), msg.getArgs());
    }

    /**
     * @param i An instant of time.
     * @return Returns formatted instant.
     */
    String format(Instant i);

}
