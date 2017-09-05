package ee.ivxv.audit;

import ee.ivxv.common.cli.AppRunner;

public class Audit {

    public static void main(String[] args) {
        AuditApp app = new AuditApp();
        AppRunner<AuditContext> runner = new AppRunner<>(app);

        if (!runner.run(args)) {
            System.exit(1);
        }
    }

}
