package ee.ivxv.common.service.i18n;

/**
 * Translatable is an interface the implementations of which are automatically translated if they
 * are present among the <tt>I18n</tt> arguments.
 */
public interface Translatable {

    /**
     * @return Returns the translation key.
     */
    Enum<?> getKey();

    /**
     * @return Returns the translation arguments, {@code null} by default.
     */
    default Object[] getArgs() {
        return null;
    }

}
