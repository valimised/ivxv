package ee.ivxv.common.cli;

import ee.ivxv.common.conf.LocaleConf;
import ee.ivxv.common.service.console.Console;
import ee.ivxv.common.service.i18n.I18n;

/**
 * InitialContext is a collection of essential non-configurable dependencies for applications. The
 * dependencies in this class do not require instances of {@code Conf} neither for creation nor
 * running.
 */
public final class InitialContext {

    public final String sessionId;
    public final LocaleConf locale;
    public final Console console;
    public final I18n i18n;

    public InitialContext(String sessionId, LocaleConf locale, Console console, I18n i18n) {
        this.locale = locale;
        this.sessionId = sessionId;
        this.console = console;
        this.i18n = i18n;
    }

}
