package ee.ivxv.common.service.smartcard.pkcs15;

import ee.ivxv.common.service.smartcard.SmartCardException;

@SuppressWarnings("serial")
public class PKCS15Exception extends SmartCardException {

    public PKCS15Exception(String message) {
        super(message);
    }

    public PKCS15Exception(String message, Throwable cause) {
        super(message, cause);
    }
}
