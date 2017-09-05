package ee.ivxv.common.cli;

import ee.ivxv.common.conf.Conf;
import ee.ivxv.common.service.bbox.BboxHelper;
import ee.ivxv.common.service.container.ContainerReader;
import ee.ivxv.common.service.report.Reporter;
import ee.ivxv.common.service.smartcard.CardService;

/**
 * AppContext is a collection of all dependencies an application or it's tools may need during their
 * work. Instances of AppContext are created after loading the configuration. Instances of
 * AppContext can be application-specific.
 * 
 * @param <T> The exact type of configuration, to support application-specific configuration.
 */
public class AppContext<T extends Conf> {

    public final InitialContext i;
    public final T conf;
    public final CommonArgs args;
    public final Reporter reporter;
    public final CardService card;
    public final ContainerReader container;
    public final BboxHelper bbox;

    public AppContext(InitialContext i, T conf, CommonArgs args) {
        this(i, conf, args, ContextFactory.get().getReporter(i.i18n),
                ContextFactory.get().getCard(i.console, i.i18n),
                ContextFactory.get().getContainer(conf, args.ct.value()));
    }

    public AppContext(InitialContext i, T conf, CommonArgs args, Reporter reporter,
            CardService card, ContainerReader container) {
        this(i, conf, args, reporter, card, container,
                ContextFactory.get().getBbox(conf, container));
    }

    public AppContext(InitialContext i, T conf, CommonArgs args, Reporter reporter,
            CardService card, ContainerReader container, BboxHelper bbox) {
        this.i = i;
        this.conf = conf;
        this.args = args;
        this.reporter = reporter;
        this.card = card;
        this.container = container;
        this.bbox = bbox;
    }

}
