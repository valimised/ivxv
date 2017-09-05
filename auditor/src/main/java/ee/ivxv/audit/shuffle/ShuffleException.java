package ee.ivxv.audit.shuffle;

public class ShuffleException extends Exception {
    public ShuffleException(String msg) {
        super(msg);
    }

    public ShuffleException(String msg, Throwable t) {
        super(msg, t);
    }
}
