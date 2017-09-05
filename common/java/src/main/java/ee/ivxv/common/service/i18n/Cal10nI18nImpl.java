package ee.ivxv.common.service.i18n;

import ch.qos.cal10n.IMessageConveyor;
import ch.qos.cal10n.MessageConveyor;
import ee.ivxv.common.conf.LocaleConf;
import java.util.Arrays;
import java.util.Locale;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class Cal10nI18nImpl extends DefaultI18n implements I18n {

    private static final Logger log = LoggerFactory.getLogger(Cal10nI18nImpl.class);

    private final LocaleConf locale;
    private IMessageConveyor mc;

    public Cal10nI18nImpl(LocaleConf locale) {
        this.locale = locale;

        setTranslatorForLocale(locale.getLocale());

        // Register to locale change events
        locale.addLocaleChangeListener(this, this::setTranslatorForLocale);
    }

    private void setTranslatorForLocale(Locale l) {
        /*
         * If translation resource bundles are to be loaded at run-time along with other
         * configuration, the usage of MessageConveyor class needs to be replaced with a
         * modification of it, where CAL10NResourceBundleFinder is invoked with a URLClassLoader
         * that loads resources from configuration bundle (directory, bdoc, zip, jar), rather than
         * the main class path.
         * 
         * Note: Similar modification has to be done with MessageKeyVerifier in I18nTest.
         */
        mc = new MessageConveyor(l);
    }

    @Override
    public String getInternal(Enum<?> key, Object... args) {
        try {
            return mc.getMessage(key, args);
        } catch (RuntimeException e) {
            String s = String.format("[Missing translation for language '%s': %s %s]",
                    locale.getLocale(), key, Arrays.asList(args));
            log.error(s, e);
            return s;
        }
    }

}
