package ee.ivxv.common.cli;

import ee.ivxv.common.service.i18n.MessageException;

public class ParseException extends MessageException {

    private static final long serialVersionUID = -5672721085726710014L;

    public ParseException(Enum<?> key, Object... args) {
        super(key, args);
    }

    public ParseException(Throwable cause, Enum<?> key, Object... args) {
        super(cause, key, args);
    }

}
