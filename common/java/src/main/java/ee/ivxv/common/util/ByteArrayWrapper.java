package ee.ivxv.common.util;

import java.util.Arrays;
import javax.xml.bind.DatatypeConverter;

/**
 * Class to wrap byte array to make it suitable for hash key.
 */
public class ByteArrayWrapper {

    private final byte[] data;

    public ByteArrayWrapper(byte[] data) {
        this.data = data;
    }

    @Override
    public boolean equals(Object other) {
        if (!(other instanceof ByteArrayWrapper)) {
            return false;
        }
        return Arrays.equals(data, ((ByteArrayWrapper) other).data);
    }

    @Override
    public int hashCode() {
        return Arrays.hashCode(data);
    }

    @Override
    public String toString() {
        return DatatypeConverter.printHexBinary(data).toLowerCase();
    }
}
