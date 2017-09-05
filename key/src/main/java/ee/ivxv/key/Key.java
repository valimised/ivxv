package ee.ivxv.key;

import ee.ivxv.common.cli.AppRunner;

public class Key {

    public static void main(String[] args) {
        KeyApp app = new KeyApp();
        AppRunner<KeyContext> runner = new AppRunner<>(app);

        if (!runner.run(args)) {
            System.exit(1);
        }
    }
}
