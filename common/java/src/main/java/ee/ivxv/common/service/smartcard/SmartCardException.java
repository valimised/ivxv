package ee.ivxv.common.service.smartcard;

@SuppressWarnings("serial")
public class SmartCardException extends Exception {
    public SmartCardException(String message) {
        super(message);
    }

    public SmartCardException(String message, Throwable cause) {
        super(message, cause);
    }
}
