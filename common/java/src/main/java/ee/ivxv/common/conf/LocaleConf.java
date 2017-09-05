package ee.ivxv.common.conf;

import ee.ivxv.common.service.i18n.MessageException;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import java.util.Locale;
import java.util.WeakHashMap;
import java.util.function.Consumer;

/**
 * LocaleConf is a class that contains the current set of locale configuration. Note that the locale
 * configuration can be modified - both the list of available locales loaded and the current locale
 * changed. Locale configuration is initialized (loaded) separately from other configuration.
 */
public class LocaleConf {

    public static final Locale DEFAULT_LOCALE = new Locale("et");

    private List<Locale> allLocales;
    private Locale locale;
    private final WeakHashMap<Object, Consumer<Locale>> localeChangeListeners = new WeakHashMap<>();

    /**
     * Uses the default locale as the only available locale.
     */
    public LocaleConf() {
        setAllLocales(Arrays.asList(DEFAULT_LOCALE));
    }

    public Locale getLocale() {
        return locale;
    }

    public void setLocale(Locale locale) {
        if (!allLocales.contains(locale)) {
            throw new MessageException(Msg.e_unsupported_locale, locale, allLocales);
        }
        Locale oldLocale = this.locale;
        this.locale = locale;
        if (oldLocale == null || !oldLocale.equals(locale)) {
            localeChangeListeners.values().forEach(l -> l.accept(locale));
        }
    }

    /**
     * @return Returns unmodifiable list of all supported locales.
     */
    public List<Locale> getAllLocales() {
        return allLocales;
    }

    /**
     * Sets the list of all available locales.
     * 
     * @param allLocales The list of all available locales. NB! Assumed to be not empty. If the list
     *        does not contain the current locale, the first element of the list is set as the
     *        current locale.
     */
    void setAllLocales(List<Locale> allLocales) {
        this.allLocales = Collections.unmodifiableList(allLocales);

        if (!allLocales.contains(locale)) {
            setLocale(allLocales.get(0));
        }
    }

    public void addLocaleChangeListener(Object key, Consumer<Locale> listener) {
        localeChangeListeners.put(key, listener);
    }

}
