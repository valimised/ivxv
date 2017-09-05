package ee.ivxv.common.util;

import ee.ivxv.common.service.i18n.Translatable;

/**
 * NameHolder creates a connection between a logical name and translatable name. This is useful e.g.
 * for command line arguments - they have static part (argument name) and translatable description.
 * It is possible that different tools of an application may want to use the same argument name, but
 * different or customized translation.
 * 
 * <p>
 * Note that <tt>NameHolder</tt> instances are <i>Translatable</i>, i.e they are automatically
 * translated when used as a message parameter.
 * </p>
 * 
 * <p>
 * Implementation hint: implement by Enum class, implement <tt>getName()</tt> with
 * <tt>extractName(name())</tt> to support different translations for the same argument name, e.g
 * constants <tt>tool1_arg</tt> and <tt>tool2_arg</tt> would have the same logical name "arg", but
 * have different translations.
 */
public interface NameHolder extends Translatable {

    /**
     * @return Returns the static short name.
     */
    String getShortName();

    /**
     * @return Returns the static name.
     */
    String getName();

    /**
     * @param name
     * @return Excludes everything before and including the first '_'.
     */
    default String extractName(String name) {
        return name.substring(name.indexOf('_') + 1);
    }
}
