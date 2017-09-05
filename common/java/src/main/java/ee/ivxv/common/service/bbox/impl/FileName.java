package ee.ivxv.common.service.bbox.impl;

import ee.ivxv.common.service.bbox.Ref;

class FileName<T extends Ref> {

    private static final char EXT_SEP = '.';
    private static final String EXPECTED_FMT = "%s" + EXT_SEP + "<ext>";

    final String path;
    final T ref;
    final String baseName;
    final String type;

    private FileName(T ref, String baseName, String type) {
        this.path = new StringBuilder().append(baseName).append(EXT_SEP).append(type).toString();
        this.ref = ref;
        this.baseName = baseName;
        this.type = type;
    }

    /**
     * @param path The file path.
     * @throws InvalidNameException If the path does not conform to expected file path pattern.
     */
    FileName(String path, RefProvider<T> provider) throws InvalidNameException {
        int i = path.lastIndexOf(EXT_SEP);
        if (i < 0) {
            throw new InvalidNameException(path, String.format(EXPECTED_FMT, "<name>"));
        }
        this.path = path;
        this.baseName = path.substring(0, i);
        try {
            this.ref = provider.get(baseName);
        } catch (InvalidNameException e) {
            throw new InvalidNameException(path, String.format(EXPECTED_FMT, e.expected));
        }
        type = path.substring(i + 1);
    }

    FileName<T> forType(Enum<?> typeEnum) {
        if (typeEnum.name().equals(type)) {
            return this;
        }
        return new FileName<>(ref, baseName, typeEnum.name());
    }

    interface RefProvider<T extends Ref> {
        T get(String name);
    }

    static class InvalidNameException extends RuntimeException {
        private static final long serialVersionUID = 6878488000246512903L;

        final String path;
        final String expected;

        InvalidNameException(String path, String expected) {
            super("Invalid path '" + path + "'. Expected: " + expected);
            this.path = path;
            this.expected = expected;
        }
    }

}
