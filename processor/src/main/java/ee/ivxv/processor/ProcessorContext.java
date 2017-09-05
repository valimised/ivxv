package ee.ivxv.processor;

import ee.ivxv.common.cli.AppContext;
import ee.ivxv.common.cli.CommonArgs;
import ee.ivxv.common.cli.InitialContext;
import ee.ivxv.common.conf.Conf;

public class ProcessorContext extends AppContext<Conf> {

    public ProcessorContext(InitialContext i, Conf conf, CommonArgs args) {
        super(i, conf, args);
    }

}
