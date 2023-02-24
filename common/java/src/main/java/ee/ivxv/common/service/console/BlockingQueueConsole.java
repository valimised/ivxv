package ee.ivxv.common.service.console;

import ee.ivxv.common.util.Util;
import java.io.PrintStream;
import java.util.Arrays;
import java.util.Optional;
import java.util.Scanner;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicBoolean;
import java.util.concurrent.atomic.AtomicReference;
import java.util.function.Supplier;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * ConsoleImpl implements Console interface and prints messages simply to <tt>System.out</tt>.
 */
public class BlockingQueueConsole implements Console {

    private static final Logger log = LoggerFactory.getLogger(BlockingQueueConsole.class);

    private static final String KEY_VALUE = "{VALUE}";
    private static final String KEY_TOTAL = "{TOTAL}";
    private static final String KEY_PERCENT = "{PERCENT}";
    private static final String KEY_BAR = "{BAR}";
    private static final int BAR_WIDTH = 50;
    private static final char BAR_CHAR = '.';
    private static final long QUEUE_CHECK_FREQ_MS = 50;
    private static final String PROGRESS_FINISHED_MSG = "shurely-unique-" + Math.random();

    private final Scanner in = new Scanner(System.in);
    private final BlockingQueue<String> queue = new LinkedBlockingQueue<>();
    private final ExecutorService executor;
    private final TextFormatter formatter;
    // Reference to 'progress supplier', i.e 'ps'.
    private final AtomicReference<Supplier<String>> psRef = new AtomicReference<>();

    public BlockingQueueConsole() {
        executor = Executors.newSingleThreadExecutor();
        formatter = new TextFormatter();
        executor.submit(() -> printQueue(System.out));
    }

    @Override
    public void println() {
        println("");
    }

    @Override
    public void println(String format, Object... args) {
        queue.add(String.format(format, args));
    }

    @Override
    public String readln() {
        return in.nextLine();
    }

    @Override
    public String readPw() {
        java.io.Console console = System.console();
        // Should warn user when console is null and password input will be visible?
        // Should be possible ONLY when running from IDE
        if (console == null) {
            return readln();
        }
        char[] chars = console.readPassword();
        String str = new String(chars);
        // Minimize the lifetime of sensitive data in memory
        Arrays.fill(chars, ' ');
        return str;
    }

    @Override
    public Progress startProgress(String format, long total) {
        ProgressImpl progress = new ProgressImpl(total, () -> queue.add(PROGRESS_FINISHED_MSG));
        psRef.set(new PbFormatter(progress, format));
        return progress;
    }

    @Override
    public Progress startInfiniteProgress(String format, long total) {
        ProgressImpl progress =
                new InfiniteProgressImpl(total, () -> queue.add(PROGRESS_FINISHED_MSG));
        psRef.set(new PbFormatter(progress, format));
        return progress;
    }

    @Override
    public void shutdown() {
        queue.add(Util.EOT);
        executor.shutdown();
        try {
            executor.awaitTermination(1, TimeUnit.DAYS);
        } catch (InterruptedException e) {
            throw new RuntimeException(e);
        }
    }

    /*
     * This can be called only through PbFormatter.get().
     */
    void progressFinished(Supplier<String> psToFinish) {
        // Only unset the reference if it has not been changed by a 'startProgress' call
        psRef.compareAndSet(psToFinish, null);
    }

    private Void printQueue(PrintStream out) {
        AtomicBoolean onProgressRow = new AtomicBoolean();
        String line = queue.poll();

        while (true) {
            try {
                for (; line != null; line = queue.poll()) {
                    if (line.equals(Util.EOT)) {
                        out.flush();
                        return null;
                    }
                    if (line.equals(PROGRESS_FINISHED_MSG)) {
                        break;
                    }
                    if (onProgressRow.getAndSet(false)) {
                        out.println();
                    }
                    String formatted = formatter.formatLine(line);
                    log.info("[CONSOLE]: " + formatted);
                    out.println(formatted);
                }

                Optional.ofNullable(psRef.get()).ifPresent(p -> {
                    out.print(p.get());
                    onProgressRow.set(true);
                });

                line = queue.poll(QUEUE_CHECK_FREQ_MS, TimeUnit.MILLISECONDS);
            } catch (Exception e) {
                log.error("Exception occurred in console printer thread", e);
            }
        }
    }

    /*
     * Not thread safe.
     */
    private class PbFormatter implements Supplier<String> {

        private final ProgressImpl progress;
        private final String format;

        private int lastPercent = -1;
        private String lastPercentStr = "";
        private String lastBarStr = "";

        PbFormatter(ProgressImpl progress, String format) {
            this.progress = progress;
            this.format = format;
        }

        @Override
        public String get() {
            return formatProgress();
        }

        private String formatProgress() {
            // For infinite progress overwrite the end of the line by spaces and move to the start
            String result = "\b\b\b\b    \r" + format;
            int percent = progress.getTotal() == 0 ? 100
                    : (int) ((progress.getValue() * 100) / progress.getTotal());

            result = result.replace(KEY_VALUE, String.valueOf(progress.getValue()));
            result = result.replace(KEY_TOTAL, String.valueOf(progress.getTotal()));
            result = result.replace(KEY_PERCENT, formatPercent(percent));
            result = result.replace(KEY_BAR, formatBar(percent));

            if (progress.isFinished()) {
                progressFinished(this);
            }

            lastPercent = percent;

            return result;
        }

        private String formatPercent(int percent) {
            if (percent == lastPercent) {
                return lastPercentStr;
            }
            lastPercentStr = String.format("%3s", percent);
            return lastPercentStr;
        }

        private String formatBar(int percent) {
            if (percent == lastPercent) {
                return lastBarStr;
            }
            StringBuffer result = new StringBuffer();
            double done = (percent * BAR_WIDTH) / 100d;

            for (int i = 0; i < BAR_WIDTH; i++) {
                result.append(i < done ? BAR_CHAR : ' ');
            }

            lastBarStr = result.toString();

            return lastBarStr;
        }
    }
}
