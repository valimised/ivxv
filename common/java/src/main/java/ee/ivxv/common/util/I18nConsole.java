package ee.ivxv.common.util;

import ee.ivxv.common.M;
import ee.ivxv.common.service.console.Console;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.service.i18n.I18n;
import ee.ivxv.common.service.i18n.Translatable;

/**
 * Convenience class to write internationalized messages to the console.
 */
public class I18nConsole {

    public final Console console;
    public final I18n i18n;

    public I18nConsole(Console console, I18n i18n) {
        this.console = console;
        this.i18n = i18n;
    }

    public void println() {
        console.println();
    }

    public void println(Translatable msg) {
        console.println(i18n.get(msg));
    }

    public void println(Enum<?> key, Object... args) {
        console.println(i18n.get(key, args));
    }

    public Progress startProgress(long total) {
        return console.startProgress(i18n.get(M.m_progress_bar), total);
    }

    public Progress startInfiniteProgress(long total) {
        return console.startInfiniteProgress(i18n.get(M.m_progress_bar), total);
    }
}
