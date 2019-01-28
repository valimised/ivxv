package ee.ivxv.common.service.smartcard;

/**
 * IndexedBlob is a blob with an index
 */
public class IndexedBlob {
    public final byte[] blob;
    public final int index;

    /**
     * Initialize using index and data
     * 
     * @param index
     * @param blob
     */
    public IndexedBlob(int index, byte[] blob) {
        this.blob = blob;
        this.index = index;
    }

    /**
     * Get blob index.
     * 
     * @return
     */
    public int getIndex() {
        return index;
    }

    /**
     * Get blob.
     * 
     * @return
     */
    public byte[] getBlob() {
        return blob;
    }

    @Override
    public String toString() {
        return String.valueOf(index);
    }
}
