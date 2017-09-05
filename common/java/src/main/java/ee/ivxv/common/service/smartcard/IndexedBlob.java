package ee.ivxv.common.service.smartcard;

public class IndexedBlob {

    public final byte[] blob;
    public final int index;

    public IndexedBlob(int index, byte[] blob) {
        this.blob = blob;
        this.index = index;
    }

    public int getIndex() {
        return index;
    }

    public byte[] getBlob() {
        return blob;
    }

}
