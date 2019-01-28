package ee.ivxv.audit.shuffle;

@SuppressWarnings("serial")
public class ShuffleException extends Exception {
    public ShuffleException(Throwable t) {
        super(t);
    }

    public ShuffleException(String msg) {
        super(msg);
    }

    public ShuffleException(String msg, Throwable t) {
        super(msg, t);
    }
}
