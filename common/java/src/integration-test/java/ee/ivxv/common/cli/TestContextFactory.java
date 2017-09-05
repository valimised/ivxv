package ee.ivxv.common.cli;

import ee.ivxv.common.conf.Conf;
import ee.ivxv.common.service.bbox.BboxHelper;
import ee.ivxv.common.service.bbox.impl.BboxHelperImpl;
import ee.ivxv.common.service.console.Console;
import ee.ivxv.common.service.container.ContainerReader;
import ee.ivxv.common.service.i18n.I18n;
import ee.ivxv.common.service.report.CsvReporterImpl;
import ee.ivxv.common.service.report.Reporter;
import ee.ivxv.common.service.smartcard.CardService;
import ee.ivxv.common.service.smartcard.dummy.DummyCardService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * This class must extend {@code ContextFactory} and have the same full name as the value of
 * {@code ContextFactory.OVERLOAD_CLASS}.
 * 
 * <p>
 * The class can override service creation methods and provide different implementation.
 */
public class TestContextFactory extends ContextFactory {

    private static final Logger log = LoggerFactory.getLogger(TestContextFactory.class);

    public TestContextFactory() {
        super();
        log.info("Creating instance of overloaded context factory {}", getClass().getName());
    }

    @Override
    public Console getConsole() {
        log.info("Creating Console by overloaded context factory {}", getClass().getName());
        return super.getConsole();
    }

    @Override
    public Reporter getReporter(I18n i18n) {
        return new CsvReporterImpl(i18n) {
            @Override
            public String getCurrentTime() {
                // Using fixed current time in reports to simplify comparing output with 'diff' tool
                return "20170401000000";
            }
        };
    }

    @Override
    public CardService getCard(Console console, I18n i18n) {
        return new DummyCardService(console, i18n, true);
    }

    @Override
    public BboxHelper getBbox(Conf conf, ContainerReader container) {
        log.info("Creating Bbox by overloaded context factory {}", getClass().getName());
        return new BboxHelperImpl(conf, container);
    }

}
