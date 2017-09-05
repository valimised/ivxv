package ee.ivxv.audit;

import ee.ivxv.common.cli.AppContext;
import ee.ivxv.common.cli.CommonArgs;
import ee.ivxv.common.cli.InitialContext;
import ee.ivxv.common.conf.Conf;

public class AuditContext extends AppContext<Conf> {

    public AuditContext(InitialContext i, Conf conf, CommonArgs args) {
        super(i, conf, args);
    }

}
