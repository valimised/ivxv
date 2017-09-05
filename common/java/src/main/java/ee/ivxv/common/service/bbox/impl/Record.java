package ee.ivxv.common.service.bbox.impl;

import ee.ivxv.common.service.bbox.Result;
import java.util.HashMap;
import java.util.Map;
import java.util.function.Consumer;

class Record<T extends Enum<T>> {

    private static final byte[] EMPTY = new byte[0];

    private final Class<T> clazz;
    final Map<T, byte[]> files = new HashMap<>();
    private int missingFiles;

    Record(Class<T> clazz) {
        this.clazz = clazz;
        missingFiles = clazz.getEnumConstants().length;
    }

    Result set(String type) {
        return set(type, EMPTY);
    }

    Result set(String type, byte[] content) {
        T t;

        try {
            t = Enum.valueOf(clazz, type);
        } catch (Exception e) {
            return Result.UNKNOWN_FILE_TYPE;
        }

        byte[] oldValue = files.put(t, content);
        if (oldValue != null) {
            missingFiles = -1;
            return Result.REPEATED_FILE;
        }

        missingFiles--;

        return Result.OK;
    }

    byte[] get(T type) {
        return files.get(type);
    }

    boolean isComplete() {
        return missingFiles == 0;
    }

    void forMissingFiles(Consumer<T> consumer) {
        for (T t : clazz.getEnumConstants()) {
            if (!files.containsKey(t)) {
                consumer.accept(t);
            }
        }
    }

}
