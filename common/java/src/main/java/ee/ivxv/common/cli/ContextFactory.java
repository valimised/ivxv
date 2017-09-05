package ee.ivxv.common.cli;

import ee.ivxv.common.conf.Conf;
import ee.ivxv.common.conf.LocaleConf;
import ee.ivxv.common.service.bbox.BboxHelper;
import ee.ivxv.common.service.bbox.impl.BboxHelperImpl;
import ee.ivxv.common.service.console.BlockingQueueConsole;
import ee.ivxv.common.service.console.Console;
import ee.ivxv.common.service.container.ContainerReader;
import ee.ivxv.common.service.container.bdoc.BdocContainerReader;
import ee.ivxv.common.service.i18n.Cal10nI18nImpl;
import ee.ivxv.common.service.i18n.I18n;
import ee.ivxv.common.service.report.CsvReporterImpl;
import ee.ivxv.common.service.report.Reporter;
import ee.ivxv.common.service.smartcard.CardService;
import ee.ivxv.common.service.smartcard.pkcs15.PKCS15CardService;
import java.math.BigInteger;
import java.security.SecureRandom;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * ContextFactory class is the "single source of truth" for creating services. The class is used
 * with the singleton pattern and the single instance of this class chooses the implementation for
 * each and every service used in an IVXV application.
 * 
 * <p>
 * It is possible to have an alternative context factory for "development mode" by extending this
 * class by a class with the name of the value of {@code OVERLOAD_CLASS}. If such class is found on
 * the classpath and creating an instance with it's default constructor succeeds, that instance is
 * used as the singleton instance. Hence it forces the behaviour of excluding the context factory
 * for development mode from the production build.
 */
public class ContextFactory {

    private static final Logger log = LoggerFactory.getLogger(ContextFactory.class);

    private static final String OVERLOAD_CLASS = "ee.ivxv.common.cli.TestContextFactory";
    private static final ContextFactory INSTANCE = createInstance();

    private final boolean testMode;

    /**
     * Creates a service factory in <u>test mode</u>.
     */
    public ContextFactory() {
        this(true);
    }

    private ContextFactory(boolean testMode) {
        this.testMode = testMode;
    }

    private static ContextFactory createInstance() {
        try {
            Class<?> type = Class.forName(OVERLOAD_CLASS);
            Object instance = type.newInstance();
            ContextFactory result = (ContextFactory) instance;
            log.info("!!!Using overladed context factory {}!!!", OVERLOAD_CLASS);
            return result;
        } catch (ClassNotFoundException e) {
            log.info("Overloaded context factory class {} not found", OVERLOAD_CLASS);
        } catch (Exception e) {
            log.error("Exception occurred while trying to create overloaded context factory {}: {}",
                    OVERLOAD_CLASS, e.getMessage(), e);
        }
        log.info("Using standard context factory {}", ContextFactory.class.getName());
        return new ContextFactory(false);
    }

    public static ContextFactory get() {
        return INSTANCE;
    }

    private static String createSessionId() {
        SecureRandom random = new SecureRandom();
        return new BigInteger(130, random).toString(32);
    }

    /**
     * @return Returns {@code false} if and only if {@code this} is an instance of the class
     *         {@code ContextFactory} and not a overloading class.
     */
    public final boolean isTestMode() {
        return testMode;
    }

    public LocaleConf getLocaleConf() {
        return new LocaleConf();
    }

    public Console getConsole() {
        return new BlockingQueueConsole();
    }

    public I18n getI18n(LocaleConf locale) {
        return new Cal10nI18nImpl(locale);
    }

    public Reporter getReporter(I18n i18n) {
        return new CsvReporterImpl(i18n);
    }

    public CardService getCard(Console console, I18n i18n) {
        return new PKCS15CardService(console, i18n);
    }

    public ContainerReader getContainer(Conf conf, int nThreads) {
        return new BdocContainerReader(conf, nThreads);
    }

    public BboxHelper getBbox(Conf conf, ContainerReader container) {
        return new BboxHelperImpl(conf, container);
    }

    public InitialContext createInitialContext() {
        String sessionId = createSessionId();
        LocaleConf locale = getLocaleConf();
        Console console = getConsole();
        I18n i18n = getI18n(locale);
        return new InitialContext(sessionId, locale, console, i18n);
    }

}
