package ee.ivxv.common.crypto.hash;

import java.util.function.Supplier;

public enum HashType {
    SHA256(Sha256::new);

    private final Supplier<HashFunction> supplier;

    HashType(Supplier<HashFunction> supplier) {
        this.supplier = supplier;
    }

    public static HashType map(String s) {
        return HashType.valueOf(s.toUpperCase().replace("-", ""));
    }

    public HashFunction getFunction() {
        return supplier.get();
    }
}
