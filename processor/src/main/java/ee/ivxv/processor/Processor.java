package ee.ivxv.processor;

import ee.ivxv.common.cli.AppRunner;

public class Processor {

    public static void main(String[] args) {
        ProcessorApp app = new ProcessorApp();
        AppRunner<ProcessorContext> runner = new AppRunner<>(app);

        if (!runner.run(args)) {
            System.exit(1);
        }
    }
}
