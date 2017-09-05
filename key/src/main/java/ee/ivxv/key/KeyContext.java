package ee.ivxv.key;

import ee.ivxv.common.cli.AppContext;
import ee.ivxv.common.cli.CommonArgs;
import ee.ivxv.common.cli.InitialContext;
import ee.ivxv.common.conf.Conf;

public class KeyContext extends AppContext<Conf> {

    public KeyContext(InitialContext i, Conf conf, CommonArgs args) {
        super(i, conf, args);
    }

}
