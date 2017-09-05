package ee.ivxv.common.service.bbox.impl;

import java.io.InputStream;
import java.util.function.BiConsumer;
import java.util.function.Consumer;

/**
 * Simple file system abstraction.
 */
public interface FileSource {

    /**
     * Iterates over all files and calls the processor on them.
     * 
     * @param processor
     */
    void processFiles(BiConsumer<String, InputStream> processor);

    /**
     * Lists all file names and calls the processor on them.
     * 
     * @param processor
     */
    default void list(Consumer<String> processor) {
        processFiles((name, in) -> {
            processor.accept(name);
        });
    }

}
