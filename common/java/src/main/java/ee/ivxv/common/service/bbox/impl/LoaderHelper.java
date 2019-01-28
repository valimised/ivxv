package ee.ivxv.common.service.bbox.impl;

import ee.ivxv.common.service.bbox.BboxHelper.Reporter;
import ee.ivxv.common.service.bbox.Ref;
import ee.ivxv.common.service.bbox.Result;
import ee.ivxv.common.service.bbox.impl.FileName.RefProvider;
import ee.ivxv.common.service.console.Progress;
import ee.ivxv.common.util.Util;
import java.io.InputStream;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Set;
import java.util.function.BiConsumer;
import java.util.function.Consumer;
import java.util.function.Predicate;
import java.util.function.Supplier;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Helper class for common logic of BboxLoader and RegDataLoader.
 * 
 * @param <T> The reference type
 */
public class LoaderHelper<T extends Ref> {

    private static final Logger log = LoggerFactory.getLogger(LoaderHelper.class);

    private final FileSource source;
    private final RefProvider<T> refProvider;
    private final Progress.Factory pf;
    private final Reporter<T> reporter;

    LoaderHelper(FileSource source, RefProvider<T> refProvider, Progress.Factory pf,
            Reporter<T> reporter) {
        this.source = source;
        this.refProvider = refProvider;
        this.pf = pf;
        this.reporter = reporter;
    }

    Set<T> getAllRefs() {
        Map<T, Boolean> initialData = new LinkedHashMap<>();

        source.list(path -> {
            try {
                FileName<T> name = new FileName<>(path, refProvider);
                initialData.put(name.ref, Boolean.TRUE);
            } catch (FileName.InvalidNameException e) {
                log.warn("Invalid file name: {}", path, e);
                reporter.report(null, Result.INVALID_FILE_NAME, e.path, e.expected);
            }
        });

        return initialData.keySet();
    }

    /**
     * Performs the integrity-check for the data being loaded.
     * 
     * @param supplier Record supplier.
     * @param total Total number of records for progress reporting.
     * @return Returns integrity-checked map from reference to record.
     */
    <U extends Record<?>> Map<T, U> checkIntegrity(Supplier<U> supplier, int total) {
        Map<T, U> result = new LinkedHashMap<>();
        Progress pb = getProgress(total);

        list(name -> {
            U record = result //
                    .computeIfAbsent(name.ref, s -> {
                        pb.increase(1);
                        return supplier.get();
                    });

            Result res = record.set(name.type);
            if (res != Result.OK) {
                report(name.ref, res, name.path);
            }
        });

        // Report missing files
        result.forEach((ref, r) -> r.forMissingFiles(t -> report(ref, Result.MISSING_FILE, t)));

        // Remove incomplete records
        result.values().removeIf(record -> !record.isComplete());

        pb.finish();

        return result;
    }

    void list(Consumer<FileName<T>> processor) {
        source.list(path -> {
            try {
                FileName<T> name = new FileName<>(path, refProvider);
                try {
                    processor.accept(name);
                } catch (Exception e) {
                    handleTechnicalError(name, e);
                }
            } catch (FileName.InvalidNameException e) {
                // Skip invalid file names. Error is reported by 'loadInitialData()'.
            }
        });
    }

    void processFiles(BiConsumer<FileName<T>, InputStream> processor) {
        source.processFiles((path, in) -> {
            try {
                FileName<T> name = new FileName<>(path, refProvider);
                try {
                    processor.accept(name, in);
                } catch (Exception e) {
                    handleTechnicalError(name, e);
                }
            } catch (FileName.InvalidNameException e) {
                // Skip invalid file names. Error is reported by 'loadInitialData()'.
            }
        });
    }

    <U extends Record<?>> void processRecords(Predicate<FileName<T>> filter, Supplier<U> supplier,
            BiConsumer<FileName<T>, U> processor) {
        processRecords(filter, false, supplier, processor);
    }

    <U extends Record<?>> void processRecords(Predicate<FileName<T>> filter,
            boolean processIncomplete, Supplier<U> supplier, BiConsumer<FileName<T>, U> processor) {
        Map<T, NameRecord<T, U>> work = new HashMap<>();
        byte[] buffer = new byte[1024];

        processFiles((name, in) -> {
            if (!filter.test(name)) {
                return;
            }
            U record = work.computeIfAbsent(name.ref,
                    x -> new NameRecord<>(name, supplier.get())).record;
            record.set(name.type, Util.toBytes(in, buffer));
            if (record.isComplete()) {
                // Release memory
                work.remove(name.ref);
                processor.accept(name, record);
            }
        });

        // Report incomplete records
        if (processIncomplete) {
            work.forEach((ref, nr) -> processor.accept(nr.name, nr.record));
        }
    }

    void handleTechnicalError(FileName<T> name, Exception e) {
        log.error("Tehcnical error occurred while processing file {}: ", name.path, e);
        report(name.ref, Result.TECHNICAL_ERROR, e);
    }

    void handleTechnicalError(T ref, Exception e) {
        log.error("Tehcnical error occurred while processing record {}: ", ref, e);
        report(ref, Result.TECHNICAL_ERROR, e);
    }

    void report(T ref, Result res, Object... args) {
        reporter.report(ref, res, args);
    }

    Progress getProgress(int total) {
        return pf.apply(total);
    }

    private static class NameRecord<T extends Ref, U extends Record<?>> {
        final FileName<T> name;
        final U record;

        public NameRecord(FileName<T> name, U record) {
            this.name = name;
            this.record = record;
        }
    }

}
