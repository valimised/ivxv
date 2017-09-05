package ee.ivxv.common.service.console;

import java.util.concurrent.atomic.LongAdder;

/**
 * Progress is a thread-safe class for tasks to report back progress.
 */
class ProgressImpl implements Progress {

    private final long total;
    private final Runnable finishedNotifier;
    private LongAdder value = new LongAdder();
    private boolean isFinished = false;

    ProgressImpl(long total, Runnable finishedNotifier) {
        this.total = Math.max(total, 0);
        this.finishedNotifier = finishedNotifier;
    }

    @Override
    public void increase(int amount) {
        if (amount <= 0) {
            return;
        }
        if (isFinished) {
            return;
        }
        value.add(amount);
        // Call update() callback here if necessary
    }

    @Override
    public void finish() {
        if (isFinished) {
            return;
        }
        isFinished = true;
        finishedNotifier.run();
    }

    long getValue() {
        return isFinished ? total : value.sum();
    }

    long getTotal() {
        return total;
    }

    boolean isFinished() {
        return isFinished;
    }

}
