package ee.ivxv.common.service.console;

import java.util.function.Function;

/**
 * Progress is a thread-safe class for tasks to report back progress.
 */
public interface Progress {

    void increase(int amount);

    void finish();

    /**
     * {@code Factory} is a named function for providing {@code Progress} instance for the 'total'.
     */
    @FunctionalInterface
    public interface Factory extends Function<Integer, Progress> {
        // Inherit only. Just for providing nicer name.
    }

}
